package interceptors_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/interceptors"
	"github.com/internal-transfers-service/internal/modules/idempotency/entities"
	"github.com/internal-transfers-service/internal/modules/idempotency/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// IdempotencyTestSuite contains tests for idempotency middleware
type IdempotencyTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *mock.MockIRepository
	ctx      context.Context
}

func TestIdempotencySuite(t *testing.T) {
	suite.Run(t, new(IdempotencyTestSuite))
}

func (s *IdempotencyTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mock.NewMockIRepository(s.ctrl)
	s.ctx = context.Background()
}

func (s *IdempotencyTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *IdempotencyTestSuite) TestMiddlewareSkipsGetRequests() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	middleware := interceptors.IdempotencyMiddleware(s.mockRepo)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/accounts/1", nil)
	req.Header.Set(constants.HeaderIdempotencyKey, "test-key")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *IdempotencyTestSuite) TestMiddlewareSkipsRequestsWithoutKey() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	})

	middleware := interceptors.IdempotencyMiddleware(s.mockRepo)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)
}

func (s *IdempotencyTestSuite) TestMiddlewareReturnsCachedResponseOnHit() {
	cachedRecord := &entities.IdempotencyRecord{
		Key:            "existing-key",
		ResponseStatus: 201,
		ResponseBody:   []byte(`{"transaction_id":"cached-123"}`),
		CreatedAt:      time.Now(),
	}

	s.mockRepo.EXPECT().
		Get(gomock.Any(), "existing-key").
		Return(cachedRecord, nil).
		Times(1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Fail("Handler should not be called on cache hit")
	})

	middleware := interceptors.IdempotencyMiddleware(s.mockRepo)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{}`))
	req.Header.Set(constants.HeaderIdempotencyKey, "existing-key")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	s.Equal(201, rec.Code)
	s.Contains(rec.Body.String(), "cached-123")
	s.Equal("true", rec.Header().Get(entities.HeaderIdempotentReplayed))
}

func (s *IdempotencyTestSuite) TestMiddlewareProcessesAndStoresOnMiss() {
	s.mockRepo.EXPECT().
		Get(gomock.Any(), "new-key").
		Return(nil, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Store(gomock.Any(), "new-key", 201, gomock.Any()).
		Return(nil).
		Times(1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"transaction_id":"new-123"}`))
	})

	middleware := interceptors.IdempotencyMiddleware(s.mockRepo)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{}`))
	req.Header.Set(constants.HeaderIdempotencyKey, "new-key")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)
	s.Contains(rec.Body.String(), "new-123")
}

func (s *IdempotencyTestSuite) TestMiddlewareRejectsKeyTooLong() {
	longKey := strings.Repeat("x", 300)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Fail("Handler should not be called for invalid key")
	})

	middleware := interceptors.IdempotencyMiddleware(s.mockRepo)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{}`))
	req.Header.Set(constants.HeaderIdempotencyKey, longKey)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Contains(rec.Body.String(), "Idempotency key too long")
}

func (s *IdempotencyTestSuite) TestMiddlewareDoesNotCache5xxErrors() {
	s.mockRepo.EXPECT().
		Get(gomock.Any(), "error-key").
		Return(nil, nil).
		Times(1)

	// Store should NOT be called for 5xx errors

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
	})

	middleware := interceptors.IdempotencyMiddleware(s.mockRepo)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{}`))
	req.Header.Set(constants.HeaderIdempotencyKey, "error-key")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *IdempotencyTestSuite) TestMiddlewareHandlesPutRequests() {
	s.mockRepo.EXPECT().
		Get(gomock.Any(), "put-key").
		Return(nil, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Store(gomock.Any(), "put-key", 200, gomock.Any()).
		Return(nil).
		Times(1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"updated":true}`))
	})

	middleware := interceptors.IdempotencyMiddleware(s.mockRepo)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodPut, "/accounts/1", bytes.NewBufferString(`{}`))
	req.Header.Set(constants.HeaderIdempotencyKey, "put-key")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *IdempotencyTestSuite) TestMiddlewareSkipsDeleteRequests() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	middleware := interceptors.IdempotencyMiddleware(s.mockRepo)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodDelete, "/accounts/1", nil)
	req.Header.Set(constants.HeaderIdempotencyKey, "delete-key")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)
}
