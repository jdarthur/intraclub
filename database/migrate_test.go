package database

import (
	"database/sql"
	"testing"
)

func openTestSqlite(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func tableExists(db *sql.DB, name string) bool {
	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", name)
	var n string
	return row.Scan(&n) == nil
}

func schemaMigrationNames(db *sql.DB) []string {
	rows, err := db.Query("SELECT name FROM schema_migrations ORDER BY name")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			names = append(names, n)
		}
	}
	return names
}

func TestMigrateAppliesInOrder(t *testing.T) {
	db := openTestSqlite(t)
	migs := []Migration{
		{Name: "001_a", SQL: "CREATE TABLE a (id TEXT PRIMARY KEY);"},
		{Name: "002_b", SQL: "CREATE TABLE b (id TEXT PRIMARY KEY);"},
	}

	if err := Migrate(db, migs); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if !tableExists(db, "a") || !tableExists(db, "b") {
		t.Fatal("expected tables a and b to exist")
	}
	names := schemaMigrationNames(db)
	if len(names) != 2 || names[0] != "001_a" || names[1] != "002_b" {
		t.Fatalf("schema_migrations = %v, want [001_a 002_b]", names)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	db := openTestSqlite(t)
	migs := []Migration{{Name: "001_a", SQL: "CREATE TABLE a (id TEXT PRIMARY KEY);"}}

	if err := Migrate(db, migs); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := Migrate(db, migs); err != nil {
		t.Fatalf("second Migrate (should be no-op): %v", err)
	}
	if names := schemaMigrationNames(db); len(names) != 1 {
		t.Fatalf("expected 1 applied migration, got %v", names)
	}
}

func TestMigrateSkipsAlreadyApplied(t *testing.T) {
	db := openTestSqlite(t)
	migs := []Migration{
		{Name: "001_a", SQL: "CREATE TABLE a (id TEXT PRIMARY KEY);"},
		{Name: "002_b", SQL: "CREATE TABLE b (id TEXT PRIMARY KEY);"},
	}

	// apply the first, then run the full set: only 002 should run
	if err := Migrate(db, migs[:1]); err != nil {
		t.Fatalf("partial Migrate: %v", err)
	}
	if err := Migrate(db, migs); err != nil {
		t.Fatalf("full Migrate: %v", err)
	}
	if names := schemaMigrationNames(db); len(names) != 2 {
		t.Fatalf("expected 2 applied migrations, got %v", names)
	}
}

func TestMigrateRollsBackOnFailure(t *testing.T) {
	db := openTestSqlite(t)
	migs := []Migration{
		{Name: "001_ok", SQL: "CREATE TABLE ok (id TEXT PRIMARY KEY);"},
		{Name: "002_bad", SQL: "THIS IS NOT SQL;"},
	}

	if err := Migrate(db, migs); err == nil {
		t.Fatal("expected error applying bad migration")
	}
	if !tableExists(db, "ok") {
		t.Fatal("expected migration 001_ok to be applied")
	}
	if names := schemaMigrationNames(db); len(names) != 1 {
		t.Fatalf("expected only 001_ok recorded, got %v", names)
	}
}

func TestEmbeddedMigrations(t *testing.T) {
	migs, err := EmbeddedMigrations()
	if err != nil {
		t.Fatalf("EmbeddedMigrations: %v", err)
	}
	if len(migs) == 0 {
		t.Fatal("expected at least one embedded migration")
	}
	if migs[0].Name != "0001_create_users.sql" {
		t.Fatalf("first migration = %q, want 0001_create_users.sql", migs[0].Name)
	}
}
