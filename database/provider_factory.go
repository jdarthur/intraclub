package database

import (
	"context"
	"errors"
	"fmt"
)

// ProviderKind identifies a supported database.Provider implementation
// that the running server can be wired to at startup.
type ProviderKind string

const (
	// ProviderMemory is an in-memory, map-backed provider. It is used by the
	// unit/contract tests and is a valid option for ephemeral local runs.
	ProviderMemory ProviderKind = "memory"

	// ProviderSqlite is the file-backed SQLite provider. Construction and
	// connection setup are implemented in issue #53; until then selecting it
	// returns an explicit "not yet implemented" error rather than failing
	// silently at startup.
	ProviderSqlite ProviderKind = "sqlite"
)

// ProviderConfig holds the settings needed to construct and open a Provider.
type ProviderConfig struct {
	Kind ProviderKind
	// Path is the SQLite database file path. Ignored for in-memory providers.
	Path string
}

// ErrUnknownProvider is returned by NewProvider when the requested
// ProviderKind is not a known provider.
var ErrUnknownProvider = errors.New("unknown database provider")

// ErrSqliteNotImplemented is returned by NewProvider when the SQLite provider
// is selected before it has been implemented (issue #53).
var ErrSqliteNotImplemented = errors.New("sqlite database provider is not implemented")

// NewProvider constructs, opens, and (for persistent providers) runs any
// pending migrations on the requested Provider, returning the ready-to-use
// Provider. The Provider interface has no Connect/Disconnect, so connection
// setup and migration live here, on the concrete type, rather than in the
// interface.
func NewProvider(ctx context.Context, cfg ProviderConfig) (Provider, error) {
	switch cfg.Kind {
	case ProviderMemory:
		return NewUnitTestDBProvider(), nil
	case ProviderSqlite:
		// TODO(#53): construct + open the SQLite provider here and run
		// pending migrations (issue #52) before returning it.
		return nil, fmt.Errorf("%w (see issue #53): %q", ErrSqliteNotImplemented, ProviderSqlite)
	default:
		return nil, fmt.Errorf("%w: %q (expected %q or %q)", ErrUnknownProvider, cfg.Kind, ProviderMemory, ProviderSqlite)
	}
}
