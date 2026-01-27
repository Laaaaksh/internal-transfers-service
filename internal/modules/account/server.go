package account

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/account/entities"
	"github.com/internal-transfers-service/pkg/apperror"
)

// Route path constants
const (
	routeAccounts   = "/accounts"
	routeAccountByID = "/accounts/{accountID}"
	paramAccountID   = "accountID"
)

// HTTPHandler handles HTTP requests for account operations
type HTTPHandler struct {
	core ICore
}

// NewHTTPHandler creates a new HTTPHandler
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes registers the account routes with the router
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Post(routeAccounts, h.CreateAccount)
	r.Get(routeAccountByID, h.GetAccount)
}

// CreateAccount handles POST /accounts
func (h *HTTPHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req entities.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, apperror.New(apperror.CodeBadRequest, err))
		return
	}

	if appErr := h.core.Create(ctx, &req); appErr != nil {
		h.writeError(w, appErr)
		return
	}

	logger.Ctx(ctx).Infow(constants.LogMsgAccountCreatedViaHTTP,
		constants.LogKeyAccountID, req.AccountID,
	)

	w.WriteHeader(http.StatusCreated)
}

// GetAccount handles GET /accounts/{accountID}
func (h *HTTPHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accountIDStr := chi.URLParam(r, paramAccountID)
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		h.writeError(w, apperror.New(apperror.CodeBadRequest, ErrInvalidAccountID).
			WithField(apperror.FieldAccountID, accountIDStr))
		return
	}

	response, appErr := h.core.GetByID(ctx, accountID)
	if appErr != nil {
		h.writeError(w, appErr)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
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

// writeError writes an error response
func (h *HTTPHandler) writeError(w http.ResponseWriter, err apperror.IError) {
	response := entities.ErrorResponse{
		Error:   err.PublicMessage(),
		Code:    err.Code().String(),
		Details: err.Fields(),
	}
	h.writeJSON(w, err.HTTPStatus(), response)
}
