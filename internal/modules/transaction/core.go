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
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Domain errors
var (
	ErrInsufficientBalance  = errors.New(entities.ErrMsgInsufficientBalance)
	ErrSameAccountTransfer  = errors.New(entities.ErrMsgSameAccountTransfer)
	ErrInvalidAmount        = errors.New(entities.ErrMsgInvalidAmount)
	ErrInvalidDecimalAmt    = errors.New(entities.ErrMsgInvalidDecimalAmt)
	ErrSourceNotFound       = errors.New(entities.ErrMsgSourceNotFound)
	ErrDestNotFound         = errors.New(entities.ErrMsgDestNotFound)
	ErrTooManyDecimalPlaces = errors.New(entities.ErrMsgTooManyDecimalPlaces)
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
	amount, appErr := c.validateTransferRequest(req)
	if appErr != nil {
		return nil, appErr
	}

	tx, err := c.beginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	committed := false
	defer c.rollbackIfNotCommitted(ctx, tx, &committed)

	sourceAccount, destAccount, appErr := c.lockAccountsInOrder(ctx, tx, req)
	if appErr != nil {
		return nil, appErr
	}

	if appErr := c.validateSufficientBalance(ctx, sourceAccount, amount, req.SourceAccountID); appErr != nil {
		return nil, appErr
	}

	txRecord, appErr := c.executeTransfer(ctx, tx, sourceAccount, destAccount, amount, req)
	if appErr != nil {
		return nil, appErr
	}

	if appErr := c.commitTransaction(ctx, tx); appErr != nil {
		return nil, appErr
	}
	committed = true

	c.logTransferCompleted(ctx, txRecord, req, amount)

	return &entities.TransferResponse{
		TransactionID: txRecord.ID.String(),
	}, nil
}

// validateTransferRequest validates the transfer request and returns the parsed amount
func (c *Core) validateTransferRequest(req *entities.TransferRequest) (decimal.Decimal, apperror.IError) {
	if req.SourceAccountID == req.DestinationAccountID {
		return decimal.Zero, apperror.NewWithMessage(apperror.CodeBadRequest, ErrSameAccountTransfer, apperror.MsgSameAccountTransfer).
			WithField(apperror.FieldSourceAccount, req.SourceAccountID).
			WithField(apperror.FieldDestAccount, req.DestinationAccountID)
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return decimal.Zero, apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidDecimalAmt, apperror.MsgInvalidAmount).
			WithField(apperror.FieldAmount, req.Amount)
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidAmount, apperror.MsgInvalidAmount).
			WithField(apperror.FieldAmount, req.Amount)
	}

	// Validate decimal precision (max 8 places to match DB schema)
	if appErr := validateDecimalPrecision(amount); appErr != nil {
		return decimal.Zero, appErr
	}

	return amount, nil
}

// beginTransaction starts a new database transaction
func (c *Core) beginTransaction(ctx context.Context) (pgx.Tx, apperror.IError) {
	tx, err := c.txRepo.BeginTx(ctx)
	if err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToBeginTx,
			constants.LogKeyError, err,
		)
		return nil, apperror.New(apperror.CodeInternalError, err)
	}
	return tx, nil
}

// rollbackIfNotCommitted rolls back the transaction if not committed
func (c *Core) rollbackIfNotCommitted(ctx context.Context, tx pgx.Tx, committed *bool) {
	if !*committed {
		_ = tx.Rollback(ctx)
	}
}

// lockAccountsInOrder locks accounts in consistent order to prevent deadlocks
func (c *Core) lockAccountsInOrder(ctx context.Context, tx pgx.Tx, req *entities.TransferRequest) (*account.Account, *account.Account, apperror.IError) {
	firstAccountID, secondAccountID := orderAccountIDs(req.SourceAccountID, req.DestinationAccountID)

	firstAccount, err := c.accountRepo.GetForUpdate(ctx, tx, firstAccountID)
	if err != nil {
		return nil, nil, c.handleAccountError(err, firstAccountID, req.SourceAccountID)
	}

	secondAccount, err := c.accountRepo.GetForUpdate(ctx, tx, secondAccountID)
	if err != nil {
		return nil, nil, c.handleAccountError(err, secondAccountID, req.SourceAccountID)
	}

	sourceAccount, destAccount := assignSourceAndDest(firstAccountID, req.SourceAccountID, firstAccount, secondAccount)
	return sourceAccount, destAccount, nil
}

