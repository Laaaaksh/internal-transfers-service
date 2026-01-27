package health

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/internal-transfers-service/internal/constants"
)

// HTTPHandler handles health check HTTP requests
type HTTPHandler struct {
	core ICore
}

// NewHTTPHandler creates a new health HTTP handler
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes registers health check routes
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(constants.RouteHealthLive, h.LivenessCheck)
	r.Get(constants.RouteHealthReady, h.ReadinessCheck)
}

// LivenessCheck handles GET /health/live
func (h *HTTPHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	status, code := h.core.RunLivenessCheck(r.Context())
	h.writeResponse(w, code, status)
}

// ReadinessCheck handles GET /health/ready
func (h *HTTPHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	status, code := h.core.RunReadinessCheck(r.Context())
	h.writeResponse(w, code, status)
}

// writeResponse writes the health check response
func (h *HTTPHandler) writeResponse(w http.ResponseWriter, statusCode int, status string) {
	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: status})
}
