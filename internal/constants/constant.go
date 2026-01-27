// Package constants provides application-wide constants.
package constants

// Application constants
const (
	// ServiceName is the name of this service
	ServiceName = "internal-transfers-service"

	// API versioning
	APIVersion = "v1"

	// HTTP headers
	HeaderRequestID     = "X-Request-ID"
	HeaderIdempotencyKey = "X-Idempotency-Key"
	HeaderContentType   = "Content-Type"

	// Content types
	ContentTypeJSON = "application/json"

	// Health check statuses
	StatusServing    = "SERVING"
	StatusNotServing = "NOT_SERVING"

	// Decimal precision for money
	DecimalPrecision = 8
)

// Database table names
const (
	TableAccounts        = "accounts"
	TableTransactions    = "transactions"
	TableIdempotencyKeys = "idempotency_keys"
)

// Logging keys - used for structured logging
const (
	LogKeyRequestID     = "request_id"
	LogKeyAccountID     = "account_id"
	LogKeySourceAccount = "source_account_id"
	LogKeyDestAccount   = "destination_account_id"
	LogKeyAmount        = "amount"
	LogKeyDuration      = "duration_ms"
	LogKeyStatusCode    = "status_code"
	LogKeyMethod        = "method"
	LogKeyPath          = "path"
	LogKeyError         = "error"
)

// Metrics names
const (
	MetricRequestDuration   = "http_request_duration_seconds"
	MetricRequestTotal      = "http_requests_total"
	MetricTransferTotal     = "transfers_total"
	MetricTransferSuccess   = "transfers_success_total"
	MetricTransferFailed    = "transfers_failed_total"
	MetricDBConnectionsOpen = "db_connections_open"
	MetricDBConnectionsIdle = "db_connections_idle"
)

// Metrics labels
const (
	LabelMethod     = "method"
	LabelPath       = "path"
	LabelStatusCode = "status_code"
	LabelReason     = "reason"
)
