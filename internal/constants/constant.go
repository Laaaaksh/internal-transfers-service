// Package constants provides application-wide constants.
package constants

// Application constants
const (
	// ServiceName is the name of this service
	ServiceName = "internal-transfers-service"

	// API versioning
	APIVersion = "v1"

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
)
