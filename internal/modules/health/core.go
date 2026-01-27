package health

import (
	"context"
	"sync/atomic"

	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/pkg/database"
)

// ICore defines the interface for health check operations
type ICore interface {
	RunLivenessCheck(ctx context.Context) (string, int)
	RunReadinessCheck(ctx context.Context) (string, int)
	MarkUnhealthy()
	IsHealthy() bool
}

// Core implements health check operations
type Core struct {
	db      *database.Database
	healthy int32 // atomic flag for health status
}

// Compile-time interface check
var _ ICore = (*Core)(nil)

// NewCore creates a new health check core
func NewCore(db *database.Database) ICore {
	return &Core{
		db:      db,
		healthy: 1, // Start as healthy
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// RunLivenessCheck checks if the service is alive
func (c *Core) RunLivenessCheck(ctx context.Context) (string, int) {
	// Liveness just checks if the process is running
	if !c.IsHealthy() {
		return constants.StatusNotServing, 503
	}
	return constants.StatusServing, 200
}

// RunReadinessCheck checks if the service is ready to accept requests
func (c *Core) RunReadinessCheck(ctx context.Context) (string, int) {
	// Check if service is marked as healthy
	if !c.IsHealthy() {
		return constants.StatusNotServing, 503
	}

	// Check database connectivity
	if c.db != nil {
		if err := c.db.Ping(ctx); err != nil {
			logger.Ctx(ctx).Warnw("Readiness check failed - database ping failed",
				"error", err,
			)
			return constants.StatusNotServing, 503
		}
	}

	return constants.StatusServing, 200
}

// MarkUnhealthy marks the service as unhealthy for graceful shutdown
func (c *Core) MarkUnhealthy() {
	atomic.StoreInt32(&c.healthy, 0)
	logger.Info("Service marked as unhealthy")
}

// IsHealthy returns whether the service is healthy
func (c *Core) IsHealthy() bool {
	return atomic.LoadInt32(&c.healthy) == 1
}
