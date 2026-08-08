package database

import (
	"context"
	"errors"
	"path/filepath"
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

func TestNewProviderSqlite(t *testing.T) {
	db, err := NewProvider(context.Background(), ProviderConfig{
		Kind: ProviderSqlite,
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("NewProvider(sqlite) returned error: %v", err)
	}
	if db == nil {
		t.Fatal("NewProvider(sqlite) returned nil provider")
	}
	p, ok := db.(*SqliteDbProvider)
	if !ok {
		t.Fatalf("NewProvider(sqlite) returned %T, want *SqliteDbProvider", db)
	}
	if err := p.Disconnect(); err != nil {
		t.Fatalf("Disconnect: %v", err)
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
