package database

import (
	"context"
	"path/filepath"
	"testing"
)

// sqliteTestRecord is a minimal CrudRecord used to exercise the sqlite
// provider's reflection-based column mapping without importing the model
// package. It covers a hex-string ID, string, int, and bool.
type sqliteTestRecord struct {
	ID      RecordId `json:"id"`
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Enabled bool     `json:"enabled"`
}

func (r *sqliteTestRecord) Type() string                          { return "test_record" }
func (r *sqliteTestRecord) GetId() RecordId                       { return r.ID }
func (r *sqliteTestRecord) SetId(id RecordId)                     { r.ID = id }
func (r *sqliteTestRecord) GetOwner() UserId                      { return SysAdminUserId }
func (r *sqliteTestRecord) SetOwner(UserId)                       {}
func (r *sqliteTestRecord) EditableBy(context.Context, Provider) []UserId { return []UserId{SysAdminUserId} }
func (r *sqliteTestRecord) AccessibleTo(context.Context, Provider) []UserId { return []UserId{EveryoneUserId} }
func (r *sqliteTestRecord) StaticallyValid() error                { return nil }
func (r *sqliteTestRecord) DynamicallyValid(context.Context, Provider) error { return nil }
func (r *sqliteTestRecord) NewRecord() CrudRecord                 { return &sqliteTestRecord{} }

const testRecordMigration = "CREATE TABLE test_record (id TEXT PRIMARY KEY, name TEXT NOT NULL, age INTEGER NOT NULL, enabled INTEGER NOT NULL);"

func newTestSqliteProvider(t *testing.T) *SqliteDbProvider {
	t.Helper()
	p, err := newSqliteDbProvider(filepath.Join(t.TempDir(), "test.db"), []Migration{
		{Name: "001_test_record", SQL: testRecordMigration},
	})
	if err != nil {
		t.Fatalf("newSqliteDbProvider: %v", err)
	}
	t.Cleanup(func() { p.Disconnect() })
	return p
}

func TestSqliteProviderCRUDRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newTestSqliteProvider(t)

	// Create assigns an id when one is not set
	created, err := p.Create(ctx, &sqliteTestRecord{Name: "alice", Age: 30, Enabled: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	rec := created.(*sqliteTestRecord)
	if rec.GetId() == InvalidRecordId {
		t.Fatal("Create did not assign a RecordId")
	}

	// GetOne round-trips every field
	got, exists, err := p.GetOne(ctx, &sqliteTestRecord{ID: rec.GetId()})
	if err != nil {
		t.Fatalf("GetOne: %v", err)
	}
	if !exists {
		t.Fatal("GetOne: record not found")
	}
	g := got.(*sqliteTestRecord)
	if g.GetId() != rec.GetId() || g.Name != "alice" || g.Age != 30 || g.Enabled != true {
		t.Fatalf("round-trip mismatch, got %+v", g)
	}

	// GetAll returns the record
	all, err := p.GetAll(ctx, &sqliteTestRecord{})
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("GetAll: expected 1 record, got %d", len(all))
	}

	// GetAllWhere filters in Go
	filtered, err := GetAllWhere(ctx, p, func(_ context.Context, r *sqliteTestRecord) bool { return r.Enabled })
	if err != nil {
		t.Fatalf("GetAllWhere: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("GetAllWhere: expected 1 record, got %d", len(filtered))
	}
	filtered, err = GetAllWhere(ctx, p, func(_ context.Context, r *sqliteTestRecord) bool { return !r.Enabled })
	if err != nil {
		t.Fatalf("GetAllWhere: %v", err)
	}
	if len(filtered) != 0 {
		t.Fatalf("GetAllWhere: expected 0 records, got %d", len(filtered))
	}

	// Update persists changes
	g.Name = "bob"
	if err := p.Update(ctx, g); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got2, _, _ := p.GetOne(ctx, &sqliteTestRecord{ID: g.GetId()})
	if got2.(*sqliteTestRecord).Name != "bob" {
		t.Fatalf("Update did not persist, got %+v", got2)
	}

	// Delete removes the record
	if err := p.Delete(ctx, g); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, exists, _ = p.GetOne(ctx, &sqliteTestRecord{ID: g.GetId()})
	if exists {
		t.Fatal("record should have been deleted")
	}
}

func TestSqliteProviderDuplicateCreate(t *testing.T) {
	ctx := context.Background()
	p := newTestSqliteProvider(t)

	created, err := p.Create(ctx, &sqliteTestRecord{Name: "alice"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	rec := created.(*sqliteTestRecord)

	// creating a second record with the same id must fail
	if _, err := p.Create(ctx, &sqliteTestRecord{ID: rec.GetId()}); err == nil {
		t.Fatal("expected duplicate-id create to fail")
	}
}

func TestSqliteProviderUpdateDeleteMissing(t *testing.T) {
	ctx := context.Background()
	p := newTestSqliteProvider(t)

	missing := &sqliteTestRecord{ID: NewRecordId()}
	if err := p.Update(ctx, missing); err == nil {
		t.Fatal("expected Update of missing record to fail")
	}
	if err := p.Delete(ctx, missing); err == nil {
		t.Fatal("expected Delete of missing record to fail")
	}
}

func TestSqliteProviderNoOpUpdateSucceeds(t *testing.T) {
	// SQLite's RowsAffected counts changed rows, so updating a record to the
	// identical values returns 0 rows affected. That must be treated as a
	// successful update (the row exists), not "does not exist".
	ctx := context.Background()
	p := newTestSqliteProvider(t)

	created, err := p.Create(ctx, &sqliteTestRecord{Name: "alice", Age: 30, Enabled: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	rec := created.(*sqliteTestRecord)

	if err := p.Update(ctx, rec); err != nil {
		t.Fatalf("no-op Update should succeed, got: %v", err)
	}
}

// sqliteSliceTestRecord carries a []UserId field to exercise the JSON TEXT
// column mapping for non-[]byte slices.
type sqliteSliceTestRecord struct {
	ID       RecordId `json:"id"`
	Owner    UserId   `json:"owner"`
	SharedTo []UserId `json:"shared_to"`
}

func (r *sqliteSliceTestRecord) Type() string                        { return "test_slice_record" }
func (r *sqliteSliceTestRecord) GetId() RecordId                     { return r.ID }
func (r *sqliteSliceTestRecord) SetId(id RecordId)                   { r.ID = id }
func (r *sqliteSliceTestRecord) GetOwner() UserId                    { return r.Owner }
func (r *sqliteSliceTestRecord) SetOwner(UserId)                     {}
func (r *sqliteSliceTestRecord) EditableBy(context.Context, Provider) []UserId { return []UserId{r.Owner} }
func (r *sqliteSliceTestRecord) AccessibleTo(context.Context, Provider) []UserId { return []UserId{r.Owner} }
func (r *sqliteSliceTestRecord) StaticallyValid() error              { return nil }
func (r *sqliteSliceTestRecord) DynamicallyValid(context.Context, Provider) error { return nil }
func (r *sqliteSliceTestRecord) NewRecord() CrudRecord               { return &sqliteSliceTestRecord{} }

const testSliceRecordMigration = "CREATE TABLE test_slice_record (id TEXT PRIMARY KEY, owner TEXT, shared_to TEXT);"

func TestSqliteProviderSliceRoundTrip(t *testing.T) {
	ctx := context.Background()
	p, err := newSqliteDbProvider(filepath.Join(t.TempDir(), "slice.db"), []Migration{
		{Name: "001_test_slice_record", SQL: testSliceRecordMigration},
	})
	if err != nil {
		t.Fatalf("newSqliteDbProvider: %v", err)
	}
	t.Cleanup(func() { p.Disconnect() })

	sharedTo := []UserId{UserId(NewRecordId()), EveryoneUserId}
	created, err := p.Create(ctx, &sqliteSliceTestRecord{Owner: UserId(NewRecordId()), SharedTo: sharedTo})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	rec := created.(*sqliteSliceTestRecord)

	got, exists, err := p.GetOne(ctx, &sqliteSliceTestRecord{ID: rec.GetId()})
	if err != nil {
		t.Fatalf("GetOne: %v", err)
	}
	if !exists {
		t.Fatal("record not found")
	}
	g := got.(*sqliteSliceTestRecord)
	if len(g.SharedTo) != len(sharedTo) || g.SharedTo[0] != sharedTo[0] || g.SharedTo[1] != sharedTo[1] {
		t.Fatalf("slice round-trip mismatch, got %+v", g.SharedTo)
	}
}
