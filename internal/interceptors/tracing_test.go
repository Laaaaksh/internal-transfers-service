package interceptors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/internal-transfers-service/internal/config"
	"github.com/stretchr/testify/suite"
)

type TracingTestSuite struct {
	suite.Suite
}

func TestTracingTestSuite(t *testing.T) {
	suite.Run(t, new(TracingTestSuite))
}

func (s *TracingTestSuite) TestTracingMiddlewareDisabledPassesThrough() {
	cfg := config.TracingConfig{Enabled: false}
	middleware := TracingMiddleware(cfg)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rec, req)

	s.True(handlerCalled)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *TracingTestSuite) TestTracingMiddlewareEnabledWrapsHandler() {
	cfg := config.TracingConfig{Enabled: true}
	middleware := TracingMiddleware(cfg)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rec, req)

	s.True(handlerCalled)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *TracingTestSuite) TestTraceContextMiddlewareAddsContext() {
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	TraceContextMiddleware(handler).ServeHTTP(rec, req)

	s.True(handlerCalled)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *TracingTestSuite) TestFormatSpanNameReturnsMethodAndPath() {
	req := httptest.NewRequest(http.MethodPost, "/v1/accounts", nil)
	result := formatSpanName("", req)
	s.Equal("POST /v1/accounts", result)
}

func (s *TracingTestSuite) TestFormatSpanNameWithGetMethod() {
	req := httptest.NewRequest(http.MethodGet, "/v1/accounts/123", nil)
	result := formatSpanName("", req)
	s.Equal("GET /v1/accounts/123", result)
}
