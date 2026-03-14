package repository

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolIface defines the methods of pgxpool.Pool used by our repositories.
// This interface enables context-based queries and better testability.
type PoolIface interface {
	pgxscan.Querier
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (commandTag pgconn.CommandTag, err error)
	Close()
}

type BeginPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// DB represents the database connection structure used by repositories
type DB struct {
	Pool PoolIface
}

// NewDB creates a new database connection structure
// In production: pass the real *pgxpool.Pool instance
// In tests: pass a mock that implements PoolIface
func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{
		Pool: pool,
	}
}

// Close calls the Close method on the pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