// validateSufficientBalance checks if source account has sufficient balance
func (c *Core) validateSufficientBalance(ctx context.Context, sourceAccount *account.Account, amount decimal.Decimal, sourceAccountID int64) apperror.IError {
	if sourceAccount.Balance.LessThan(amount) {
		logger.Ctx(ctx).Warnw(constants.LogMsgInsufficientBalance,
			constants.LogKeySourceAccount, sourceAccountID,
			constants.LogFieldCurrentBalance, sourceAccount.Balance.String(),
			constants.LogFieldRequestedAmt, amount.String(),
		)
		return apperror.NewWithMessage(apperror.CodeInsufficientFunds, ErrInsufficientBalance, apperror.MsgInsufficientBalance).
			WithField(apperror.FieldSourceAccount, sourceAccountID).
			WithField(constants.LogFieldCurrentBalance, sourceAccount.Balance.String()).
			WithField(constants.LogFieldRequestedAmt, amount.String())
	}
	return nil
}

// executeTransfer updates balances and creates the transaction record
func (c *Core) executeTransfer(ctx context.Context, tx pgx.Tx, sourceAccount, destAccount *account.Account, amount decimal.Decimal, req *entities.TransferRequest) (*Transaction, apperror.IError) {
	if appErr := c.updateSourceBalance(ctx, tx, sourceAccount, amount); appErr != nil {
		return nil, appErr
	}

	if appErr := c.updateDestBalance(ctx, tx, destAccount, amount); appErr != nil {
		return nil, appErr
	}

	txRecord, appErr := c.createTransactionRecord(ctx, tx, req, amount)
	if appErr != nil {
		return nil, appErr
	}

	return txRecord, nil
}

// updateSourceBalance debits the source account
func (c *Core) updateSourceBalance(ctx context.Context, tx pgx.Tx, sourceAccount *account.Account, amount decimal.Decimal) apperror.IError {
	newBalance := sourceAccount.Balance.Sub(amount)
	if err := c.accountRepo.UpdateBalance(ctx, tx, sourceAccount.AccountID, newBalance); err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToUpdateSourceBal,
			constants.LogKeyAccountID, sourceAccount.AccountID,
			constants.LogKeyError, err,
		)
		return apperror.New(apperror.CodeInternalError, err)
	}
	return nil
}

// updateDestBalance credits the destination account
func (c *Core) updateDestBalance(ctx context.Context, tx pgx.Tx, destAccount *account.Account, amount decimal.Decimal) apperror.IError {
	newBalance := destAccount.Balance.Add(amount)
	if err := c.accountRepo.UpdateBalance(ctx, tx, destAccount.AccountID, newBalance); err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToUpdateDestBal,
			constants.LogKeyAccountID, destAccount.AccountID,
			constants.LogKeyError, err,
		)
		return apperror.New(apperror.CodeInternalError, err)
	}
	return nil
}

// createTransactionRecord creates the transaction audit record
func (c *Core) createTransactionRecord(ctx context.Context, tx pgx.Tx, req *entities.TransferRequest, amount decimal.Decimal) (*Transaction, apperror.IError) {
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

	return txRecord, nil
}

// commitTransaction commits the database transaction
func (c *Core) commitTransaction(ctx context.Context, tx pgx.Tx) apperror.IError {
	if err := tx.Commit(ctx); err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgFailedToCommitTx,
			constants.LogKeyError, err,
		)
		return apperror.New(apperror.CodeInternalError, err)
	}
	return nil
}

// logTransferCompleted logs successful transfer completion
func (c *Core) logTransferCompleted(ctx context.Context, txRecord *Transaction, req *entities.TransferRequest, amount decimal.Decimal) {
	logger.Ctx(ctx).Infow(constants.LogMsgTransferCompleted,
		constants.LogFieldTransactionID, txRecord.ID.String(),
		constants.LogKeySourceAccount, req.SourceAccountID,
		constants.LogKeyDestAccount, req.DestinationAccountID,
		constants.LogKeyAmount, amount.String(),
	)
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

// validateDecimalPrecision checks if the value exceeds the maximum allowed decimal places.
// The database uses DECIMAL(19,8) so we limit to 8 decimal places.
func validateDecimalPrecision(value decimal.Decimal) apperror.IError {
	exp := value.Exponent()
	if exp >= -constants.MaxDecimalPlaces {
		return nil
	}

	actualPlaces := -exp
	return apperror.NewWithMessage(apperror.CodeBadRequest, ErrTooManyDecimalPlaces, apperror.MsgTooManyDecimalPlaces).
		WithField(apperror.FieldAmount, value.String()).
		WithField(apperror.FieldDecimalPlaces, actualPlaces).
		WithField(apperror.FieldMaxAllowed, constants.MaxDecimalPlaces)
}
