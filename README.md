# Internal Transfers Service

A production-grade microservice for facilitating internal fund transfers between accounts.

## Features

- **Account Management** - Create accounts with initial balances and query account information
- **Fund Transfers** - Transfer funds between accounts with ACID guarantees
- **Decimal Precision** - All monetary values use 8 decimal places for financial accuracy
- **Idempotency Support** - Safe retries using X-Idempotency-Key header
- **API Versioning** - All endpoints prefixed with `/v1` for future compatibility
- **Distributed Tracing** - OpenTelemetry integration for request tracing
- **Rate Limiting** - Token bucket rate limiting to prevent API abuse
- **Health Checks** - Kubernetes-ready liveness and readiness probes
- **Metrics** - Prometheus-compatible metrics endpoint
- **Structured Logging** - JSON-formatted logs with trace/span IDs
- **Graceful Shutdown** - Proper connection draining and cleanup

---

## Prerequisites

Before you begin, ensure you have the following installed:

| Tool | Version | Installation Guide |
|------|---------|-------------------|
| Git | Any recent | [Install Git](https://git-scm.com/downloads) |
| Go | 1.21+ | [Install Go](https://golang.org/dl/) |
| Docker | Latest | [Install Docker](https://docs.docker.com/get-docker/) |
| Docker Compose | Latest | Included with Docker Desktop |
| Make | Any | See below |
| curl | Any | Pre-installed on macOS/Linux |

### Quick Install by OS

<details>
<summary><strong>macOS</strong> (click to expand)</summary>

```bash
# Install Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install prerequisites
xcode-select --install          # Includes Make and Git
brew install go                 # Go programming language
brew install --cask docker      # Docker Desktop

# Start Docker Desktop from Applications
# Add Go to PATH
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc && source ~/.zshrc
```

</details>

<details>
<summary><strong>Linux (Ubuntu/Debian)</strong> (click to expand)</summary>

```bash
# Update packages
sudo apt update && sudo apt upgrade -y

# Install build tools, Git, and curl
sudo apt install -y build-essential git curl wget

# Install Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin' >> ~/.bashrc && source ~/.bashrc

# Install Docker (see https://docs.docker.com/engine/install/ubuntu/)
curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh
sudo usermod -aG docker $USER && newgrp docker
```

</details>

<details>
<summary><strong>Windows</strong> (click to expand)</summary>

**Important:** This project requires WSL2 (Windows Subsystem for Linux).

```powershell
# In PowerShell (as Administrator)
wsl --install
# Restart computer, then open Ubuntu from Start menu
```

After restart, open **Ubuntu** and follow the Linux instructions above.

Install [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop/) and enable WSL2 integration in Settings.

</details>

### Verify Prerequisites

```bash
git --version          # git version 2.x.x
go version             # go version go1.21.x or higher
docker --version       # Docker version 24.x.x
docker compose version # Docker Compose version v2.x.x
make --version         # GNU Make 3.x or 4.x
```

> **Need detailed instructions?** See [wiki/getting-started.md](wiki/getting-started.md) for step-by-step setup for each operating system.

---

## Quick Start

```bash
# Clone the repository
git clone https://github.com/your-org/internal-transfers-service.git
cd internal-transfers-service

# First-time setup (installs tools, starts DB, runs migrations)
make setup

# Start the service
make run
```

The service starts on:
- **Main API**: http://localhost:8080
- **Ops (health/metrics)**: http://localhost:8081

### Test the API

```bash
# Create accounts (note: all API endpoints are prefixed with /v1)
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "1000.00"}'

curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 2, "initial_balance": "500.00"}'

# Get account balance
curl http://localhost:8080/v1/accounts/1
# Response: {"account_id":1,"balance":"1000"}

# Transfer funds (with idempotency key for safe retries)
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: transfer-001" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'
# Response: {"transaction_id":"550e8400-e29b-41d4-a716-446655440000"}
```

### Ops Endpoints

```bash
# Liveness probe - check if process is running
curl http://localhost:8081/health/live
# Response: {"status":"SERVING"}

# Readiness probe - check if service is ready for traffic
curl http://localhost:8081/health/ready
# Response: {"status":"SERVING"} (healthy)
# Response: {"status":"NOT_SERVING"} (unhealthy - e.g., DB down)

# Prometheus metrics
curl http://localhost:8081/metrics
# Response: Prometheus text format with http_request_duration_seconds, 
#           transfers_total, db_connections_open, etc.
```

---

## Documentation

Detailed documentation is available in the [wiki/](wiki/) folder:

| Document | Description |
|----------|-------------|
| [Getting Started](wiki/getting-started.md) | **Complete setup guide for fresh machines** |
| [API Reference](wiki/api-reference.md) | Complete API documentation |
| [Development Guide](wiki/development.md) | Development commands and workflow |
| [Database Guide](wiki/database.md) | Database schema and data inspection |
| [Configuration](wiki/configuration.md) | Configuration reference |
| [Architecture](wiki/architecture.md) | System architecture and design patterns |
| [Deployment](wiki/deployment.md) | Production deployment guide |
| [Troubleshooting](wiki/troubleshooting.md) | Common issues and solutions |

---

## Common Commands

```bash
make help          # Show all available commands
make setup         # First-time setup (tools, deps, db, mocks)
make run           # Run the service
make test          # Run all tests
make test-short    # Run tests (faster, less verbose)
make docker-up     # Start PostgreSQL
make docker-down   # Stop PostgreSQL
make migrate-up    # Run database migrations
make mock          # Regenerate mocks
make lint          # Run linter
make fmt           # Format code
```

---

## Project Structure

```
internal-transfers-service/
├── cmd/api/           # Application entry point
├── internal/          # Application code
│   ├── boot/          # Application bootstrap
│   ├── config/        # Configuration management
│   ├── interceptors/  # HTTP middleware
│   ├── modules/       # Business modules (account, transaction, health)
│   └── ...
├── pkg/               # Shared libraries
│   ├── apperror/      # Error handling
│   └── database/      # Database connection
├── config/            # Configuration files (TOML)
├── deployment/        # Docker and deployment configs
│   ├── dev/           # Development (docker-compose, Dockerfile)
│   └── prod/          # Production (Dockerfile)
├── wiki/              # Documentation
└── memory-bank/       # Project context
```

---

## Docker

### Development

```bash
# Build development image
docker build -t internal-transfers-service:dev -f deployment/dev/Dockerfile .

# Run with docker-compose (PostgreSQL + app)
cd deployment/dev && docker-compose up -d
```

### Production

```bash
# Build production image (multi-stage, optimized)
docker build -t internal-transfers-service:latest -f deployment/prod/Dockerfile .

# Run production container
docker run -d \
  -p 8080:8080 -p 8081:8081 \
  -e APP_DATABASE_HOST=your-db-host \
  -e APP_DATABASE_PASSWORD=your-password \
  internal-transfers-service:latest
```

---

## API Design Decisions

### Transaction Response Enhancement

The original specification suggests returning an empty response for successful transactions. This implementation intentionally returns a `transaction_id` in the response:

```json
{"transaction_id": "550e8400-e29b-41d4-a716-446655440000"}
```

**Rationale**: Returning a unique transaction identifier provides significant value:
- Clients can track and reference specific transactions
- Enables easier debugging and support
- Supports audit trail requirements
- Follows RESTful best practices for resource creation

### Decimal Precision

All monetary values are validated to a maximum of 8 decimal places, matching the database schema `DECIMAL(19,8)`. Requests exceeding this precision will receive a 400 Bad Request error.

---

## Security Features

This service implements production-grade security measures:

| Feature | Description |
|---------|-------------|
| **Request Size Limit** | Maximum 1MB request body to prevent DoS attacks |
| **Content-Type Validation** | Strict `application/json` validation on mutating requests |
| **Security Headers** | `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Cache-Control: no-store` |
| **Request Timeout** | 30-second timeout on all requests |
| **Rate Limiting** | Token bucket algorithm (100 req/s, burst 200) to prevent API abuse |
| **Idempotency** | Safe retries with `X-Idempotency-Key` header |
| **Pessimistic Locking** | Ordered row locks to prevent deadlocks and race conditions |

## Observability

| Feature | Description |
|---------|-------------|
| **Distributed Tracing** | OpenTelemetry integration with OTLP exporter |
| **Trace Context** | Trace/span IDs propagated through context and logs |
| **Metrics** | Prometheus metrics for HTTP requests, transfers, DB connections |
| **Structured Logging** | JSON logs with request_id, trace_id, span_id |

### Enable Tracing

To enable distributed tracing, set environment variables:

```bash
export TRACING_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317  # Your OTLP collector
```

Or configure in `config/prod.toml`:

```toml
[tracing]
enabled = true
endpoint = "your-otlp-collector:4317"
sample_rate = 0.1  # Sample 10% of requests
```

---

## Troubleshooting Quick Reference

| Problem | Solution |
|---------|----------|
| `Cannot connect to Docker daemon` | Start Docker Desktop (macOS/Windows) or `sudo systemctl start docker` (Linux) |
| `Port 8080 already in use` | Kill process: `lsof -i :8080` then `kill -9 <PID>` |
| `command not found: go` | Add to PATH: `export PATH=$PATH:/usr/local/go/bin` |
| `command not found: make` | macOS: `xcode-select --install`, Linux: `sudo apt install build-essential` |
| `permission denied` (Docker on Linux) | `sudo usermod -aG docker $USER && newgrp docker` |
| Tests fail with mock errors | Run `make mock` to regenerate mocks |

> **More troubleshooting?** See [wiki/troubleshooting.md](wiki/troubleshooting.md) or [wiki/getting-started.md#troubleshooting](wiki/getting-started.md#troubleshooting)

---

## Assumptions

The following assumptions were made during the design and implementation:

### Business Logic
- **Account IDs are provided by clients** - The system expects clients to provide unique account IDs rather than auto-generating them
- **Transfers are synchronous** - All fund transfers are processed immediately and synchronously
- **Single currency** - The system handles a single currency; multi-currency support is not implemented
- **No negative balances** - Accounts cannot have negative balances; transfers are rejected if insufficient funds
- **Decimal precision** - All monetary values use 8 decimal places for financial accuracy

### Technical
- **PostgreSQL required** - The service requires PostgreSQL 12+ for pessimistic locking support
- **Single instance initially** - While designed for horizontal scaling, the current implementation assumes single-instance deployment
- **Idempotency keys are client-provided** - Clients must generate and provide idempotency keys for safe retries
- **No authentication** - The API does not implement authentication/authorization (assumed to be handled by API gateway)
- **UTC timestamps** - All timestamps are stored and returned in UTC

### Operational
- **Health checks for Kubernetes** - Liveness and readiness probes are designed for Kubernetes orchestration
- **Prometheus metrics** - Metrics are exposed in Prometheus format for observability
- **Graceful shutdown** - The service implements graceful shutdown with configurable delay for load balancer draining

---

## License

MIT
