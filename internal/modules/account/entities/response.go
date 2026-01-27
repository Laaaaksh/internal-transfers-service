package entities

// AccountResponse represents the response for account operations
type AccountResponse struct {
	AccountID int64  `json:"account_id"`
	Balance   string `json:"balance"`
}

// NOTE: ErrorResponse has been consolidated to pkg/apperror/response.go
// Import github.com/internal-transfers-service/pkg/apperror for ErrorResponse
