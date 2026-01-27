package transaction_test

import (
	"context"
	"errors"
	"testing"

	"github.com/internal-transfers-service/internal/modules/account"
	accountMock "github.com/internal-transfers-service/internal/modules/account/mock"
	"github.com/internal-transfers-service/internal/modules/transaction"
	"github.com/internal-transfers-service/internal/modules/transaction/entities"
	txMock "github.com/internal-transfers-service/internal/modules/transaction/mock"
	"github.com/internal-transfers-service/pkg/apperror"
	dbMock "github.com/internal-transfers-service/pkg/database/mock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// Test constants
const (
	testSourceAccountID      = int64(100)
	testDestinationAccountID = int64(200)
	testValidAmount          = "50.00"
	testLargeAmount          = "1000.00"
)

// Test error constants - used for simulating database errors in tests
var (
	errDatabaseConnectionFailed = errors.New("database connection failed")
	errUpdateFailed             = errors.New("update failed")
	errInsertFailed             = errors.New("insert failed")
	errCommitFailed             = errors.New("commit failed")
)

// CoreTestSuite contains tests for transaction Core
type CoreTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockTxRepo      *txMock.MockIRepository
	mockAccountRepo *accountMock.MockIRepository
	mockPgxTx       *dbMock.MockTx
	core            transaction.ICore
	ctx             context.Context
}

func TestCoreSuite(t *testing.T) {
	suite.Run(t, new(CoreTestSuite))
}

func (s *CoreTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockTxRepo = txMock.NewMockIRepository(s.ctrl)
	s.mockAccountRepo = accountMock.NewMockIRepository(s.ctrl)
	s.mockPgxTx = dbMock.NewMockTx(s.ctrl)
	s.ctx = context.Background()
	s.core = transaction.NewCoreWithRepo(s.ctx, s.mockTxRepo, s.mockAccountRepo)
}

func (s *CoreTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Helper method to create source account with balance
func (s *CoreTestSuite) createSourceAccount(balance string) *account.Account {
	bal, _ := decimal.NewFromString(balance)
	return &account.Account{
		AccountID: testSourceAccountID,
		Balance:   bal,
	}
}

// Helper method to create destination account with balance
func (s *CoreTestSuite) createDestAccount(balance string) *account.Account {
	bal, _ := decimal.NewFromString(balance)
	return &account.Account{
		AccountID: testDestinationAccountID,
		Balance:   bal,
	}
}

// Test Transfer - Success Cases

func (s *CoreTestSuite) TestTransferWithValidDataSucceeds() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               testValidAmount,
	}

	sourceAccount := s.createSourceAccount("100.00")
	destAccount := s.createDestAccount("50.00")

	// Mock transaction begin
	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	// Lock accounts in order (source=100 < dest=200, so source first)
	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(sourceAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID).
		Return(destAccount, nil).
		Times(1)

	// Update source balance (100 - 50 = 50)
	expectedSourceBalance, _ := decimal.NewFromString("50.00")
	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testSourceAccountID, expectedSourceBalance).
		Return(nil).
		Times(1)

	// Update destination balance (50 + 50 = 100)
	expectedDestBalance, _ := decimal.NewFromString("100.00")
	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testDestinationAccountID, expectedDestBalance).
		Return(nil).
		Times(1)

	// Create transaction record
	s.mockTxRepo.EXPECT().
		Create(s.ctx, s.mockPgxTx, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ any, txRecord *transaction.Transaction) error {
			s.Equal(testSourceAccountID, txRecord.SourceAccountID)
			s.Equal(testDestinationAccountID, txRecord.DestinationAccountID)
			expectedAmount, _ := decimal.NewFromString(testValidAmount)
			s.True(txRecord.Amount.Equal(expectedAmount))
			return nil
		}).
		Times(1)

	// Commit transaction
	s.mockPgxTx.EXPECT().
		Commit(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.Nil(err)
	s.NotNil(response)
	s.NotEmpty(response.TransactionID)
}

