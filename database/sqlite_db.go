package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SqliteDbProvider implements Provider backed by a SQLite database.
//
// Every record type maps to a table named record.Type(), created by a
// migration (see database/migrations and #54+). The primary key column is
// always "id" and is handled via GetId/SetId; all other scalar fields are
// mapped by reflection using their JSON tag (snake_case) as the column name.
//
// Encoding conventions are defined in docs/schema-conventions.md (see #54):
//   - ID-family uint kinds -> TEXT hex (avoids signed-64 overflow)
//   - string kinds         -> TEXT
//   - bool                 -> INTEGER (0/1)
//   - int kinds            -> INTEGER
//   - time.Time            -> RFC3339 TEXT
//   - pointers             -> nullable (NULL when nil)
//
// Tables are created by model subtask migrations (#55+) layered on the
// skeleton migration; see database/migrations and docs/schema-conventions.md.
type SqliteDbProvider struct {
	db   *sql.DB
	path string
}

// NewSqliteDbProvider opens (creating if needed) the SQLite database at path,
// enables WAL + busy timeout, and runs any pending migrations before
// returning a ready-to-use Provider.
func NewSqliteDbProvider(path string) (Provider, error) {
	migs, err := EmbeddedMigrations()
	if err != nil {
		return nil, err
	}
	p, err := newSqliteDbProvider(path, migs)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func newSqliteDbProvider(path string, migrations []Migration) (*SqliteDbProvider, error) {
	if path == "" {
		path = "intraclub.db"
	}

	dsn := "file:" + path +
		"?_pragma=busy_timeout(5000)" + // concurrent reads from Gin handlers
		"&_pragma=journal_mode(WAL)" + // single-writer, concurrent readers
		"&_txlock=immediate" // acquire the write lock up front
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if err := Migrate(db, migrations); err != nil {
		db.Close()
		return nil, err
	}
	return &SqliteDbProvider{db: db, path: path}, nil
}

// Disconnect closes the underlying database connection.
func (s *SqliteDbProvider) Disconnect() error {
	return s.db.Close()
}

func (s *SqliteDbProvider) GetOne(ctx context.Context, record CrudRecord) (CrudRecord, bool, error) {
	table := record.Type()
	rows, err := s.db.QueryContext(ctx, "SELECT * FROM "+table+" WHERE id = ?", record.GetId().String())
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, false, err
	}
	if !rows.Next() {
		return nil, false, rows.Err()
	}

	blank := record.NewRecord()
	if err := scanRow(rows, cols, blank); err != nil {
		return nil, false, err
	}
	return blank, true, rows.Err()
}

func (s *SqliteDbProvider) GetAll(ctx context.Context, recordType CrudRecord) ([]CrudRecord, error) {
	return s.GetAllWhere(ctx, recordType, nil)
}

func (s *SqliteDbProvider) GetAllWhere(ctx context.Context, recordType CrudRecord, where WhereFunc) ([]CrudRecord, error) {
	table := recordType.Type()
	rows, err := s.db.QueryContext(ctx, "SELECT * FROM "+table+" ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	output := make([]CrudRecord, 0)
	for rows.Next() {
		blank := recordType.NewRecord()
		if err := scanRow(rows, cols, blank); err != nil {
			return nil, err
		}
		if where == nil || where(ctx, blank) {
			output = append(output, blank)
		}
	}
	return output, rows.Err()
}

func (s *SqliteDbProvider) Create(ctx context.Context, record CrudRecord) (CrudRecord, error) {
	if !record.GetId().ValidRecordId() {
		record.SetId(NewRecordId())
	}

	cols, vals, err := insertParts(record)
	if err != nil {
		return nil, err
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(cols)), ",")
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		record.Type(), strings.Join(cols, ", "), placeholders)

	_, err = s.db.ExecContext(ctx, query, vals...)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, fmt.Errorf("A %s record with ID %s already exists", record.Type(), record.GetId())
		}
		return nil, err
	}
	return record, nil
}

func (s *SqliteDbProvider) Update(ctx context.Context, record CrudRecord) error {
	cols, vals, err := insertParts(record)
	if err != nil {
		return err
	}
	// cols[0] == "id" (handled via GetId/SetId); exclude it from SET and use
	// it in the WHERE clause.
	idVal := vals[0]
	sets := make([]string, 0, len(cols)-1)
	args := make([]any, 0, len(cols))
	for i, col := range cols[1:] {
		sets = append(sets, col+" = ?")
		args = append(args, vals[i+1])
	}
	args = append(args, idVal)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
		record.Type(), strings.Join(sets, ", "))
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// SQLite reports changed rows, so a no-op update (identical values)
		// returns 0 even though the row exists. Check existence to
		// distinguish that from a genuinely missing row.
		exists, err := s.recordExists(ctx, record)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("%s with ID %s does not exist", record.Type(), record.GetId())
		}
	}
	return nil
}

