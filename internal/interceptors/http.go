// Package interceptors provides HTTP middleware for the application.
package interceptors

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/metrics"
)

const (
	// StackSize is the size of the stack trace buffer for panic recovery
	StackSize = 4 << 10 // 4 KB
)

// RecoveryMiddleware recovers from panics and logs the stack trace.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(r, w)
		next.ServeHTTP(w, r)
	})
}

// recoverFromPanic handles panic recovery with stack trace logging
func recoverFromPanic(r *http.Request, w http.ResponseWriter) {
	if err := recover(); err != nil {
		stack := captureStackTrace()
		logPanicRecovery(r, err, stack)
		sendInternalServerError(w, r)
	}
}

// captureStackTrace captures the current stack trace
func captureStackTrace() string {
	stack := make([]byte, StackSize)
	length := runtime.Stack(stack, false)
	return string(stack[:length])
}

// logPanicRecovery logs the panic with stack trace
func logPanicRecovery(r *http.Request, err interface{}, stack string) {
	logger.Ctx(r.Context()).Errorw(constants.LogMsgPanicRecovered,
		constants.LogKeyError, err,
		constants.LogKeyStack, stack,
		constants.LogKeyPath, r.URL.Path,
		constants.LogKeyMethod, r.Method,
	)
}

// sendInternalServerError sends a 500 response
func sendInternalServerError(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(http.StatusInternalServerError)

	response := `{"error":"Internal server error","code":"INTERNAL_ERROR","request_id":"` + requestID + `"}`
	w.Write([]byte(response))
}

// RequestLoggerMiddleware logs HTTP requests with timing information.
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		logHTTPRequest(r, ww, time.Since(start))
	})
}

// logHTTPRequest logs the completed HTTP request details
func logHTTPRequest(r *http.Request, ww middleware.WrapResponseWriter, duration time.Duration) {
	logger.Ctx(r.Context()).Infow(constants.LogMsgHTTPRequestCompleted,
		constants.LogKeyMethod, r.Method,
		constants.LogKeyPath, r.URL.Path,
		constants.LogKeyStatusCode, ww.Status(),
		constants.LogKeyDuration, duration.Milliseconds(),
		constants.LogFieldBytesWritten, ww.BytesWritten(),
	)
}

// MetricsMiddleware records HTTP request metrics to Prometheus.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		recordRequestMetrics(r, ww.Status(), time.Since(start))
	})
}

// recordRequestMetrics records the HTTP request metrics
func recordRequestMetrics(r *http.Request, statusCode int, duration time.Duration) {
	path := normalizePath(r.URL.Path)
	metrics.RecordHTTPRequest(r.Method, path, statusCode, duration.Seconds())
}

// normalizePath normalizes the path to avoid high cardinality metrics.
// Replaces dynamic segments like IDs with placeholders.
func normalizePath(path string) string {
	// For now, return the path as-is
	// In production, you might want to replace numeric IDs with :id
	// e.g., /accounts/123 -> /accounts/:id
	return path
}

// RequestIDMiddleware adds a request ID to the response header.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		if requestID != "" {
			w.Header().Set(constants.HeaderRequestID, requestID)
		}
		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware adds a timeout to requests.
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, timeoutErrorResponse())
	}
}

// timeoutErrorResponse returns the timeout error response body
func timeoutErrorResponse() string {
	return `{"error":"Request timeout","code":"TIMEOUT"}`
}

// CORSMiddleware adds basic CORS headers.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Idempotency-Key, X-Request-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ContentTypeMiddleware ensures JSON content type for API responses.
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set default content type for responses
		w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
		next.ServeHTTP(w, r)
	})
}

// statusCodeToString converts HTTP status code to string
func statusCodeToString(code int) string {
	return strconv.Itoa(code)
}
