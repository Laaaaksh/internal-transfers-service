// Package interceptors provides HTTP middleware for the application.
package interceptors

import (
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/metrics"
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
	stack := make([]byte, constants.StackSizeBytes)
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

	response := buildErrorResponseWithRequestID(constants.ErrMsgInternalServerError, constants.ErrCodeInternalError, requestID)
	_, _ = w.Write([]byte(response))
}

// buildErrorResponseWithRequestID builds a JSON error response with request ID
func buildErrorResponseWithRequestID(errMsg, errCode, requestID string) string {
	return constants.JSONErrorPrefix + errMsg +
		constants.JSONCodePrefix + errCode +
		constants.JSONRequestIDPrefix + requestID +
		constants.JSONSuffix
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
	return normalizePathSegments(path)
}

// normalizePathSegments replaces numeric and UUID segments with placeholders.
func normalizePathSegments(path string) string {
	segments := splitPath(path)
	for i, segment := range segments {
		segments[i] = normalizeSegment(segment)
	}
	return joinPath(segments)
}

// normalizeSegment replaces a single segment with a placeholder if needed.
func normalizeSegment(segment string) string {
	if isNumericSegment(segment) {
		return constants.PathPlaceholderID
	}
	if isUUIDSegment(segment) {
		return constants.PathPlaceholderUUID
	}
	return segment
}

// splitPath splits a path into segments.
func splitPath(path string) []string {
	if path == "" || path == "/" {
		return []string{}
	}
	// Remove leading slash
	if path[0] == '/' {
		path = path[1:]
	}
	// Split by /
	result := []string{}
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if i > start {
				result = append(result, path[start:i])
			}
			start = i + 1
		}
	}
	return result
}

// joinPath joins segments back into a path.
func joinPath(segments []string) string {
	if len(segments) == 0 {
		return "/"
	}
	result := ""
	for _, s := range segments {
		result += "/" + s
	}
	return result
}

// isNumericSegment checks if a segment is purely numeric.
func isNumericSegment(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isUUIDSegment checks if a segment looks like a UUID.
// UUIDs are 36 characters: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func isUUIDSegment(s string) bool {
	const uuidLength = 36
	if len(s) != uuidLength {
		return false
	}
	return validateUUIDFormat(s)
}

// validateUUIDFormat validates that a string follows UUID format.
func validateUUIDFormat(s string) bool {
	for i, c := range s {
		if !isValidUUIDChar(i, c) {
			return false
		}
	}
	return true
}

// isValidUUIDChar checks if a character is valid at the given UUID position.
func isValidUUIDChar(position int, c rune) bool {
	if isUUIDDashPosition(position) {
		return c == '-'
	}
	return isHexChar(c)
}

// isUUIDDashPosition checks if the position should have a dash in a UUID.
func isUUIDDashPosition(position int) bool {
	return position == 8 || position == 13 || position == 18 || position == 23
}

// isHexChar checks if a character is a valid hex character.
func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
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
	return buildErrorResponse(constants.ErrMsgRequestTimeout, constants.ErrCodeTimeout)
}

// buildErrorResponse builds a JSON error response
func buildErrorResponse(errMsg, errCode string) string {
	return constants.JSONErrorPrefix + errMsg +
		constants.JSONCodePrefix + errCode +
		constants.JSONSuffix
}

// CORSMiddleware adds basic CORS headers with default origin (*).
// For production, use CORSMiddlewareWithOrigin with a specific origin.
func CORSMiddleware(next http.Handler) http.Handler {
	return CORSMiddlewareWithOrigin(constants.CORSAllowOriginAll)(next)
}

// CORSMiddlewareWithOrigin creates a CORS middleware with a configurable origin.
// Use this in production with a specific origin instead of "*".
func CORSMiddlewareWithOrigin(allowOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(constants.HeaderAccessControlAllowOrigin, allowOrigin)
			w.Header().Set(constants.HeaderAccessControlAllowMethods, constants.CORSAllowMethodsAll)
			w.Header().Set(constants.HeaderAccessControlAllowHeaders, constants.CORSAllowHeadersCommon)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeMiddleware ensures JSON content type for API responses.
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set default content type for responses
		w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds standard security headers to responses.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)
		next.ServeHTTP(w, r)
	})
}

// setSecurityHeaders sets the standard security headers
func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set(constants.HeaderXContentTypeOptions, constants.ValueNoSniff)
	w.Header().Set(constants.HeaderXFrameOptions, constants.ValueDeny)
	w.Header().Set(constants.HeaderCacheControl, constants.ValueNoStore)
}

// MaxBytesMiddleware limits the size of request bodies to prevent DoS attacks.
func MaxBytesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, constants.MaxRequestBodySize)
		next.ServeHTTP(w, r)
	})
}

// ContentTypeValidationMiddleware validates Content-Type header for mutating requests.
func ContentTypeValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isMutatingRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		if !isValidJSONContentType(r) {
			writeUnsupportedMediaTypeError(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isMutatingRequest checks if the request is a mutating HTTP method
func isMutatingRequest(r *http.Request) bool {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

// isValidJSONContentType checks if the Content-Type header indicates JSON
func isValidJSONContentType(r *http.Request) bool {
	contentType := r.Header.Get(constants.HeaderContentType)
	return strings.HasPrefix(contentType, constants.ContentTypeJSONPrefix)
}

// writeUnsupportedMediaTypeError writes a 415 error response
func writeUnsupportedMediaTypeError(w http.ResponseWriter, r *http.Request) {
	logger.Ctx(r.Context()).Warnw(constants.LogMsgInvalidContentType,
		constants.LogKeyMethod, r.Method,
		constants.LogKeyPath, r.URL.Path,
		constants.HeaderContentType, r.Header.Get(constants.HeaderContentType),
	)

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(constants.HTTPStatusUnsupportedMediaType)
	_, _ = w.Write([]byte(buildUnsupportedMediaTypeResponse()))
}

// buildUnsupportedMediaTypeResponse builds the error response JSON
func buildUnsupportedMediaTypeResponse() string {
	return constants.JSONErrorPrefix + constants.ErrMsgUnsupportedMediaType +
		constants.JSONCodePrefix + constants.ErrCodeUnsupportedMedia +
		constants.JSONSuffix
}
