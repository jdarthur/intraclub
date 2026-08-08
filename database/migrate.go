package database

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration is a single ordered schema migration. Name uniquely identifies
// the migration and is recorded in schema_migrations once applied; SQL is
// executed (within a transaction) when the migration is applied.
type Migration struct {
	Name string
	SQL  string
}

// EmbeddedMigrations loads every migration from the embedded migrations/
// directory, sorted by file name (which should be zero-padded and ordered,
// e.g. 0001_create_users.sql). Adding a new migration file to that directory
// is the documented path for a schema change.
func EmbeddedMigrations() ([]Migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	migs := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		b, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, err
		}
		migs = append(migs, Migration{Name: entry.Name(), SQL: string(b)})
	}

	sort.Slice(migs, func(i, j int) bool {
		return migs[i].Name < migs[j].Name
	})
	return migs, nil
}

// Migrate applies all pending migrations to db in order, recording each
// applied version in the schema_migrations table. Migrations already applied
// are skipped, so re-running Migrate is a no-op. Each migration runs in its
// own transaction; if it fails, that transaction (and its version record)
// rolls back while previously applied migrations remain in place.
func Migrate(db *sql.DB, migrations []Migration) error {
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return err
	}
	applied, err := appliedMigrations(db)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.Name] {
			continue
		}
		if err := applyMigration(db, m); err != nil {
			return fmt.Errorf("migration %q failed: %w", m.Name, err)
		}
	}
	return nil
}

func ensureSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name       TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL
	)`)
	return err
}

// appliedMigrations returns the set of migration names already recorded in
// schema_migrations.
func appliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT name FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

func applyMigration(db *sql.DB, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(m.SQL); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (name, applied_at) VALUES (?, ?)`,
		m.Name, time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		return err
	}
	return tx.Commit()
}
