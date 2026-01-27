# Development Guide

This guide covers the development workflow and available commands.

## Available Commands

Run `make help` to see all available commands.

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

---

## Project Structure

```
internal-transfers-service/
├── cmd/api/               # Application entry point
├── internal/
│   ├── boot/              # Application bootstrap
│   ├── config/            # Configuration management
│   ├── constants/         # Application constants
│   ├── database/          # Database migrations
│   ├── interceptors/      # HTTP middleware (rate limiting, tracing, etc.)
│   ├── logger/            # Structured logging (with trace/span IDs)
│   ├── metrics/           # Prometheus metrics
│   ├── tracing/           # OpenTelemetry distributed tracing
│   ├── modules/           # Business modules
│   │   ├── account/       # Account module
│   │   ├── health/        # Health check module
│   │   ├── idempotency/   # Idempotency key management
│   │   └── transaction/   # Transaction module
│   └── utils/             # Utility functions
├── pkg/
│   ├── apperror/          # Custom error handling
│   └── database/          # Database connection (with retry)
├── config/                # Configuration files (TOML)
├── deployment/            # Deployment configs
├── wiki/                  # Documentation
└── memory-bank/           # Project documentation
```

---

## Development Workflow

### 1. Start Development Environment

```bash
# Start database
make docker-up

# Run migrations
make migrate-up

# Generate mocks
make mock
```

### 2. Run the Service

```bash
# Standard run
make run

# Or with hot reload (requires air)
make run-dev
```

### 3. Run Tests

```bash
# Quick test run
make test-short

# Full test with verbose output
make test

# With coverage report
make test-coverage
```

### 4. Before Committing

```bash
# Format code
make fmt

# Run linter
make lint

# Run all tests
make test-short
```

---

## Adding New Features

### Adding a New Module

1. Create a new directory under `internal/modules/`:
   ```
   internal/modules/newmodule/
   ├── init.go         # Module initialization
   ├── core.go         # Business logic
   ├── repository.go   # Data access
   ├── server.go       # HTTP handlers
   ├── entities/
   │   ├── request.go
   │   └── response.go
   └── mock/
       └── mock_*.go
   ```

2. Define interfaces in `core.go` and `repository.go`

3. Add mock generation to Makefile:
   ```makefile
   @mockgen -source=internal/modules/newmodule/repository.go -destination=internal/modules/newmodule/mock/mock_repository.go -package=mock
   ```

4. Register routes in the module's `init.go`

### Adding Database Migrations

```bash
make migrate-create
# Enter migration name: add_new_table

# This creates:
# - internal/database/migrations/000004_add_new_table.up.sql
# - internal/database/migrations/000004_add_new_table.down.sql
```

Edit the generated files and run:
```bash
make migrate-up
```

---

## Testing Guidelines

### Test Structure

- Use `testify/suite` for all tests
- One test method per scenario (no table-driven tests)
- Use camelCase for test method names

```go
type CoreTestSuite struct {
    suite.Suite
    ctrl       *gomock.Controller
    mockRepo   *mock.MockIRepository
    core       *Core
}

func (s *CoreTestSuite) SetupTest() {
    s.ctrl = gomock.NewController(s.T())
    s.mockRepo = mock.NewMockIRepository(s.ctrl)
    s.core = NewCore(s.mockRepo)
}

func (s *CoreTestSuite) TearDownTest() {
    s.ctrl.Finish()
}

func (s *CoreTestSuite) TestGetAccountSuccess() {
    // Test implementation
}

func (s *CoreTestSuite) TestGetAccountNotFound() {
    // Test implementation
}
```

### Running Specific Tests

```bash
# Run tests for a specific package
go test -v ./internal/modules/account/...

# Run a specific test
go test -v -run TestCoreSuite/TestGetAccountSuccess ./internal/modules/account/
```

---

## Debugging

### View Service Logs

The service outputs structured JSON logs. For development, the format is set to `console` for readability.

### View PostgreSQL Logs

```bash
make docker-logs
```

### Check Database Connection

```bash
# Connect to PostgreSQL
docker exec -it transfers-postgres psql -U postgres -d transfers
```

### Check Application Health

```bash
# Liveness (is the process running?)
curl http://localhost:8081/health/live

# Readiness (is the database connected?)
curl http://localhost:8081/health/ready
```
