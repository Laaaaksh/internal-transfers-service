// Package interceptors provides HTTP middleware chain configuration.
package interceptors

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/internal-transfers-service/internal/config"
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
// This version does not include idempotency support.
func GetChiMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		RecoveryMiddleware,
		RequestIDMiddleware,
		SecurityHeadersMiddleware,
		MaxBytesMiddleware,
		ContentTypeValidationMiddleware,
		TimeoutMiddleware(DefaultRequestTimeout),
		MetricsMiddleware,
		RequestLoggerMiddleware,
	}
}

// GetChiMiddlewareWithIdempotency returns the production-ready middleware chain.
// Middleware order is carefully designed for security and proper request handling:
// 1. RequestID - Generate/propagate request ID first for tracing
// 2. RealIP - Extract real client IP from proxy headers
// 3. RecoveryMiddleware - Panic recovery (must be early to catch all panics)
// 4. RequestIDMiddleware - Add request ID to response headers
// 5. SecurityHeadersMiddleware - Add security headers early
// 6. RateLimitMiddleware - Rate limiting early (before heavy processing)
// 7. MaxBytesMiddleware - Limit body size before parsing (DoS protection)
// 8. ContentTypeValidationMiddleware - Validate content-type before parsing
// 9. TimeoutMiddleware - Request timeout protection
// 10. MetricsMiddleware - Record metrics (after timeout to measure actual time)
// 11. RequestLoggerMiddleware - Log requests (captures response details)
// 12. IdempotencyMiddleware - Handle idempotent requests last
func GetChiMiddlewareWithIdempotency(idempotencyRepo idempotency.IRepository) []func(http.Handler) http.Handler {
	return GetChiMiddlewareWithConfig(idempotencyRepo, config.RateLimitConfig{})
}

// GetChiMiddlewareWithConfig returns the production-ready middleware chain with full config.
// Use this when rate limiting configuration is available.
func GetChiMiddlewareWithConfig(idempotencyRepo idempotency.IRepository, rateLimitCfg config.RateLimitConfig) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		RecoveryMiddleware,
		RequestIDMiddleware,
		SecurityHeadersMiddleware,
		RateLimitMiddleware(rateLimitCfg),
		MaxBytesMiddleware,
		ContentTypeValidationMiddleware,
		TimeoutMiddleware(DefaultRequestTimeout),
		MetricsMiddleware,
		RequestLoggerMiddleware,
		IdempotencyMiddleware(idempotencyRepo),
	}
}
