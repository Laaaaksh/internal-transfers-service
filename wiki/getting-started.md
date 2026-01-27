# Getting Started

This guide will help you set up the Internal Transfers Service for local development.

## Prerequisites

Before you begin, ensure you have the following installed:

| Tool | Version | Installation |
|------|---------|--------------|
| Go | 1.21+ | [Download Go](https://golang.org/dl/) |
| Docker | Latest | [Install Docker](https://docs.docker.com/get-docker/) |
| Docker Compose | Latest | Included with Docker Desktop |
| Make | Any | Pre-installed on macOS/Linux |

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

## Quick Start (Recommended)

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
NAME                 STATUS
transfers-postgres   Up X seconds (healthy)
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

### Step 6: Run Tests

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

## Next Steps

- [API Reference](api-reference.md) - Learn about all available endpoints
- [Development Guide](development.md) - Learn about development commands
- [Database Guide](database.md) - Learn how to inspect the database