// recordExists reports whether a record of the given type exists with the
// record's id.
func (s *SqliteDbProvider) recordExists(ctx context.Context, record CrudRecord) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, "SELECT 1 FROM "+record.Type()+" WHERE id = ?", record.GetId().String()).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SqliteDbProvider) Delete(ctx context.Context, record CrudRecord) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM "+record.Type()+" WHERE id = ?", record.GetId().String())
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%s with ID %s does not exist", record.Type(), record.GetId())
	}
	return nil
}

// insertParts returns the ordered column list and values for INSERT/UPDATE of
// record. "id" is always first (from GetId/SetId); remaining columns come
// from the struct's fields by reflection.
func insertParts(record CrudRecord) ([]string, []any, error) {
	cols := []string{"id"}
	vals := []any{record.GetId().String()}

	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	err := walkFields(v, "", func(colName string, field reflect.Value) error {
		if colName == "id" {
			return nil
		}
		val, err := encodeValue(field)
		if err != nil {
			return fmt.Errorf("column %q: %w", colName, err)
		}
		cols = append(cols, colName)
		vals = append(vals, val)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return cols, vals, nil
}

// walkFields visits every scalar field of a struct, flattening nested (non-time)
// struct fields into prefixed column names (e.g. Team.Color.Name -> "color_name").
// emit is called with the final column name and the leaf reflect.Value for each
// scalar field. This lets a single value-struct like TeamColor live in the same
// table as flattened columns rather than a child table.
func walkFields(v reflect.Value, prefix string, emit func(colName string, field reflect.Value) error) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}
		colName := prefix + columnName(field)
		fv := v.Field(i)
		if !isTimeLike(fv.Type()) && fv.Kind() == reflect.Struct {
			if err := walkFields(fv, colName+"_", emit); err != nil {
				return err
			}
			continue
		}
		if err := emit(colName, fv); err != nil {
			return err
		}
	}
	return nil
}

// scanRow scans the current row into record, mapping columns back to struct
// fields by name. The "id" column is written via SetId.
func scanRow(rows *sql.Rows, cols []string, record CrudRecord) error {
	raw := make([]any, len(cols))
	dests := make([]any, len(cols))
	for i := range cols {
		dests[i] = &raw[i]
	}
	if err := rows.Scan(dests...); err != nil {
		return err
	}

	byColumn := fieldByColumn(record)
	for i, name := range cols {
		if name == "id" {
			id, err := RecordIdFromString(asString(raw[i]))
			if err != nil {
				return err
			}
			record.SetId(id)
			continue
		}
		field := byColumn[name]
		if !field.IsValid() {
			// column exists in the table but not on the struct; skip it
			continue
		}
		if field.Kind() == reflect.Interface {
			// Interface-valued field (e.g. Draft.DraftOrderPattern): let the
			// record reconstruct the concrete value from its stored string.
			if setter, ok := record.(InterfaceFieldSetter); ok {
				if err := setter.SetInterfaceField(name, asString(raw[i])); err != nil {
					return fmt.Errorf("column %q: %w", name, err)
				}
				continue
			}
			return fmt.Errorf("column %q: unsupported interface field %s", name, field.Type())
		}
		if err := setField(field, raw[i]); err != nil {
			return fmt.Errorf("column %q: %w", name, err)
		}
	}
	return nil
}

// fieldByColumn returns a map from JSON-tag column name to the struct's
// settable leaf field value (excluding the id column). Nested non-time struct
// fields are flattened into prefixed column names, mirroring insertParts.
func fieldByColumn(record CrudRecord) map[string]reflect.Value {
	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	out := map[string]reflect.Value{}
	walkFieldMap(v, "", out)
	return out
}

func walkFieldMap(v reflect.Value, prefix string, out map[string]reflect.Value) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		colName := prefix + columnName(field)
		fv := v.Field(i)
		if !isTimeLike(fv.Type()) && fv.Kind() == reflect.Struct {
			walkFieldMap(fv, colName+"_", out)
			continue
		}
		if colName == "id" {
			continue
		}
		out[colName] = fv
	}
}

