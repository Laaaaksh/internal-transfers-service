package app_context_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/internal-transfers-service/internal/boot/app_context"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/stretchr/testify/suite"
)

// ContextTestSuite tests app context functions
type ContextTestSuite struct {
	suite.Suite
}

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}

// TestGetRequestIDReturnsEmptyForEmptyContext verifies empty context returns empty string
func (s *ContextTestSuite) TestGetRequestIDReturnsEmptyForEmptyContext() {
	ctx := context.Background()
	reqID := app_context.GetRequestID(ctx)
	s.Empty(reqID)
}

// TestSetRequestIDSetsValue verifies request ID is set in context
func (s *ContextTestSuite) TestSetRequestIDSetsValue() {
	ctx := context.Background()
	expectedID := "test-request-id-123"

	ctx = app_context.SetRequestID(ctx, expectedID)
	reqID := app_context.GetRequestID(ctx)

	s.Equal(expectedID, reqID)
}

// TestGetRequestIDFromChiMiddleware verifies chi middleware ID is retrieved
func (s *ContextTestSuite) TestGetRequestIDFromChiMiddleware() {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := app_context.GetRequestID(r.Context())
		s.NotEmpty(reqID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

// TestGetRequestIDFromRequestHeader verifies request ID from header is retrieved
func (s *ContextTestSuite) TestGetRequestIDFromRequestHeader() {
	expectedID := "header-request-id-456"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(constants.HeaderRequestID, expectedID)

	reqID := app_context.GetRequestIDFromRequest(req)

	s.Equal(expectedID, reqID)
}

// TestGetRequestIDFromRequestFallsBackToChiMiddleware verifies fallback to chi middleware
func (s *ContextTestSuite) TestGetRequestIDFromRequestFallsBackToChiMiddleware() {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := app_context.GetRequestIDFromRequest(r)
		s.NotEmpty(reqID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

// TestWithRequestIDAddsToContext verifies WithRequestID adds ID to context
func (s *ContextTestSuite) TestWithRequestIDAddsToContext() {
	expectedID := "with-request-id-789"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(constants.HeaderRequestID, expectedID)

	ctx := app_context.WithRequestID(req.Context(), req)
	reqID := app_context.GetRequestID(ctx)

	s.Equal(expectedID, reqID)
}

// TestWithRequestIDReturnsContextWhenNoID verifies context is returned when no ID
func (s *ContextTestSuite) TestWithRequestIDReturnsContextWhenNoID() {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	originalCtx := req.Context()
	newCtx := app_context.WithRequestID(originalCtx, req)

	// Should return same context since no ID to add
	s.NotNil(newCtx)
}

// TestAddRequestIDToResponseAddsHeader verifies header is added to response
func (s *ContextTestSuite) TestAddRequestIDToResponseAddsHeader() {
	expectedID := "response-request-id-101"
	ctx := app_context.SetRequestID(context.Background(), expectedID)
	rec := httptest.NewRecorder()

	app_context.AddRequestIDToResponse(ctx, rec)

	s.Equal(expectedID, rec.Header().Get(constants.HeaderRequestID))
}

// TestAddRequestIDToResponseDoesNothingWhenNoID verifies no header when no ID
func (s *ContextTestSuite) TestAddRequestIDToResponseDoesNothingWhenNoID() {
	ctx := context.Background()
	rec := httptest.NewRecorder()

	app_context.AddRequestIDToResponse(ctx, rec)

	s.Empty(rec.Header().Get(constants.HeaderRequestID))
}
