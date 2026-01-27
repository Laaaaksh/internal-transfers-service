package interceptors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/internal-transfers-service/internal/config"
	"github.com/internal-transfers-service/internal/interceptors"
	"github.com/stretchr/testify/suite"
)

// RateLimitTestSuite tests the rate limiting middleware
type RateLimitTestSuite struct {
	suite.Suite
}

func TestRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RateLimitTestSuite))
}

// TestRateLimitMiddlewareAllowsRequestWithinLimit verifies requests within limit pass through
func (s *RateLimitTestSuite) TestRateLimitMiddlewareAllowsRequestWithinLimit() {
	cfg := config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 10.0,
		BurstSize:      10,
	}

	handler := interceptors.RateLimitMiddleware(cfg)(createOKHandler())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("OK", rec.Body.String())
}

// TestRateLimitMiddlewareReturns429WhenLimitExceeded verifies 429 is returned when rate exceeded
func (s *RateLimitTestSuite) TestRateLimitMiddlewareReturns429WhenLimitExceeded() {
	cfg := config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 1.0,
		BurstSize:      1,
	}

	handler := interceptors.RateLimitMiddleware(cfg)(createOKHandler())

	// First request should pass
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	s.Equal(http.StatusOK, rec1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	s.Equal(http.StatusTooManyRequests, rec2.Code)
}

// TestRateLimitMiddlewareReturnsCorrectErrorResponse verifies error response format
func (s *RateLimitTestSuite) TestRateLimitMiddlewareReturnsCorrectErrorResponse() {
	cfg := config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 1.0,
		BurstSize:      1,
	}

	handler := interceptors.RateLimitMiddleware(cfg)(createOKHandler())

	// Exhaust the limit
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	// Get the rate limited response
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	s.Equal(http.StatusTooManyRequests, rec2.Code)
	s.Contains(rec2.Body.String(), "Rate limit exceeded")
	s.Contains(rec2.Body.String(), "RATE_LIMIT_EXCEEDED")
}

// TestRateLimitMiddlewareSetsRetryAfterHeader verifies Retry-After header is set
func (s *RateLimitTestSuite) TestRateLimitMiddlewareSetsRetryAfterHeader() {
	cfg := config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 1.0,
		BurstSize:      1,
	}

	handler := interceptors.RateLimitMiddleware(cfg)(createOKHandler())

	// Exhaust the limit
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	// Get the rate limited response
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	s.Equal(http.StatusTooManyRequests, rec2.Code)
	s.Equal("1", rec2.Header().Get("Retry-After"))
}

// TestRateLimitMiddlewareSetsContentTypeHeader verifies Content-Type is set in error response
func (s *RateLimitTestSuite) TestRateLimitMiddlewareSetsContentTypeHeader() {
	cfg := config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 1.0,
		BurstSize:      1,
	}

	handler := interceptors.RateLimitMiddleware(cfg)(createOKHandler())

	// Exhaust the limit
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	// Get the rate limited response
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	s.Equal(http.StatusTooManyRequests, rec2.Code)
	s.Equal("application/json", rec2.Header().Get("Content-Type"))
}

// TestRateLimitMiddlewareDisabledPassesAllRequests verifies disabled rate limiter passes all
func (s *RateLimitTestSuite) TestRateLimitMiddlewareDisabledPassesAllRequests() {
	cfg := config.RateLimitConfig{
		Enabled:        false,
		RequestsPerSec: 1.0,
		BurstSize:      1,
	}

	handler := interceptors.RateLimitMiddleware(cfg)(createOKHandler())

	// All requests should pass when disabled
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code)
	}
}

// TestRateLimitMiddlewareAllowsBurstUpToLimit verifies burst requests up to limit pass
func (s *RateLimitTestSuite) TestRateLimitMiddlewareAllowsBurstUpToLimit() {
	cfg := config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 1.0,
		BurstSize:      5,
	}

	handler := interceptors.RateLimitMiddleware(cfg)(createOKHandler())

	// First 5 requests should pass (burst size)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		s.Equal(http.StatusOK, rec.Code, "Request %d should pass", i+1)
	}

	// 6th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	s.Equal(http.StatusTooManyRequests, rec.Code, "Request 6 should be rate limited")
}

// TestRateLimitMiddlewareWithZeroBurstSizeRejectsAll verifies zero burst rejects immediately
func (s *RateLimitTestSuite) TestRateLimitMiddlewareWithZeroBurstSizeRejectsAll() {
	cfg := config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 1.0,
		BurstSize:      0,
	}

	handler := interceptors.RateLimitMiddleware(cfg)(createOKHandler())

	// With burst size 0, all requests should be rejected
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusTooManyRequests, rec.Code)
}

// TestRateLimitMiddlewareWorksWithPOSTMethod verifies rate limiting works with POST
func (s *RateLimitTestSuite) TestRateLimitMiddlewareWorksWithPOSTMethod() {
	cfg := config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 1.0,
		BurstSize:      1,
	}

	handler := interceptors.RateLimitMiddleware(cfg)(createOKHandler())

	// First POST request should pass
	req1 := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	s.Equal(http.StatusOK, rec1.Code)

	// Second POST request should be rate limited
	req2 := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	s.Equal(http.StatusTooManyRequests, rec2.Code)
}

// TestRateLimitMiddlewarePreservesHandlerBehavior verifies handler response is preserved
func (s *RateLimitTestSuite) TestRateLimitMiddlewarePreservesHandlerBehavior() {
	cfg := config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 100.0,
		BurstSize:      100,
	}

	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123}`))
	})

	handler := interceptors.RateLimitMiddleware(cfg)(customHandler)

	req := httptest.NewRequest(http.MethodPost, "/accounts", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)
	s.Equal("test-value", rec.Header().Get("X-Custom-Header"))
	s.Equal(`{"id": 123}`, rec.Body.String())
}

// createOKHandler creates a simple handler that returns 200 OK
func createOKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
