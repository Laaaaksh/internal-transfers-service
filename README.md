# Internal Transfers Service

A production-grade microservice for facilitating internal fund transfers between accounts.

## Features

- **Account Management** - Create accounts with initial balances and query account information
- **Fund Transfers** - Transfer funds between accounts with ACID guarantees
- **Decimal Precision** - All monetary values use 8 decimal places for financial accuracy
- **Idempotency Support** - Safe retries using X-Idempotency-Key header
- **Health Checks** - Kubernetes-ready liveness and readiness probes
- **Metrics** - Prometheus-compatible metrics endpoint
- **Structured Logging** - JSON-formatted logs with request tracing
- **Graceful Shutdown** - Proper connection draining and cleanup

## Quick Start

```bash
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
# Create accounts
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "1000.00"}'

curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 2, "initial_balance": "500.00"}'

# Get account balance
curl http://localhost:8080/accounts/1
# Response: {"account_id":1,"balance":"1000"}

# Transfer funds (with idempotency key for safe retries)
curl -X POST http://localhost:8080/transactions \
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

## Documentation

Detailed documentation is available in the [wiki/](wiki/) folder:

| Document | Description |
|----------|-------------|
| [Getting Started](wiki/getting-started.md) | Prerequisites and setup guide |
| [API Reference](wiki/api-reference.md) | Complete API documentation |
| [Development Guide](wiki/development.md) | Development commands and workflow |
| [Database Guide](wiki/database.md) | Database schema and data inspection |
| [Configuration](wiki/configuration.md) | Configuration reference |
| [Architecture](wiki/architecture.md) | System architecture and design patterns |
| [Deployment](wiki/deployment.md) | Production deployment guide |
| [Troubleshooting](wiki/troubleshooting.md) | Common issues and solutions |

## Common Commands

```bash
make help          # Show all available commands
make run           # Run the service
make test          # Run all tests
make docker-up     # Start PostgreSQL
make docker-down   # Stop PostgreSQL
make migrate-up    # Run database migrations
make mock          # Regenerate mocks
```

## Project Structure

```
internal-transfers-service/
├── cmd/api/           # Application entry point
├── internal/          # Application code
│   ├── modules/       # Business modules (account, transaction, health)
│   ├── config/        # Configuration management
│   └── ...
├── pkg/               # Shared libraries
├── config/            # Configuration files (TOML)
├── deployment/        # Docker and deployment configs
│   ├── dev/           # Development (docker-compose, Dockerfile)
│   └── prod/          # Production (Dockerfile)
├── wiki/              # Documentation
└── memory-bank/       # Project context
```

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

## Security Features

This service implements production-grade security measures:

| Feature | Description |
|---------|-------------|
| **Request Size Limit** | Maximum 1MB request body to prevent DoS attacks |
| **Content-Type Validation** | Strict `application/json` validation on mutating requests |
| **Security Headers** | `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Cache-Control: no-store` |
| **Request Timeout** | 30-second timeout on all requests |
| **Idempotency** | Safe retries with `X-Idempotency-Key` header |
| **Pessimistic Locking** | Ordered row locks to prevent deadlocks and race conditions |

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

## License

MIT
