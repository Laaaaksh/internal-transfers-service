package interceptors

import (
	"bytes"
	"net/http"

	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/idempotency"
	"github.com/internal-transfers-service/internal/modules/idempotency/entities"
)

// IdempotencyMiddleware provides idempotent request handling.
// It caches responses by idempotency key and returns cached responses on retry.
func IdempotencyMiddleware(repo idempotency.IRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !shouldApplyIdempotency(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get(constants.HeaderIdempotencyKey)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !isValidIdempotencyKey(key) {
				writeIdempotencyError(w, r, key)
				return
			}

			if handleCachedResponse(w, r, repo, key) {
				return
			}

			captureAndStoreResponse(w, r, next, repo, key)
		})
	}
}

// shouldApplyIdempotency determines if idempotency should be applied to this request.
// Only applies to mutating HTTP methods (POST, PUT, PATCH).
func shouldApplyIdempotency(r *http.Request) bool {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

// isValidIdempotencyKey validates the idempotency key format.
func isValidIdempotencyKey(key string) bool {
	return len(key) <= entities.MaxKeyLength
}

// writeIdempotencyError writes an error response for invalid idempotency key.
func writeIdempotencyError(w http.ResponseWriter, r *http.Request, key string) {
	truncatedKey := truncateKey(key)
	logger.Ctx(r.Context()).Warnw(entities.LogMsgIdempotencyKeyTooLong,
		entities.LogFieldIdempotencyKey, truncatedKey,
		entities.LogFieldKeyLength, len(key),
	)

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(buildIdempotencyErrorResponse()))
}

// truncateKey truncates long keys for safe logging
func truncateKey(key string) string {
	const maxLogKeyLength = 50
	if len(key) > maxLogKeyLength {
		return key[:maxLogKeyLength] + "..."
	}
	return key
}

// buildIdempotencyErrorResponse builds the error response JSON
func buildIdempotencyErrorResponse() string {
	return entities.ErrResponsePrefix + constants.ErrMsgIdempotencyKeyTooLong +
		entities.ErrResponseCode + constants.ErrCodeInvalidIdempotency + entities.ErrResponseSuffix
}

// handleCachedResponse checks for a cached response and returns it if found.
// Returns true if a cached response was returned, false otherwise.
func handleCachedResponse(w http.ResponseWriter, r *http.Request, repo idempotency.IRepository, key string) bool {
	record, err := repo.Get(r.Context(), key)
	if err != nil {
		logger.Ctx(r.Context()).Warnw(entities.LogMsgIdempotencyGetFailed,
			entities.LogFieldIdempotencyKey, key,
			constants.LogKeyError, err,
		)
		return false
	}

	if record == nil {
		return false
	}

	logger.Ctx(r.Context()).Infow(entities.LogMsgIdempotencyHit,
		entities.LogFieldIdempotencyKey, key,
		entities.LogFieldCachedStatus, record.ResponseStatus,
	)

	writeCachedResponse(w, record)
	return true
}

// writeCachedResponse writes the cached response to the client.
func writeCachedResponse(w http.ResponseWriter, record *entities.IdempotencyRecord) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.Header().Set(entities.HeaderIdempotentReplayed, entities.HeaderValueTrue)
	w.WriteHeader(record.ResponseStatus)
	_, _ = w.Write(record.ResponseBody)
}

// captureAndStoreResponse captures the response and stores it for future requests.
func captureAndStoreResponse(w http.ResponseWriter, r *http.Request, next http.Handler, repo idempotency.IRepository, key string) {
	recorder := newResponseRecorder(w)
	next.ServeHTTP(recorder, r)

	if shouldCacheResponse(recorder) {
		storeResponse(r, repo, key, recorder)
		return
	}

	if recorder.bodyCapped {
		logResponseTooBig(r, key, recorder.statusCode)
	}
}

// logResponseTooBig logs when a response is too large to cache
func logResponseTooBig(r *http.Request, key string, statusCode int) {
	logger.Ctx(r.Context()).Debugw(entities.LogMsgIdempotencyResponseTooBig,
		entities.LogFieldIdempotencyKey, key,
		constants.LogKeyStatusCode, statusCode,
	)
}

// shouldCacheResponse determines if the response should be cached.
// Only caches 2xx-4xx responses (not 5xx server errors) and responses under the size limit.
func shouldCacheResponse(recorder *responseRecorder) bool {
	statusCodeCacheable := recorder.statusCode >= 200 && recorder.statusCode < 500
	return statusCodeCacheable && recorder.isCacheable()
}

// storeResponse stores the response in the idempotency cache.
func storeResponse(r *http.Request, repo idempotency.IRepository, key string, recorder *responseRecorder) {
	err := repo.Store(r.Context(), key, recorder.statusCode, recorder.body.Bytes())
	if err != nil {
		logger.Ctx(r.Context()).Warnw(entities.LogMsgIdempotencyStoreFailed,
			entities.LogFieldIdempotencyKey, key,
			constants.LogKeyError, err,
		)
		return
	}

	logger.Ctx(r.Context()).Debugw(entities.LogMsgIdempotencyStored,
		entities.LogFieldIdempotencyKey, key,
		constants.LogKeyStatusCode, recorder.statusCode,
	)
}

// responseRecorder captures the response for storage.
// It limits the cached body size to prevent memory issues with large responses.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	bodyCapped bool // true if we stopped capturing due to size limit
}

// newResponseRecorder creates a new response recorder.
func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
		bodyCapped:     false,
	}
}

// WriteHeader captures the status code and writes it to the underlying writer.
func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// Write captures the body and writes it to the underlying writer.
// Only captures up to MaxCachedResponseSize bytes to prevent memory issues.
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.captureBody(b)
	return r.ResponseWriter.Write(b)
}

// captureBody captures the response body up to the max size limit
func (r *responseRecorder) captureBody(b []byte) {
	if r.bodyCapped {
		return
	}

	remaining := entities.MaxCachedResponseSize - r.body.Len()
	if remaining <= 0 {
		r.bodyCapped = true
		return
	}

	if len(b) <= remaining {
		r.body.Write(b)
		return
	}

	r.body.Write(b[:remaining])
	r.bodyCapped = true
}

// isCacheable returns true if the response can be cached.
// Returns false if the body exceeded the max size limit.
func (r *responseRecorder) isCacheable() bool {
	return !r.bodyCapped
}
