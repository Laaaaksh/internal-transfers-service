package account

//go:generate mockgen -source=core.go -destination=mock/mock_core.go -package=mock

import (
	"context"
	"errors"

	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/account/entities"
	"github.com/internal-transfers-service/pkg/apperror"
	"github.com/shopspring/decimal"
)

// Domain errors
var (
	ErrAccountNotFound  = errors.New(entities.ErrMsgAccountNotFound)
	ErrAccountExists    = errors.New(entities.ErrMsgAccountExists)
	ErrInvalidAccountID = errors.New(entities.ErrMsgInvalidAccountID)
	ErrInvalidBalance   = errors.New(entities.ErrMsgInvalidBalance)
	ErrInvalidDecimal   = errors.New(entities.ErrMsgInvalidDecimal)
)

// ICore defines the interface for account business logic
type ICore interface {
	Create(ctx context.Context, req *entities.CreateAccountRequest) apperror.IError
	GetByID(ctx context.Context, accountID int64) (*entities.AccountResponse, apperror.IError)
}

// Core implements ICore
type Core struct {
	repo IRepository
}

// Compile-time interface check
var _ ICore = (*Core)(nil)

// coreInstance is the singleton instance
var coreInstance ICore

// NewCore creates a new Core instance
func NewCore(_ context.Context, repo IRepository) ICore {
	coreInstance = &Core{
		repo: repo,
	}
	return coreInstance
}

// NewCoreWithRepo creates a new Core instance with the given repository (for testing)
func NewCoreWithRepo(_ context.Context, repo IRepository) ICore {
	return &Core{
		repo: repo,
	}
}

// GetCore returns the singleton Core instance
func GetCore() ICore {
	return coreInstance
}

// Create creates a new account with the given initial balance
func (c *Core) Create(ctx context.Context, req *entities.CreateAccountRequest) apperror.IError {
	// Validate account ID
	if req.AccountID <= 0 {
		return apperror.New(apperror.CodeBadRequest, ErrInvalidAccountID).
			WithField("account_id", req.AccountID)
	}

	// Parse and validate initial balance
	balance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		return apperror.New(apperror.CodeBadRequest, ErrInvalidDecimal).
			WithField("initial_balance", req.InitialBalance)
	}

	if balance.IsNegative() {
		return apperror.New(apperror.CodeBadRequest, ErrInvalidBalance).
			WithField("initial_balance", req.InitialBalance)
	}

	// Check if account already exists
	exists, err := c.repo.Exists(ctx, req.AccountID)
	if err != nil {
		logger.Ctx(ctx).Errorw("Failed to check account existence",
			"account_id", req.AccountID,
			"error", err,
		)
		return apperror.New(apperror.CodeInternalError, err).
			WithField("account_id", req.AccountID)
	}

	if exists {
		return apperror.NewWithMessage(apperror.CodeConflict, ErrAccountExists, apperror.MsgDuplicateAccount).
			WithField("account_id", req.AccountID)
	}

	// Create the account
	account := &Account{
		AccountID: req.AccountID,
		Balance:   balance,
	}

	if err := c.repo.Create(ctx, account); err != nil {
		logger.Ctx(ctx).Errorw("Failed to create account",
			"account_id", req.AccountID,
			"error", err,
		)
		return apperror.New(apperror.CodeInternalError, err).
			WithField("account_id", req.AccountID)
	}

	logger.Ctx(ctx).Infow("Account created successfully",
		"account_id", req.AccountID,
		"initial_balance", balance.String(),
	)

	return nil
}

// GetByID retrieves an account by its ID
func (c *Core) GetByID(ctx context.Context, accountID int64) (*entities.AccountResponse, apperror.IError) {
	// Validate account ID
	if accountID <= 0 {
		return nil, apperror.New(apperror.CodeBadRequest, ErrInvalidAccountID).
			WithField("account_id", accountID)
	}

	account, err := c.repo.GetByID(ctx, accountID)
	if err != nil {
		// Check if it's already an apperror
		var appErr *apperror.Error
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, apperror.New(apperror.CodeInternalError, err).
			WithField("account_id", accountID)
	}

	return &entities.AccountResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance.String(),
	}, nil
}
