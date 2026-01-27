# =============================================================================
# Internal Transfers Service - Development Dockerfile
# =============================================================================
# Optimized for development with debugging support and faster rebuilds
#
# Build: docker build -t internal-transfers-service:dev -f deployment/dev/Dockerfile .
# Run:   docker run -p 8080:8080 -p 8081:8081 internal-transfers-service:dev
# =============================================================================

FROM golang:1.21-alpine

# Install development dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    curl \
    bash

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (with debug symbols for development)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /app/bin/internal-transfers-service \
    ./cmd/api

# Expose ports
# 8080: Main API
# 8081: Ops (health, metrics)
EXPOSE 8080 8081

# Set environment defaults for development
ENV APP_ENV=dev \
    APP_APP_PORT=:8080 \
    APP_APP_OPS_PORT=:8081

# Health check
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/health/live || exit 1

# Run the application
CMD ["/app/bin/internal-transfers-service"]
