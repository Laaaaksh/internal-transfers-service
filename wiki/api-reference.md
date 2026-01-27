# API Reference

This document provides complete API documentation for the Internal Transfers Service.

## Base URLs

| Environment | URL |
|-------------|-----|
| Development | http://localhost:8080 |
| Ops/Health | http://localhost:8081 |

## API Versioning

All API endpoints are versioned with a `/v1` prefix. This ensures backward compatibility when new versions are introduced.

| Version | Prefix | Status |
|---------|--------|--------|
| v1 | `/v1` | Current (Active) |

## Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /v1/accounts | Create a new account |
| GET | /v1/accounts/{accountID} | Get account details |
| POST | /v1/transactions | Transfer funds between accounts |
| GET | /health/live | Liveness probe |
| GET | /health/ready | Readiness probe |
| GET | /metrics | Prometheus metrics |

> **Note:** Health and metrics endpoints are NOT versioned as they are ops endpoints on port 8081.

---

## Account Endpoints

### Create Account

Creates a new account with an initial balance.

**Request:**
```http
POST /v1/accounts
Content-Type: application/json

{
    "account_id": 123,
    "initial_balance": "1000.50"
}
```

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| account_id | integer | Yes | Unique account identifier (positive integer) |
| initial_balance | string | Yes | Initial balance (decimal string, >= 0) |

**Response:**

| Status | Description |
|--------|-------------|
| 201 Created | Account created successfully (empty body) |
| 400 Bad Request | Invalid request body |
| 409 Conflict | Account already exists |
| 500 Internal Server Error | Server error |

**Examples:**

```bash
# Create account with ID 1
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "1000.00"}'

# Create account with decimal precision
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 2, "initial_balance": "500.12345678"}'
```

---

### Get Account

Retrieves account details including current balance.

**Request:**
```http
GET /v1/accounts/{accountID}
```

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| accountID | integer | Account identifier |

**Response:**

| Status | Description |
|--------|-------------|
| 200 OK | Account details |
| 400 Bad Request | Invalid account ID |
| 404 Not Found | Account not found |
| 500 Internal Server Error | Server error |

**Success Response Body:**
```json
{
    "account_id": 123,
    "balance": "1000.50"
}
```

**Examples:**

```bash
# Get account 1
curl http://localhost:8080/v1/accounts/1

# Response:
# {"account_id":1,"balance":"1000"}
```

---

## Transaction Endpoints

### Create Transaction (Transfer)

Transfers funds from one account to another atomically.

**Request:**
```http
POST /v1/transactions
Content-Type: application/json
X-Idempotency-Key: unique-key-123 (optional)

{
    "source_account_id": 1,
    "destination_account_id": 2,
    "amount": "100.00"
}
```

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| source_account_id | integer | Yes | Source account ID |
| destination_account_id | integer | Yes | Destination account ID |
| amount | string | Yes | Transfer amount (decimal string, > 0) |

**Headers:**

| Header | Required | Description |
|--------|----------|-------------|
| X-Idempotency-Key | No | Unique key for idempotent requests |

**Response:**

| Status | Description |
|--------|-------------|
| 201 Created | Transfer successful |
| 400 Bad Request | Invalid request or insufficient balance |
| 404 Not Found | Account not found |
| 500 Internal Server Error | Server error |

**Success Response Body:**
```json
{
    "transaction_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Examples:**

```bash
# Transfer 100 from account 1 to account 2
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'

