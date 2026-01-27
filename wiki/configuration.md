# Configuration Guide

This document describes the configuration system and all available settings.

## Configuration Loading

Configuration is loaded in layers:

1. **default.toml** - Base configuration with all default values
2. **{env}.toml** - Environment-specific overrides (dev.toml, test.toml, prod.toml)
3. **Environment Variables** - Final overrides using `APP_` prefix

The environment is determined by the `APP_ENV` variable (defaults to `dev`).

---

## Configuration Files

### default.toml (Base Configuration)

```toml
[app]
env = "${APP_ENV:-dev}"
name = "internal-transfers-service"
port = ":8080"
ops_port = ":8081"
shutdown_delay = 5
shutdown_timeout = 30

[database]
host = "${DB_HOST:-localhost}"
port = 5432
user = "${DB_USER:-postgres}"
password = "${DB_PASSWORD:-postgres}"
name = "${DB_NAME:-transfers}"
ssl_mode = "disable"
max_connections = 25
min_connections = 5
max_conn_lifetime = "1h"
max_conn_idle_time = "30m"

[logging]
level = "info"
format = "json"

[metrics]
enabled = true
path = "/metrics"

[idempotency]
ttl = "24h"
```

### dev.toml (Development)

```toml
[app]
env = "dev"

[database]
host = "localhost"
port = 5432
user = "postgres"
password = "postgres"
name = "transfers"
ssl_mode = "disable"
max_connections = 10
min_connections = 2

[logging]
level = "debug"
format = "console"

[metrics]
enabled = true
```

### prod.toml (Production)

```toml
[app]
env = "prod"
shutdown_delay = 10
shutdown_timeout = 60

[database]
host = "${DB_HOST}"
port = 5432
user = "${DB_USER}"
password = "${DB_PASSWORD}"
name = "${DB_NAME:-transfers}"
ssl_mode = "require"
max_connections = 50
min_connections = 10
max_conn_lifetime = "30m"
max_conn_idle_time = "10m"

[logging]
level = "info"
format = "json"

[metrics]
enabled = true

[idempotency]
ttl = "48h"
```

---

## Configuration Reference

### Application Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| app.env | string | dev | Environment name |
| app.name | string | internal-transfers-service | Service name |
| app.port | string | :8080 | Main API port |
| app.ops_port | string | :8081 | Ops/health port |
| app.shutdown_delay | int | 5 | Seconds to wait before shutdown |
| app.shutdown_timeout | int | 30 | Max seconds for graceful shutdown |

### Database Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| database.host | string | localhost | PostgreSQL host |
| database.port | int | 5432 | PostgreSQL port |
| database.user | string | postgres | Database user |
| database.password | string | postgres | Database password |
| database.name | string | transfers | Database name |
| database.ssl_mode | string | disable | SSL mode (disable/require/verify-full) |
| database.max_connections | int | 25 | Max pool connections |
| database.min_connections | int | 5 | Min pool connections |
| database.max_conn_lifetime | duration | 1h | Max connection lifetime |
| database.max_conn_idle_time | duration | 30m | Max idle time before closing |

### Logging Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| logging.level | string | info | Log level (debug/info/warn/error) |
| logging.format | string | json | Log format (json/console) |

### Metrics Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| metrics.enabled | bool | true | Enable Prometheus metrics |
| metrics.path | string | /metrics | Metrics endpoint path |

### Idempotency Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| idempotency.ttl | duration | 24h | How long to keep idempotency keys |

---

## Environment Variables

Any configuration value can be overridden using environment variables with the `APP_` prefix:

```bash
# Override database settings
export APP_DATABASE_HOST=production-db.internal
export APP_DATABASE_PASSWORD=secret-password
export APP_DATABASE_NAME=transfers_prod

# Override logging
export APP_LOGGING_LEVEL=warn

# Override ports
export APP_APP_PORT=:3000
export APP_APP_OPS_PORT=:3001
```

### Variable Naming Convention

Convert the TOML path to environment variable:
- Replace `.` with `_`
- Convert to UPPERCASE
- Add `APP_` prefix

Examples:
| TOML Path | Environment Variable |
|-----------|---------------------|
| database.host | APP_DATABASE_HOST |
| database.max_connections | APP_DATABASE_MAX_CONNECTIONS |
| logging.level | APP_LOGGING_LEVEL |
| app.shutdown_timeout | APP_APP_SHUTDOWN_TIMEOUT |

---

## Template Variables

Configuration files support environment variable expansion with defaults:

```toml
# Use DB_HOST if set, otherwise use "localhost"
host = "${DB_HOST:-localhost}"

# Require DB_PASSWORD (no default)
password = "${DB_PASSWORD}"
```

Syntax:
- `${VAR}` - Use environment variable VAR (empty if not set)
- `${VAR:-default}` - Use VAR if set, otherwise use "default"

---

## Switching Environments

### Set Environment

```bash
# Use development configuration
export APP_ENV=dev

# Use production configuration
export APP_ENV=prod

# Use test configuration
export APP_ENV=test
```

### Verify Configuration

Start the service and check the logs for configuration values:

```bash
APP_ENV=dev make run
```

The startup log shows:
```json
{"level":"info","msg":"Starting service","name":"internal-transfers-service","env":"dev","port":":8080","ops_port":":8081"}
{"level":"info","msg":"Database connection pool initialized","host":"localhost","port":5432,"database":"transfers","max_connections":10}
```

---

## Production Configuration Checklist

Before deploying to production:

- [ ] Set `APP_ENV=prod`
- [ ] Set `APP_DATABASE_HOST` to production database
- [ ] Set `APP_DATABASE_PASSWORD` (use secrets management)
- [ ] Ensure `ssl_mode = "require"` or `verify-full`
- [ ] Adjust `max_connections` based on load
- [ ] Set `shutdown_delay` for load balancer drain time
- [ ] Verify `logging.level = "info"` (not debug)
- [ ] Ensure `logging.format = "json"` for log aggregation
