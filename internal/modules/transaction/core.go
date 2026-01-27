package transaction

//go:generate mockgen -source=core.go -destination=mock/mock_core.go -package=mock

import (
	"context"
	"errors"

	"github.com/internal-transfers-service/internal/constants"
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
			WithField(apperror.FieldSourceAccount, req.SourceAccountID).
			WithField(apperror.FieldDestAccount, req.DestinationAccountID)
	}

	// Parse and validate amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidDecimalAmt, apperror.MsgInvalidAmount).
			WithField(apperror.FieldAmount, req.Amount)
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidAmount, apperror.MsgInvalidAmount).
			WithField(apperror.FieldAmount, req.Amount)
	}

	// Begin transaction
	tx, err := c.txRepo.BeginTx(ctx)
	if err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToBeginTx,
			constants.LogKeyError, err,
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
	firstAccountID, secondAccountID := orderAccountIDs(req.SourceAccountID, req.DestinationAccountID)

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
	sourceAccount, destAccount := assignSourceAndDest(firstAccountID, req.SourceAccountID, firstAccount, secondAccount)

	// Check sufficient balance
	if sourceAccount.Balance.LessThan(amount) {
		logger.Ctx(ctx).Warnw(constants.LogMsgInsufficientBalance,
			constants.LogKeySourceAccount, req.SourceAccountID,
			constants.LogFieldCurrentBalance, sourceAccount.Balance.String(),
			constants.LogFieldRequestedAmt, amount.String(),
		)
		return nil, apperror.NewWithMessage(apperror.CodeInsufficientFunds, ErrInsufficientBalance, apperror.MsgInsufficientBalance).
			WithField(apperror.FieldSourceAccount, req.SourceAccountID).
			WithField(constants.LogFieldCurrentBalance, sourceAccount.Balance.String()).
			WithField(constants.LogFieldRequestedAmt, amount.String())
	}

	// Calculate new balances
	newSourceBalance := sourceAccount.Balance.Sub(amount)
	newDestBalance := destAccount.Balance.Add(amount)

	// Update source account balance
	if err := c.accountRepo.UpdateBalance(ctx, tx, sourceAccount.AccountID, newSourceBalance); err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToUpdateSourceBal,
			constants.LogKeyAccountID, sourceAccount.AccountID,
			constants.LogKeyError, err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err)
	}

	// Update destination account balance
	if err := c.accountRepo.UpdateBalance(ctx, tx, destAccount.AccountID, newDestBalance); err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToUpdateDestBal,
			constants.LogKeyAccountID, destAccount.AccountID,
			constants.LogKeyError, err,
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
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToCreateTxRecord,
			constants.LogKeyError, err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToCommitTx,
			constants.LogKeyError, err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err)
	}
	committed = true

	logger.Ctx(ctx).Infow(constants.LogMsgTransferCompleted,
		constants.LogFieldTransactionID, txRecord.ID.String(),
		constants.LogKeySourceAccount, req.SourceAccountID,
		constants.LogKeyDestAccount, req.DestinationAccountID,
		constants.LogKeyAmount, amount.String(),
	)

	return &entities.TransferResponse{
		TransactionID: txRecord.ID.String(),
	}, nil
}

// orderAccountIDs returns account IDs in ascending order for consistent locking
func orderAccountIDs(sourceID, destID int64) (first, second int64) {
	if sourceID < destID {
		return sourceID, destID
	}
	return destID, sourceID
}

// assignSourceAndDest assigns accounts to source and destination based on the first account ID
func assignSourceAndDest(firstAccountID, sourceAccountID int64, firstAccount, secondAccount *account.Account) (*account.Account, *account.Account) {
	if firstAccountID == sourceAccountID {
		return firstAccount, secondAccount
	}
	return secondAccount, firstAccount
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
