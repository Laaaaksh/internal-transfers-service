# Troubleshooting Guide

This guide covers common issues and their solutions, organized by category.

---

## Table of Contents

1. [Setup & Installation Issues](#setup--installation-issues)
2. [Docker Issues](#docker-issues)
3. [Database Issues](#database-issues)
4. [Service Issues](#service-issues)
5. [API Errors](#api-errors)
6. [Test Failures](#test-failures)
7. [Migration Issues](#migration-issues)
8. [Debugging Commands](#debugging-commands)
9. [Getting Help](#getting-help)

---

## Setup & Installation Issues

### Command Not Found: go

**Error:**
```
zsh: command not found: go
```

**Cause:** Go is not installed or not in your PATH.

**Solutions:**

1. **Verify Go is installed:**
   ```bash
   ls /usr/local/go/bin/go  # Linux
   ls /opt/homebrew/bin/go   # macOS Apple Silicon
   ls /usr/local/bin/go      # macOS Intel
   ```

2. **Add Go to PATH:**
   ```bash
   # For bash (~/.bashrc)
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
   source ~/.bashrc

   # For zsh (~/.zshrc)
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
   echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
   source ~/.zshrc
   ```

3. **Install Go if missing:**
   - macOS: `brew install go`
   - Linux: See [Getting Started - Install Go](getting-started.md#install-go)

---

### Command Not Found: make

**Error:**
```
zsh: command not found: make
```

**Solutions:**

**macOS:**
```bash
xcode-select --install
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt install -y build-essential
```

**Windows (in WSL/Ubuntu):**
```bash
sudo apt install -y build-essential
```

---

### Command Not Found: docker

**Error:**
```
zsh: command not found: docker
```

**Solutions:**

1. **Install Docker:**
   - macOS: `brew install --cask docker`
   - Linux: See [Getting Started - Install Docker](getting-started.md#install-docker)
   - Windows: Install Docker Desktop

2. **Start Docker:**
   - macOS/Windows: Open Docker Desktop application
   - Linux: `sudo systemctl start docker`

---

### Cannot Connect to Docker Daemon

**Error:**
```
Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
```

**Cause:** Docker daemon is not running.

**Solutions:**

**macOS:**
1. Open Docker Desktop from Applications
2. Wait for the whale icon to stop animating
3. Try the command again

**Linux:**
```bash
# Start Docker service
sudo systemctl start docker

# Enable Docker to start on boot
sudo systemctl enable docker

# Verify it's running
sudo systemctl status docker
```

**Windows:**
1. Open Docker Desktop
2. Wait for it to fully start
3. Ensure WSL integration is enabled (Settings → Resources → WSL Integration)

---

### Docker Permission Denied (Linux)

**Error:**
```
Got permission denied while trying to connect to the Docker daemon socket
```

**Cause:** User is not in the `docker` group.

**Solution:**
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Apply changes (or log out and log back in)
newgrp docker

# Verify
docker run hello-world
```

---

### Homebrew Not Found (macOS)

**Error:**
```
zsh: command not found: brew
```

**Solution:**
```bash
# Install Homebrew
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Add to PATH (Apple Silicon)
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/opt/homebrew/bin/brew shellenv)"

# Add to PATH (Intel Mac)
echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/usr/local/bin/brew shellenv)"
```

---

### Git Not Installed

**Error:**
```
zsh: command not found: git
```

**Solutions:**

**macOS:**
```bash
xcode-select --install
# Or
brew install git
```

**Linux:**
```bash
sudo apt install -y git
```

---

### curl Not Installed (Windows)

**Error in PowerShell:**
```
curl : The term 'curl' is not recognized
```

**Solution:** Use WSL (Ubuntu) instead of PowerShell:
```bash
wsl
# Then run curl commands inside Ubuntu
```

Or install curl for Windows via [curl.se](https://curl.se/windows/).

---

### WSL Not Installed (Windows)

**Error:**
```
'wsl' is not recognized as an internal or external command
```

**Solution:**
Open PowerShell as Administrator:
```powershell
wsl --install
```
Restart computer when prompted.

---

## Docker Issues

### Container Won't Start

**Check logs:**
```bash
make docker-logs
```

**Common fixes:**

1. **Port conflict:**
   ```bash
   # Find what's using port 5432
   lsof -i :5432
   
   # Stop conflicting process
   kill -9 <PID>
   
   # Or stop conflicting container
   docker stop <container-name>
   ```

2. **Clean restart:**
   ```bash
   make docker-clean
   make docker-up
   ```

3. **Check Docker has resources:**
   - Docker Desktop → Settings → Resources
   - Increase memory if needed

---

### Port 5432 Already in Use

**Error:**
```
Bind for 0.0.0.0:5432 failed: port is already allocated
```

**Cause:** Another PostgreSQL or container is using the port.

**Solutions:**

1. **Find and stop the process:**
   ```bash
   lsof -i :5432
   kill -9 <PID>
   ```

2. **Stop other PostgreSQL service:**
   ```bash
   # macOS (Homebrew PostgreSQL)
   brew services stop postgresql
   
   # Linux
   sudo systemctl stop postgresql
   ```

3. **Use a different port:**
   
   Edit `deployment/dev/docker-compose.yml`:
   ```yaml
   ports:
     - "5433:5432"  # Use 5433 on host
   ```
   
   Update `config/dev.toml`:
   ```toml
   [database]
   port = 5433
   ```

---

### Data Persistence

**Issue:** Data lost after `docker-clean`

**Note:** `make docker-clean` removes volumes (intentionally). Use `make docker-down` to preserve data between runs.

---

## Database Issues

### Database Connection Failed

**Error:**
```
Failed to initialize database: failed to connect to host=localhost: server error
```

**Solutions:**

1. **Check if PostgreSQL is running:**
   ```bash
   make docker-status
   ```

2. **Start PostgreSQL if not running:**
   ```bash
   make docker-up
   ```

3. **Verify connection details in config match docker-compose:**
   - Host: localhost
   - Port: 5432
   - User: postgres
   - Password: postgres
   - Database: transfers

4. **Wait for PostgreSQL to be ready:**
   ```bash
   # Check container health
   docker ps
   # Wait until STATUS shows "(healthy)"
   ```

---

### Database Does Not Exist

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

### Tables Not Found

**Error:**
```
relation "accounts" does not exist
```

**Cause:** Migrations haven't been run.

**Solution:**
```bash
make migrate-up
```

---

## Service Issues

### Port Already in Use

**Error:**
```
listen tcp :8080: bind: address already in use
```

**Solutions:**

1. **Find and kill the process:**
   ```bash
   # Find process using port
   lsof -i :8080
   
   # Kill it
   kill -9 <PID>
   
   # One-liner
   lsof -ti:8080 | xargs kill -9
   ```

2. **Use a different port:**
   
   Edit `config/dev.toml`:
   ```toml
   [app]
   port = ":8090"
   ops_port = ":8091"
   ```

---

### Service Crashes on Start

**Debug steps:**

1. **Check logs:** Look at the terminal output for errors

2. **Check configuration:**
   ```bash
   cat config/dev.toml
   ```

3. **Check database is accessible:**
   ```bash
   make docker-status
   curl http://localhost:8081/health/ready
   ```

4. **Try verbose mode:**
   ```bash
   go run ./cmd/api 2>&1 | head -50
   ```

---

## API Errors

### 500 Internal Server Error

**Possible Causes:**
1. Database connection lost
2. Migration not applied
3. Configuration error

**Debug Steps:**

1. **Check health endpoint:**
   ```bash
   curl http://localhost:8081/health/ready
   ```

2. **Check service logs** for the specific error

3. **Verify database is accessible:**
   ```bash
   docker exec -it transfers-postgres psql -U postgres -d transfers -c "SELECT 1"
   ```

---

### 404 Account Not Found

**Cause:** Account doesn't exist in the database

**Solution:**

1. **Create the account first:**
   ```bash
   curl -X POST http://localhost:8080/v1/accounts \
     -H "Content-Type: application/json" \
     -d '{"account_id": 1, "initial_balance": "1000.00"}'
   ```

2. **Verify account exists:**
   ```bash
   docker exec -it transfers-postgres psql -U postgres -d transfers \
     -c "SELECT * FROM accounts WHERE account_id = 1"
   ```

---

### 400 Insufficient Balance

**Cause:** Source account doesn't have enough funds

**Debug:**
```bash
# Check current balance
curl http://localhost:8080/v1/accounts/1
```

---

### 409 Duplicate Account

**Cause:** Account with that ID already exists

**Solution:** Use a different account_id or delete the existing account:
```bash
docker exec -it transfers-postgres psql -U postgres -d transfers \
  -c "DELETE FROM accounts WHERE account_id = 1"
```

---

### 415 Unsupported Media Type

**Cause:** Missing or incorrect Content-Type header

**Solution:** Include the header in your request:
```bash
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "1000.00"}'
```

---

## Test Failures

### Mocks Not Generated

**Error:**
```
undefined: mock.MockIRepository
```

**Solution:**
```bash
make mock
```

---

### Import Cycle

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

### Test Database Connection

**Error:** Tests fail with database connection errors

**Cause:** Tests may be trying to connect to real database

**Note:** Unit tests should use mocks, not real database. Check that mocks are properly set up.

---

## Migration Issues

### Dirty Migration State

**Error:**
```
error: Dirty database version X. Fix and force version.
```

**Solution:**
```bash
# Force to last successful version
make migrate-force
# Enter the version number when prompted (e.g., 2)

# Then retry
make migrate-up
```

---

### Migration Already Applied

**Error:**
```
no change
```

**This is normal** - means all migrations are already applied.

---

### Migration Syntax Error

**Error:**
```
migration failed: <SQL error>
```

**Solution:**

1. Fix the SQL in the migration file
2. Force to previous version
3. Retry the migration

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

# Exit psql
\q
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

### Check System Resources

```bash
# Check if ports are in use
lsof -i :8080
lsof -i :8081
lsof -i :5432

# Check Docker containers
docker ps -a

# Check Docker resources
docker system df
```

### Full System Verification

```bash
# Run this to check everything
echo "=== Prerequisites ===" && \
git --version && \
go version && \
docker --version && \
docker compose version && \
make --version && \
echo "" && \
echo "=== Docker Status ===" && \
docker ps && \
echo "" && \
echo "=== Service Health ===" && \
curl -s http://localhost:8081/health/live && \
echo "" && \
curl -s http://localhost:8081/health/ready
```

---

## Getting Help

If you're still stuck:

1. **Check the documentation:**
   - [Getting Started](getting-started.md)
   - [API Reference](api-reference.md)
   - [Configuration Guide](configuration.md)
   - [Database Guide](database.md)

2. **Review the service logs** for specific error messages

3. **Search for the error message** online

4. **Open an issue** with:
   - Error message (full output)
   - Steps to reproduce
   - Operating system and version
   - Output of `make --version`, `go version`, `docker --version`
   - Output of `make docker-status`
