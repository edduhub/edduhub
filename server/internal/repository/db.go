package repository

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolIface defines the methods of pgxpool.Pool used by our repositories.
// This interface enables context-based queries and better testability.
type PoolIface interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (commandTag pgconn.CommandTag, err error)
	Close()
	// Add other methods from pgxpool.Pool if needed by any repository
}

// DB represents the database connection structure used by repositories
type DB struct {
	Pool *pgxpool.Pool // Direct pgxpool.Pool reference for connection pooling
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