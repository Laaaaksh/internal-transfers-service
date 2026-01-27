// Package entities provides types and constants for the idempotency module.
package entities

// Error messages for the idempotency module
const (
	ErrMsgKeyNotFound  = "idempotency key not found"
	ErrMsgKeyTooLong   = "idempotency key exceeds maximum length"
	ErrMsgStoreFailed  = "failed to store idempotency key"
	ErrMsgGetFailed    = "failed to get idempotency key"
	ErrMsgDeleteFailed = "failed to delete expired keys"
)

// Error response JSON template parts
const (
	ErrResponsePrefix = `{"error":"`
	ErrResponseCode   = `","code":"`
	ErrResponseSuffix = `"}`
)

// Validation constants
const (
	// MaxKeyLength is the maximum allowed length for idempotency keys (matches DB VARCHAR(255))
	MaxKeyLength = 255

	// MaxCachedResponseSize is the maximum response body size to cache (64 KB)
	// Larger responses will not be cached for idempotency
	MaxCachedResponseSize = 64 * 1024
)

// Database column names
const (
	ColKey            = "key"
	ColResponseStatus = "response_status"
	ColResponseBody   = "response_body"
	ColCreatedAt      = "created_at"
)

// Log message constants
const (
	LogMsgIdempotencyHit            = "Idempotency cache hit, returning cached response"
	LogMsgIdempotencyMiss           = "Idempotency cache miss, processing request"
	LogMsgIdempotencyStored         = "Stored idempotency response"
	LogMsgIdempotencyStoreFailed    = "Failed to store idempotency response"
	LogMsgIdempotencyGetFailed      = "Failed to check idempotency key"
	LogMsgIdempotencyKeyTooLong     = "Idempotency key exceeds maximum length"
	LogMsgIdempotencyCleanup        = "Cleaned up expired idempotency keys"
	LogMsgIdempotencyCleanupFailed  = "Failed to cleanup expired idempotency keys"
	LogMsgIdempotencyResponseTooBig = "Response too large to cache for idempotency"
)

// Log field keys
const (
	LogFieldIdempotencyKey = "idempotency_key"
	LogFieldCachedStatus   = "cached_status"
	LogFieldDeletedCount   = "deleted_count"
	LogFieldKeyLength      = "key_length"
)

// Header constants
const (
	HeaderIdempotentReplayed = "X-Idempotent-Replayed"
	HeaderValueTrue          = "true"
)
