package account

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/constants/contextkeys"
	"github.com/internal-transfers-service/internal/logger"
	"github.com/internal-transfers-service/internal/modules/account/entities"
	"github.com/internal-transfers-service/pkg/apperror"
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
	r.Post(entities.RouteAccounts, h.CreateAccount)
	r.Get(entities.RouteAccountByID, h.GetAccount)
}

// CreateAccount handles POST /accounts
func (h *HTTPHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req entities.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorWithContext(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, err, apperror.MsgInvalidJSONBody))
		return
	}

	if appErr := h.core.Create(ctx, &req); appErr != nil {
		h.writeErrorWithContext(w, r, appErr)
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

	accountIDStr := chi.URLParam(r, entities.ParamAccountID)
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		h.writeErrorWithContext(w, r, apperror.NewWithMessage(apperror.CodeBadRequest, ErrInvalidAccountID, apperror.MsgInvalidAccountID).
			WithField(apperror.FieldAccountID, accountIDStr))
		return
	}

	response, appErr := h.core.GetByID(ctx, accountID)
	if appErr != nil {
		h.writeErrorWithContext(w, r, appErr)
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
