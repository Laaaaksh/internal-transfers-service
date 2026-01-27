package entities

// TransferResponse represents the response for a successful transfer
type TransferResponse struct {
	TransactionID string `json:"transaction_id"`
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}
