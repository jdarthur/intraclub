package database

import (
	"context"
	"errors"
	"testing"
)

func TestNewProviderMemory(t *testing.T) {
	db, err := NewProvider(context.Background(), ProviderConfig{Kind: ProviderMemory})
	if err != nil {
		t.Fatalf("NewProvider(memory) returned error: %v", err)
	}
	if db == nil {
		t.Fatal("NewProvider(memory) returned nil provider")
	}
	if _, ok := db.(*UnitTestDbProvider); !ok {
		t.Fatalf("NewProvider(memory) returned %T, want *UnitTestDbProvider", db)
	}
}

func TestNewProviderSqliteNotImplemented(t *testing.T) {
	// NOTE: this branch is implemented in issue #53; this test should be
	// updated to assert a working provider once the SQLite provider lands.
	_, err := NewProvider(context.Background(), ProviderConfig{Kind: ProviderSqlite, Path: ":memory:"})
	if err == nil {
		t.Fatal("NewProvider(sqlite) returned nil error, want not-implemented error")
	}
	if !errors.Is(err, ErrSqliteNotImplemented) {
		t.Fatalf("NewProvider(sqlite) error = %v, want ErrSqliteNotImplemented", err)
	}
}

func TestNewProviderUnknownKind(t *testing.T) {
	_, err := NewProvider(context.Background(), ProviderConfig{Kind: ProviderKind("bogus")})
	if err == nil {
		t.Fatal("NewProvider(unknown) returned nil error, want error")
	}
	if !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("NewProvider(unknown) error = %v, want ErrUnknownProvider", err)
	}
}
