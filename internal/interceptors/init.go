// Package interceptors provides HTTP middleware chain configuration.
package interceptors

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/modules/idempotency"
)

// DefaultRequestTimeout is the default timeout for HTTP requests
var DefaultRequestTimeout = time.Duration(constants.DefaultRequestTimeoutSeconds) * time.Second

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain represents an ordered list of middleware
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{middlewares: middlewares}
}

// Then wraps the final handler with all middleware in the chain
func (c *Chain) Then(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// Append adds middleware to the chain
func (c *Chain) Append(middlewares ...Middleware) *Chain {
	newMiddlewares := make([]Middleware, 0, len(c.middlewares)+len(middlewares))
	newMiddlewares = append(newMiddlewares, c.middlewares...)
	newMiddlewares = append(newMiddlewares, middlewares...)
	return &Chain{middlewares: newMiddlewares}
}

// DefaultMiddleware returns the standard middleware chain for the main API
func DefaultMiddleware() []Middleware {
	return []Middleware{
		middleware.RequestID,       // Generate/propagate request ID
		middleware.RealIP,          // Extract real IP from headers
		RecoveryMiddleware,         // Panic recovery with logging
		RequestIDMiddleware,        // Add request ID to response header
		MetricsMiddleware,          // Record Prometheus metrics
		RequestLoggerMiddleware,    // Log requests with timing
	}
}

// DefaultMiddlewareWithTimeout returns the standard middleware chain with timeout
func DefaultMiddlewareWithTimeout(timeout time.Duration) []Middleware {
	return []Middleware{
		middleware.RequestID,
		middleware.RealIP,
		RecoveryMiddleware,
		RequestIDMiddleware,
		TimeoutMiddleware(timeout),
		MetricsMiddleware,
		RequestLoggerMiddleware,
	}
}

// ApplyMiddleware applies a list of middleware to a chi router
func ApplyMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	chain := NewChain(middlewares...)
	return chain.Then(handler)
}

// GetChiMiddleware returns middleware compatible with chi.Router.Use()
func GetChiMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		RecoveryMiddleware,
		RequestIDMiddleware,
		MetricsMiddleware,
		RequestLoggerMiddleware,
	}
}

// GetChiMiddlewareWithIdempotency returns middleware with idempotency support.
// The idempotency middleware is placed after logging but before the handler
// so that idempotent requests are properly logged.
func GetChiMiddlewareWithIdempotency(idempotencyRepo idempotency.IRepository) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		RecoveryMiddleware,
		RequestIDMiddleware,
		MetricsMiddleware,
		RequestLoggerMiddleware,
		IdempotencyMiddleware(idempotencyRepo),
	}
}
