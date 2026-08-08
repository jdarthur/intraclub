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
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}
		name := columnName(field)
		if name == "id" {
			continue
		}
		val, err := encodeValue(v.Field(i))
		if err != nil {
			return nil, nil, fmt.Errorf("column %q: %w", name, err)
		}
		cols = append(cols, name)
		vals = append(vals, val)
	}
	return cols, vals, nil
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
		if err := setField(field, raw[i]); err != nil {
			return fmt.Errorf("column %q: %w", name, err)
		}
	}
	return nil
}

// fieldByColumn returns a map from JSON-tag column name to the struct's
// settable field value (excluding the id column).
func fieldByColumn(record CrudRecord) map[string]reflect.Value {
	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	out := map[string]reflect.Value{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := columnName(field)
		if name == "id" {
			continue
		}
		out[name] = v.Field(i)
	}
	return out
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
		if field.Type() == reflect.TypeOf(time.Time{}) {
			return field.Interface().(time.Time).UTC().Format(time.RFC3339Nano), nil
		}
		return nil, fmt.Errorf("unsupported struct field %s", field.Type())
	default:
		return nil, fmt.Errorf("unsupported field kind %s", field.Kind())
	}
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
		if field.Type() == reflect.TypeOf(time.Time{}) {
			ts, err := time.Parse(time.RFC3339Nano, asString(raw))
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(ts))
			return nil
		}
		return fmt.Errorf("unsupported struct field %s", field.Type())
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
