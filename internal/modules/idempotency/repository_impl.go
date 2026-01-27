package idempotency

import (
	"context"
	"errors"
	"time"

	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/modules/idempotency/entities"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements IRepository using PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// Compile-time interface check
var _ IRepository = (*Repository)(nil)

// NewRepository creates a new idempotency repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Get retrieves a cached response by key.
// Returns nil, nil if the key doesn't exist.
func (r *Repository) Get(ctx context.Context, key string) (*entities.IdempotencyRecord, error) {
	query := `
		SELECT key, response_status, response_body, created_at
		FROM idempotency_keys
		WHERE key = $1
	`

	record := &entities.IdempotencyRecord{}
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&record.Key,
		&record.ResponseStatus,
		&record.ResponseBody,
		&record.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}

// Store saves a response for future retrieval.
func (r *Repository) Store(ctx context.Context, key string, status int, body []byte) error {
	query := `
		INSERT INTO idempotency_keys (key, response_status, response_body, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, key, status, body, time.Now())
	return err
}

// DeleteExpired removes keys older than the specified TTL.
// Returns the number of deleted keys.
func (r *Repository) DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error) {
	query := `
		DELETE FROM ` + constants.TableIdempotencyKeys + `
		WHERE created_at < $1
	`

	cutoff := time.Now().Add(-ttl)
	result, err := r.pool.Exec(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
