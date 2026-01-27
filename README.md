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

Before you begin, ensure you have the following installed:

| Tool | Version | Installation |
|------|---------|--------------|
| Go | 1.21+ | [Download Go](https://golang.org/dl/) |
| Docker | Latest | [Install Docker](https://docs.docker.com/get-docker/) |
| Docker Compose | Latest | Included with Docker Desktop |
| Make | Any | Pre-installed on macOS/Linux, or use [GnuWin32](http://gnuwin32.sourceforge.net/packages/make.htm) on Windows |

### Verify Prerequisites

```bash
# Check Go version
go version
# Expected: go version go1.21+ ...

# Check Docker
docker --version
# Expected: Docker version 24.x.x or later

# Check Docker Compose
docker-compose --version
# Expected: docker-compose version 2.x.x or later

# Check Make
make --version
# Expected: GNU Make 3.x or later
```

## Quick Start (First-Time Setup)

For first-time setup, run the automated setup command:

```bash
make setup
```

This command will:
1. Install development tools (golangci-lint, mockgen, migrate, goimports)
2. Download Go dependencies
3. Start PostgreSQL via Docker
4. Run database migrations
5. Generate mock files for testing

After setup completes, start the service:

```bash
make run
```

The service will start on:
- **Main API**: http://localhost:8080
- **Ops (health/metrics)**: http://localhost:8081

## Manual Setup (Step-by-Step)

If you prefer to set up manually or need to troubleshoot:

### Step 1: Install Development Tools

```bash
# Install all required development tools
make deps-install

# This installs:
# - golangci-lint (linter)
# - mockgen (mock generation)
# - migrate (database migrations)
# - goimports (import formatting)
```

Or install individually:

```bash
# golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# mockgen
go install go.uber.org/mock/mockgen@latest

# golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# goimports
go install golang.org/x/tools/cmd/goimports@latest
```

**macOS Alternative (Homebrew):**
```bash
brew install golangci-lint golang-migrate
```

### Step 2: Download Dependencies

```bash
make deps
```

### Step 3: Start PostgreSQL

```bash
# Start PostgreSQL container
make docker-up

# Verify it's running
make docker-status
```

Expected output:
```
NAME                         STATUS
internal-transfers-postgres  Up X seconds
```

### Step 4: Run Database Migrations

```bash
make migrate-up
```

Expected output:
```
Running migrations...
1/u create_accounts (xx.xxxms)
2/u create_transactions (xx.xxxms)
3/u create_idempotency_keys (xx.xxxms)
Migrations complete
```

### Step 5: Generate Mocks (for testing)

```bash
make mock
```

### Step 6: Run Tests (verify everything works)

```bash
make test-short
```

All tests should pass:
```
ok  github.com/internal-transfers-service/internal/modules/account
ok  github.com/internal-transfers-service/internal/modules/health
ok  github.com/internal-transfers-service/internal/modules/transaction
ok  github.com/internal-transfers-service/pkg/apperror
```

### Step 7: Run the Service

```bash
make run
```

### Step 8: Test the API

```bash
# Create an account
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "1000.00"}'

# Get account balance
curl http://localhost:8080/accounts/1

# Check health
curl http://localhost:8081/health/live
curl http://localhost:8081/health/ready
```

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

Configuration is loaded from `config/default.toml` with environment-specific overrides. The environment is determined by the `APP_ENV` variable (defaults to `dev`).

```bash
# Set environment (dev, test, prod)
export APP_ENV=dev
```

Environment variables can override any setting using the `APP_` prefix:

```bash
APP_DATABASE_HOST=production-db.internal
APP_DATABASE_PASSWORD=secret
APP_LOGGING_LEVEL=warn
```

## Development Commands

### All Available Commands

Run `make help` to see all available commands:

```bash
make help
```

### Build & Run

| Command | Description |
|---------|-------------|
| `make build` | Build the application binary to `bin/` |
| `make run` | Run the application |
| `make run-dev` | Run with hot reload (requires `air`) |

### Testing

| Command | Description |
|---------|-------------|
| `make test` | Run all tests with verbose output |
| `make test-short` | Run tests without verbose (faster) |
| `make test-coverage` | Run tests and generate HTML coverage report |

### Code Quality

| Command | Description |
|---------|-------------|
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with gofmt and goimports |
| `make verify` | Verify go.mod dependencies |

### Docker

| Command | Description |
|---------|-------------|
| `make docker-up` | Start PostgreSQL container |
| `make docker-down` | Stop PostgreSQL container |
| `make docker-clean` | Stop and remove volumes |
| `make docker-status` | Show container status |
| `make docker-logs` | View PostgreSQL logs |

### Database Migrations

| Command | Description |
|---------|-------------|
| `make migrate-up` | Run all pending migrations |
| `make migrate-down` | Rollback all migrations |
| `make migrate-down-one` | Rollback one migration |
| `make migrate-create` | Create new migration files |
| `make migrate-version` | Show current migration version |

### Mock Generation

| Command | Description |
|---------|-------------|
| `make mock` | Clean and regenerate all mocks |
| `make mock-clean` | Remove generated mock files |

### Dependencies

| Command | Description |
|---------|-------------|
| `make deps` | Download and tidy go modules |
| `make deps-install` | Install development tools |

### Cleanup

| Command | Description |
|---------|-------------|
| `make clean` | Clean build artifacts |
| `make clean-all` | Clean all including mocks |

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
