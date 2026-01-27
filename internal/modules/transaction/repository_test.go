package transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internal-transfers-service/internal/modules/transaction"
	dbmock "github.com/internal-transfers-service/pkg/database/mock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// Test error constants - used for simulating database errors in repository tests
var (
	errRepoTxDBConnectionFailed = errors.New("database connection failed")
	errRepoTxForeignKey         = errors.New("violates foreign key constraint")
	errRepoTxAborted            = errors.New("current transaction is aborted")
	errRepoTxTooManyConns       = errors.New("too many connections")
	errRepoTxConnTimeout        = errors.New("connection timeout")
)

// RepositoryTestSuite contains tests for transaction Repository
type RepositoryTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockPool *dbmock.MockPool
	mockTx   *dbmock.MockTx
	repo     transaction.IRepository
	ctx      context.Context
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (s *RepositoryTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockPool = dbmock.NewMockPool(s.ctrl)
	s.mockTx = dbmock.NewMockTx(s.ctrl)
	s.ctx = context.Background()
	s.repo = transaction.NewRepository(s.mockPool)
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test Create - Success Cases

func (s *RepositoryTestSuite) TestCreateTransactionSucceeds() {
	tx := &transaction.Transaction{
		SourceAccountID:      123,
		DestinationAccountID: 456,
		Amount:               decimal.NewFromFloat(100.50),
	}

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), int64(123), int64(456), tx.Amount, gomock.Any()).
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil).
		Times(1)

	err := s.repo.Create(s.ctx, s.mockTx, tx)
	s.Nil(err)
	s.NotEqual(uuid.Nil, tx.ID)
	s.False(tx.CreatedAt.IsZero())
}

func (s *RepositoryTestSuite) TestCreateTransactionWithExistingIDSucceeds() {
	existingID := uuid.New()
	tx := &transaction.Transaction{
		ID:                   existingID,
		SourceAccountID:      123,
		DestinationAccountID: 456,
		Amount:               decimal.NewFromFloat(200.00),
	}

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), existingID, int64(123), int64(456), tx.Amount, gomock.Any()).
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil).
		Times(1)

	err := s.repo.Create(s.ctx, s.mockTx, tx)
	s.Nil(err)
	s.Equal(existingID, tx.ID)
}

func (s *RepositoryTestSuite) TestCreateTransactionGeneratesNewIDWhenNil() {
	tx := &transaction.Transaction{
		ID:                   uuid.Nil,
		SourceAccountID:      100,
		DestinationAccountID: 200,
		Amount:               decimal.NewFromFloat(50.00),
	}

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Not(uuid.Nil), int64(100), int64(200), tx.Amount, gomock.Any()).
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil).
		Times(1)

	err := s.repo.Create(s.ctx, s.mockTx, tx)
	s.Nil(err)
	s.NotEqual(uuid.Nil, tx.ID)
}

func (s *RepositoryTestSuite) TestCreateTransactionSetsCreatedAtTimestamp() {
	tx := &transaction.Transaction{
		SourceAccountID:      123,
		DestinationAccountID: 456,
		Amount:               decimal.NewFromFloat(75.25),
	}

	before := time.Now().UTC()

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil).
		Times(1)

	err := s.repo.Create(s.ctx, s.mockTx, tx)
	s.Nil(err)

	after := time.Now().UTC()

	s.True(tx.CreatedAt.After(before) || tx.CreatedAt.Equal(before))
	s.True(tx.CreatedAt.Before(after) || tx.CreatedAt.Equal(after))
}

func (s *RepositoryTestSuite) TestCreateTransactionWithHighPrecisionAmount() {
	highPrecisionAmount, _ := decimal.NewFromString("123.45678901")
	tx := &transaction.Transaction{
		SourceAccountID:      111,
		DestinationAccountID: 222,
		Amount:               highPrecisionAmount,
	}

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), int64(111), int64(222), highPrecisionAmount, gomock.Any()).
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil).
		Times(1)

	err := s.repo.Create(s.ctx, s.mockTx, tx)
	s.Nil(err)
}

func (s *RepositoryTestSuite) TestCreateTransactionWithSmallAmount() {
	smallAmount, _ := decimal.NewFromString("0.00000001")
	tx := &transaction.Transaction{
		SourceAccountID:      333,
		DestinationAccountID: 444,
		Amount:               smallAmount,
	}

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), int64(333), int64(444), smallAmount, gomock.Any()).
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil).
		Times(1)

	err := s.repo.Create(s.ctx, s.mockTx, tx)
	s.Nil(err)
}

// Test Create - Error Cases

func (s *RepositoryTestSuite) TestCreateTransactionWhenExecFailsReturnsError() {
	tx := &transaction.Transaction{
		SourceAccountID:      123,
		DestinationAccountID: 456,
		Amount:               decimal.NewFromFloat(100.00),
	}

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.CommandTag{}, errRepoTxDBConnectionFailed).
		Times(1)

	err := s.repo.Create(s.ctx, s.mockTx, tx)
	s.NotNil(err)
	s.Equal(errRepoTxDBConnectionFailed, err)
}

func (s *RepositoryTestSuite) TestCreateTransactionWhenConstraintViolationReturnsError() {
	tx := &transaction.Transaction{
		SourceAccountID:      123,
		DestinationAccountID: 456,
		Amount:               decimal.NewFromFloat(100.00),
	}

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.CommandTag{}, errRepoTxForeignKey).
		Times(1)

	err := s.repo.Create(s.ctx, s.mockTx, tx)
	s.NotNil(err)
	s.Contains(err.Error(), "foreign key")
}

func (s *RepositoryTestSuite) TestCreateTransactionWhenTransactionAbortedReturnsError() {
	tx := &transaction.Transaction{
		SourceAccountID:      123,
		DestinationAccountID: 456,
		Amount:               decimal.NewFromFloat(100.00),
	}

	s.mockTx.EXPECT().
		Exec(s.ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.CommandTag{}, errRepoTxAborted).
		Times(1)

	err := s.repo.Create(s.ctx, s.mockTx, tx)
	s.NotNil(err)
	s.Contains(err.Error(), "aborted")
}

// Test BeginTx - Success Cases

func (s *RepositoryTestSuite) TestBeginTxSucceeds() {
	s.mockPool.EXPECT().
		Begin(s.ctx).
		Return(s.mockTx, nil).
		Times(1)

	tx, err := s.repo.BeginTx(s.ctx)
	s.Nil(err)
	s.NotNil(tx)
}

// Test BeginTx - Error Cases

func (s *RepositoryTestSuite) TestBeginTxWhenPoolFailsReturnsError() {
	s.mockPool.EXPECT().
		Begin(s.ctx).
		Return(nil, errRepoTxTooManyConns).
		Times(1)

	tx, err := s.repo.BeginTx(s.ctx)
	s.NotNil(err)
	s.Nil(tx)
	s.Equal(errRepoTxTooManyConns, err)
}

func (s *RepositoryTestSuite) TestBeginTxWhenConnectionTimeoutReturnsError() {
	s.mockPool.EXPECT().
		Begin(s.ctx).
		Return(nil, errRepoTxConnTimeout).
		Times(1)

	tx, err := s.repo.BeginTx(s.ctx)
	s.NotNil(err)
	s.Nil(tx)
}
