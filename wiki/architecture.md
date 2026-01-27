# Architecture Guide

This document describes the system architecture, design patterns, and key technical decisions.

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Internal Transfers Service                       │
│                                                                      │
│  ┌──────────────────┐    ┌──────────────────┐                       │
│  │   HTTP Server    │    │   Ops Server     │                       │
│  │   (Port 8080)    │    │   (Port 8081)    │                       │
│  │                  │    │                  │                       │
│  │  /accounts       │    │  /health/live    │                       │
│  │  /transactions   │    │  /health/ready   │                       │
│  │                  │    │  /metrics        │                       │
│  └────────┬─────────┘    └──────────────────┘                       │
│           │                                                          │
│  ┌────────▼─────────────────────────────────────────────────────┐   │
│  │                      Middleware Layer                         │   │
│  │  RequestID │ Logger │ Metrics │ Recovery │ Idempotency       │   │
│  └────────┬──────────────────────────────────────────────────────┘   │
│           │                                                          │
│  ┌────────▼─────────────────────────────────────────────────────┐   │
│  │                      Handler Layer                            │   │
│  │           AccountHandler │ TransactionHandler                 │   │
│  └────────┬──────────────────────────────────────────────────────┘   │
│           │                                                          │
│  ┌────────▼─────────────────────────────────────────────────────┐   │
│  │                      Core Layer                               │   │
│  │           AccountCore │ TransactionCore                       │   │
│  └────────┬──────────────────────────────────────────────────────┘   │
│           │                                                          │
│  ┌────────▼─────────────────────────────────────────────────────┐   │
│  │                     Repository Layer                          │   │
│  │    AccountRepo │ TransactionRepo │ IdempotencyRepo           │   │
│  └────────┬──────────────────────────────────────────────────────┘   │
│           │                                                          │
└───────────┼──────────────────────────────────────────────────────────┘
            │
┌───────────▼──────────────────────────────────────────────────────────┐
│                         PostgreSQL                                    │
│         accounts │ transactions │ idempotency_keys                   │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Layered Architecture

The service follows **Clean Architecture** principles with clear separation of concerns:

### 1. Handler Layer (`server.go`)

**Responsibility**: HTTP request/response handling

- Parse HTTP requests
- Validate request format
- Call core layer
- Format HTTP responses
- Handle HTTP-specific errors

```go
func (s *Server) CreateAccount(w http.ResponseWriter, r *http.Request) {
    var req entities.CreateAccountRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // Handle error
    }
    
    if err := s.core.Create(r.Context(), &req); err != nil {
        // Handle error
    }
    
    w.WriteHeader(http.StatusCreated)
}
```

### 2. Core Layer (`core.go`)

**Responsibility**: Business logic

- Implement business rules
- Orchestrate operations
- Manage transactions
- Enforce constraints

```go
func (c *Core) Transfer(ctx context.Context, req *TransferRequest) (*Response, error) {
    amount, err := c.validateTransferRequest(req)
    tx, err := c.beginTransaction(ctx)
    defer c.rollbackIfNotCommitted(ctx, tx, &committed)
    
    source, dest, err := c.lockAccountsInOrder(ctx, tx, req)
    err = c.validateSufficientBalance(ctx, source, amount, req.SourceAccountID)
    txRecord, err := c.executeTransfer(ctx, tx, source, dest, amount, req)
    
    return &Response{TransactionID: txRecord.ID.String()}, nil
}
```

### 3. Repository Layer (`repository.go`)

**Responsibility**: Data access

- Database queries
- SQL execution
- Data mapping

```go
func (r *Repository) GetByID(ctx context.Context, id int64) (*models.Account, error) {
    row := r.pool.QueryRow(ctx, selectAccountByIDQuery, id)
    var account models.Account
    err := row.Scan(&account.AccountID, &account.Balance, &account.CreatedAt, &account.UpdatedAt)
    return &account, nil
}
```

---

## Key Design Patterns

### 1. Repository Pattern

Abstracts data access from business logic, enabling:
- Easy testing with mock repositories
- Database technology independence
- Clear separation of concerns

```go
type IRepository interface {
    Create(ctx context.Context, account *models.Account) error
    GetByID(ctx context.Context, id int64) (*models.Account, error)
    GetForUpdate(ctx context.Context, tx pgx.Tx, id int64) (*models.Account, error)
    UpdateBalance(ctx context.Context, tx pgx.Tx, id int64, newBalance decimal.Decimal) error
}
```

### 2. Dependency Injection

All dependencies are injected via interfaces:
- Enables unit testing with mocks
- Loose coupling between components
- Easy to swap implementations

```go
type Core struct {
    repo IRepository
    pool IPool
}

func NewCore(repo IRepository, pool IPool) *Core {
    return &Core{repo: repo, pool: pool}
}
```

### 3. Pessimistic Locking

Prevents race conditions in concurrent transfers:

```go
func (r *Repository) GetForUpdate(ctx context.Context, tx pgx.Tx, id int64) (*models.Account, error) {
    row := tx.QueryRow(ctx, selectAccountForUpdateQuery, id)
    // Uses: SELECT ... FOR UPDATE
}
```

