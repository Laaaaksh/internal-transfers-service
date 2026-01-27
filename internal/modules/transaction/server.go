package transaction

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/transaction/entities"
	"github.com/internal-transfers-service/pkg/apperror"
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
	r.Post("/transactions", h.CreateTransaction)
}

// CreateTransaction handles POST /transactions
func (h *HTTPHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req entities.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, apperror.New(apperror.CodeBadRequest, err))
		return
	}

	response, appErr := h.core.Transfer(ctx, &req)
	if appErr != nil {
		h.writeError(w, appErr)
		return
	}

	logger.Ctx(ctx).Infow("Transaction created via HTTP",
		"transaction_id", response.TransactionID,
		"source_account_id", req.SourceAccountID,
		"destination_account_id", req.DestinationAccountID,
	)

	h.writeJSON(w, http.StatusCreated, response)
}

// writeJSON writes a JSON response
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Error("Failed to encode response", "error", err)
		}
	}
}

// writeError writes an error response
func (h *HTTPHandler) writeError(w http.ResponseWriter, err apperror.IError) {
	response := entities.ErrorResponse{
		Error:   err.PublicMessage(),
		Code:    err.Code().String(),
		Details: err.Fields(),
	}
	h.writeJSON(w, err.HTTPStatus(), response)
}