# Idempotent transfer (safe to retry)
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: transfer-001" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'
```

---

## Health & Ops Endpoints

All ops endpoints are served on port **8081** (separate from the main API on port 8080).

### Liveness Probe

Indicates if the service process is running. This is a simple check that always returns success if the process is alive.

**Request:**
```http
GET /health/live
```

**Response:**
```json
{"status":"SERVING"}
```

**Curl Example:**
```bash
curl http://localhost:8081/health/live
# Response: {"status":"SERVING"}
```

**HTTP Status Codes:**
| Status | Description |
|--------|-------------|
| 200 OK | Process is running |

**Usage:** Use this for Kubernetes liveness probes. If this fails, Kubernetes should restart the pod.

---

### Readiness Probe

Indicates if the service is ready to accept traffic. This checks database connectivity and other dependencies.

**Request:**
```http
GET /health/ready
```

**Response (healthy):**
```json
{"status":"SERVING"}
```

**Response (unhealthy):**
```json
{"status":"NOT_SERVING"}
```

**Curl Examples:**
```bash
# Check readiness
curl http://localhost:8081/health/ready
# Response (healthy): {"status":"SERVING"}
# Response (unhealthy): {"status":"NOT_SERVING"}

# Check with HTTP status code
curl -w "\nHTTP Status: %{http_code}\n" http://localhost:8081/health/ready
# HTTP Status: 200 (healthy) or 503 (unhealthy)
```

**HTTP Status Codes:**
| Status | Description |
|--------|-------------|
| 200 OK | Service is ready to accept traffic |
| 503 Service Unavailable | Service is not ready (e.g., DB connection lost) |

**Usage:** Use this for Kubernetes readiness probes. If this fails, Kubernetes should stop sending traffic to the pod.

---

### Metrics

Prometheus-compatible metrics endpoint exposing application metrics.

**Request:**
```http
GET /metrics
```

**Response:** Prometheus text format metrics.

**Curl Example:**
```bash
# Get all metrics
curl http://localhost:8081/metrics

# Filter for specific metrics
curl -s http://localhost:8081/metrics | grep http_request

# Example output:
# http_request_duration_seconds_bucket{method="POST",path="/accounts",status_code="201",le="0.005"} 10
# http_requests_total{method="GET",path="/accounts/1",status_code="200"} 25
```

**Available Metrics:**

| Metric | Type | Description |
|--------|------|-------------|
| `http_request_duration_seconds` | Histogram | HTTP request duration in seconds |
| `http_requests_total` | Counter | Total HTTP requests by method, path, status |
| `transfers_total` | Counter | Total transfer attempts |
| `transfers_success_total` | Counter | Successful transfers |
| `transfers_failed_total` | Counter | Failed transfers |
| `db_connections_open` | Gauge | Open database connections |
| `db_connections_idle` | Gauge | Idle database connections |

**Usage:** Configure Prometheus to scrape `http://service:8081/metrics`

---

## Error Responses

All errors follow a consistent structure:

```json
{
    "error": "User-friendly error message",
    "code": "ERROR_CODE",
    "details": {
        "field": "additional context"
    }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| INVALID_REQUEST | 400 | Invalid request body or parameters |
| INSUFFICIENT_BALANCE | 400 | Source account has insufficient funds |
| ACCOUNT_NOT_FOUND | 404 | Account does not exist |
| ACCOUNT_ALREADY_EXISTS | 409 | Account with this ID already exists |
| INTERNAL_ERROR | 500 | Internal server error |

---

## Idempotency

Idempotency ensures that retrying a request with the same key produces the same result without re-executing the operation. This is critical for financial transactions where network failures could cause duplicate processing.

### How It Works

1. Client includes `X-Idempotency-Key` header with a unique identifier
2. Server checks if this key was seen before (stored in PostgreSQL)
3. **If key exists**: Returns the cached response without processing again
4. **If key is new**: Processes the request, stores the response, and returns it

### Supported Endpoints

| Endpoint | Idempotency Supported |
|----------|----------------------|
| POST /v1/accounts | ✅ Yes |
| GET /v1/accounts/{id} | ❌ N/A (GET is inherently idempotent) |
| POST /v1/transactions | ✅ Yes |

### Request Headers

| Header | Required | Description |
|--------|----------|-------------|
| `X-Idempotency-Key` | No | Unique key for idempotent requests (max 255 chars) |

### Response Headers

| Header | Description |
|--------|-------------|
| `X-Idempotent-Replayed` | Set to `true` if response was returned from cache |

### Examples

**First Request (Processed Normally):**
```bash
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: transfer-abc-123" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'

