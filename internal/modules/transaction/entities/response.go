package entities

// TransferResponse represents the response for a successful transfer
type TransferResponse struct {
	TransactionID string `json:"transaction_id"`
}

// NOTE: ErrorResponse has been consolidated to pkg/apperror/response.go
// Import github.com/internal-transfers-service/pkg/apperror for ErrorResponse
