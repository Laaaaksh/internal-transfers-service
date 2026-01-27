package transaction

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/constants/contextkeys"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/transaction/entities"
	"github.com/internal-transfers-service/pkg/apperror"
)

// Route path constants
const (
	routeTransactions = "/transactions"
)

// HTTPHandler handles HTTP requests for transaction operations
type HTTPHandler struct {
	core ICore
}

// NewHTTPHandler creates a new HTTPHandler
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes registers the transaction routes with the router
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Post(routeTransactions, h.CreateTransaction)
}

// CreateTransaction handles POST /transactions
func (h *HTTPHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req entities.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorWithContext(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, err, apperror.MsgInvalidJSONBody))
		return
	}

	response, appErr := h.core.Transfer(ctx, &req)
	if appErr != nil {
		h.writeErrorWithContext(w, r, appErr)
		return
	}

	logger.Ctx(ctx).Infow(constants.LogMsgTransactionCreatedHTTP,
		constants.LogFieldTransactionID, response.TransactionID,
		constants.LogKeySourceAccount, req.SourceAccountID,
		constants.LogKeyDestAccount, req.DestinationAccountID,
	)

	h.writeJSON(w, http.StatusCreated, response)
}

// writeJSON writes a JSON response
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Error(constants.LogMsgFailedToEncodeResponse, constants.LogKeyError, err)
		}
	}
}

// writeErrorWithContext writes an error response with request ID for tracing
func (h *HTTPHandler) writeErrorWithContext(w http.ResponseWriter, r *http.Request, err apperror.IError) {
	requestID := ""
	if id, ok := r.Context().Value(contextkeys.RequestID).(string); ok {
		requestID = id
	}

	response := entities.ErrorResponse{
		Error:     err.PublicMessage(),
		Code:      err.Code().String(),
		RequestID: requestID,
		Details:   err.Fields(),
	}

	// Log error for debugging
	logger.Ctx(r.Context()).Errorw(constants.LogMsgRequestFailed,
		constants.LogKeyError, err.Error(),
		constants.LogKeyStatusCode, err.HTTPStatus(),
	)

	h.writeJSON(w, err.HTTPStatus(), response)
}

// writeError writes an error response (deprecated, use writeErrorWithContext)
func (h *HTTPHandler) writeError(w http.ResponseWriter, err apperror.IError) {
	response := entities.ErrorResponse{
		Error:   err.PublicMessage(),
		Code:    err.Code().String(),
		Details: err.Fields(),
	}
	h.writeJSON(w, err.HTTPStatus(), response)
}
