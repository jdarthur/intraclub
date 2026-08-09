package database

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"
)

// This file implements the provider-parameterization harness for the
// database.Provider contract tests (issue #62). The contract tests in
// database_provider_test.go, db_with_access_control_test.go,
// validatable_test.go, and unique_constraint_test.go each exercise the same
// CRUD / access-control / validation / uniqueness behavior; this harness lets
// a single test body run against every supported provider: the in-memory
// map-backed provider, SQLite in-memory, and SQLite on-disk.

// testProviderKind identifies a database.Provider backend the contract tests
// run against.
type testProviderKind string

const (
	testProviderMemory       testProviderKind = "memory"
	testProviderSqliteMemory testProviderKind = "sqlite-memory"
	testProviderSqliteDisk   testProviderKind = "sqlite-disk"
)

// contractTestKinds returns every provider kind the contract tests run against.
func contractTestKinds() []testProviderKind {
	return []testProviderKind{
		testProviderMemory,
		testProviderSqliteMemory,
		testProviderSqliteDisk,
	}
}

// contractTestSchema creates the tables needed by the contract-test record
// types (testRecord, testUnique, staticFailRecord, dynamicFailRecord,
// PrivateTestRecord). Table names equal record.Type() and column names/types
// follow the reflection-based column mapping in sqlite_db.go (see
// docs/schema-conventions.md). The private_record.shared_to column stores the
// []UserId slice as JSON TEXT.
const contractTestSchema = `
CREATE TABLE test_record (
	id         TEXT PRIMARY KEY,
	owner      TEXT,
	created_at TEXT,
	updated_at TEXT
);
CREATE TABLE test_unique (
	id            TEXT PRIMARY KEY,
	reference_id1 TEXT,
	reference_id2 TEXT
);
CREATE TABLE static_fail (
	id    TEXT PRIMARY KEY,
	value TEXT
);
CREATE TABLE dynamic_fail (
	id    TEXT PRIMARY KEY,
	value TEXT
);
CREATE TABLE private_record (
	id        TEXT PRIMARY KEY,
	owner     TEXT,
	shared_to TEXT,
	value     TEXT
);
`

// contractSqliteMemorySeq ensures each SQLite in-memory contract-test provider
// gets a unique shared-cache database name, keeping provider instances isolated.
var contractSqliteMemorySeq int64

// newContractTestDB returns a fresh, isolated database.Provider of the given
// kind with the contract-test schema applied (SQLite kinds) and disconnected
// when the test completes. Each call returns an independent provider, so no
// records leak between contract tests.
func newContractTestDB(t *testing.T, kind testProviderKind) Provider {
	t.Helper()
	switch kind {
	case testProviderMemory:
		return NewUnitTestDBProvider()
	case testProviderSqliteMemory, testProviderSqliteDisk:
		var path string
		if kind == testProviderSqliteMemory {
			// A uniquely-named shared-cache in-memory database: cache=shared
			// makes every connection in the provider's pool observe the same
			// data, and the unique name keeps each provider instance isolated
			// (an unnamed ":memory:" shared cache is process-wide and would
			// leak records between tests).
			path = fmt.Sprintf("contract-mem-%d?mode=memory&cache=shared", atomic.AddInt64(&contractSqliteMemorySeq, 1))
		} else {
			path = filepath.Join(t.TempDir(), "contract.db")
		}
		p, err := newSqliteDbProvider(path, []Migration{
			{Name: "001_contract_tests", SQL: contractTestSchema},
		})
		if err != nil {
			t.Fatalf("newSqliteDbProvider: %v", err)
		}
		t.Cleanup(func() { p.Disconnect() })
		return p
	default:
		t.Fatalf("unknown contract provider kind %q", kind)
		return nil
	}
}

// contractTest pairs a contract-test name with its body. The body takes the
// provider to run against, so it can be shared across every provider kind.
type contractTest struct {
	name string
	fn   func(t *testing.T, db Provider)
}

// runContractTests runs every contract test body against every provider kind,
// giving each (kind, test) a fresh, isolated provider. This is the driver that
// makes the database.Provider contract tests run against memory, SQLite
// in-memory, and SQLite on-disk (issue #62).
func runContractTests(t *testing.T, tests []contractTest) {
	t.Helper()
	for _, kind := range contractTestKinds() {
		t.Run(string(kind), func(t *testing.T) {
			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					tc.fn(t, newContractTestDB(t, kind))
				})
			}
		})
	}
}
