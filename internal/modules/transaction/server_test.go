package transaction_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/internal-transfers-service/internal/constants"
	accountMock "github.com/internal-transfers-service/internal/modules/account/mock"
	"github.com/internal-transfers-service/internal/modules/transaction"
	"github.com/internal-transfers-service/internal/modules/transaction/entities"
	"github.com/internal-transfers-service/internal/modules/transaction/mock"
	"github.com/internal-transfers-service/pkg/apperror"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// ServerTestSuite contains tests for transaction HTTPHandler
type ServerTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockCore *mock.MockICore
	handler  *transaction.HTTPHandler
	router   chi.Router
	ctx      context.Context
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (s *ServerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockCore = mock.NewMockICore(s.ctrl)
	s.handler = transaction.NewHTTPHandler(s.mockCore)
	s.router = chi.NewRouter()
	s.handler.RegisterRoutes(s.router)
	s.ctx = context.Background()
}

func (s *ServerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// TestNewHTTPHandlerCreatesHandler verifies handler creation
func (s *ServerTestSuite) TestNewHTTPHandlerCreatesHandler() {
	handler := transaction.NewHTTPHandler(s.mockCore)
	s.NotNil(handler)
}

// TestRegisterRoutesAddsRoutes verifies routes registration
func (s *ServerTestSuite) TestRegisterRoutesAddsRoutes() {
	expectedRequest := &entities.TransferRequest{
		SourceAccountID:      int64(1),
		DestinationAccountID: int64(2),
		Amount:               "100.00",
	}
	expectedResponse := &entities.TransferResponse{
		TransactionID: "txn-123",
	}

	s.mockCore.EXPECT().
		Transfer(gomock.Any(), expectedRequest).
		Return(expectedResponse, nil).
		Times(1)

	body := `{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(body))
	req.Header.Set(constants.HeaderContentType, constants.ContentTypeJSON)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)
}

// CreateTransaction Tests

func (s *ServerTestSuite) TestCreateTransactionSuccessReturnsCreated() {
	expectedResponse := &entities.TransferResponse{
		TransactionID: "abc-123-def-456",
	}

	s.mockCore.EXPECT().
		Transfer(gomock.Any(), &entities.TransferRequest{
			SourceAccountID:      int64(100),
			DestinationAccountID: int64(200),
			Amount:               "50.00",
		}).
		Return(expectedResponse, nil).
		Times(1)

	body := `{"source_account_id": 100, "destination_account_id": 200, "amount": "50.00"}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(body))
	req.Header.Set(constants.HeaderContentType, constants.ContentTypeJSON)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)

	var response entities.TransferResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	s.NoError(err)
	s.Equal("abc-123-def-456", response.TransactionID)
}

func (s *ServerTestSuite) TestCreateTransactionWithInvalidJSONReturnsBadRequest() {
	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(body))
	req.Header.Set(constants.HeaderContentType, constants.ContentTypeJSON)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)

	var response apperror.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	s.NoError(err)
	s.NotEmpty(response.Error)
}

func (s *ServerTestSuite) TestCreateTransactionInsufficientBalanceReturnsError() {
	expectedRequest := &entities.TransferRequest{
		SourceAccountID:      int64(100),
		DestinationAccountID: int64(200),
		Amount:               "999999.00",
	}
	coreError := apperror.NewWithMessage(apperror.CodeInsufficientFunds, transaction.ErrInsufficientBalance, "Insufficient balance")

	s.mockCore.EXPECT().
		Transfer(gomock.Any(), expectedRequest).
		Return(nil, coreError).
		Times(1)

	body := `{"source_account_id": 100, "destination_account_id": 200, "amount": "999999.00"}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(body))
	req.Header.Set(constants.HeaderContentType, constants.ContentTypeJSON)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusUnprocessableEntity, rec.Code)
}

func (s *ServerTestSuite) TestCreateTransactionAccountNotFoundReturnsNotFound() {
	expectedRequest := &entities.TransferRequest{
		SourceAccountID:      int64(999),
		DestinationAccountID: int64(200),
		Amount:               "50.00",
	}
	coreError := apperror.NewWithMessage(apperror.CodeNotFound, transaction.ErrSourceNotFound, "Account not found")

	s.mockCore.EXPECT().
		Transfer(gomock.Any(), expectedRequest).
		Return(nil, coreError).
		Times(1)

	body := `{"source_account_id": 999, "destination_account_id": 200, "amount": "50.00"}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(body))
	req.Header.Set(constants.HeaderContentType, constants.ContentTypeJSON)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *ServerTestSuite) TestCreateTransactionSameAccountReturnsError() {
	expectedRequest := &entities.TransferRequest{
		SourceAccountID:      int64(100),
		DestinationAccountID: int64(100),
		Amount:               "50.00",
	}
	coreError := apperror.NewWithMessage(apperror.CodeBadRequest, transaction.ErrSameAccountTransfer, "Cannot transfer to same account")

	s.mockCore.EXPECT().
		Transfer(gomock.Any(), expectedRequest).
		Return(nil, coreError).
		Times(1)

	body := `{"source_account_id": 100, "destination_account_id": 100, "amount": "50.00"}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(body))
	req.Header.Set(constants.HeaderContentType, constants.ContentTypeJSON)
	rec := httptest.NewRecorder()

	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

// InitTestSuite contains tests for transaction module initialization
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
	mockAcctRepo := accountMock.NewMockIRepository(ctrl)
	core := transaction.NewCoreWithRepo(s.ctx, mockRepo, mockAcctRepo)
	handler := transaction.NewHTTPHandler(core)

	module := &transaction.Module{
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
	mockAcctRepo := accountMock.NewMockIRepository(ctrl)
	core := transaction.NewCore(s.ctx, mockRepo, mockAcctRepo)

	s.NotNil(core)
}

// TestGetCoreReturnsSingleton verifies GetCore returns singleton
func (s *InitTestSuite) TestGetCoreReturnsSingleton() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	mockRepo := mock.NewMockIRepository(ctrl)
	mockAcctRepo := accountMock.NewMockIRepository(ctrl)

	core1 := transaction.NewCore(s.ctx, mockRepo, mockAcctRepo)
	core2 := transaction.GetCore()

	s.Equal(core1, core2)
}
