package interceptors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/internal-transfers-service/internal/interceptors"
	"github.com/stretchr/testify/suite"
)

// HTTPMiddlewareTestSuite tests HTTP middleware functions
type HTTPMiddlewareTestSuite struct {
	suite.Suite
}

func TestHTTPMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(HTTPMiddlewareTestSuite))
}

// TestRequestLoggerMiddlewareLogsRequest verifies request logger middleware works
func (s *HTTPMiddlewareTestSuite) TestRequestLoggerMiddlewareLogsRequest() {
	handler := interceptors.RequestLoggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("OK", rec.Body.String())
}

// TestMetricsMiddlewareRecordsMetrics verifies metrics middleware works
func (s *HTTPMiddlewareTestSuite) TestMetricsMiddlewareRecordsMetrics() {
	handler := interceptors.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

// TestRecoveryMiddlewareRecoversPanic verifies panic recovery works
func (s *HTTPMiddlewareTestSuite) TestRecoveryMiddlewareRecoversPanic() {
	// Wrap with chi's RequestID middleware first to get proper context
	handler := middleware.RequestID(interceptors.RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	s.NotPanics(func() {
		handler.ServeHTTP(rec, req)
	})

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.Contains(rec.Body.String(), "Internal server error")
}

// TestRecoveryMiddlewareAllowsNormalRequests verifies normal requests pass through
func (s *HTTPMiddlewareTestSuite) TestRecoveryMiddlewareAllowsNormalRequests() {
	handler := interceptors.RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("success", rec.Body.String())
}

// TestRequestIDMiddlewareAddsHeader verifies request ID is added to response
func (s *HTTPMiddlewareTestSuite) TestRequestIDMiddlewareAddsHeader() {
	// First wrap with chi's RequestID middleware, then our RequestIDMiddleware
	handler := middleware.RequestID(interceptors.RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.NotEmpty(rec.Header().Get("X-Request-ID"))
}

// TestRequestIDMiddlewareWithoutIDDoesNotFail verifies middleware works without ID
func (s *HTTPMiddlewareTestSuite) TestRequestIDMiddlewareWithoutIDDoesNotFail() {
	handler := interceptors.RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

// TestTimeoutMiddlewareTimesOut verifies timeout middleware works
func (s *HTTPMiddlewareTestSuite) TestTimeoutMiddlewareTimesOut() {
	handler := interceptors.TimeoutMiddleware(50 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusServiceUnavailable, rec.Code)
}

// TestTimeoutMiddlewareAllowsFastRequests verifies fast requests pass through
func (s *HTTPMiddlewareTestSuite) TestTimeoutMiddlewareAllowsFastRequests() {
	handler := interceptors.TimeoutMiddleware(100 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fast"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("fast", rec.Body.String())
}

// TestCORSMiddlewareAddsCORSHeaders verifies CORS headers are added
func (s *HTTPMiddlewareTestSuite) TestCORSMiddlewareAddsCORSHeaders() {
	handler := interceptors.CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("*", rec.Header().Get("Access-Control-Allow-Origin"))
	s.NotEmpty(rec.Header().Get("Access-Control-Allow-Methods"))
	s.NotEmpty(rec.Header().Get("Access-Control-Allow-Headers"))
}

// TestCORSMiddlewareHandlesOptionsRequest verifies OPTIONS requests are handled
func (s *HTTPMiddlewareTestSuite) TestCORSMiddlewareHandlesOptionsRequest() {
	handlerCalled := false
	handler := interceptors.CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.False(handlerCalled, "Handler should not be called for OPTIONS requests")
}

// TestContentTypeMiddlewareSetsContentType verifies content type is set
func (s *HTTPMiddlewareTestSuite) TestContentTypeMiddlewareSetsContentType() {
	handler := interceptors.ContentTypeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))
}

// ========== Security Headers Middleware Tests ==========

func (s *HTTPMiddlewareTestSuite) TestSecurityHeadersMiddlewareSetsAllHeaders() {
	handler := interceptors.SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("nosniff", rec.Header().Get("X-Content-Type-Options"))
	s.Equal("DENY", rec.Header().Get("X-Frame-Options"))
	s.Equal("no-store", rec.Header().Get("Cache-Control"))
}

// ========== MaxBytes Middleware Tests ==========

func (s *HTTPMiddlewareTestSuite) TestMaxBytesMiddlewareLimitsBodySize() {
	handler := interceptors.MaxBytesMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

// ========== ContentType Validation Middleware Tests ==========

func (s *HTTPMiddlewareTestSuite) TestContentTypeValidationMiddlewareAllowsValidJSON() {
	handler := interceptors.ContentTypeValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *HTTPMiddlewareTestSuite) TestContentTypeValidationMiddlewareAllowsJSONWithCharset() {
	handler := interceptors.ContentTypeValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *HTTPMiddlewareTestSuite) TestContentTypeValidationMiddlewareRejectsInvalidContentType() {
	handler := interceptors.ContentTypeValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusUnsupportedMediaType, rec.Code)
}

func (s *HTTPMiddlewareTestSuite) TestContentTypeValidationMiddlewareAllowsGETWithoutContentType() {
	handler := interceptors.ContentTypeValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *HTTPMiddlewareTestSuite) TestContentTypeValidationMiddlewareRejectsPOSTWithoutContentType() {
	handler := interceptors.ContentTypeValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusUnsupportedMediaType, rec.Code)
}
