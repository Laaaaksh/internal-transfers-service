// Package contextkeys provides typed context keys for the application.
package contextkeys

// ContextKey is a typed key for context values
type ContextKey string

// Context keys used throughout the application
const (
	// RequestID is the key for the request ID in context
	RequestID ContextKey = "request_id"

	// Logger is the key for the logger in context
	Logger ContextKey = "logger"

	// StartTime is the key for the request start time in context
	StartTime ContextKey = "start_time"

	// IdempotencyKey is the key for the idempotency key in context
	IdempotencyKey ContextKey = "idempotency_key"
)

// String returns the string representation of the context key
func (c ContextKey) String() string {
	return string(c)
}
