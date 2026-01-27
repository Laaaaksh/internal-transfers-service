# API Reference

This document provides complete API documentation for the Internal Transfers Service.

## Base URLs

| Environment | URL |
|-------------|-----|
| Development | http://localhost:8080 |
| Ops/Health | http://localhost:8081 |

## Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /accounts | Create a new account |
| GET | /accounts/{accountID} | Get account details |
| POST | /transactions | Transfer funds between accounts |
| GET | /health/live | Liveness probe |
| GET | /health/ready | Readiness probe |
| GET | /metrics | Prometheus metrics |

---

## Account Endpoints

### Create Account

Creates a new account with an initial balance.

**Request:**
```http
POST /accounts
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
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "1000.00"}'

# Create account with decimal precision
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 2, "initial_balance": "500.12345678"}'
```

---

### Get Account

Retrieves account details including current balance.

**Request:**
```http
GET /accounts/{accountID}
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
curl http://localhost:8080/accounts/1

# Response:
# {"account_id":1,"balance":"1000"}
```

---

## Transaction Endpoints

### Create Transaction (Transfer)

Transfers funds from one account to another atomically.

**Request:**
```http
POST /transactions
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
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'

# Idempotent transfer (safe to retry)
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: transfer-001" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'
```

---

## Health Endpoints

### Liveness Probe

Indicates if the service process is running.

**Request:**
```http
GET /health/live
```

**Response:**
```json
{"status":"SERVING"}
```

**Usage:** Use this for Kubernetes liveness probes.

---

### Readiness Probe

Indicates if the service is ready to accept traffic (database connected).

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

**Usage:** Use this for Kubernetes readiness probes.

---

### Metrics

Prometheus-compatible metrics endpoint.

**Request:**
```http
GET /metrics
```

**Response:** Prometheus text format metrics.

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

For safe retries on transaction requests, include the `X-Idempotency-Key` header:

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: my-unique-key-123" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'
```

If you retry the same request with the same idempotency key:
- The operation will not be executed again
- You will receive the same response as the original request

Idempotency keys expire after 24 hours.
