package account

//go:generate mockgen -source=core.go -destination=mock/mock_core.go -package=mock

import (
	"context"
	"errors"

	"github.com/internal-transfers-service/internal/constants"
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
		logger.Ctx(ctx).Debugw(constants.LogMsgInvalidAccountIDCreate,
			constants.LogKeyAccountID, req.AccountID,
		)
		return apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidAccountID, apperror.MsgInvalidAccountID).
			WithField(apperror.FieldAccountID, req.AccountID)
	}

	// Parse and validate initial balance
	balance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		logger.Ctx(ctx).Debugw(constants.LogMsgInvalidDecimalFormat,
			constants.LogFieldInitialBalance, req.InitialBalance,
			constants.LogKeyError, err,
		)
		return apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidDecimal, apperror.MsgInvalidDecimalFormat).
			WithField(constants.LogFieldInitialBalance, req.InitialBalance)
	}

	if balance.IsNegative() {
		logger.Ctx(ctx).Debugw(constants.LogMsgNegativeBalanceProvided,
			constants.LogFieldInitialBalance, req.InitialBalance,
		)
		return apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidBalance, apperror.MsgNegativeBalance).
			WithField(constants.LogFieldInitialBalance, req.InitialBalance)
	}

	// Check if account already exists
	exists, err := c.repo.Exists(ctx, req.AccountID)
	if err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToCheckAcctExist,
			constants.LogKeyAccountID, req.AccountID,
			constants.LogKeyError, err,
		)
		return apperror.New(apperror.CodeInternalError, err).
			WithField(apperror.FieldAccountID, req.AccountID)
	}

	if exists {
		return apperror.NewWithMessage(apperror.CodeConflict, ErrAccountExists, apperror.MsgDuplicateAccount).
			WithField(apperror.FieldAccountID, req.AccountID)
	}

	// Create the account
	account := &Account{
		AccountID: req.AccountID,
		Balance:   balance,
	}

	if err := c.repo.Create(ctx, account); err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToCreateAccount,
			constants.LogKeyAccountID, req.AccountID,
			constants.LogKeyError, err,
		)
		return apperror.New(apperror.CodeInternalError, err).
			WithField(apperror.FieldAccountID, req.AccountID)
	}

	logger.Ctx(ctx).Infow(constants.LogMsgAccountCreated,
		constants.LogKeyAccountID, req.AccountID,
		constants.LogFieldInitialBalance, balance.String(),
	)

	return nil
}

// GetByID retrieves an account by its ID
func (c *Core) GetByID(ctx context.Context, accountID int64) (*entities.AccountResponse, apperror.IError) {
	// Validate account ID
	if accountID <= 0 {
		logger.Ctx(ctx).Debugw(constants.LogMsgInvalidAccountIDGet,
			constants.LogKeyAccountID, accountID,
		)
		return nil, apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidAccountID, apperror.MsgInvalidAccountID).
			WithField(apperror.FieldAccountID, accountID)
	}

	account, err := c.repo.GetByID(ctx, accountID)
	if err != nil {
		// Check if it's already an apperror
		var appErr *apperror.Error
		if errors.As(err, &appErr) {
			if appErr.Code() == apperror.CodeNotFound {
				logger.Ctx(ctx).Debugw(constants.LogMsgAccountNotFoundDebug,
					constants.LogKeyAccountID, accountID,
				)
			}
			return nil, appErr
		}
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToGetAccount,
			constants.LogKeyAccountID, accountID,
			constants.LogKeyError, err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err).
			WithField(apperror.FieldAccountID, accountID)
	}

	return &entities.AccountResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance.String(),
	}, nil
}
