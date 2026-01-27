// Package apperror provides custom error handling with error codes and contextual information.
package apperror

// ErrorResponse is the standardized API error response format.
// This struct is used across all modules for consistent error responses.
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Code      string                 `json:"code,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}