### 4. Consistent Lock Ordering

Prevents deadlocks by always locking accounts in the same order:

```go
func (c *Core) lockAccountsInOrder(ctx context.Context, tx pgx.Tx, req *TransferRequest) (*Account, *Account, error) {
    firstID, secondID := orderAccountIDs(req.SourceAccountID, req.DestinationAccountID)
    
    first, err := c.repo.GetForUpdate(ctx, tx, firstID)
    second, err := c.repo.GetForUpdate(ctx, tx, secondID)
    
    // Return in correct order
    if req.SourceAccountID < req.DestinationAccountID {
        return first, second, nil
    }
    return second, first, nil
}

func orderAccountIDs(a, b int64) (int64, int64) {
    if a < b {
        return a, b
    }
    return b, a
}
```

### 5. Unit of Work (Database Transactions)

Ensures atomic multi-table operations:

```go
func (c *Core) Transfer(ctx context.Context, req *TransferRequest) (*Response, error) {
    tx, err := c.beginTransaction(ctx)
    committed := false
    defer c.rollbackIfNotCommitted(ctx, tx, &committed)
    
    // All operations within the same transaction
    err = c.updateSourceBalance(ctx, tx, source, amount)
    err = c.updateDestBalance(ctx, tx, dest, amount)
    err = c.createTransactionRecord(ctx, tx, req, amount)
    
    err = c.commitTransaction(ctx, tx)
    committed = true
    
    return response, nil
}
```

### 6. Small Function Decomposition

Functions are kept small (20-30 lines) and focused:

```go
// Main function reads like documentation
func (c *Core) Transfer(ctx context.Context, req *TransferRequest) (*Response, error) {
    amount, err := c.validateTransferRequest(req)
    tx, err := c.beginTransaction(ctx)
    source, dest, err := c.lockAccountsInOrder(ctx, tx, req)
    err = c.validateSufficientBalance(ctx, source, amount, req.SourceAccountID)
    txRecord, err := c.executeTransfer(ctx, tx, source, dest, amount, req)
    return &Response{TransactionID: txRecord.ID.String()}, nil
}

// Each helper function does ONE thing
func (c *Core) validateTransferRequest(req *TransferRequest) (decimal.Decimal, error) { ... }
func (c *Core) beginTransaction(ctx context.Context) (pgx.Tx, error) { ... }
func (c *Core) lockAccountsInOrder(ctx, tx, req) (*Account, *Account, error) { ... }
func (c *Core) validateSufficientBalance(ctx, source, amount, sourceID) error { ... }
func (c *Core) executeTransfer(ctx, tx, source, dest, amount, req) (*Transaction, error) { ... }
```

---

## Error Handling

Consistent error handling using custom error types:

```go
type AppError struct {
    Code       string                 // Machine-readable code
    Message    string                 // Public message
    Internal   error                  // Internal error (not exposed)
    HTTPStatus int                    // HTTP status code
    Fields     map[string]interface{} // Context
}

// Predefined errors
var (
    ErrAccountNotFound     = NewError(CodeNotFound, "Account not found")
    ErrInsufficientBalance = NewError(CodeInvalidRequest, "Insufficient balance")
    ErrDuplicateAccount    = NewError(CodeConflict, "Account already exists")
)
```

---

## Observability

### Structured Logging

```go
logger.Ctx(ctx).Infow("Transfer completed",
    "source_account_id", req.SourceAccountID,
    "destination_account_id", req.DestinationAccountID,
    "amount", amount.String(),
    "transaction_id", txRecord.ID.String(),
)
```

### Prometheus Metrics

- `http_request_duration_seconds` - Request latency histogram
- `http_requests_total` - Request counter by method/path/status
- `transfers_total` - Transfer counter by status
- `db_connections_active` - Active database connections

### Health Checks

- **Liveness**: `/health/live` - Is the process running?
- **Readiness**: `/health/ready` - Is the database connected?

---

## Graceful Shutdown

```go
func main() {
    // Start servers...
    
    // Wait for shutdown signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
    <-sigChan
    
    // Mark unhealthy immediately
    healthChecker.MarkUnhealthy()
    
    // Wait for load balancer to drain
    time.Sleep(shutdownDelay)
    
    // Graceful shutdown with timeout
    shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
    defer cancel()
    
    httpServer.Shutdown(shutdownCtx)
    pool.Close()
}
```

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Locking Strategy | Pessimistic (SELECT FOR UPDATE) | Industry standard for financial systems |
| Decimal Handling | DECIMAL(19,8) + shopspring/decimal | Avoids floating-point errors |
| Idempotency | X-Idempotency-Key header | Safe retries for financial operations |
| Observability | Prometheus + Structured JSON | Cost-optimized, production-ready |
| HTTP Framework | Chi router | Lightweight, production-proven |
| Database Driver | pgx v5 | Best Go PostgreSQL driver |
| Configuration | TOML + Viper | Simple, environment overrides |
| Testing | testify/suite + gomock | Comprehensive, maintainable tests |
