// Package idempotency provides idempotency key management for safe request retries.
package idempotency

import (
	"context"
	"time"

	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/idempotency/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IModule defines the interface for the idempotency module.
type IModule interface {
	GetRepository() IRepository
	StartCleanupWorker(ctx context.Context, ttl time.Duration, interval time.Duration)
	StopCleanupWorker()
}

// Module implements IModule.
type Module struct {
	Repo       IRepository
	cancelFunc context.CancelFunc
}

// Compile-time interface check
var _ IModule = (*Module)(nil)

// NewModule creates a new idempotency module.
func NewModule(_ context.Context, pool *pgxpool.Pool) IModule {
	repo := NewRepository(pool)
	return &Module{
		Repo: repo,
	}
}

// GetRepository returns the idempotency repository.
func (m *Module) GetRepository() IRepository {
	return m.Repo
}

// StartCleanupWorker starts a background goroutine that periodically cleans up expired keys.
func (m *Module) StartCleanupWorker(ctx context.Context, ttl time.Duration, interval time.Duration) {
	workerCtx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel

	go m.runCleanupLoop(workerCtx, ttl, interval)
}

// StopCleanupWorker stops the cleanup worker.
func (m *Module) StopCleanupWorker() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
}

// runCleanupLoop runs the periodic cleanup of expired idempotency keys.
func (m *Module) runCleanupLoop(ctx context.Context, ttl time.Duration, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.cleanupExpiredKeys(ctx, ttl)
		}
	}
}

// cleanupExpiredKeys deletes expired idempotency keys and logs the result.
func (m *Module) cleanupExpiredKeys(ctx context.Context, ttl time.Duration) {
	deleted, err := m.Repo.DeleteExpired(ctx, ttl)
	if err != nil {
		logger.Error(entities.LogMsgIdempotencyCleanupFailed, "error", err)
		return
	}

	if deleted > 0 {
		logger.Info(entities.LogMsgIdempotencyCleanup, entities.LogFieldDeletedCount, deleted)
	}
}
