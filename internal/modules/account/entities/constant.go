// Package entities provides request/response types and constants for the account module.
package entities

// Error messages for the account module
const (
	ErrMsgAccountNotFound    = "account not found"
	ErrMsgAccountExists      = "account already exists"
	ErrMsgInvalidAccountID   = "invalid account ID"
	ErrMsgInvalidBalance     = "initial balance must be non-negative"
	ErrMsgInvalidDecimal     = "invalid decimal format for balance"
)
