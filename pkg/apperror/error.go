package apperror

import (
	"errors"
	"fmt"
)

// IError is the interface for application errors
type IError interface {
	error
	Code() Code
	PublicMessage() string
	HTTPStatus() int
	Fields() map[string]interface{}
	Unwrap() error
	WithField(key string, value interface{}) IError
	WithFields(fields map[string]interface{}) IError
}

// Error represents an application error with code, message, and contextual fields
type Error struct {
	code          Code
	cause         error
	publicMessage string
	fields        map[string]interface{}
}

// Compile-time interface check
var _ IError = (*Error)(nil)

// New creates a new application error with the given code and cause
func New(code Code, cause error) *Error {
	return &Error{
		code:          code,
		cause:         cause,
		publicMessage: codeToPublicMessage(code),
		fields:        make(map[string]interface{}),
	}
}

// NewWithMessage creates a new application error with a custom public message
func NewWithMessage(code Code, cause error, publicMessage string) *Error {
	return &Error{
		code:          code,
		cause:         cause,
		publicMessage: publicMessage,
		fields:        make(map[string]interface{}),
	}
}

// Error returns the error message (includes cause for internal logging)
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.code, e.publicMessage, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.publicMessage)
}

// Code returns the error code
func (e *Error) Code() Code {
	return e.code
}

// PublicMessage returns the user-facing error message
func (e *Error) PublicMessage() string {
	return e.publicMessage
}

// HTTPStatus returns the HTTP status code for this error
func (e *Error) HTTPStatus() int {
	return e.code.HTTPStatus()
}

// Fields returns the contextual fields attached to this error
func (e *Error) Fields() map[string]interface{} {
	return e.fields
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.cause
}

// WithField adds a contextual field to the error
func (e *Error) WithField(key string, value interface{}) IError {
	if e.fields == nil {
		e.fields = make(map[string]interface{})
	}
	e.fields[key] = value
	return e
}

// WithFields adds multiple contextual fields to the error
func (e *Error) WithFields(fields map[string]interface{}) IError {
	if e.fields == nil {
		e.fields = make(map[string]interface{})
	}
	for k, v := range fields {
		e.fields[k] = v
	}
	return e
}

// codeToPublicMessage maps error codes to default public messages
func codeToPublicMessage(code Code) string {
	switch code {
	case CodeBadRequest, CodeValidationError:
		return MsgInvalidRequest
	case CodeNotFound:
		return MsgAccountNotFound
	case CodeConflict:
		return MsgDuplicateAccount
	case CodeInsufficientFunds:
		return MsgInsufficientBalance
	case CodeServiceUnavailable:
		return MsgServiceUnavailable
	case CodeInternalError:
		return MsgInternalError
	default:
		return MsgInternalError
	}
}

// Is checks if the error matches the target error
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As attempts to convert the error to the target type
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
