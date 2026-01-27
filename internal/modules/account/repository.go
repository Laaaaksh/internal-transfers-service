package account

//go:generate mockgen -source=repository.go -destination=mock/mock_repository.go -package=mock

import (
	"context"
	"errors"
	"time"

	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/pkg/apperror"
	"github.com/internal-transfers-service/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Account represents the account domain model
type Account struct {
	AccountID int64           `json:"account_id"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// IRepository defines the interface for account data access
type IRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, accountID int64) (*Account, error)
	GetForUpdate(ctx context.Context, tx pgx.Tx, accountID int64) (*Account, error)
	UpdateBalance(ctx context.Context, tx pgx.Tx, accountID int64, newBalance decimal.Decimal) error
	Exists(ctx context.Context, accountID int64) (bool, error)
}

// Repository implements IRepository
type Repository struct {
	pool database.IPool
}

// Compile-time interface check
var _ IRepository = (*Repository)(nil)

// NewRepository creates a new account repository
func NewRepository(pool database.IPool) *Repository {
	return &Repository{pool: pool}
}

// SQL queries
const (
	queryInsertAccount = `
		INSERT INTO accounts (account_id, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`

	querySelectByID = `
		SELECT account_id, balance, created_at, updated_at
		FROM accounts
		WHERE account_id = $1`

	querySelectForUpdate = `
		SELECT account_id, balance, created_at, updated_at
		FROM accounts
		WHERE account_id = $1
		FOR UPDATE`

	queryUpdateBalance = `
		UPDATE accounts
		SET balance = $2, updated_at = $3
		WHERE account_id = $1`

	queryExists = `
		SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id = $1)`
)

// Create inserts a new account into the database
func (r *Repository) Create(ctx context.Context, account *Account) error {
	now := time.Now().UTC()
	account.CreatedAt = now
	account.UpdatedAt = now

	_, err := r.pool.Exec(ctx, queryInsertAccount,
		account.AccountID,
		account.Balance,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		logger.Ctx(ctx).Errorw("Failed to create account",
			"account_id", account.AccountID,
			"error", err,
		)
		return err
	}

	logger.Ctx(ctx).Infow("Account created",
		"account_id", account.AccountID,
		"initial_balance", account.Balance.String(),
	)
	return nil
}

// GetByID retrieves an account by its ID
func (r *Repository) GetByID(ctx context.Context, accountID int64) (*Account, error) {
	var account Account
	err := r.pool.QueryRow(ctx, querySelectByID, accountID).Scan(
		&account.AccountID,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New(apperror.CodeNotFound, err).
				WithField("account_id", accountID)
		}
		logger.Ctx(ctx).Errorw("Failed to get account",
			"account_id", accountID,
			"error", err,
		)
		return nil, err
	}

	return &account, nil
}

// GetForUpdate retrieves an account with a row-level lock for update
func (r *Repository) GetForUpdate(ctx context.Context, tx pgx.Tx, accountID int64) (*Account, error) {
	var account Account
	err := tx.QueryRow(ctx, querySelectForUpdate, accountID).Scan(
		&account.AccountID,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New(apperror.CodeNotFound, err).
				WithField("account_id", accountID)
		}
		logger.Ctx(ctx).Errorw("Failed to get account for update",
			"account_id", accountID,
			"error", err,
		)
		return nil, err
	}

	return &account, nil
}

// UpdateBalance updates the balance of an account within a transaction
func (r *Repository) UpdateBalance(ctx context.Context, tx pgx.Tx, accountID int64, newBalance decimal.Decimal) error {
	now := time.Now().UTC()
	_, err := tx.Exec(ctx, queryUpdateBalance, accountID, newBalance, now)
	if err != nil {
		logger.Ctx(ctx).Errorw("Failed to update account balance",
			"account_id", accountID,
			"new_balance", newBalance.String(),
			"error", err,
		)
		return err
	}
	return nil
}

// Exists checks if an account exists
func (r *Repository) Exists(ctx context.Context, accountID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, queryExists, accountID).Scan(&exists)
	if err != nil {
		logger.Ctx(ctx).Errorw("Failed to check account existence",
			"account_id", accountID,
			"error", err,
		)
		return false, err
	}
	return exists, nil
}
