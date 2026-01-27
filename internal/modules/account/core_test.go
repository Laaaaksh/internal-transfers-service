package account_test

import (
	"context"
	"errors"
	"testing"

	"github.com/internal-transfers-service/internal/modules/account"
	"github.com/internal-transfers-service/internal/modules/account/entities"
	"github.com/internal-transfers-service/internal/modules/account/mock"
	"github.com/internal-transfers-service/pkg/apperror"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// CoreTestSuite contains tests for account Core
type CoreTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *mock.MockIRepository
	core     account.ICore
	ctx      context.Context
}

func TestCoreSuite(t *testing.T) {
	suite.Run(t, new(CoreTestSuite))
}

func (s *CoreTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mock.NewMockIRepository(s.ctrl)
	s.ctx = context.Background()
	s.core = account.NewCoreWithRepo(s.ctx, s.mockRepo)
}

func (s *CoreTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test Create Account - Success Cases

func (s *CoreTestSuite) TestCreateAccountWithValidDataSucceeds() {
	req := &entities.CreateAccountRequest{
		AccountID:      123,
		InitialBalance: "100.50",
	}

	s.mockRepo.EXPECT().
		Exists(s.ctx, int64(123)).
		Return(false, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, acc *account.Account) error {
			s.Equal(int64(123), acc.AccountID)
			s.True(acc.Balance.Equal(decimal.NewFromFloat(100.50)))
			return nil
		}).
		Times(1)

	err := s.core.Create(s.ctx, req)
	s.Nil(err)
}

func (s *CoreTestSuite) TestCreateAccountWithZeroBalanceSucceeds() {
	req := &entities.CreateAccountRequest{
		AccountID:      456,
		InitialBalance: "0",
	}

	s.mockRepo.EXPECT().
		Exists(s.ctx, int64(456)).
		Return(false, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		Return(nil).
		Times(1)

	err := s.core.Create(s.ctx, req)
	s.Nil(err)
}

func (s *CoreTestSuite) TestCreateAccountWithHighPrecisionSucceeds() {
	req := &entities.CreateAccountRequest{
		AccountID:      789,
		InitialBalance: "123.45678901",
	}

	s.mockRepo.EXPECT().
		Exists(s.ctx, int64(789)).
		Return(false, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, acc *account.Account) error {
			expectedBalance, _ := decimal.NewFromString("123.45678901")
			s.True(acc.Balance.Equal(expectedBalance))
			return nil
		}).
		Times(1)

	err := s.core.Create(s.ctx, req)
	s.Nil(err)
}

// Test Create Account - Validation Errors

func (s *CoreTestSuite) TestCreateAccountWithZeroAccountIDFails() {
	req := &entities.CreateAccountRequest{
		AccountID:      0,
		InitialBalance: "100.00",
	}

	err := s.core.Create(s.ctx, req)
	s.NotNil(err)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

func (s *CoreTestSuite) TestCreateAccountWithNegativeAccountIDFails() {
	req := &entities.CreateAccountRequest{
		AccountID:      -1,
		InitialBalance: "100.00",
	}

	err := s.core.Create(s.ctx, req)
	s.NotNil(err)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

func (s *CoreTestSuite) TestCreateAccountWithNegativeBalanceFails() {
	req := &entities.CreateAccountRequest{
		AccountID:      123,
		InitialBalance: "-50.00",
	}

	err := s.core.Create(s.ctx, req)
	s.NotNil(err)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

func (s *CoreTestSuite) TestCreateAccountWithInvalidDecimalFormatFails() {
	req := &entities.CreateAccountRequest{
		AccountID:      123,
		InitialBalance: "not-a-number",
	}

	err := s.core.Create(s.ctx, req)
	s.NotNil(err)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

// Test Create Account - Conflict Error

func (s *CoreTestSuite) TestCreateAccountWhenAccountExistsFails() {
	req := &entities.CreateAccountRequest{
		AccountID:      123,
		InitialBalance: "100.00",
	}

	s.mockRepo.EXPECT().
		Exists(s.ctx, int64(123)).
		Return(true, nil).
		Times(1)

	err := s.core.Create(s.ctx, req)
	s.NotNil(err)
	s.Equal(apperror.CodeConflict, err.Code())
}

// Test Create Account - Repository Errors

func (s *CoreTestSuite) TestCreateAccountWhenExistsCheckFailsReturnsError() {
	req := &entities.CreateAccountRequest{
		AccountID:      123,
		InitialBalance: "100.00",
	}

	s.mockRepo.EXPECT().
		Exists(s.ctx, int64(123)).
		Return(false, errors.New("database connection failed")).
		Times(1)

	err := s.core.Create(s.ctx, req)
	s.NotNil(err)
	s.Equal(apperror.CodeInternalError, err.Code())
}

func (s *CoreTestSuite) TestCreateAccountWhenCreateFailsReturnsError() {
	req := &entities.CreateAccountRequest{
		AccountID:      123,
		InitialBalance: "100.00",
	}

	s.mockRepo.EXPECT().
		Exists(s.ctx, int64(123)).
		Return(false, nil).
		Times(1)

	s.mockRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		Return(errors.New("insert failed")).
		Times(1)

	err := s.core.Create(s.ctx, req)
	s.NotNil(err)
	s.Equal(apperror.CodeInternalError, err.Code())
}

// Test Get Account By ID - Success Cases

func (s *CoreTestSuite) TestGetByIDWithValidAccountReturnsAccount() {
	expectedAccount := &account.Account{
		AccountID: 123,
		Balance:   decimal.NewFromFloat(250.75),
	}

	s.mockRepo.EXPECT().
		GetByID(s.ctx, int64(123)).
		Return(expectedAccount, nil).
		Times(1)

	response, err := s.core.GetByID(s.ctx, 123)
	s.Nil(err)
	s.NotNil(response)
	s.Equal(int64(123), response.AccountID)
	s.Equal("250.75", response.Balance)
}

func (s *CoreTestSuite) TestGetByIDWithZeroBalanceReturnsCorrectBalance() {
	expectedAccount := &account.Account{
		AccountID: 456,
		Balance:   decimal.Zero,
	}

	s.mockRepo.EXPECT().
		GetByID(s.ctx, int64(456)).
		Return(expectedAccount, nil).
		Times(1)

	response, err := s.core.GetByID(s.ctx, 456)
	s.Nil(err)
	s.NotNil(response)
	s.Equal("0", response.Balance)
}

// Test Get Account By ID - Validation Errors

func (s *CoreTestSuite) TestGetByIDWithZeroAccountIDFails() {
	response, err := s.core.GetByID(s.ctx, 0)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

func (s *CoreTestSuite) TestGetByIDWithNegativeAccountIDFails() {
	response, err := s.core.GetByID(s.ctx, -1)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

// Test Get Account By ID - Not Found Error

func (s *CoreTestSuite) TestGetByIDWhenAccountNotFoundReturnsNotFoundError() {
	notFoundErr := apperror.New(apperror.CodeNotFound, errors.New("account not found"))

	s.mockRepo.EXPECT().
		GetByID(s.ctx, int64(999)).
		Return(nil, notFoundErr).
		Times(1)

	response, err := s.core.GetByID(s.ctx, 999)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeNotFound, err.Code())
}

// Test Get Account By ID - Repository Errors

func (s *CoreTestSuite) TestGetByIDWhenRepoFailsReturnsInternalError() {
	s.mockRepo.EXPECT().
		GetByID(s.ctx, int64(123)).
		Return(nil, errors.New("database error")).
		Times(1)

	response, err := s.core.GetByID(s.ctx, 123)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeInternalError, err.Code())
}
