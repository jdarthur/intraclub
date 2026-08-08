package database

import (
	"testing"
)

// TestSkeletonMigrationApplies confirms the embedded skeleton migration executes
// cleanly through the migration runner and is recorded in schema_migrations.
func TestSkeletonMigrationApplies(t *testing.T) {
	migs, err := EmbeddedMigrations()
	if err != nil {
		t.Fatalf("EmbeddedMigrations: %v", err)
	}
	if len(migs) == 0 || migs[0].Name != "0001_schema_skeleton.sql" {
		t.Fatalf("first migration = %v, want 0001_schema_skeleton.sql", migs)
	}
	db := openTestSqlite(t)
	if err := Migrate(db, migs); err != nil {
		t.Fatalf("Migrate(skeleton): %v", err)
	}
	if names := schemaMigrationNames(db); !contains(names, "0001_schema_skeleton.sql") {
		t.Fatalf("schema_migrations = %v, want it to contain 0001_schema_skeleton.sql", names)
	}
	// idempotent on re-run
	if err := Migrate(db, migs); err != nil {
		t.Fatalf("re-Migrate(skeleton): %v", err)
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
