# Database Guide

This guide covers the database schema, how to inspect data, and common database operations.

## Database Schema

### Accounts Table

Stores account information with balance tracking.

```sql
CREATE TABLE accounts (
    account_id BIGINT PRIMARY KEY,
    balance DECIMAL(19, 8) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT positive_balance CHECK (balance >= 0)
);
```

| Column | Type | Description |
|--------|------|-------------|
| account_id | BIGINT | Primary key, client-provided |
| balance | DECIMAL(19,8) | Current balance (8 decimal places) |
| created_at | TIMESTAMPTZ | Account creation timestamp |
| updated_at | TIMESTAMPTZ | Last update timestamp |

### Transactions Table

Records all fund transfers between accounts.

```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_account_id BIGINT NOT NULL REFERENCES accounts(account_id),
    destination_account_id BIGINT NOT NULL REFERENCES accounts(account_id),
    amount DECIMAL(19, 8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT positive_amount CHECK (amount > 0),
    CONSTRAINT different_accounts CHECK (source_account_id != destination_account_id)
);

CREATE INDEX idx_transactions_source ON transactions(source_account_id);
CREATE INDEX idx_transactions_destination ON transactions(destination_account_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
```

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key, auto-generated |
| source_account_id | BIGINT | Account funds came from |
| destination_account_id | BIGINT | Account funds went to |
| amount | DECIMAL(19,8) | Transfer amount |
| created_at | TIMESTAMPTZ | Transaction timestamp |

### Idempotency Keys Table

Stores idempotency keys for safe request retries.

```sql
CREATE TABLE idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    response_status INT NOT NULL,
    response_body JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_idempotency_created_at ON idempotency_keys(created_at);
```

| Column | Type | Description |
|--------|------|-------------|
| key | VARCHAR(255) | Client-provided idempotency key |
| response_status | INT | Cached HTTP status code |
| response_body | JSONB | Cached response body |
| created_at | TIMESTAMPTZ | Key creation timestamp |

---

## Connecting to the Database

### Using Docker Exec (Recommended)

```bash
# Connect to the PostgreSQL container
docker exec -it transfers-postgres psql -U postgres -d transfers
```

### Using psql Directly

```bash
# If you have psql installed locally
psql -h localhost -U postgres -d transfers
# Password: postgres
```

### Connection Details

| Parameter | Value |
|-----------|-------|
| Host | localhost |
| Port | 5432 |
| Database | transfers |
| User | postgres |
| Password | postgres |
| SSL Mode | disable |

---

## Common Queries

### View All Accounts

```sql
SELECT * FROM accounts ORDER BY account_id;
```

### View Account Balances with Formatting

```sql
SELECT 
    account_id,
    to_char(balance, 'FM999,999,999,990.00000000') as balance,
    created_at
FROM accounts 
ORDER BY account_id;
```

### View Recent Transactions

```sql
SELECT 
    id,
    source_account_id,
    destination_account_id,
    to_char(amount, 'FM999,999,990.00000000') as amount,
    created_at
FROM transactions 
ORDER BY created_at DESC 
LIMIT 20;
```

### View Transaction History for an Account

```sql
-- All transactions involving account 1
SELECT 
    id,
    CASE 
        WHEN source_account_id = 1 THEN 'DEBIT'
        ELSE 'CREDIT'
    END as type,
    source_account_id,
    destination_account_id,
    to_char(amount, 'FM999,999,990.00000000') as amount,
    created_at
FROM transactions 
WHERE source_account_id = 1 OR destination_account_id = 1
ORDER BY created_at DESC;
```

### Calculate Total Balances (Integrity Check)

```sql
-- Sum of all account balances
SELECT 
    COUNT(*) as total_accounts,
    to_char(SUM(balance), 'FM999,999,999,990.00000000') as total_balance
FROM accounts;
```

### View Idempotency Keys

```sql
SELECT 
    key,
    response_status,
    response_body,
    created_at
FROM idempotency_keys 
ORDER BY created_at DESC 
LIMIT 10;
```

---

## Database Management

### Check Migration Status

```bash
make migrate-version
```

### Run Pending Migrations

```bash
make migrate-up
```

### Rollback Last Migration

```bash
make migrate-down-one
```

### Rollback All Migrations

```bash
make migrate-down
```

### Create New Migration

```bash
make migrate-create
# Enter migration name when prompted
```

---

## Reset Database

To completely reset the database:

```bash
# Stop and remove container with volumes
make docker-clean

# Start fresh
make docker-up

# Wait for PostgreSQL to be ready
sleep 5

# Run migrations
make migrate-up
```

---

## Database Constraints

The database enforces the following constraints:

1. **Positive Balance**: Account balance cannot be negative
   ```sql
   CONSTRAINT positive_balance CHECK (balance >= 0)
   ```

2. **Positive Amount**: Transaction amount must be positive
   ```sql
   CONSTRAINT positive_amount CHECK (amount > 0)
   ```

3. **Different Accounts**: Source and destination must be different
   ```sql
   CONSTRAINT different_accounts CHECK (source_account_id != destination_account_id)
   ```

4. **Referential Integrity**: Transaction accounts must exist
   ```sql
   REFERENCES accounts(account_id)
   ```

---

## Performance Considerations

### Indexes

The following indexes are created for performance:

```sql
-- Primary keys (automatically indexed)
accounts.account_id
transactions.id
idempotency_keys.key

-- Additional indexes
idx_transactions_source       -- For source account lookups
idx_transactions_destination  -- For destination account lookups
idx_transactions_created_at   -- For time-based queries
idx_idempotency_created_at    -- For cleanup queries
```

### Connection Pool

The application uses connection pooling with these defaults:

| Setting | Development | Production |
|---------|-------------|------------|
| Max Connections | 10 | 50 |
| Min Connections | 2 | 10 |
| Max Lifetime | 1h | 30m |
| Max Idle Time | 30m | 10m |
