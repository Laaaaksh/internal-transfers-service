# Internal Transfers Service

A production-grade microservice for facilitating internal fund transfers between accounts.

## Features

- **Account Management**: Create accounts with initial balances and query account information
- **Fund Transfers**: Transfer funds between accounts with ACID guarantees
- **Decimal Precision**: All monetary values use 8 decimal places for financial accuracy
- **Idempotency Support**: Safe retries using X-Idempotency-Key header
- **Health Checks**: Kubernetes-ready liveness and readiness probes
- **Metrics**: Prometheus-compatible metrics endpoint
- **Structured Logging**: JSON-formatted logs with request tracing
- **Graceful Shutdown**: Proper connection draining and cleanup

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make
- golang-migrate (for migrations)

## Quick Start

### 1. Start PostgreSQL

```bash
make docker-up
```

### 2. Run Database Migrations

```bash
# Install golang-migrate if not already installed
# macOS: brew install golang-migrate
# Linux: See https://github.com/golang-migrate/migrate

make migrate-up
```

### 3. Run the Service

```bash
make run
```

The service will start on:
- **Main API**: http://localhost:8080
- **Ops (health/metrics)**: http://localhost:8081

## API Endpoints

### Create Account

```bash
POST /accounts
Content-Type: application/json

{
    "account_id": 123,
    "initial_balance": "100.23344"
}
```

**Response**: `201 Created` (empty body on success)

### Get Account

```bash
GET /accounts/{accountID}
```

**Response**:
```json
{
    "account_id": 123,
    "balance": "100.23344"
}
```

### Create Transaction

```bash
POST /transactions
Content-Type: application/json

{
    "source_account_id": 123,
    "destination_account_id": 456,
    "amount": "50.00000"
}
```

**Response**:
```json
{
    "transaction_id": "uuid-here"
}
```

### Health Endpoints

```bash
GET /health/live   # Liveness probe
GET /health/ready  # Readiness probe (includes DB check)
GET /metrics       # Prometheus metrics
```

## Configuration

Configuration is loaded from `config/config.toml`. Environment variables can override any setting using the `APP_` prefix:

```bash
APP_DATABASE_HOST=production-db.internal
APP_DATABASE_PASSWORD=secret
APP_LOGGING_LEVEL=warn
```

## Development

### Run Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make test-coverage
```

### Generate Mocks

```bash
make mock
```

### Format Code

```bash
make fmt
```

### Run Linter

```bash
make lint
```

## Project Structure

```
internal-transfers-service/
├── cmd/api/               # Application entry point
├── internal/
│   ├── boot/              # Application bootstrap
│   ├── config/            # Configuration management
│   ├── constants/         # Application constants
│   ├── database/          # Database migrations
│   ├── interceptors/      # HTTP middleware
│   ├── logger/            # Structured logging
│   ├── metrics/           # Prometheus metrics
│   ├── modules/           # Business modules
│   │   ├── account/       # Account module
│   │   ├── health/        # Health check module
│   │   └── transaction/   # Transaction module
│   └── utils/             # Utility functions
├── pkg/
│   ├── apperror/          # Custom error handling
│   └── database/          # Database connection
├── config/                # Configuration files
├── deployment/            # Deployment configs
└── memory-bank/           # Project documentation
```

## Architecture

The service follows a **Clean Architecture** pattern:

1. **Handler Layer** (`server.go`): HTTP request handling
2. **Core Layer** (`core.go`): Business logic
3. **Repository Layer** (`repository.go`): Data access

Key design patterns:
- **Dependency Injection**: All dependencies are injected via interfaces
- **Pessimistic Locking**: Uses `SELECT FOR UPDATE` for atomic transfers
- **Consistent Lock Ordering**: Prevents deadlocks by always locking accounts in ascending ID order

## Error Handling

All errors follow a consistent structure:

```json
{
    "error": "User-friendly message",
    "code": "ERROR_CODE",
    "details": {
        "account_id": 123
    }
}
```

## Assumptions

1. All accounts use the same currency
2. No authentication/authorization required
3. Account IDs are positive integers provided by the client
4. Balance cannot go negative (enforced by DB constraint and application logic)
5. Idempotency keys are optional but recommended for transaction requests

## License

MIT
