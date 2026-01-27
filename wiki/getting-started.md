# Getting Started

This guide will help you set up the Internal Transfers Service for local development on a **completely fresh machine**.

---

## Table of Contents

1. [Prerequisites Overview](#prerequisites-overview)
2. [Operating System Setup](#operating-system-setup)
   - [macOS Setup](#macos-setup)
   - [Linux Setup](#linux-setup)
   - [Windows Setup](#windows-setup)
3. [Install Git](#install-git)
4. [Install Go](#install-go)
5. [Install Docker](#install-docker)
6. [Install Make](#install-make)
7. [Clone the Repository](#clone-the-repository)
8. [Quick Start](#quick-start-recommended)
9. [Manual Setup](#manual-setup-step-by-step)
10. [Verify Installation](#verify-installation)
11. [IDE Setup (Recommended)](#ide-setup-recommended)
12. [Troubleshooting](#troubleshooting)
13. [Next Steps](#next-steps)

---

## Prerequisites Overview

Before you begin, you'll need to install the following tools:

| Tool | Version | Purpose |
|------|---------|---------|
| Git | Any recent | Clone the repository |
| Go | 1.24+ | Build and run the service |
| Docker | Latest | Run PostgreSQL database |
| Docker Compose | Latest | Orchestrate containers |
| Make | Any | Run project commands |
| curl | Any | Test API endpoints |

---

## Operating System Setup

Choose your operating system below for specific instructions.

### macOS Setup

#### Step 1: Install Xcode Command Line Tools

Open Terminal and run:

```bash
xcode-select --install
```

A popup will appear. Click "Install" and wait for completion (may take 5-10 minutes).

#### Step 2: Install Homebrew (Package Manager)

Homebrew makes installing software on macOS easy.

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

After installation, follow the instructions printed to add Homebrew to your PATH. Typically:

```bash
# For Apple Silicon (M1/M2/M3) Macs:
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/opt/homebrew/bin/brew shellenv)"

# For Intel Macs:
echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/usr/local/bin/brew shellenv)"
```

Verify installation:

```bash
brew --version
# Expected: Homebrew 4.x.x
```

#### Step 3: Install All Prerequisites via Homebrew

```bash
# Install Git (may already be installed with Xcode tools)
brew install git

# Install Go
brew install go

# Install Docker Desktop (includes Docker Compose)
brew install --cask docker

# Make is included with Xcode Command Line Tools
# curl is pre-installed on macOS
```

#### Step 4: Start Docker Desktop

1. Open **Docker Desktop** from Applications
2. Wait for Docker to start (whale icon in menu bar stops animating)
3. You may need to grant permissions when prompted

#### Step 5: Configure Go Environment

Add Go to your PATH (if not already done):

```bash
# Add to ~/.zshrc (or ~/.bashrc if using bash)
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

**macOS setup complete!** Skip to [Clone the Repository](#clone-the-repository).

---

### Linux Setup

These instructions are for **Ubuntu/Debian**. For other distributions, adapt the package manager commands.

#### Step 1: Update System Packages

```bash
sudo apt update && sudo apt upgrade -y
```

#### Step 2: Install Essential Build Tools

```bash
sudo apt install -y build-essential curl wget git
```

This installs:
- `build-essential` - Includes Make and other build tools
- `curl` - For testing APIs
- `wget` - For downloading files
- `git` - Version control

#### Step 3: Install Go

```bash
# Download Go 1.24 (or latest from https://golang.org/dl/)
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz

# Remove any previous Go installation and extract
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz

# Clean up downloaded file
rm go1.24.0.linux-amd64.tar.gz
```

Add Go to your PATH:

```bash
# Add to ~/.bashrc (or ~/.zshrc if using zsh)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

Verify:

```bash
go version
# Expected: go version go1.24.x linux/amd64
```

#### Step 4: Install Docker Engine

```bash
# Install prerequisites
sudo apt install -y ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up the repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
```

#### Step 5: Configure Docker (Run Without sudo)

**Important:** Without this step, you'll need `sudo` for all Docker commands.

```bash
# Create docker group (may already exist)
sudo groupadd docker

# Add your user to the docker group
sudo usermod -aG docker $USER

# Activate the changes (or log out and log back in)
newgrp docker
```

Verify Docker works without sudo:

```bash
docker run hello-world
# Should print "Hello from Docker!"
```

#### Step 6: Start Docker on Boot

```bash
sudo systemctl enable docker
sudo systemctl start docker
```

**Linux setup complete!** Skip to [Clone the Repository](#clone-the-repository).

---

### Windows Setup

**Important:** This project requires **WSL2** (Windows Subsystem for Linux) for proper compatibility.

#### Step 1: Enable WSL2

Open **PowerShell as Administrator** and run:

```powershell
wsl --install
```

This installs WSL2 with Ubuntu by default. **Restart your computer** when prompted.

After restart, Ubuntu will automatically open and ask you to create a username and password.

#### Step 2: Update Ubuntu in WSL

Open **Ubuntu** from the Start menu:

```bash
sudo apt update && sudo apt upgrade -y
```

#### Step 3: Install Docker Desktop for Windows

1. Download [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop/)
2. Run the installer
3. **Important:** During installation, ensure "Use WSL 2 instead of Hyper-V" is checked
4. Restart your computer if prompted

After restart:
1. Open Docker Desktop
2. Go to Settings → Resources → WSL Integration
3. Enable integration with your Ubuntu distro
4. Click "Apply & Restart"

#### Step 4: Install Development Tools in WSL

Open **Ubuntu** terminal and follow the [Linux Setup](#linux-setup) steps starting from Step 1.

**All development should be done inside WSL, not Windows PowerShell/CMD.**

#### Step 5: Clone Repository in WSL

Store your code in the WSL filesystem for better performance:

```bash
# Inside Ubuntu terminal
cd ~
mkdir -p projects
cd projects
# Clone repository here (see next section)
```

**Windows setup complete!** Continue to [Clone the Repository](#clone-the-repository).

---

## Install Git

If you haven't installed Git yet:

### macOS
```bash
brew install git
# Or: Git is included with Xcode Command Line Tools
```

### Linux (Ubuntu/Debian)
```bash
sudo apt install -y git
```

### Verify Git Installation
```bash
git --version
# Expected: git version 2.x.x
```

### Configure Git (First Time Only)

```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

---

## Install Go

### macOS
```bash
brew install go
```

### Linux
See [Linux Setup Step 3](#step-3-install-go) above.

### Verify Go Installation

```bash
go version
# Expected: go version go1.24.x or higher
```

### Configure Go Environment

Ensure Go binaries are in your PATH:

```bash
# Check current GOPATH
go env GOPATH
# Usually: /Users/yourname/go (macOS) or /home/yourname/go (Linux)

# Add to your shell config (~/.zshrc, ~/.bashrc)
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

Verify:

```bash
echo $PATH | tr ':' '\n' | grep go
# Should show go binary paths
```

---

## Install Docker

### macOS
```bash
brew install --cask docker
# Then open Docker Desktop from Applications
```

### Linux
See [Linux Setup Step 4-6](#step-4-install-docker-engine) above.

### Windows
See [Windows Setup Step 3](#step-3-install-docker-desktop-for-windows) above.

### Verify Docker Installation

```bash
# Check Docker version
docker --version
# Expected: Docker version 24.x.x or later

# Check Docker Compose version
docker compose version
# Expected: Docker Compose version v2.x.x

# Test Docker is running
docker run hello-world
# Should print "Hello from Docker!"
```

### Ensure Docker is Running

If you get "Cannot connect to the Docker daemon":

**macOS:** Open Docker Desktop from Applications and wait for it to start.

**Linux:**
```bash
sudo systemctl start docker
```

---

## Install Make

### macOS
Make is included with Xcode Command Line Tools. Verify:
```bash
make --version
# Expected: GNU Make 3.x or 4.x
```

### Linux (Ubuntu/Debian)
```bash
sudo apt install -y build-essential
```

### Windows (WSL)
Inside Ubuntu WSL:
```bash
sudo apt install -y build-essential
```

---

## Clone the Repository

```bash
# Navigate to your projects directory
cd ~/projects  # or wherever you keep code

# Clone the repository
git clone https://github.com/your-org/internal-transfers-service.git

# Enter the project directory
cd internal-transfers-service
```

---

## Quick Start (Recommended)

For first-time setup, run the automated setup command:

```bash
make setup
```

This command will:
1. ✅ Install development tools (golangci-lint, mockgen, migrate, goimports)
2. ✅ Generate mock files for testing
3. ✅ Download Go dependencies
4. ✅ Start PostgreSQL via Docker
5. ✅ Run database migrations

After setup completes, start the service:

```bash
make run
```

The service will start on:
- **Main API**: http://localhost:8080
- **Ops (health/metrics)**: http://localhost:8081

### Test the Service

```bash
# Health check
curl http://localhost:8081/health/live
# Expected: {"status":"SERVING"}

# Create an account (note: all API endpoints use /v1 prefix)
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "1000.00"}'
# Expected: 201 Created

# Get account
curl http://localhost:8080/v1/accounts/1
# Expected: {"account_id":1,"balance":"1000"}
```

**Setup complete!** 🎉

---

## Manual Setup (Step-by-Step)

If `make setup` fails or you prefer manual setup:

### Step 1: Install Development Tools

```bash
make deps-install
```

This installs:
- `golangci-lint` - Code linter
- `mockgen` - Mock generation for tests
- `migrate` - Database migration tool
- `goimports` - Import formatting

**Alternative (Homebrew on macOS):**
```bash
brew install golangci-lint golang-migrate
go install go.uber.org/mock/mockgen@latest
go install golang.org/x/tools/cmd/goimports@latest
```

### Step 2: Generate Mocks

Mocks must be generated before downloading dependencies (test files import mock packages).

```bash
make mock
```

### Step 3: Download Go Dependencies

```bash
make deps
```

### Step 4: Start PostgreSQL

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

### Step 5: Run Database Migrations

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

### Step 6: Run Tests

```bash
make test-short
```

All tests should pass.

### Step 7: Run the Service

```bash
make run
```

---

## Verify Installation

Run these commands to verify everything is set up correctly:

```bash
# 1. Check all prerequisites
echo "=== Prerequisites ===" && \
git --version && \
go version && \
docker --version && \
docker compose version && \
make --version && \
curl --version | head -1

# 2. Check Docker is running
docker ps

# 3. Check PostgreSQL container
make docker-status

# 4. Run tests
make test-short

# 5. Start service and test
make run &
sleep 3
curl http://localhost:8081/health/ready
```

---

## IDE Setup (Recommended)

### Visual Studio Code

1. Download [VS Code](https://code.visualstudio.com/)

2. Install the **Go extension**:
   - Open VS Code
   - Press `Cmd+Shift+X` (macOS) or `Ctrl+Shift+X` (Linux)
   - Search for "Go" by the Go Team at Google
   - Click Install

3. Install Go tools when prompted (or manually):
   ```bash
   # VS Code will prompt to install these
   go install golang.org/x/tools/gopls@latest
   go install github.com/go-delve/delve/cmd/dlv@latest
   ```

4. Open the project:
   ```bash
   code ~/projects/internal-transfers-service
   ```

### GoLand (JetBrains)

1. Download [GoLand](https://www.jetbrains.com/go/)
2. Open the project directory
3. GoLand will automatically detect the Go project and configure it

### Recommended VS Code Settings

Create `.vscode/settings.json` in the project (if not exists):

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "editor.formatOnSave": true,
  "go.testFlags": ["-v"],
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  }
}
```

---

## Troubleshooting

### Docker: "Cannot connect to the Docker daemon"

**Cause:** Docker is not running.

**Fix:**
- **macOS:** Open Docker Desktop from Applications and wait for it to start
- **Linux:** Run `sudo systemctl start docker`
- **Windows:** Open Docker Desktop and ensure WSL integration is enabled

### Port Already in Use (8080, 8081, or 5432)

**Cause:** Another service is using the port.

**Find what's using the port:**
```bash
# macOS/Linux
lsof -i :8080
lsof -i :5432

# Kill the process
kill -9 <PID>
```

**Or change the port in config:**

Edit `config/dev.toml`:
```toml
[app]
port = ":8090"      # Change from 8080
ops_port = ":8091"  # Change from 8081
```

For PostgreSQL, edit `deployment/dev/docker-compose.yml`:
```yaml
ports:
  - "5433:5432"  # Use 5433 on host
```

And update `config/dev.toml`:
```toml
[database]
port = 5433
```

### Go: "command not found: go"

**Cause:** Go is not in your PATH.

**Fix:**
```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$(go env GOPATH)/bin

# Reload
source ~/.zshrc  # or ~/.bashrc
```

### Make: "command not found: make"

**macOS:**
```bash
xcode-select --install
```

**Linux:**
```bash
sudo apt install -y build-essential
```

### Linux: "permission denied" for Docker

**Cause:** User not in docker group.

**Fix:**
```bash
sudo usermod -aG docker $USER
newgrp docker
# Or log out and log back in
```

### Migrations: "dirty database version"

**Cause:** A migration failed halfway.

**Fix:**
```bash
# Check current version
make migrate-version

# Force to specific version (e.g., 2)
migrate -path internal/database/migrations -database "postgres://postgres:postgres@localhost:5432/transfers?sslmode=disable" force 2

# Retry migrations
make migrate-up
```

### Tests: "mock files not found"

**Cause:** Mocks haven't been generated.

**Fix:**
```bash
make mock
```

### Windows: "make is not recognized"

**Cause:** Running in PowerShell instead of WSL.

**Fix:** Open Ubuntu (WSL) and run commands there:
```bash
wsl
cd ~/projects/internal-transfers-service
make setup
```

### Firewall Blocking Docker

**Symptoms:** Docker pulls fail, containers can't connect.

**Fix:** Add exceptions for Docker in your firewall/antivirus settings, or temporarily disable it for testing.

---

## Next Steps

- [API Reference](api-reference.md) - Learn about all available endpoints
- [Development Guide](development.md) - Learn about development commands
- [Database Guide](database.md) - Learn how to inspect the database
- [Configuration Guide](configuration.md) - Customize settings
- [Architecture](architecture.md) - Understand the system design
