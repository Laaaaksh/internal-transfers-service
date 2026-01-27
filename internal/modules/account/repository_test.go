package account_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/internal-transfers-service/internal/modules/account"
	dbmock "github.com/internal-transfers-service/pkg/database/mock"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// Test error constants - used for simulating database errors in repository tests
var (
	errRepoDBConnectionFailed = errors.New("database connection failed")
	errRepoDuplicateKey       = errors.New("duplicate key value violates unique constraint")
	errRepoConnectionTimeout  = errors.New("connection timeout")
	errRepoLockTimeout        = errors.New("lock timeout")
	errRepoTxAborted          = errors.New("transaction aborted")
	errRepoQueryFailed        = errors.New("query execution failed")
)

// RepositoryTestSuite contains tests for account Repository
type RepositoryTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockPool *dbmock.MockIPool
	mockRow  *dbmock.MockRow
	mockTx   *dbmock.MockTx
	repo     account.IRepository
	ctx      context.Context
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (s *RepositoryTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockPool = dbmock.NewMockIPool(s.ctrl)
	s.mockRow = dbmock.NewMockRow(s.ctrl)
	s.mockTx = dbmock.NewMockTx(s.ctrl)
	s.ctx = context.Background()
	s.repo = account.NewRepository(s.mockPool)
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test Create - Success Cases

func (s *RepositoryTestSuite) TestCreateAccountSucceeds() {
	acc := &account.Account{
		AccountID: 123,
		Balance:   decimal.NewFromFloat(100.50),
	}

	s.mockPool.EXPECT().
		Exec(s.ctx, gomock.Any(), int64(123), acc.Balance, gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil).
		Times(1)

	err := s.repo.Create(s.ctx, acc)
	s.Nil(err)
	s.False(acc.CreatedAt.IsZero())
	s.False(acc.UpdatedAt.IsZero())
}

func (s *RepositoryTestSuite) TestCreateAccountSetsTimestamps() {
	acc := &account.Account{
		AccountID: 456,
		Balance:   decimal.Zero,
	}

	before := time.Now().UTC()

	s.mockPool.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil).
		Times(1)

	err := s.repo.Create(s.ctx, acc)
	s.Nil(err)

	after := time.Now().UTC()

	s.True(acc.CreatedAt.After(before) || acc.CreatedAt.Equal(before))
	s.True(acc.CreatedAt.Before(after) || acc.CreatedAt.Equal(after))
	s.Equal(acc.CreatedAt, acc.UpdatedAt)
}

// Test Create - Error Cases

func (s *RepositoryTestSuite) TestCreateAccountWhenExecFailsReturnsError() {
	acc := &account.Account{
		AccountID: 123,
		Balance:   decimal.NewFromFloat(100.50),
	}

	s.mockPool.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.CommandTag{}, errRepoDBConnectionFailed).
		Times(1)

	err := s.repo.Create(s.ctx, acc)
	s.NotNil(err)
	s.Equal(errRepoDBConnectionFailed, err)
}

func (s *RepositoryTestSuite) TestCreateAccountWithDuplicateKeyReturnsError() {
	acc := &account.Account{
		AccountID: 123,
		Balance:   decimal.NewFromFloat(100.50),
	}

	s.mockPool.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.CommandTag{}, errRepoDuplicateKey).
		Times(1)

	err := s.repo.Create(s.ctx, acc)
	s.NotNil(err)
	s.Contains(err.Error(), "duplicate key")
}

// Test GetByID - Success Cases

func (s *RepositoryTestSuite) TestGetByIDSucceeds() {
	expectedBalance := decimal.NewFromFloat(250.75)
	expectedCreatedAt := time.Now().UTC()
	expectedUpdatedAt := time.Now().UTC()

	s.mockPool.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(123)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = 123
			*dest[1].(*decimal.Decimal) = expectedBalance
			*dest[2].(*time.Time) = expectedCreatedAt
			*dest[3].(*time.Time) = expectedUpdatedAt
			return nil
		}).
		Times(1)

	result, err := s.repo.GetByID(s.ctx, 123)
	s.Nil(err)
	s.NotNil(result)
	s.Equal(int64(123), result.AccountID)
	s.True(result.Balance.Equal(expectedBalance))
}

func (s *RepositoryTestSuite) TestGetByIDWithZeroBalance() {
	s.mockPool.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(456)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = 456
			*dest[1].(*decimal.Decimal) = decimal.Zero
			*dest[2].(*time.Time) = time.Now().UTC()
			*dest[3].(*time.Time) = time.Now().UTC()
			return nil
		}).
		Times(1)

	result, err := s.repo.GetByID(s.ctx, 456)
	s.Nil(err)
	s.NotNil(result)
	s.True(result.Balance.IsZero())
}

// Test GetByID - Not Found Cases

func (s *RepositoryTestSuite) TestGetByIDWhenNotFoundReturnsNotFoundError() {
	s.mockPool.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(999)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgx.ErrNoRows).
		Times(1)

	result, err := s.repo.GetByID(s.ctx, 999)
	s.NotNil(err)
	s.Nil(result)
}

