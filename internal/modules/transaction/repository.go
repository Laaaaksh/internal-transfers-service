package transaction

//go:generate mockgen -source=repository.go -destination=mock/mock_repository.go -package=mock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Transaction represents the transaction domain model
type Transaction struct {
	ID                   uuid.UUID       `json:"id"`
	SourceAccountID      int64           `json:"source_account_id"`
	DestinationAccountID int64           `json:"destination_account_id"`
	Amount               decimal.Decimal `json:"amount"`
	CreatedAt            time.Time       `json:"created_at"`
}

// IRepository defines the interface for transaction data access
type IRepository interface {
	Create(ctx context.Context, tx pgx.Tx, transaction *Transaction) error
	BeginTx(ctx context.Context) (pgx.Tx, error)
}

// Repository implements IRepository
type Repository struct {
	pool database.IPool
}

// Compile-time interface check
var _ IRepository = (*Repository)(nil)

// NewRepository creates a new transaction repository
func NewRepository(pool database.IPool) *Repository {
	return &Repository{pool: pool}
}

// SQL queries
const (
	queryInsertTransaction = `
		INSERT INTO transactions (id, source_account_id, destination_account_id, amount, created_at)
		VALUES ($1, $2, $3, $4, $5)`
)

// Create inserts a new transaction into the database
func (r *Repository) Create(ctx context.Context, tx pgx.Tx, transaction *Transaction) error {
	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New()
	}
	transaction.CreatedAt = time.Now().UTC()

	_, err := tx.Exec(ctx, queryInsertTransaction,
		transaction.ID,
		transaction.SourceAccountID,
		transaction.DestinationAccountID,
		transaction.Amount,
		transaction.CreatedAt,
	)

	if err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToCreateTx,
			constants.LogFieldTransactionID, transaction.ID.String(),
			constants.LogFieldSourceAccount, transaction.SourceAccountID,
			constants.LogFieldDestAccount, transaction.DestinationAccountID,
			constants.LogKeyAmount, transaction.Amount.String(),
			constants.LogKeyError, err,
		)
		return err
	}

	logger.Ctx(ctx).Infow(constants.LogMsgTransactionCreated,
		constants.LogFieldTransactionID, transaction.ID.String(),
		constants.LogFieldSourceAccount, transaction.SourceAccountID,
		constants.LogFieldDestAccount, transaction.DestinationAccountID,
		constants.LogKeyAmount, transaction.Amount.String(),
	)
	return nil
}

// BeginTx starts a new database transaction with explicit isolation level.
// Uses ReadCommitted isolation which is appropriate for financial transactions
// when combined with pessimistic locking (SELECT ... FOR UPDATE).
func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
}