func (s *CoreTestSuite) TestTransferWithReversedAccountOrderLocksCorrectly() {
	// When destination ID < source ID, destination should be locked first
	req := &entities.TransferRequest{
		SourceAccountID:      testDestinationAccountID, // 200
		DestinationAccountID: testSourceAccountID,      // 100
		Amount:               "25.00",
	}

	sourceAccount := &account.Account{
		AccountID: testDestinationAccountID,
		Balance:   decimal.NewFromFloat(100.00),
	}
	destAccount := &account.Account{
		AccountID: testSourceAccountID,
		Balance:   decimal.NewFromFloat(50.00),
	}

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	// Lock in order: 100 first, then 200
	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID). // 100 first
		Return(destAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID). // 200 second
		Return(sourceAccount, nil).
		Times(1)

	// Update balances
	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testDestinationAccountID, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testSourceAccountID, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockTxRepo.EXPECT().
		Create(s.ctx, s.mockPgxTx, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockPgxTx.EXPECT().
		Commit(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.Nil(err)
	s.NotNil(response)
}

func (s *CoreTestSuite) TestTransferWithHighPrecisionAmountSucceeds() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               "123.45678901",
	}

	sourceAccount := s.createSourceAccount("500.00000000")
	destAccount := s.createDestAccount("100.00000000")

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(sourceAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID).
		Return(destAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testSourceAccountID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ any, _ int64, newBalance decimal.Decimal) error {
			expected, _ := decimal.NewFromString("376.54321099")
			s.True(newBalance.Equal(expected), "Expected %s but got %s", expected.String(), newBalance.String())
			return nil
		}).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testDestinationAccountID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ any, _ int64, newBalance decimal.Decimal) error {
			expected, _ := decimal.NewFromString("223.45678901")
			s.True(newBalance.Equal(expected), "Expected %s but got %s", expected.String(), newBalance.String())
			return nil
		}).
		Times(1)

	s.mockTxRepo.EXPECT().
		Create(s.ctx, s.mockPgxTx, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockPgxTx.EXPECT().
		Commit(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.Nil(err)
	s.NotNil(response)
}

// Test Transfer - Validation Errors

func (s *CoreTestSuite) TestTransferWithSameSourceAndDestinationFails() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testSourceAccountID, // Same as source
		Amount:               testValidAmount,
	}

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

func (s *CoreTestSuite) TestTransferWithInvalidDecimalFormatFails() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               "not-a-number",
	}

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

func (s *CoreTestSuite) TestTransferWithZeroAmountFails() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               "0",
	}

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

func (s *CoreTestSuite) TestTransferWithNegativeAmountFails() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               "-50.00",
	}

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeBadRequest, err.Code())
}

// Test Transfer - Transaction Begin Error

func (s *CoreTestSuite) TestTransferWhenBeginTxFailsReturnsError() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               testValidAmount,
	}

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(nil, errDatabaseConnectionFailed).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeInternalError, err.Code())
}

// Test Transfer - Account Not Found Errors

func (s *CoreTestSuite) TestTransferWhenSourceAccountNotFoundFails() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               testValidAmount,
	}

	notFoundErr := apperror.New(apperror.CodeNotFound, account.ErrAccountNotFound)

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	// Source ID (100) < Dest ID (200), so source is locked first
	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(nil, notFoundErr).
		Times(1)

	s.mockPgxTx.EXPECT().
		Rollback(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeNotFound, err.Code())
}

func (s *CoreTestSuite) TestTransferWhenDestinationAccountNotFoundFails() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               testValidAmount,
	}

	sourceAccount := s.createSourceAccount("100.00")
	notFoundErr := apperror.New(apperror.CodeNotFound, account.ErrAccountNotFound)

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(sourceAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID).
		Return(nil, notFoundErr).
		Times(1)

	s.mockPgxTx.EXPECT().
		Rollback(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeNotFound, err.Code())
}

// Test Transfer - Insufficient Balance Error

func (s *CoreTestSuite) TestTransferWithInsufficientBalanceFails() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               testLargeAmount, // 1000.00, more than source has
	}

	sourceAccount := s.createSourceAccount("100.00") // Only 100.00
	destAccount := s.createDestAccount("50.00")

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(sourceAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID).
		Return(destAccount, nil).
		Times(1)

	s.mockPgxTx.EXPECT().
		Rollback(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeInsufficientFunds, err.Code())
}