// Test GetByID - Error Cases

func (s *RepositoryTestSuite) TestGetByIDWhenDatabaseFailsReturnsError() {
	dbError := errRepoConnectionTimeout

	s.mockPool.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(123)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(dbError).
		Times(1)

	result, err := s.repo.GetByID(s.ctx, 123)
	s.NotNil(err)
	s.Nil(result)
	s.Equal(dbError, err)
}

// Test GetForUpdate - Success Cases

func (s *RepositoryTestSuite) TestGetForUpdateSucceeds() {
	expectedBalance := decimal.NewFromFloat(500.00)

	s.mockTx.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(123)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = 123
			*dest[1].(*decimal.Decimal) = expectedBalance
			*dest[2].(*time.Time) = time.Now().UTC()
			*dest[3].(*time.Time) = time.Now().UTC()
			return nil
		}).
		Times(1)

	result, err := s.repo.GetForUpdate(s.ctx, s.mockTx, 123)
	s.Nil(err)
	s.NotNil(result)
	s.Equal(int64(123), result.AccountID)
	s.True(result.Balance.Equal(expectedBalance))
}

// Test GetForUpdate - Not Found Cases

func (s *RepositoryTestSuite) TestGetForUpdateWhenNotFoundReturnsNotFoundError() {
	s.mockTx.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(999)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgx.ErrNoRows).
		Times(1)

	result, err := s.repo.GetForUpdate(s.ctx, s.mockTx, 999)
	s.NotNil(err)
	s.Nil(result)
}

// Test GetForUpdate - Error Cases

func (s *RepositoryTestSuite) TestGetForUpdateWhenDatabaseFailsReturnsError() {
	dbError := errRepoLockTimeout

	s.mockTx.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(123)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(dbError).
		Times(1)

	result, err := s.repo.GetForUpdate(s.ctx, s.mockTx, 123)
	s.NotNil(err)
	s.Nil(result)
	s.Equal(dbError, err)
}

// Test UpdateBalance - Success Cases

func (s *RepositoryTestSuite) TestUpdateBalanceSucceeds() {
	newBalance := decimal.NewFromFloat(750.50)

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), int64(123), newBalance, gomock.Any()).
		Return(pgconn.NewCommandTag("UPDATE 1"), nil).
		Times(1)

	err := s.repo.UpdateBalance(s.ctx, s.mockTx, 123, newBalance)
	s.Nil(err)
}

func (s *RepositoryTestSuite) TestUpdateBalanceWithZeroSucceeds() {
	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), int64(456), decimal.Zero, gomock.Any()).
		Return(pgconn.NewCommandTag("UPDATE 1"), nil).
		Times(1)

	err := s.repo.UpdateBalance(s.ctx, s.mockTx, 456, decimal.Zero)
	s.Nil(err)
}

func (s *RepositoryTestSuite) TestUpdateBalanceWithHighPrecision() {
	highPrecisionBalance, _ := decimal.NewFromString("123.45678901")

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), int64(789), highPrecisionBalance, gomock.Any()).
		Return(pgconn.NewCommandTag("UPDATE 1"), nil).
		Times(1)

	err := s.repo.UpdateBalance(s.ctx, s.mockTx, 789, highPrecisionBalance)
	s.Nil(err)
}

// Test UpdateBalance - Error Cases

func (s *RepositoryTestSuite) TestUpdateBalanceWhenExecFailsReturnsError() {
	newBalance := decimal.NewFromFloat(100.00)
	dbError := errRepoTxAborted

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), int64(123), newBalance, gomock.Any()).
		Return(pgconn.CommandTag{}, dbError).
		Times(1)

	err := s.repo.UpdateBalance(s.ctx, s.mockTx, 123, newBalance)
	s.NotNil(err)
	s.Equal(dbError, err)
}

// Test Exists - Success Cases

func (s *RepositoryTestSuite) TestExistsWhenAccountExistsReturnsTrue() {
	s.mockPool.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(123)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any()).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*bool) = true
			return nil
		}).
		Times(1)

	exists, err := s.repo.Exists(s.ctx, 123)
	s.Nil(err)
	s.True(exists)
}

func (s *RepositoryTestSuite) TestExistsWhenAccountDoesNotExistReturnsFalse() {
	s.mockPool.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(999)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any()).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*bool) = false
			return nil
		}).
		Times(1)

	exists, err := s.repo.Exists(s.ctx, 999)
	s.Nil(err)
	s.False(exists)
}

// Test Exists - Error Cases

func (s *RepositoryTestSuite) TestExistsWhenDatabaseFailsReturnsError() {
	dbError := errRepoQueryFailed

	s.mockPool.EXPECT().
		QueryRow(s.ctx, gomock.Any(), int64(123)).
		Return(s.mockRow).
		Times(1)

	s.mockRow.EXPECT().
		Scan(gomock.Any()).
		Return(dbError).
		Times(1)

	exists, err := s.repo.Exists(s.ctx, 123)
	s.NotNil(err)
	s.False(exists)
	s.Equal(dbError, err)
}
