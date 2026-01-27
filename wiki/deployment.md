# Deployment Guide

This guide covers deployment of the Internal Transfers Service.

## Dockerfile Locations

The project has separate Dockerfiles for different environments:

| Environment | Dockerfile | Purpose |
|-------------|------------|---------|
| Development | `deployment/dev/Dockerfile` | Includes debug tools, faster builds |
| Production | `deployment/prod/Dockerfile` | Multi-stage, optimized, minimal |

---

## Development Docker Setup

### Build Development Image

```bash
# From project root
docker build -t internal-transfers-service:dev -f deployment/dev/Dockerfile .
```

### Local Development with Docker Compose

The `deployment/dev/docker-compose.yml` starts PostgreSQL for local development:

```bash
# Start PostgreSQL
docker-compose -f deployment/dev/docker-compose.yml up -d

# Check status
docker-compose -f deployment/dev/docker-compose.yml ps

# Stop
docker-compose -f deployment/dev/docker-compose.yml down
```

---

## Production Docker Deployment

### Build Production Image

```bash
# Build optimized production image
docker build -t internal-transfers-service:latest -f deployment/prod/Dockerfile .

# Build with specific tag
docker build -t internal-transfers-service:v1.0.0 -f deployment/prod/Dockerfile .
```

### Run Production Container

```bash
docker run -d \
  --name internal-transfers-service \
  -p 8080:8080 \
  -p 8081:8081 \
  -e APP_ENV=prod \
  -e APP_DATABASE_HOST=your-db-host \
  -e APP_DATABASE_USER=your-db-user \
  -e APP_DATABASE_PASSWORD=your-db-password \
  -e APP_DATABASE_NAME=transfers \
  internal-transfers-service:latest
```

### Production Image Features

The production Dockerfile (`deployment/prod/Dockerfile`) includes:

- **Multi-stage build** - Separate build and runtime stages
- **Minimal base image** - Alpine Linux (~5MB)
- **Non-root user** - Runs as unprivileged user for security
- **Static binary** - No external dependencies
- **Stripped symbols** - Smaller binary size
- **Health checks** - Built-in container health monitoring
- **Reproducible builds** - Trimmed paths for consistent builds

---

## Kubernetes Deployment

### Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: internal-transfers-service
  labels:
    app: internal-transfers-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: internal-transfers-service
  template:
    metadata:
      labels:
        app: internal-transfers-service
    spec:
      containers:
      - name: internal-transfers-service
        image: internal-transfers-service:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8081
          name: ops
        env:
        - name: APP_ENV
          value: "prod"
        - name: APP_DATABASE_HOST
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: host
        - name: APP_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
```

### Service Manifest

```yaml
apiVersion: v1
kind: Service
metadata:
  name: internal-transfers-service
spec:
  selector:
    app: internal-transfers-service
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: ops
    port: 8081
    targetPort: 8081
  type: ClusterIP
```

---

## Environment Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| APP_ENV | Environment name | prod |
| APP_DATABASE_HOST | PostgreSQL host | db.internal |
| APP_DATABASE_USER | Database user | app_user |
| APP_DATABASE_PASSWORD | Database password | (secret) |
| APP_DATABASE_NAME | Database name | transfers |

### Optional Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| APP_APP_PORT | :8080 | Main API port |
| APP_APP_OPS_PORT | :8081 | Ops/health port |
| APP_DATABASE_SSL_MODE | require | SSL mode |
| APP_DATABASE_MAX_CONNECTIONS | 50 | Max pool size |
| APP_LOGGING_LEVEL | info | Log level |
| TRACING_ENABLED | false | Enable OpenTelemetry tracing |
| OTEL_EXPORTER_OTLP_ENDPOINT | localhost:4317 | OTLP collector endpoint |
| APP_RATE_LIMIT_REQUESTS_PER_SECOND | 100 | Rate limit (requests/sec) |
| APP_RATE_LIMIT_BURST_SIZE | 200 | Rate limit burst size |
| CORS_ALLOW_ORIGIN | * | CORS allowed origin |

---

## Health Checks

### Liveness Probe

```
GET /health/live
```

Returns 200 if the process is running. Use this for Kubernetes liveness probes.

### Readiness Probe

```
GET /health/ready
```

Returns 200 if the service is ready to accept traffic (database connected). Use this for Kubernetes readiness probes.

---

## Graceful Shutdown

The service supports graceful shutdown:

1. Receives SIGTERM/SIGINT
2. Marks service as unhealthy (readiness returns 503)
3. Waits `shutdown_delay` seconds for load balancer drain
4. Stops accepting new requests
5. Waits up to `shutdown_timeout` seconds for in-flight requests
6. Closes database connections
7. Exits

Configure these values based on your environment:
- `APP_APP_SHUTDOWN_DELAY=10` (seconds to wait for LB drain)
- `APP_APP_SHUTDOWN_TIMEOUT=60` (max shutdown wait time)

---

## Database Migrations

Run migrations before deploying:

```bash
migrate -path internal/database/migrations \
  -database "postgres://user:pass@host:5432/transfers?sslmode=require" \
  up
```

Or use a Kubernetes Job:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: migrate/migrate:latest
        command:
        - /bin/sh
        - -c
        - |
          migrate -path /migrations \
            -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:5432/${DB_NAME}?sslmode=require" \
            up
        envFrom:
        - secretRef:
            name: db-credentials
        volumeMounts:
        - name: migrations
          mountPath: /migrations
      volumes:
      - name: migrations
        configMap:
          name: db-migrations
      restartPolicy: OnFailure
```

---

## Monitoring

### Prometheus Metrics

Scrape metrics from `/metrics` on the ops port:

```yaml
- job_name: 'internal-transfers-service'
  static_configs:
  - targets: ['internal-transfers-service:8081']
  metrics_path: '/metrics'
```

### Key Metrics to Monitor

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| http_request_duration_seconds | Request latency | p99 > 500ms |
| http_requests_total{status=~"5.."} | Error rate | > 1% of total |
| transfers_total{status="error"} | Failed transfers | Any increase |
| db_connections_active | Active DB connections | > 80% of max |

---

## Production Checklist

### Core
- [ ] Database migrations applied
- [ ] Environment variables configured
- [ ] SSL enabled for database (`ssl_mode=require`)
- [ ] Health checks configured in orchestrator
- [ ] Resource limits set

### Security
- [ ] Rate limiting enabled (`APP_RATE_LIMIT_ENABLED=true`)
- [ ] CORS configured for specific origin (`CORS_ALLOW_ORIGIN`)
- [ ] Secrets management in place
- [ ] Network policies configured

### Observability
- [ ] Prometheus scraping configured
- [ ] Distributed tracing enabled (`TRACING_ENABLED=true`)
- [ ] OTLP collector endpoint configured
- [ ] Log aggregation configured (JSON format)

### Resilience
- [ ] Database connection retry enabled
- [ ] Horizontal pod autoscaling configured (optional)
- [ ] Backup strategy for database
