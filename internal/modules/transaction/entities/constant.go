// Package entities provides request/response types and constants for the transaction module.
package entities

// Error messages for the transaction module
const (
	ErrMsgInsufficientBalance  = "insufficient balance for this transaction"
	ErrMsgSameAccountTransfer  = "source and destination accounts must be different"
	ErrMsgInvalidAmount        = "amount must be a positive number"
	ErrMsgInvalidDecimalAmt    = "invalid decimal format for amount"
	ErrMsgSourceNotFound       = "source account not found"
	ErrMsgDestNotFound         = "destination account not found"
	ErrMsgTooManyDecimalPlaces = "amount exceeds maximum precision"
)

// Route path constants for the transaction module
const (
	RouteTransactions = "/transactions"
)
