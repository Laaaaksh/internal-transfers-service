# Internal Transfers Service Makefile

.PHONY: all build run test test-coverage lint clean docker-up docker-down migrate-up migrate-down mock help

# Variables
BINARY_NAME=internal-transfers-service
BUILD_DIR=bin
MIGRATION_DIR=internal/database/migrations
DB_URL=postgres://postgres:postgres@localhost:5432/transfers?sslmode=disable

# Default target
all: lint test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@go run ./cmd/api

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Start Docker containers
docker-up:
	@echo "Starting Docker containers..."
	@docker-compose -f deployment/dev/docker-compose.yml up -d

# Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose -f deployment/dev/docker-compose.yml down

# Run database migrations up
migrate-up:
	@echo "Running migrations..."
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up

# Run database migrations down
migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $$name

# Generate mocks
mock:
	@echo "Generating mocks..."
	@mockgen -source=internal/modules/account/repository.go -destination=internal/modules/account/mock/repository.go -package=mock
	@mockgen -source=internal/modules/account/core.go -destination=internal/modules/account/mock/core.go -package=mock
	@mockgen -source=internal/modules/transaction/repository.go -destination=internal/modules/transaction/mock/repository.go -package=mock
	@mockgen -source=internal/modules/transaction/core.go -destination=internal/modules/transaction/mock/core.go -package=mock

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	@go mod verify

# Show help
help:
	@echo "Available targets:"
	@echo "  all           - Run lint, test, and build"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-up     - Start Docker containers"
	@echo "  docker-down   - Stop Docker containers"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  migrate-create- Create a new migration"
	@echo "  mock          - Generate mocks"
	@echo "  deps          - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  verify        - Verify dependencies"
	@echo "  help          - Show this help message"
