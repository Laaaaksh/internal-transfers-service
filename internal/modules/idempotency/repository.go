package idempotency

//go:generate mockgen -source=repository.go -destination=mock/mock_repository.go -package=mock

import (
	"context"
	"time"

	"github.com/internal-transfers-service/internal/modules/idempotency/entities"
)

// IRepository defines the interface for idempotency storage operations.
type IRepository interface {
	// Get retrieves a cached response by key.
	// Returns nil, nil if the key doesn't exist.
	Get(ctx context.Context, key string) (*entities.IdempotencyRecord, error)

	// Store saves a response for future retrieval.
	Store(ctx context.Context, key string, status int, body []byte) error

	// DeleteExpired removes keys older than the specified TTL.
	// Returns the number of deleted keys.
	DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error)
}
