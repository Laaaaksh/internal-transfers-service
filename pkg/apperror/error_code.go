package apperror

import "net/http"

// Code represents an error code type
type Code string

// Error codes used throughout the application
const (
	CodeBadRequest         Code = "BAD_REQUEST"
	CodeNotFound           Code = "NOT_FOUND"
	CodeConflict           Code = "CONFLICT"
	CodeInsufficientFunds  Code = "INSUFFICIENT_FUNDS"
	CodeInternalError      Code = "INTERNAL_ERROR"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeValidationError    Code = "VALIDATION_ERROR"
	CodeDuplicateRequest   Code = "DUPLICATE_REQUEST"
)

// HTTPStatus returns the HTTP status code for an error code
func (c Code) HTTPStatus() int {
	switch c {
	case CodeBadRequest, CodeValidationError:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeDuplicateRequest:
		return http.StatusConflict
	case CodeInsufficientFunds:
		return http.StatusUnprocessableEntity
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case CodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// String returns the string representation of the error code
func (c Code) String() string {
	return string(c)
}
