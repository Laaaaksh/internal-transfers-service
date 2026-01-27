package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/internal-transfers-service/internal/modules/health"
	"github.com/stretchr/testify/suite"
)

// ServerTestSuite contains tests for health HTTPHandler
type ServerTestSuite struct {
	suite.Suite
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

// TestNewHTTPHandlerCreatesHandler verifies handler creation
func (s *ServerTestSuite) TestNewHTTPHandlerCreatesHandler() {
	core := health.NewCoreForTesting(nil, true)
	handler := health.NewHTTPHandler(core)

	s.NotNil(handler)
}

// TestRegisterRoutesAddsRoutes verifies routes are registered
func (s *ServerTestSuite) TestRegisterRoutesAddsRoutes() {
	core := health.NewCoreForTesting(nil, true)
	handler := health.NewHTTPHandler(core)
	router := chi.NewRouter()

	handler.RegisterRoutes(router)

	// Verify routes are registered by making requests
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

// TestLivenessCheckReturnsServingWhenHealthy verifies healthy response
func (s *ServerTestSuite) TestLivenessCheckReturnsServingWhenHealthy() {
	core := health.NewCoreForTesting(nil, true)
	handler := health.NewHTTPHandler(core)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))

	var response health.HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	s.NoError(err)
	s.Equal("SERVING", response.Status)
}

// TestLivenessCheckReturnsNotServingWhenUnhealthy verifies unhealthy response
func (s *ServerTestSuite) TestLivenessCheckReturnsNotServingWhenUnhealthy() {
	core := health.NewCoreForTesting(nil, false)
	handler := health.NewHTTPHandler(core)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	s.Equal(http.StatusServiceUnavailable, rec.Code)

	var response health.HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	s.NoError(err)
	s.Equal("NOT_SERVING", response.Status)
}

// TestReadinessCheckReturnsServingWhenHealthy verifies ready response
func (s *ServerTestSuite) TestReadinessCheckReturnsServingWhenHealthy() {
	core := health.NewCoreForTesting(nil, true)
	handler := health.NewHTTPHandler(core)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))

	var response health.HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	s.NoError(err)
	s.Equal("SERVING", response.Status)
}

// TestReadinessCheckReturnsNotServingWhenUnhealthy verifies not ready response
func (s *ServerTestSuite) TestReadinessCheckReturnsNotServingWhenUnhealthy() {
	core := health.NewCoreForTesting(nil, false)
	handler := health.NewHTTPHandler(core)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	s.Equal(http.StatusServiceUnavailable, rec.Code)

	var response health.HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	s.NoError(err)
	s.Equal("NOT_SERVING", response.Status)
}

// TestLivenessCheckDirectCall tests handler directly
func (s *ServerTestSuite) TestLivenessCheckDirectCall() {
	core := health.NewCoreForTesting(nil, true)
	handler := health.NewHTTPHandler(core)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler.LivenessCheck(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

// TestReadinessCheckDirectCall tests handler directly
func (s *ServerTestSuite) TestReadinessCheckDirectCall() {
	core := health.NewCoreForTesting(nil, true)
	handler := health.NewHTTPHandler(core)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	handler.ReadinessCheck(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}
