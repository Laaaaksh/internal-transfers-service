// Package database provides PostgreSQL connection management.
package database

import (
	"context"
	"fmt"

	"github.com/internal-transfers-service/internal/config"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the global database connection pool
var Pool *pgxpool.Pool

// IDatabase is the interface for database operations
type IDatabase interface {
	GetPool() *pgxpool.Pool
	Close()
	Ping(ctx context.Context) error
}

// Database implements IDatabase
type Database struct {
	pool *pgxpool.Pool
}

// Initialize creates and configures the database connection pool
func Initialize(ctx context.Context, cfg *config.DatabaseConfig) (*Database, error) {
	connString := cfg.GetConnectionString()

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrFmtFailedToParseConnString, err)
	}

	// Configure pool settings
	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MinConns = cfg.MinConnections
	poolConfig.MaxConnLifetime = cfg.GetMaxConnLifetime()
	poolConfig.MaxConnIdleTime = cfg.GetMaxConnIdleTime()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrFmtFailedToCreateConnPool, err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf(constants.ErrFmtFailedToPingDB, err)
	}

	Pool = pool
	logger.Info(constants.LogMsgDBPoolInitialized,
		constants.LogFieldHost, cfg.Host,
		constants.LogFieldPort, cfg.Port,
		constants.LogFieldDatabase, cfg.Name,
		constants.LogFieldMaxConnections, cfg.MaxConnections,
	)

	return &Database{pool: pool}, nil
}

// GetPool returns the connection pool
func (d *Database) GetPool() *pgxpool.Pool {
	return d.pool
}

// Close closes the database connection pool
func (d *Database) Close() {
	if d.pool != nil {
		d.pool.Close()
		logger.Info(constants.LogMsgDBPoolClosed)
	}
}

// Ping checks the database connection
func (d *Database) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

// GetStats returns the current pool statistics
func (d *Database) GetStats() *pgxpool.Stat {
	if d.pool != nil {
		return d.pool.Stat()
	}
	return nil
}
