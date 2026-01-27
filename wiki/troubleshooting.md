# Troubleshooting Guide

This guide covers common issues and their solutions.

## Common Issues

### Service Won't Start

#### Port Already in Use

**Error:**
```
listen tcp :8080: bind: address already in use
```

**Solution:**
```bash
# Find process using the port
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use this one-liner
lsof -ti:8080 | xargs kill -9
```

#### Database Connection Failed

**Error:**
```
Failed to initialize database: failed to connect to host=localhost: server error
```

**Solutions:**

1. Check if PostgreSQL is running:
   ```bash
   make docker-status
   ```

2. Start PostgreSQL if not running:
   ```bash
   make docker-up
   ```

3. Verify connection details in config match docker-compose:
   - Host: localhost
   - Port: 5432
   - User: postgres
   - Password: postgres
   - Database: transfers

#### Database Does Not Exist

**Error:**
```
database "transfers" does not exist
```

**Solution:**
Run migrations:
```bash
make migrate-up
```

---

### API Errors

#### 500 Internal Server Error

**Possible Causes:**
1. Database connection lost
2. Migration not applied
3. Configuration error

**Debug Steps:**

1. Check health endpoint:
   ```bash
   curl http://localhost:8081/health/ready
   ```

2. Check service logs for errors

3. Verify database is accessible:
   ```bash
   docker exec -it transfers-postgres psql -U postgres -d transfers -c "SELECT 1"
   ```

#### 404 Account Not Found

**Cause:** Account doesn't exist in the database

**Solution:**
1. Create the account first:
   ```bash
   curl -X POST http://localhost:8080/accounts \
     -H "Content-Type: application/json" \
     -d '{"account_id": 1, "initial_balance": "1000.00"}'
   ```

2. Verify account exists:
   ```bash
   docker exec -it transfers-postgres psql -U postgres -d transfers \
     -c "SELECT * FROM accounts WHERE account_id = 1"
   ```

#### 400 Insufficient Balance

**Cause:** Source account doesn't have enough funds

**Debug:**
```bash
# Check current balance
curl http://localhost:8080/accounts/1
```

---

### Test Failures

#### Mocks Not Generated

**Error:**
```
undefined: mock.MockIRepository
```

**Solution:**
```bash
make mock
```

#### Import Cycle

**Error:**
```
import cycle not allowed in test
```

**Solution:**
Use black-box testing pattern:
```go
package account_test  // NOT package account
```

---

### Docker Issues

#### Container Won't Start

**Check logs:**
```bash
make docker-logs
```

**Common fixes:**

1. Port conflict:
   ```bash
   # Stop conflicting container
   docker stop <container-using-5432>
   ```

2. Clean restart:
   ```bash
   make docker-clean
   make docker-up
   ```

#### Data Persistence

**Issue:** Data lost after docker-clean

**Note:** `docker-clean` removes volumes. Use `docker-down` to preserve data.

---

### Migration Issues

#### Dirty Migration State

**Error:**
```
error: Dirty database version X. Fix and force version.
```

**Solution:**
```bash
# Force to last successful version
make migrate-force
# Enter the version number when prompted
```

#### Migration Already Applied

**Error:**
```
no change
```

**This is normal** - means all migrations are already applied.

---

## Debugging Commands

### Check Service Status

```bash
# Is the service running?
curl http://localhost:8081/health/live

# Is the database connected?
curl http://localhost:8081/health/ready

# View metrics
curl http://localhost:8081/metrics
```

### Check Database

```bash
# Connect to database
docker exec -it transfers-postgres psql -U postgres -d transfers

# View accounts
SELECT * FROM accounts;

# View recent transactions
SELECT * FROM transactions ORDER BY created_at DESC LIMIT 10;

# Check migration version
SELECT * FROM schema_migrations;
```

### View Logs

```bash
# PostgreSQL logs
make docker-logs

# Application logs (when running with make run)
# Logs appear in the terminal
```

### Check Configuration

```bash
# View loaded configuration
APP_ENV=dev go run ./cmd/api
# Check the startup log for config values
```

---

## Getting Help

If you're still stuck:

1. Check the [API Reference](api-reference.md)
2. Check the [Configuration Guide](configuration.md)
3. Check the [Database Guide](database.md)
4. Review the service logs
5. Open an issue with:
   - Error message
   - Steps to reproduce
   - Environment details
