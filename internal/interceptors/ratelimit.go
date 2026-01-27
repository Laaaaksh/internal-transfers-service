// Package interceptors provides HTTP middleware for the application.
package interceptors

import (
	"net/http"

	"github.com/internal-transfers-service/internal/config"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware creates a rate limiting middleware using token bucket algorithm.
// It allows a burst of requests up to burstSize, while enforcing an average rate.
// If rate limiting is disabled in config, it returns a pass-through middleware.
func RateLimitMiddleware(cfg config.RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return createPassThroughMiddleware()
	}

	limiter := createLimiter(cfg)
	return createRateLimitHandler(limiter)
}

// createPassThroughMiddleware returns a middleware that does nothing
func createPassThroughMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}

// createLimiter creates a new rate limiter from config
func createLimiter(cfg config.RateLimitConfig) *rate.Limiter {
	return rate.NewLimiter(rate.Limit(cfg.RequestsPerSec), cfg.BurstSize)
}

// createRateLimitHandler creates the actual rate limiting handler
func createRateLimitHandler(limiter *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				handleRateLimitExceeded(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// handleRateLimitExceeded writes the 429 Too Many Requests response
func handleRateLimitExceeded(w http.ResponseWriter, r *http.Request) {
	logRateLimitExceeded(r)
	writeRateLimitResponse(w)
}

// logRateLimitExceeded logs the rate limit exceeded event
func logRateLimitExceeded(r *http.Request) {
	logger.Ctx(r.Context()).Warnw(constants.LogMsgRateLimitExceeded,
		constants.LogKeyMethod, r.Method,
		constants.LogKeyPath, r.URL.Path,
	)
}

// writeRateLimitResponse writes the 429 response with appropriate headers
func writeRateLimitResponse(w http.ResponseWriter) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.Header().Set(constants.HeaderRetryAfter, constants.DefaultRetryAfterSeconds)
	w.WriteHeader(constants.HTTPStatusTooManyRequests)
	_, _ = w.Write([]byte(buildRateLimitErrorResponse()))
}

// buildRateLimitErrorResponse builds the JSON error response for rate limit exceeded
func buildRateLimitErrorResponse() string {
	return constants.JSONErrorPrefix + constants.ErrMsgRateLimitExceeded +
		constants.JSONCodePrefix + constants.ErrCodeRateLimitExceeded +
		constants.JSONSuffix
}
