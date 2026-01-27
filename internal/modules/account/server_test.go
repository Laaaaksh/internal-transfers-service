package account_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/internal-transfers-service/internal/modules/account"
	"github.com/internal-transfers-service/internal/modules/account/entities"
	"github.com/internal-transfers-service/internal/modules/account/mock"
	"github.com/internal-transfers-service/pkg/apperror"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// ServerTestSuite contains tests for account HTTPHandler
type ServerTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockCore *mock.MockICore
	handler  *account.HTTPHandler
	router   chi.Router
	ctx      context.Context
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (s *ServerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockCore = mock.NewMockICore(s.ctrl)
	s.handler = account.NewHTTPHandler(s.mockCore)
	s.router = chi.NewRouter()
	s.handler.RegisterRoutes(s.router)
	s.ctx = context.Background()
}

func (s *ServerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// TestNewHTTPHandlerCreatesHandler verifies handler creation
func (s *ServerTestSuite) TestNewHTTPHandlerCreatesHandler() {
	handler := account.NewHTTPHandler(s.mockCore)
	s.NotNil(handler)
}

// TestRegisterRoutesAddsRoutes verifies routes registration
func (s *ServerTestSuite) TestRegisterRoutesAddsRoutes() {
	expectedRequest := &entities.CreateAccountRequest{
		AccountID:      int64(1),
		InitialBalance: "100.00",
	}

	s.mockCore.EXPECT().
		Create(gomock.Any(), expectedRequest).
		Return(nil).
		Times(1)

	body := `{"account_id": 1, "initial_balance": "100.00"}`
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)
}

// CreateAccount Tests

func (s *ServerTestSuite) TestCreateAccountSuccessReturnsCreated() {
	s.mockCore.EXPECT().
		Create(gomock.Any(), &entities.CreateAccountRequest{
			AccountID:      int64(123),
			InitialBalance: "500.00",
		}).
		Return(nil).
		Times(1)

	body := `{"account_id": 123, "initial_balance": "500.00"}`
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)
}

func (s *ServerTestSuite) TestCreateAccountWithInvalidJSONReturnsBadRequest() {
	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)

	var response apperror.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *ServerTestSuite) TestCreateAccountWhenCoreReturnsErrorReturnsError() {
	expectedRequest := &entities.CreateAccountRequest{
		AccountID:      int64(123),
		InitialBalance: "100.00",
	}
	coreError := apperror.NewWithMessage(apperror.CodeConflict, account.ErrAccountExists, "Account already exists")

	s.mockCore.EXPECT().
		Create(gomock.Any(), expectedRequest).
		Return(coreError).
		Times(1)

	body := `{"account_id": 123, "initial_balance": "100.00"}`
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusConflict, rec.Code)
}

// GetAccount Tests

func (s *ServerTestSuite) TestGetAccountSuccessReturnsAccount() {
	expectedResponse := &entities.AccountResponse{
		AccountID: 123,
		Balance:   "500.00",
	}

	s.mockCore.EXPECT().
		GetByID(gomock.Any(), int64(123)).
		Return(expectedResponse, nil).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/accounts/123", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var response entities.AccountResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	s.NoError(err)
	s.Equal(int64(123), response.AccountID)
	s.Equal("500.00", response.Balance)
}

func (s *ServerTestSuite) TestGetAccountWithInvalidIDReturnsBadRequest() {
	req := httptest.NewRequest(http.MethodGet, "/accounts/invalid", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *ServerTestSuite) TestGetAccountNotFoundReturnsNotFound() {
	notFoundErr := apperror.NewWithMessage(apperror.CodeNotFound, account.ErrAccountNotFound, "Account not found")

	s.mockCore.EXPECT().
		GetByID(gomock.Any(), int64(999)).
		Return(nil, notFoundErr).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/accounts/999", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *ServerTestSuite) TestGetAccountWithZeroBalanceReturnsCorrectBalance() {
	expectedResponse := &entities.AccountResponse{
		AccountID: 456,
		Balance:   "0",
	}

	s.mockCore.EXPECT().
		GetByID(gomock.Any(), int64(456)).
		Return(expectedResponse, nil).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/accounts/456", nil)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var response entities.AccountResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	s.NoError(err)
	s.Equal("0", response.Balance)
}

// InitTestSuite contains tests for account module initialization
type InitTestSuite struct {
	suite.Suite
	ctx context.Context
}

func TestInitSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (s *InitTestSuite) SetupTest() {
	s.ctx = context.Background()
}

// TestModuleMethodsReturnCorrectValues verifies module methods
func (s *InitTestSuite) TestModuleMethodsReturnCorrectValues() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	mockRepo := mock.NewMockIRepository(ctrl)
	core := account.NewCoreWithRepo(s.ctx, mockRepo)
	handler := account.NewHTTPHandler(core)

	module := &account.Module{
		Core:    core,
		Handler: handler,
		Repo:    mockRepo,
	}

	s.Equal(core, module.GetCore())
	s.Equal(handler, module.GetHandler())
	s.Equal(mockRepo, module.GetRepository())
}

// TestNewCoreCreatesCore verifies NewCore function
func (s *InitTestSuite) TestNewCoreCreatesCore() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	mockRepo := mock.NewMockIRepository(ctrl)
	core := account.NewCore(s.ctx, mockRepo)

	s.NotNil(core)
}

// TestGetCoreReturnsSingleton verifies GetCore returns singleton
func (s *InitTestSuite) TestGetCoreReturnsSingleton() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	mockRepo := mock.NewMockIRepository(ctrl)

	core1 := account.NewCore(s.ctx, mockRepo)
	core2 := account.GetCore()

	s.Equal(core1, core2)
}

// Helper to create a test account
func createTestAccount(id int64, balance string) *account.Account {
	bal, _ := decimal.NewFromString(balance)
	return &account.Account{
		AccountID: id,
		Balance:   bal,
	}
}
