# =============================================================================
# Internal Transfers Service - Production Dockerfile
# =============================================================================
# Multi-stage build for minimal, secure production image
#
# Build: docker build -t internal-transfers-service:latest -f deployment/prod/Dockerfile .
# Run:   docker run -p 8080:8080 -p 8081:8081 internal-transfers-service:latest
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build Stage
# -----------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0: Static binary (no C dependencies)
# -ldflags: Strip debug info and disable symbol table for smaller binary
# -trimpath: Remove file system paths from binary for reproducible builds
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -trimpath \
    -o /app/bin/internal-transfers-service \
    ./cmd/api

# -----------------------------------------------------------------------------
# Stage 2: Production Stage
# -----------------------------------------------------------------------------
FROM alpine:3.19 AS production

# Add labels for container metadata
LABEL maintainer="Internal Transfers Service Team"
LABEL version="1.0"
LABEL description="Production-grade internal transfers microservice"

# Install runtime dependencies
# ca-certificates: For HTTPS connections
# tzdata: For timezone support
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user for security
# Running as non-root is a security best practice
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/internal-transfers-service /app/internal-transfers-service

# Copy configuration files
COPY --from=builder /app/config /app/config

# Set ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
# 8080: Main API
# 8081: Ops (health, metrics)
EXPOSE 8080 8081

# Health check
# Checks liveness endpoint every 30s, fails after 3 retries
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health/live || exit 1

# Set environment defaults
ENV APP_ENV=prod \
    APP_APP_PORT=:8080 \
    APP_APP_OPS_PORT=:8081

# Run the application
ENTRYPOINT ["/app/internal-transfers-service"]
