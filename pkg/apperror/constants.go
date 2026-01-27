// Package apperror provides custom error handling with error codes and contextual information.
package apperror

// Error message constants - all error messages must be defined here
const (
	// InternalErrorDetails is the key for internal error details in error fields
	InternalErrorDetails = "internal_error_details"

	// PublicErrorDetails is the key for public-facing error details
	PublicErrorDetails = "public_error_details"

	// ErrorField keys
	FieldAccountID      = "account_id"
	FieldSourceAccount  = "source_account_id"
	FieldDestAccount    = "destination_account_id"
	FieldAmount         = "amount"
	FieldIdempotencyKey = "idempotency_key"
	FieldRequestID      = "request_id"
)

// Public error messages - user-facing messages
const (
	MsgInternalError       = "An internal error occurred. Please try again later."
	MsgInvalidRequest      = "The request is invalid."
	MsgAccountNotFound     = "The specified account was not found."
	MsgSourceNotFound      = "Source account not found."
	MsgDestNotFound        = "Destination account not found."
	MsgInsufficientBalance = "Insufficient balance for this transaction."
	MsgDuplicateAccount    = "An account with this ID already exists."
	MsgInvalidAmount       = "The amount must be a positive number."
	MsgSameAccountTransfer = "Source and destination accounts must be different."
	MsgServiceUnavailable  = "Service is temporarily unavailable."

	// More descriptive validation messages
	MsgInvalidAccountID      = "Account ID must be a positive integer."
	MsgInvalidInitialBalance = "Initial balance must be a valid non-negative decimal number."
	MsgInvalidDecimalFormat  = "The provided value is not a valid decimal number."
	MsgNegativeBalance       = "Balance cannot be negative."
	MsgInvalidJSONBody       = "Invalid JSON in request body."
	MsgTooManyDecimalPlaces  = "Value exceeds maximum precision of 8 decimal places."
)

// Additional field keys
const (
	FieldDecimalPlaces = "decimal_places"
	FieldMaxAllowed    = "max_allowed"
)
