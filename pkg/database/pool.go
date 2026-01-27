// Package database provides PostgreSQL connection management.
package database

//go:generate mockgen -source=pool.go -destination=mock/mock_pool.go -package=mock

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IPool defines the interface for database pool operations used by repositories.
// This interface enables dependency injection and easier testing.
type IPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// PoolWrapper wraps *pgxpool.Pool to implement IPool interface.
type PoolWrapper struct {
	pool *pgxpool.Pool
}

// Compile-time check to ensure PoolWrapper implements IPool
var _ IPool = (*PoolWrapper)(nil)

// NewPoolWrapper creates a new PoolWrapper from a pgxpool.Pool
func NewPoolWrapper(pool *pgxpool.Pool) *PoolWrapper {
	return &PoolWrapper{pool: pool}
}

// Exec executes a query that doesn't return rows
func (p *PoolWrapper) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, sql, arguments...)
}

// QueryRow executes a query that returns at most one row
func (p *PoolWrapper) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

// Query executes a query that returns rows
func (p *PoolWrapper) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

// Begin starts a transaction
func (p *PoolWrapper) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.pool.Begin(ctx)
}

// GetUnderlyingPool returns the underlying pgxpool.Pool
func (p *PoolWrapper) GetUnderlyingPool() *pgxpool.Pool {
	return p.pool
}