// columnName returns the DB column name for a struct field: its JSON tag (if
// present and non-empty) else the snake_case field name.
func columnName(field reflect.StructField) string {
	if tag, ok := field.Tag.Lookup("json"); ok {
		name := strings.Split(tag, ",")[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return snakeCase(field.Name)
}

// isTimeLike reports whether t is time.Time or a defined type over time.Time
// (e.g. model.StartTime). Such structs are stored as RFC3339 TEXT columns; all
// other structs are treated as nested value-structs and flattened by
// walkFields / walkFieldMap.
func isTimeLike(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && t.ConvertibleTo(reflect.TypeOf(time.Time{}))
}

func snakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func encodeValue(field reflect.Value) (any, error) {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil, nil
		}
		return encodeValue(field.Elem())
	}

	switch field.Kind() {
	case reflect.String:
		return field.String(), nil
	case reflect.Bool:
		if field.Bool() {
			return int64(1), nil
		}
		return int64(0), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		// ID family: store as 16-hex TEXT to avoid signed-64 overflow.
		return fmt.Sprintf("%016x", field.Uint()), nil
	case reflect.Struct:
		if isTimeLike(field.Type()) {
			ts := field.Convert(reflect.TypeOf(time.Time{})).Interface().(time.Time)
			return ts.UTC().Format(time.RFC3339Nano), nil
		}
		return nil, fmt.Errorf("unsupported struct field %s", field.Type())
	case reflect.Interface:
		// Interface fields are persisted as a single TEXT column holding a
		// stable string identity. Concrete values exposing Name() (e.g.
		// model.DraftOrderPattern) are encoded by name; nil interfaces are
		// stored as an empty string.
		if field.IsNil() {
			return "", nil
		}
		if n, ok := field.Interface().(StringKeyed); ok {
			return n.Name(), nil
		}
		return nil, fmt.Errorf("unsupported interface field %s", field.Type())
	case reflect.Slice, reflect.Array:
		// []byte (and [N]byte) are stored as a SQLite BLOB column (e.g.
		// Photo.Contents).
		if field.Type().Elem().Kind() == reflect.Uint8 {
			return field.Bytes(), nil
		}
		return nil, fmt.Errorf("unsupported slice/array field %s", field.Type())
	default:
		return nil, fmt.Errorf("unsupported field kind %s", field.Kind())
	}
}

// StringKeyed is implemented by interface values whose concrete types expose a
// stable string name (e.g. model.DraftOrderPattern). The SQLite mapper stores
// such interface fields as a single TEXT column holding Name().
type StringKeyed interface {
	Name() string
}

// InterfaceFieldSetter is optionally implemented by CrudRecords that contain an
// interface-valued field (e.g. Draft.DraftOrderPattern) persisted as a TEXT
// column. On read, the mapper calls SetInterfaceField so the record can
// reconstruct the concrete value from its stored string. It is needed because
// the database package cannot import the model packages (import cycle).
type InterfaceFieldSetter interface {
	SetInterfaceField(field, value string) error
}

func setField(field reflect.Value, raw any) error {
	if field.Kind() == reflect.Ptr {
		if raw == nil {
			return nil
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setField(field.Elem(), raw)
	}
	if raw == nil {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(asString(raw))
	case reflect.Bool:
		field.SetBool(asInt(raw) != 0)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(asInt(raw))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(asString(raw), 16, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)
	case reflect.Struct:
		if isTimeLike(field.Type()) {
			ts, err := time.Parse(time.RFC3339Nano, asString(raw))
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(ts).Convert(field.Type()))
			return nil
		}
		return fmt.Errorf("unsupported struct field %s", field.Type())
	case reflect.Slice, reflect.Array:
		// []byte round-trips from a SQLite BLOB column (e.g. Photo.Contents).
		if field.Type().Elem().Kind() == reflect.Uint8 {
			if b, ok := raw.([]byte); ok {
				field.SetBytes(b)
			} else {
				field.SetBytes([]byte(asString(raw)))
			}
			return nil
		}
		return fmt.Errorf("unsupported slice/array field %s", field.Type())
	default:
		return fmt.Errorf("unsupported field kind %s", field.Kind())
	}
	return nil
}

func asString(raw any) string {
	switch v := raw.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

func asInt(raw any) int64 {
	switch v := raw.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	case []byte:
		i, _ := strconv.ParseInt(string(v), 10, 64)
		return i
	default:
		return 0
	}
}

func isUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