# Response: 201 Created
# {"transaction_id":"550e8400-e29b-41d4-a716-446655440000"}
```

**Retry with Same Key (Returns Cached Response):**
```bash
curl -v -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: transfer-abc-123" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'

# Response Headers include: X-Idempotent-Replayed: true
# Response: 201 Created (same as original)
# {"transaction_id":"550e8400-e29b-41d4-a716-446655440000"}
```

**Account Creation with Idempotency:**
```bash
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: create-account-456" \
  -d '{"account_id": 123, "initial_balance": "1000.00"}'

# Safe to retry - won't create duplicate accounts
```

### Best Practices

1. **Generate unique keys**: Use UUIDs or combine timestamp + operation type
   ```bash
   # Good key patterns:
   X-Idempotency-Key: txn-$(uuidgen)
   X-Idempotency-Key: transfer-1706384400-src1-dst2
   X-Idempotency-Key: create-account-user123-$(date +%s)
   ```

2. **Store keys client-side**: Keep track of keys for potential retries

3. **Don't reuse keys**: Each logical operation should have a unique key

4. **Handle the replayed header**: Log when `X-Idempotent-Replayed: true` for debugging

### Key Expiration

- **TTL**: 24 hours from first request
- **Storage**: PostgreSQL `idempotency_keys` table
- **Cleanup**: Expired keys are automatically removed

### Error Handling

| Scenario | Behavior |
|----------|----------|
| Key too long (> 255 chars) | Returns 400 Bad Request |
| Key not provided | Request processed normally (no idempotency) |
| Database error checking key | Request processed normally (fail-open) |
| Same key, different body | Returns cached response (body not validated) |

### Important Notes

- **Idempotency keys are NOT request-body sensitive**: Sending the same key with a different body will return the cached response from the first request, not process the new body
- **Only 2xx-4xx responses are cached**: 5xx server errors are not cached, allowing retries to potentially succeed
- **Keys are global**: The same key across different endpoints is treated as the same key

---

## Rate Limiting

The API implements token bucket rate limiting to prevent abuse and ensure fair usage.

### Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `requests_per_second` | 100 | Average requests per second allowed |
| `burst_size` | 200 | Maximum burst of requests allowed |

### Response Headers

When rate limited, the response includes:

| Header | Value | Description |
|--------|-------|-------------|
| `Retry-After` | `1` | Seconds to wait before retrying |
| `Content-Type` | `application/json` | Response content type |

### Rate Limit Response

```bash
# When rate limited (HTTP 429)
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "100.00"}'

# Response: 429 Too Many Requests
# {"error":"rate limit exceeded","code":"RATE_LIMITED"}
```

### Best Practices

1. **Implement exponential backoff**: On 429, wait and retry with increasing delays
2. **Check Retry-After header**: Honor the suggested wait time
3. **Use idempotency keys**: Ensures safe retries after rate limiting

---

## Distributed Tracing

The service supports OpenTelemetry distributed tracing for observability.

### Trace Context Propagation

The service automatically propagates W3C TraceContext headers:

| Header | Description |
|--------|-------------|
| `traceparent` | W3C trace context parent header |
| `tracestate` | W3C trace context state header |

### Trace and Span IDs in Logs

When tracing is enabled, all logs include:

```json
{
  "level": "info",
  "msg": "Request completed",
  "request_id": "abc123",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7"
}
```

### Enabling Tracing

Set environment variables:

```bash
export TRACING_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

Or configure in `config/prod.toml`:

```toml
[tracing]
enabled = true
endpoint = "your-otlp-collector:4317"
sample_rate = 0.1  # Sample 10% of requests in production
insecure = false
```

### Correlation Example

```bash
# Make request with trace context
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'

# The trace ID will appear in all related logs and spans
```
