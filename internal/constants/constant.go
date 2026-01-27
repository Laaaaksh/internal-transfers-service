// Package constants provides application-wide constants.
package constants

// Application constants
const (
	// ServiceName is the name of this service
	ServiceName = "internal-transfers-service"

	// API versioning
	APIVersion       = "v1"
	APIVersionPrefix = "/v1"

	// HTTP headers
	HeaderRequestID      = "X-Request-ID"
	HeaderIdempotencyKey = "X-Idempotency-Key"
	HeaderContentType    = "Content-Type"

	// Content types
	ContentTypeJSON = "application/json"

	// Health check statuses
	StatusServing    = "SERVING"
	StatusNotServing = "NOT_SERVING"

	// Decimal precision for money
	DecimalPrecision = 8
)

// Log format types
const (
	LogFormatJSON    = "json"
	LogFormatConsole = "console"
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Log encoder configuration keys
const (
	LogEncoderTimeKey    = "timestamp"
	LogEncoderMessageKey = "message"
	LogEncoderLevelKey   = "level"
	LogEncoderCallerKey  = "caller"
)

// Log output paths
const (
	LogOutputStdout = "stdout"
	LogOutputStderr = "stderr"
)

// HTTP status codes
const (
	HTTPStatusOK                  = 200
	HTTPStatusCreated             = 201
	HTTPStatusBadRequest          = 400
	HTTPStatusNotFound            = 404
	HTTPStatusConflict            = 409
	HTTPStatusUnprocessableEntity = 422
	HTTPStatusInternalServerError = 500
	HTTPStatusServiceUnavailable  = 503
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

// Log message constants - static log messages
const (
	LogMsgStartingService          = "Starting service"
	LogMsgFailedToInitDB           = "Failed to initialize database"
	LogMsgMainServerStarting       = "Main HTTP server starting"
	LogMsgOpsServerStarting        = "Ops HTTP server starting"
	LogMsgMainServerFailed         = "Main server failed"
	LogMsgOpsServerFailed          = "Ops server failed"
	LogMsgShutdownSignalReceived   = "Shutdown signal received, initiating graceful shutdown"
	LogMsgWaitingForShutdownDelay  = "Waiting for shutdown delay"
	LogMsgMainServerShutdownErr    = "Main server shutdown error"
	LogMsgMainServerShutdownDone   = "Main server shutdown complete"
	LogMsgOpsServerShutdownErr     = "Ops server shutdown error"
	LogMsgOpsServerShutdownDone    = "Ops server shutdown complete"
	LogMsgGracefulShutdownComplete = "Graceful shutdown complete"
	LogMsgShutdownTimeoutExceeded  = "Shutdown timeout exceeded, forcing exit"
	LogMsgServiceStopped           = "Service stopped"
	LogMsgHTTPRequestCompleted     = "HTTP request completed"
	LogMsgAccountCreatedViaHTTP    = "Account created via HTTP"
	LogMsgTransactionCreatedHTTP   = "Transaction created via HTTP"
	LogMsgFailedToEncodeResponse   = "Failed to encode response"
	LogMsgReadinessCheckFailed     = "Readiness check failed - database ping failed"
	LogMsgServiceMarkedUnhealthy   = "Service marked as unhealthy"

	// Transaction core log messages
	LogMsgFailedToBeginTx         = "Failed to begin transaction"
	LogMsgInsufficientBalance     = "Insufficient balance for transfer"
	LogMsgFailedToUpdateSourceBal = "Failed to update source account balance"
	LogMsgFailedToUpdateDestBal   = "Failed to update destination account balance"
	LogMsgFailedToCreateTxRecord  = "Failed to create transaction record"
	LogMsgFailedToCommitTx        = "Failed to commit transaction"
	LogMsgTransferCompleted       = "Transfer completed successfully"

	// Account core log messages
	LogMsgFailedToCheckAcctExist = "Failed to check account existence"
	LogMsgFailedToCreateAccount  = "Failed to create account"
	LogMsgAccountCreated         = "Account created successfully"
	LogMsgFailedToGetAccount     = "Failed to get account"
	LogMsgFailedToGetForUpdate   = "Failed to get account for update"
	LogMsgFailedToUpdateBalance  = "Failed to update account balance"

	// Validation debug log messages
	LogMsgInvalidAccountIDCreate  = "Invalid account ID in create request"
	LogMsgInvalidAccountIDGet     = "Invalid account ID in get request"
	LogMsgInvalidDecimalFormat    = "Invalid decimal format for initial balance"
	LogMsgNegativeBalanceProvided = "Negative initial balance provided"
	LogMsgAccountNotFoundDebug    = "Account not found"

	// HTTP handler log messages
	LogMsgRequestFailed   = "Request failed"
	LogMsgPanicRecovered  = "Panic recovered"
	LogMsgRequestReceived = "Request received"
)

// Additional log field keys for interceptors
const (
	LogKeyStack = "stack"
)

// Interceptor constants
const (
	// StackSizeBytes is the size of the stack trace buffer for panic recovery (4 KB)
	StackSizeBytes = 4 << 10

	// DefaultRequestTimeoutSeconds is the default timeout for HTTP requests in seconds
	DefaultRequestTimeoutSeconds = 30
)

// CORS headers
const (
	HeaderAccessControlAllowOrigin  = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders = "Access-Control-Allow-Headers"

	CORSAllowOriginAll     = "*"
	CORSAllowMethodsAll    = "GET, POST, PUT, DELETE, OPTIONS"
	CORSAllowHeadersCommon = "Content-Type, Authorization, X-Idempotency-Key, X-Request-ID"
)

// Error response messages for interceptors
const (
	ErrMsgInternalServerError   = "Internal server error"
	ErrMsgRequestTimeout        = "Request timeout"
	ErrMsgIdempotencyKeyTooLong = "Idempotency key too long"
	ErrCodeInternalError        = "INTERNAL_ERROR"
	ErrCodeTimeout              = "TIMEOUT"
	ErrCodeInvalidIdempotency   = "INVALID_IDEMPOTENCY_KEY"
)

// Path normalization placeholders
const (
	PathPlaceholderID   = ":id"
	PathPlaceholderUUID = ":uuid"
)

// JSON response building constants
const (
	JSONErrorPrefix     = `{"error":"`
	JSONCodePrefix      = `","code":"`
	JSONRequestIDPrefix = `","request_id":"`
	JSONSuffix          = `"}`
)

// Log field key constants
const (
	LogFieldName           = "name"
	LogFieldEnv            = "env"
	LogFieldPort           = "port"
	LogFieldOpsPort        = "ops_port"
	LogFieldAddr           = "addr"
	LogFieldDelaySeconds   = "delay_seconds"
	LogFieldBytesWritten   = "bytes_written"
	LogFieldTransactionID  = "transaction_id"
	LogFieldCurrentBalance = "current_balance"
	LogFieldRequestedAmt   = "requested_amount"
	LogFieldNewBalance     = "new_balance"
	LogFieldInitialBalance = "initial_balance"
	LogFieldHost           = "host"
	LogFieldDatabase       = "database"
	LogFieldMaxConnections = "max_connections"
	LogFieldSourceAccount  = "source_account"
	LogFieldDestAccount    = "destination_account"
)

// Database log messages
const (
	LogMsgDBPoolInitialized = "Database connection pool initialized"
	LogMsgDBPoolClosed      = "Database connection pool closed"
)

// Database error format strings
const (
	ErrFmtFailedToParseConnString = "failed to parse connection string: %w"
	ErrFmtFailedToCreateConnPool  = "failed to create connection pool: %w"
	ErrFmtFailedToPingDB          = "failed to ping database: %w"
	ErrFmtDBConnectionFailedRetry = "database connection failed after %d attempts: %w"
)

// Database retry constants
const (
	// Default retry values
	DefaultDBRetryMaxRetries     = 5
	DefaultDBRetryInitialBackoff = "1s"
	DefaultDBRetryMaxBackoff     = "30s"
)

// Database retry log messages
const (
	LogMsgDBConnectionAttempt = "Attempting database connection"
	LogMsgDBConnectionRetry   = "Database connection failed, retrying"
	LogMsgDBConnectionSuccess = "Database connection established"
	LogMsgDBConnectionFailed  = "Database connection failed after max retries"
)

// Database retry log field keys
const (
	LogFieldAttempt     = "attempt"
	LogFieldMaxRetries  = "max_retries"
	LogFieldBackoff     = "backoff"
	LogFieldNextBackoff = "next_backoff"
)

// Transaction repository log messages
const (
	LogMsgFailedToCreateTx   = "Failed to create transaction"
	LogMsgTransactionCreated = "Transaction created"
)

// Health module route paths
const (
	RouteHealthLive  = "/health/live"
	RouteHealthReady = "/health/ready"
	RouteMetrics     = "/metrics"
)

// Server name constants
const (
	ServerNameMain = "main"
	ServerNameOps  = "ops"
)

// Environment variable keys
const (
	EnvKeyAppEnv  = "APP_ENV"
	EnvDefaultDev = "dev"
)

// Security middleware constants
const (
	// MaxRequestBodySize is the maximum allowed request body size (1 MB)
	MaxRequestBodySize = 1 << 20

	// Security headers
	HeaderXContentTypeOptions = "X-Content-Type-Options"
	HeaderXFrameOptions       = "X-Frame-Options"
	HeaderCacheControl        = "Cache-Control"

	// Security header values
	ValueNoSniff = "nosniff"
	ValueDeny    = "DENY"
	ValueNoStore = "no-store"

	// Content-Type validation
	ContentTypeJSONPrefix = "application/json"
)

// HTTP status code for unsupported media type
const (
	HTTPStatusUnsupportedMediaType = 415
)

// Error messages for security middlewares
const (
	ErrMsgUnsupportedMediaType = "Content-Type must be application/json"
	ErrMsgRequestBodyTooLarge  = "Request body too large"
	ErrCodeUnsupportedMedia    = "UNSUPPORTED_MEDIA_TYPE"
	ErrCodeBodyTooLarge        = "REQUEST_BODY_TOO_LARGE"
)

// Log messages for security middlewares
const (
	LogMsgInvalidContentType  = "Invalid Content-Type header"
	LogMsgRequestBodyTooLarge = "Request body exceeds size limit"
)

// Decimal validation constants
const (
	// MaxDecimalPlaces is the maximum number of decimal places allowed (matches DB DECIMAL(19,8))
	MaxDecimalPlaces = 8
)

// Log messages for decimal validation
const (
	LogMsgTooManyDecimalPlaces = "Too many decimal places in value"
)

// Rate limiting constants
const (
	// Default rate limit values
	DefaultRateLimitRequestsPerSec = 100.0
	DefaultRateLimitBurstSize      = 200

	// HTTP status code for rate limit exceeded
	HTTPStatusTooManyRequests = 429

	// Rate limit headers
	HeaderRetryAfter = "Retry-After"

	// Default retry after value in seconds
	DefaultRetryAfterSeconds = "1"
)

// Rate limit error messages
const (
	ErrMsgRateLimitExceeded  = "Rate limit exceeded"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)

// Rate limit log messages
const (
	LogMsgRateLimitExceeded = "Rate limit exceeded for request"
)

// Tracing constants
const (
	// Default tracing configuration values
	DefaultTracingEndpoint           = "localhost:4317"
	DefaultTracingSampleRate         = 1.0
	DefaultTracingBatchTimeout       = "5s"
	DefaultTracingBatchTimeoutSecond = 5

	// Tracing span name format
	TracingSpanNameSeparator = " "

	// Tracing attribute keys
	TracingAttrHTTPMethod     = "http.method"
	TracingAttrHTTPRoute      = "http.route"
	TracingAttrHTTPStatusCode = "http.status_code"
	TracingAttrHTTPRequestID  = "http.request_id"
	TracingAttrHTTPUserAgent  = "http.user_agent"
	TracingAttrHTTPClientIP   = "http.client_ip"
	TracingAttrDBSystem       = "db.system"
	TracingAttrDBStatement    = "db.statement"
	TracingAttrDBOperation    = "db.operation"

	// Tracing span names
	TracingSpanHTTPRequest = "http.request"
	TracingSpanDBQuery     = "db.query"
)

// Tracing log messages
const (
	LogMsgTracerInitialized    = "OpenTelemetry tracer initialized"
	LogMsgTracerInitFailed     = "Failed to initialize tracer"
	LogMsgTracerShutdown       = "Tracer shutdown complete"
	LogMsgTracerShutdownFailed = "Failed to shutdown tracer"
	LogMsgTracerDisabled       = "Tracing is disabled"
)

// Tracing log field keys
const (
	LogFieldTraceID    = "trace_id"
	LogFieldSpanID     = "span_id"
	LogFieldEndpoint   = "endpoint"
	LogFieldSampleRate = "sample_rate"
)
