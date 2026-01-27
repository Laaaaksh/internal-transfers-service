# Internal Transfers Service Makefile

.PHONY: all build run test test-coverage test-short lint clean docker-up docker-down docker-status \
        migrate-up migrate-down migrate-create mock mock-clean deps deps-install fmt verify help setup

# Variables
BINARY_NAME=internal-transfers-service
BUILD_DIR=bin
MIGRATION_DIR=internal/database/migrations
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= transfers
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Colors for output
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
NC     := \033[0m # No Color

# Default target
all: lint test build

## ==================== Build & Run ====================

# Build the application
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Run the application
run:
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	@go run ./cmd/api

# Run the application with hot reload (requires air: go install github.com/air-verse/air@latest)
run-dev:
	@echo "$(GREEN)Running $(BINARY_NAME) with hot reload...$(NC)"
	@air

## ==================== Testing ====================

# Run all tests with verbose output
test:
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -v -race ./...

# Run tests without verbose (faster output)
test-short:
	@echo "$(GREEN)Running tests (short)...$(NC)"
	@go test -race ./...

# Run tests with coverage
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

## ==================== Code Quality ====================

# Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@golangci-lint run ./...

# Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@gofmt -s -w .
	@goimports -w . 2>/dev/null || true

# Verify dependencies
verify:
	@echo "$(GREEN)Verifying dependencies...$(NC)"
	@go mod verify

## ==================== Clean ====================

# Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Clean all including mocks
clean-all: clean mock-clean
	@echo "$(YELLOW)All artifacts cleaned$(NC)"

## ==================== Docker ====================

# Start Docker containers
docker-up:
	@echo "$(GREEN)Starting Docker containers...$(NC)"
	@docker-compose -f deployment/dev/docker-compose.yml up -d
	@echo "$(GREEN)Waiting for PostgreSQL to be ready...$(NC)"
	@sleep 3
	@make docker-status

# Stop Docker containers
docker-down:
	@echo "$(YELLOW)Stopping Docker containers...$(NC)"
	@docker-compose -f deployment/dev/docker-compose.yml down

# Stop Docker containers and remove volumes
docker-clean:
	@echo "$(RED)Stopping Docker containers and removing volumes...$(NC)"
	@docker-compose -f deployment/dev/docker-compose.yml down -v

# Show Docker container status
docker-status:
	@echo "$(GREEN)Docker container status:$(NC)"
	@docker-compose -f deployment/dev/docker-compose.yml ps

# View PostgreSQL logs
docker-logs:
	@docker-compose -f deployment/dev/docker-compose.yml logs -f postgres

## ==================== Database ====================

# Run database migrations up
migrate-up:
	@echo "$(GREEN)Running migrations...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "$(GREEN)Migrations complete$(NC)"

# Run database migrations down (rollback all)
migrate-down:
	@echo "$(YELLOW)Rolling back all migrations...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down -all

# Rollback one migration
migrate-down-one:
	@echo "$(YELLOW)Rolling back one migration...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down 1

# Force set migration version (use with caution)
migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" force $$version

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $$name

# Show current migration version
migrate-version:
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" version

## ==================== Mock Generation ====================

# Generate all mocks (8 files total)
mock: mock-clean
	@echo "$(GREEN)Generating mocks...$(NC)"
	@echo "  Generating account mocks..."
	@mockgen -source=internal/modules/account/repository.go -destination=internal/modules/account/mock/mock_repository.go -package=mock
	@mockgen -source=internal/modules/account/core.go -destination=internal/modules/account/mock/mock_core.go -package=mock
	@echo "  Generating transaction mocks..."
	@mockgen -source=internal/modules/transaction/repository.go -destination=internal/modules/transaction/mock/mock_repository.go -package=mock
	@mockgen -source=internal/modules/transaction/core.go -destination=internal/modules/transaction/mock/mock_core.go -package=mock
	@echo "  Generating idempotency mocks..."
	@mockgen -source=internal/modules/idempotency/repository.go -destination=internal/modules/idempotency/mock/mock_repository.go -package=mock
	@echo "  Generating database mocks..."
	@mockgen -source=pkg/database/pool.go -destination=pkg/database/mock/mock_pool.go -package=mock
	@mockgen -destination=pkg/database/mock/mock_row.go -package=mock github.com/jackc/pgx/v5 Row
	@mockgen -destination=pkg/database/mock/mock_tx.go -package=mock github.com/jackc/pgx/v5 Tx
	@echo "$(GREEN)Mocks generated successfully (8 files)$(NC)"

# Clean generated mocks (removes all .go files in mock directories)
mock-clean:
	@echo "$(YELLOW)Cleaning generated mocks...$(NC)"
	@rm -f internal/modules/account/mock/*.go
	@rm -f internal/modules/transaction/mock/*.go
	@rm -f internal/modules/idempotency/mock/*.go
	@rm -f pkg/database/mock/*.go

## ==================== Dependencies ====================

# Download and tidy dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy

# Install development tools
deps-install:
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install go.uber.org/mock/mockgen@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)Development tools installed$(NC)"

## ==================== Setup (First Time) ====================

# Full setup for first-time development
# Order: install tools -> generate mocks -> tidy deps -> start db -> migrate
setup:
	@echo ""
	@echo "$(GREEN)============================================$(NC)"
	@echo "$(GREEN)  Internal Transfers Service - Setup$(NC)"
	@echo "$(GREEN)============================================$(NC)"
	@echo ""
	@echo "$(GREEN)[1/5] Installing development tools...$(NC)"
	@$(MAKE) deps-install --no-print-directory
	@echo ""
	@echo "$(GREEN)[2/5] Generating mock files...$(NC)"
	@$(MAKE) mock --no-print-directory
	@echo ""
	@echo "$(GREEN)[3/5] Downloading dependencies...$(NC)"
	@$(MAKE) deps --no-print-directory
	@echo ""
	@echo "$(GREEN)[4/5] Starting PostgreSQL database...$(NC)"
	@$(MAKE) docker-up --no-print-directory
	@echo ""
	@echo "$(GREEN)[5/5] Running database migrations...$(NC)"
	@echo "$(GREEN)Waiting for database to be ready...$(NC)"
	@sleep 5
	@$(MAKE) migrate-up --no-print-directory
	@echo ""
	@echo "$(GREEN)============================================$(NC)"
	@echo "$(GREEN)  Setup complete!$(NC)"
	@echo "$(GREEN)============================================$(NC)"
	@echo ""
	@echo "$(GREEN)Next steps:$(NC)"
	@echo "  $(YELLOW)make run$(NC)   - Start the service (API on :8080, Ops on :8081)"
	@echo "  $(YELLOW)make test$(NC)  - Run all tests"
	@echo "  $(YELLOW)make help$(NC)  - Show all available commands"
	@echo ""
	@echo "$(GREEN)Quick test:$(NC)"
	@echo "  curl http://localhost:8081/health/live"
	@echo ""

## ==================== Help ====================

# Show help
help:
	@echo ""
	@echo "$(GREEN)============================================$(NC)"
	@echo "$(GREEN)  Internal Transfers Service - Help$(NC)"
	@echo "$(GREEN)============================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Quick Start (New Users):$(NC)"
	@echo "  1. make setup    - First-time setup (installs tools, db, mocks)"
	@echo "  2. make run      - Start the service"
	@echo "  3. curl http://localhost:8081/health/live"
	@echo ""
	@echo "$(YELLOW)Build & Run:$(NC)"
	@echo "  make build         - Build the application binary"
	@echo "  make run           - Run the application"
	@echo "  make run-dev       - Run with hot reload (requires 'air')"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  make test          - Run all tests with verbose output"
	@echo "  make test-short    - Run tests without verbose"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo ""
	@echo "$(YELLOW)Code Quality:$(NC)"
	@echo "  make lint          - Run golangci-lint"
	@echo "  make fmt           - Format code with gofmt and goimports"
	@echo "  make verify        - Verify go.mod dependencies"
	@echo ""
	@echo "$(YELLOW)Docker:$(NC)"
	@echo "  make docker-up     - Start PostgreSQL container"
	@echo "  make docker-down   - Stop PostgreSQL container"
	@echo "  make docker-clean  - Stop and remove volumes"
	@echo "  make docker-status - Show container status"
	@echo "  make docker-logs   - View PostgreSQL logs"
	@echo ""
	@echo "$(YELLOW)Database:$(NC)"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback all migrations"
	@echo "  make migrate-down-one - Rollback one migration"
	@echo "  make migrate-create   - Create new migration files"
	@echo "  make migrate-version  - Show current migration version"
	@echo ""
	@echo "$(YELLOW)Mocks:$(NC)"
	@echo "  make mock          - Clean and regenerate all mocks"
	@echo "  make mock-clean    - Remove generated mock files"
	@echo ""
	@echo "$(YELLOW)Dependencies:$(NC)"
	@echo "  make deps          - Download and tidy go modules"
	@echo "  make deps-install  - Install development tools"
	@echo ""
	@echo "$(YELLOW)Setup:$(NC)"
	@echo "  make setup         - Full first-time setup (tools, deps, db, mocks)"
	@echo ""
	@echo "$(YELLOW)Clean:$(NC)"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make clean-all     - Clean all including mocks"
	@echo ""
