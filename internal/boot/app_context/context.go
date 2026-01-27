// Package app_context provides context helpers for the application.
package app_context

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/internal-transfers-service/internal/constants"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
)

// GetRequestID extracts the request ID from context.
// First tries chi middleware's request ID, then falls back to our custom key.
func GetRequestID(ctx context.Context) string {
	// Try chi middleware's request ID first
	if reqID := middleware.GetReqID(ctx); reqID != "" {
		return reqID
	}

	// Fallback to our custom key
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}

	return ""
}

// SetRequestID adds a request ID to the context.
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestIDFromRequest extracts or generates request ID from HTTP request.
func GetRequestIDFromRequest(r *http.Request) string {
	// Check header first
	if reqID := r.Header.Get(constants.HeaderRequestID); reqID != "" {
		return reqID
	}

	// Fall back to chi middleware's request ID
	return middleware.GetReqID(r.Context())
}

// WithRequestID adds request ID to context from HTTP request.
func WithRequestID(ctx context.Context, r *http.Request) context.Context {
	reqID := GetRequestIDFromRequest(r)
	if reqID != "" {
		return SetRequestID(ctx, reqID)
	}
	return ctx
}

// AddRequestIDToResponse adds the request ID to the response header.
func AddRequestIDToResponse(ctx context.Context, w http.ResponseWriter) {
	reqID := GetRequestID(ctx)
	if reqID != "" {
		w.Header().Set(constants.HeaderRequestID, reqID)
	}
}
