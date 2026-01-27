// Package database provides PostgreSQL connection management.
package database

import (
	"context"
	"fmt"
	"time"

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

// InitializeWithRetry attempts to connect to the database with exponential backoff.
// If retry is disabled in config, it falls back to single attempt Initialize.
func InitializeWithRetry(ctx context.Context, cfg *config.DatabaseConfig) (*Database, error) {
	if !cfg.Retry.Enabled {
		return Initialize(ctx, cfg)
	}

	return connectWithRetry(ctx, cfg)
}

// connectWithRetry implements the retry loop with exponential backoff
func connectWithRetry(ctx context.Context, cfg *config.DatabaseConfig) (*Database, error) {
	var lastErr error
	backoff := cfg.Retry.GetInitialBackoff()
	maxBackoff := cfg.Retry.GetMaxBackoff()

	for attempt := 1; attempt <= cfg.Retry.MaxRetries; attempt++ {
		logConnectionAttempt(attempt, cfg.Retry.MaxRetries)

		db, err := Initialize(ctx, cfg)
		if err == nil {
			logConnectionSuccess(attempt)
			return db, nil
		}

		lastErr = err
		if attempt < cfg.Retry.MaxRetries {
			logRetryWithBackoff(attempt, cfg.Retry.MaxRetries, backoff, err)
			sleepWithContext(ctx, backoff)
			backoff = calculateNextBackoff(backoff, maxBackoff)
		}
	}

	logConnectionFailed(cfg.Retry.MaxRetries, lastErr)
	return nil, fmt.Errorf(constants.ErrFmtDBConnectionFailedRetry, cfg.Retry.MaxRetries, lastErr)
}

// logConnectionAttempt logs the connection attempt
func logConnectionAttempt(attempt, maxRetries int) {
	logger.Info(constants.LogMsgDBConnectionAttempt,
		constants.LogFieldAttempt, attempt,
		constants.LogFieldMaxRetries, maxRetries,
	)
}

// logConnectionSuccess logs successful connection
func logConnectionSuccess(attempt int) {
	logger.Info(constants.LogMsgDBConnectionSuccess,
		constants.LogFieldAttempt, attempt,
	)
}

// logRetryWithBackoff logs the retry attempt with backoff duration
func logRetryWithBackoff(attempt, maxRetries int, backoff time.Duration, err error) {
	logger.Warn(constants.LogMsgDBConnectionRetry,
		constants.LogFieldAttempt, attempt,
		constants.LogFieldMaxRetries, maxRetries,
		constants.LogFieldNextBackoff, backoff.String(),
		constants.LogKeyError, err,
	)
}

// logConnectionFailed logs the final connection failure
func logConnectionFailed(maxRetries int, err error) {
	logger.Error(constants.LogMsgDBConnectionFailed,
		constants.LogFieldMaxRetries, maxRetries,
		constants.LogKeyError, err,
	)
}

// sleepWithContext sleeps for the given duration or until context is cancelled
func sleepWithContext(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(duration):
		return
	}
}

// calculateNextBackoff doubles the backoff up to maxBackoff
func calculateNextBackoff(current, max time.Duration) time.Duration {
	next := current * 2
	if next > max {
		return max
	}
	return next
}
