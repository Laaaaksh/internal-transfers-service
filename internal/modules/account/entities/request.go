package entities

// CreateAccountRequest represents the request to create a new account
type CreateAccountRequest struct {
	AccountID      int64  `json:"account_id"`
	InitialBalance string `json:"initial_balance"`
}

// GetAccountRequest represents the request to get an account by ID
type GetAccountRequest struct {
	AccountID int64 `json:"account_id"`
}