func (s *CoreTestSuite) TestTransferWithExactBalanceSucceeds() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               "100.00", // Exactly what source has
	}

	sourceAccount := s.createSourceAccount("100.00")
	destAccount := s.createDestAccount("50.00")

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(sourceAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID).
		Return(destAccount, nil).
		Times(1)

	// Source balance should be 0
	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testSourceAccountID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ any, _ int64, newBalance decimal.Decimal) error {
			s.True(newBalance.IsZero(), "Expected zero balance but got %s", newBalance.String())
			return nil
		}).
		Times(1)

	// Dest balance should be 150
	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testDestinationAccountID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ any, _ int64, newBalance decimal.Decimal) error {
			expected, _ := decimal.NewFromString("150.00")
			s.True(newBalance.Equal(expected), "Expected %s but got %s", expected.String(), newBalance.String())
			return nil
		}).
		Times(1)

	s.mockTxRepo.EXPECT().
		Create(s.ctx, s.mockPgxTx, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockPgxTx.EXPECT().
		Commit(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.Nil(err)
	s.NotNil(response)
}

// Test Transfer - Update Balance Errors

func (s *CoreTestSuite) TestTransferWhenUpdateSourceBalanceFailsReturnsError() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               testValidAmount,
	}

	sourceAccount := s.createSourceAccount("100.00")
	destAccount := s.createDestAccount("50.00")

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(sourceAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID).
		Return(destAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testSourceAccountID, gomock.Any()).
		Return(errUpdateFailed).
		Times(1)

	s.mockPgxTx.EXPECT().
		Rollback(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeInternalError, err.Code())
}

func (s *CoreTestSuite) TestTransferWhenUpdateDestBalanceFailsReturnsError() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               testValidAmount,
	}

	sourceAccount := s.createSourceAccount("100.00")
	destAccount := s.createDestAccount("50.00")

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(sourceAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID).
		Return(destAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testSourceAccountID, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testDestinationAccountID, gomock.Any()).
		Return(errUpdateFailed).
		Times(1)

	s.mockPgxTx.EXPECT().
		Rollback(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeInternalError, err.Code())
}

// Test Transfer - Transaction Record Creation Error

func (s *CoreTestSuite) TestTransferWhenCreateTransactionRecordFailsReturnsError() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               testValidAmount,
	}

	sourceAccount := s.createSourceAccount("100.00")
	destAccount := s.createDestAccount("50.00")

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(sourceAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID).
		Return(destAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testSourceAccountID, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testDestinationAccountID, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockTxRepo.EXPECT().
		Create(s.ctx, s.mockPgxTx, gomock.Any()).
		Return(errInsertFailed).
		Times(1)

	s.mockPgxTx.EXPECT().
		Rollback(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeInternalError, err.Code())
}

// Test Transfer - Commit Error

func (s *CoreTestSuite) TestTransferWhenCommitFailsReturnsError() {
	req := &entities.TransferRequest{
		SourceAccountID:      testSourceAccountID,
		DestinationAccountID: testDestinationAccountID,
		Amount:               testValidAmount,
	}

	sourceAccount := s.createSourceAccount("100.00")
	destAccount := s.createDestAccount("50.00")

	s.mockTxRepo.EXPECT().
		BeginTx(s.ctx).
		Return(s.mockPgxTx, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testSourceAccountID).
		Return(sourceAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		GetForUpdate(s.ctx, s.mockPgxTx, testDestinationAccountID).
		Return(destAccount, nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testSourceAccountID, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockAccountRepo.EXPECT().
		UpdateBalance(s.ctx, s.mockPgxTx, testDestinationAccountID, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockTxRepo.EXPECT().
		Create(s.ctx, s.mockPgxTx, gomock.Any()).
		Return(nil).
		Times(1)

	s.mockPgxTx.EXPECT().
		Commit(s.ctx).
		Return(errCommitFailed).
		Times(1)

	s.mockPgxTx.EXPECT().
		Rollback(s.ctx).
		Return(nil).
		Times(1)

	response, err := s.core.Transfer(s.ctx, req)
	s.NotNil(err)
	s.Nil(response)
	s.Equal(apperror.CodeInternalError, err.Code())
}
