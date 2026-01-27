package interceptors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/internal-transfers-service/internal/interceptors"
	"github.com/stretchr/testify/suite"
)

// InitTestSuite tests the middleware chain functions
type InitTestSuite struct {
	suite.Suite
}

func TestInitSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

// TestNewChainCreatesChain verifies chain creation
func (s *InitTestSuite) TestNewChainCreatesChain() {
	chain := interceptors.NewChain()
	s.NotNil(chain)
}

// TestChainThenWrapsHandler verifies Then wraps the handler
func (s *InitTestSuite) TestChainThenWrapsHandler() {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	chain := interceptors.NewChain()
	wrappedHandler := chain.Then(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	s.True(called)
	s.Equal(http.StatusOK, rec.Code)
}

// TestChainAppendAddsMiddleware verifies middleware can be appended
func (s *InitTestSuite) TestChainAppendAddsMiddleware() {
	middleware1Called := false
	middleware2Called := false

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware1Called = true
			next.ServeHTTP(w, r)
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware2Called = true
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	chain := interceptors.NewChain(mw1)
	chain = chain.Append(mw2)
	wrappedHandler := chain.Then(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	s.True(middleware1Called)
	s.True(middleware2Called)
	s.Equal(http.StatusOK, rec.Code)
}

// TestChainExecutesMiddlewareInOrder verifies middleware order is preserved
func (s *InitTestSuite) TestChainExecutesMiddlewareInOrder() {
	order := []int{}

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 1)
			next.ServeHTTP(w, r)
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 2)
			next.ServeHTTP(w, r)
		})
	}

	mw3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 3)
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, 4)
		w.WriteHeader(http.StatusOK)
	})

	chain := interceptors.NewChain(mw1, mw2, mw3)
	wrappedHandler := chain.Then(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	s.Equal([]int{1, 2, 3, 4}, order)
}

// TestDefaultMiddlewareReturnsMiddleware verifies default middleware is returned
func (s *InitTestSuite) TestDefaultMiddlewareReturnsMiddleware() {
	middlewares := interceptors.DefaultMiddleware()
	s.NotEmpty(middlewares)
	s.Greater(len(middlewares), 0)
}

// TestGetChiMiddlewareReturnsMiddleware verifies chi middleware is returned
func (s *InitTestSuite) TestGetChiMiddlewareReturnsMiddleware() {
	middlewares := interceptors.GetChiMiddleware()
	s.NotEmpty(middlewares)
	s.Greater(len(middlewares), 0)
}

// TestApplyMiddlewareAppliesAllMiddleware verifies all middleware is applied
func (s *InitTestSuite) TestApplyMiddlewareAppliesAllMiddleware() {
	callCount := 0

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := interceptors.ApplyMiddleware(handler, mw, mw, mw)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	s.Equal(3, callCount)
	s.Equal(http.StatusOK, rec.Code)
}
