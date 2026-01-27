package transaction

//go:generate mockgen -source=core.go -destination=mock/mock_core.go -package=mock

import (
	"context"
	"errors"

	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/account"
	"github.com/internal-transfers-service/internal/modules/transaction/entities"
	"github.com/internal-transfers-service/pkg/apperror"
	"github.com/shopspring/decimal"
)

// Domain errors
var (
	ErrInsufficientBalance = errors.New(entities.ErrMsgInsufficientBalance)
	ErrSameAccountTransfer = errors.New(entities.ErrMsgSameAccountTransfer)
	ErrInvalidAmount       = errors.New(entities.ErrMsgInvalidAmount)
	ErrInvalidDecimalAmt   = errors.New(entities.ErrMsgInvalidDecimalAmt)
	ErrSourceNotFound      = errors.New(entities.ErrMsgSourceNotFound)
	ErrDestNotFound        = errors.New(entities.ErrMsgDestNotFound)
)

// ICore defines the interface for transaction business logic
type ICore interface {
	Transfer(ctx context.Context, req *entities.TransferRequest) (*entities.TransferResponse, apperror.IError)
}

// Core implements ICore
type Core struct {
	txRepo      IRepository
	accountRepo account.IRepository
}

// Compile-time interface check
var _ ICore = (*Core)(nil)

// coreInstance is the singleton instance
var coreInstance ICore

// NewCore creates a new Core instance
func NewCore(_ context.Context, txRepo IRepository, accountRepo account.IRepository) ICore {
	coreInstance = &Core{
		txRepo:      txRepo,
		accountRepo: accountRepo,
	}
	return coreInstance
}

// NewCoreWithRepo creates a new Core instance with the given repositories (for testing)
func NewCoreWithRepo(_ context.Context, txRepo IRepository, accountRepo account.IRepository) ICore {
	return &Core{
		txRepo:      txRepo,
		accountRepo: accountRepo,
	}
}

// GetCore returns the singleton Core instance
func GetCore() ICore {
	return coreInstance
}

// Transfer executes a fund transfer between two accounts
func (c *Core) Transfer(ctx context.Context, req *entities.TransferRequest) (*entities.TransferResponse, apperror.IError) {
	// Validate same account transfer
	if req.SourceAccountID == req.DestinationAccountID {
		return nil, apperror.NewWithMessage(apperror.CodeBadRequest, ErrSameAccountTransfer, apperror.MsgSameAccountTransfer).
			WithField("source_account_id", req.SourceAccountID).
			WithField("destination_account_id", req.DestinationAccountID)
	}

	// Parse and validate amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidDecimalAmt, apperror.MsgInvalidAmount).
			WithField("amount", req.Amount)
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidAmount, apperror.MsgInvalidAmount).
			WithField("amount", req.Amount)
	}

	// Begin transaction
	tx, err := c.txRepo.BeginTx(ctx)
	if err != nil {
		logger.Ctx(ctx).Errorw("Failed to begin transaction",
			"error", err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err)
	}

	// Use committed flag to ensure rollback on any error path
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock accounts in consistent order to prevent deadlocks
	// Always lock the account with the smaller ID first
	var firstAccountID, secondAccountID int64
	if req.SourceAccountID < req.DestinationAccountID {
		firstAccountID = req.SourceAccountID
		secondAccountID = req.DestinationAccountID
	} else {
		firstAccountID = req.DestinationAccountID
		secondAccountID = req.SourceAccountID
	}

	// Lock first account
	firstAccount, err := c.accountRepo.GetForUpdate(ctx, tx, firstAccountID)
	if err != nil {
		return nil, c.handleAccountError(err, firstAccountID, req.SourceAccountID)
	}

	// Lock second account
	secondAccount, err := c.accountRepo.GetForUpdate(ctx, tx, secondAccountID)
	if err != nil {
		return nil, c.handleAccountError(err, secondAccountID, req.SourceAccountID)
	}

	// Determine which is source and which is destination
	var sourceAccount, destAccount *account.Account
	if firstAccountID == req.SourceAccountID {
		sourceAccount = firstAccount
		destAccount = secondAccount
	} else {
		sourceAccount = secondAccount
		destAccount = firstAccount
	}

	// Check sufficient balance
	if sourceAccount.Balance.LessThan(amount) {
		logger.Ctx(ctx).Warnw("Insufficient balance for transfer",
			"source_account_id", req.SourceAccountID,
			"current_balance", sourceAccount.Balance.String(),
			"requested_amount", amount.String(),
		)
		return nil, apperror.NewWithMessage(apperror.CodeInsufficientFunds, ErrInsufficientBalance, apperror.MsgInsufficientBalance).
			WithField("source_account_id", req.SourceAccountID).
			WithField("current_balance", sourceAccount.Balance.String()).
			WithField("requested_amount", amount.String())
	}

	// Calculate new balances
	newSourceBalance := sourceAccount.Balance.Sub(amount)
	newDestBalance := destAccount.Balance.Add(amount)

	// Update source account balance
	if err := c.accountRepo.UpdateBalance(ctx, tx, sourceAccount.AccountID, newSourceBalance); err != nil {
		logger.Ctx(ctx).Errorw("Failed to update source account balance",
			"account_id", sourceAccount.AccountID,
			"error", err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err)
	}

	// Update destination account balance
	if err := c.accountRepo.UpdateBalance(ctx, tx, destAccount.AccountID, newDestBalance); err != nil {
		logger.Ctx(ctx).Errorw("Failed to update destination account balance",
			"account_id", destAccount.AccountID,
			"error", err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err)
	}

	// Create transaction record
	txRecord := &Transaction{
		SourceAccountID:      req.SourceAccountID,
		DestinationAccountID: req.DestinationAccountID,
		Amount:               amount,
	}

	if err := c.txRepo.Create(ctx, tx, txRecord); err != nil {
		logger.Ctx(ctx).Errorw("Failed to create transaction record",
			"error", err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		logger.Ctx(ctx).Errorw("Failed to commit transaction",
			"error", err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err)
	}
	committed = true

	logger.Ctx(ctx).Infow("Transfer completed successfully",
		"transaction_id", txRecord.ID.String(),
		"source_account_id", req.SourceAccountID,
		"destination_account_id", req.DestinationAccountID,
		"amount", amount.String(),
	)

	return &entities.TransferResponse{
		TransactionID: txRecord.ID.String(),
	}, nil
}

// handleAccountError converts account errors to appropriate API errors
func (c *Core) handleAccountError(err error, accountID int64, sourceAccountID int64) apperror.IError {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		if appErr.Code() == apperror.CodeNotFound {
			if accountID == sourceAccountID {
				return apperror.NewWithMessage(apperror.CodeNotFound, ErrSourceNotFound, apperror.MsgSourceNotFound).
					WithField(apperror.FieldSourceAccount, accountID)
			}
			return apperror.NewWithMessage(apperror.CodeNotFound, ErrDestNotFound, apperror.MsgDestNotFound).
				WithField(apperror.FieldDestAccount, accountID)
		}
		return appErr
	}
	return apperror.New(apperror.CodeInternalError, err)
}
