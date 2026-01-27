// Package config provides configuration management for the application.
package config

import (
	"fmt"
	"time"

	pkgconfig "github.com/internal-transfers-service/pkg/config"
)

// Environment constants
const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

// Config holds all application configuration
type Config struct {
	App         AppConfig         `mapstructure:"app"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Metrics     MetricsConfig     `mapstructure:"metrics"`
	Idempotency IdempotencyConfig `mapstructure:"idempotency"`
	Security    SecurityConfig    `mapstructure:"security"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limit"`
	Tracing     TracingConfig     `mapstructure:"tracing"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Env             string `mapstructure:"env"`
	Name            string `mapstructure:"name"`
	Port            string `mapstructure:"port"`
	OpsPort         string `mapstructure:"ops_port"`
	ShutdownDelay   int    `mapstructure:"shutdown_delay"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string              `mapstructure:"host"`
	Port            int                 `mapstructure:"port"`
	User            string              `mapstructure:"user"`
	Password        string              `mapstructure:"password"`
	Name            string              `mapstructure:"name"`
	SSLMode         string              `mapstructure:"ssl_mode"`
	MaxConnections  int32               `mapstructure:"max_connections"`
	MinConnections  int32               `mapstructure:"min_connections"`
	MaxConnLifetime string              `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime string              `mapstructure:"max_conn_idle_time"`
	Retry           DatabaseRetryConfig `mapstructure:"retry"`
}

// DatabaseRetryConfig holds database connection retry configuration
type DatabaseRetryConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	MaxRetries     int    `mapstructure:"max_retries"`
	InitialBackoff string `mapstructure:"initial_backoff"`
	MaxBackoff     string `mapstructure:"max_backoff"`
}

// GetInitialBackoff returns the initial backoff duration
func (c *DatabaseRetryConfig) GetInitialBackoff() time.Duration {
	d, err := time.ParseDuration(c.InitialBackoff)
	if err != nil {
		return time.Second
	}
	return d
}

// GetMaxBackoff returns the max backoff duration
func (c *DatabaseRetryConfig) GetMaxBackoff() time.Duration {
	d, err := time.ParseDuration(c.MaxBackoff)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// IdempotencyConfig holds idempotency configuration
type IdempotencyConfig struct {
	TTL string `mapstructure:"ttl"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	CORSAllowOrigin string `mapstructure:"cors_allow_origin"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool    `mapstructure:"enabled"`
	RequestsPerSec float64 `mapstructure:"requests_per_second"`
	BurstSize      int     `mapstructure:"burst_size"`
}

// TracingConfig holds distributed tracing configuration
type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	Endpoint     string  `mapstructure:"endpoint"`
	ServiceName  string  `mapstructure:"service_name"`
	SampleRate   float64 `mapstructure:"sample_rate"`
	Insecure     bool    `mapstructure:"insecure"`
	BatchTimeout string  `mapstructure:"batch_timeout"`
}

// C is the global configuration instance
var C *Config

// Load loads configuration for the specified environment.
// It first loads default.toml, then merges environment-specific overrides.
// Environment is determined by APP_ENV env var or defaults to "dev".
func Load() (*Config, error) {
	env := pkgconfig.GetEnv()
	return LoadForEnv(env)
}

// LoadForEnv loads configuration for a specific environment.
// It first loads default.toml, then merges the env-specific config (e.g., dev.toml).
func LoadForEnv(env string) (*Config, error) {
	loader := pkgconfig.NewDefaultConfig()

	var cfg Config
	if err := loader.Load(env, &cfg, "APP"); err != nil {
		return nil, fmt.Errorf("failed to load config for env %s: %w", env, err)
	}

	C = &cfg
	return &cfg, nil
}

// LoadFromPath loads configuration from a custom path (for backward compatibility).
// Deprecated: Use Load() or LoadForEnv() instead.
func LoadFromPath(configPath string, env string) (*Config, error) {
	opts := pkgconfig.NewOptions(pkgconfig.DefaultConfigType, configPath, pkgconfig.DefaultConfigFileName)
	loader := pkgconfig.NewConfig(opts)

	var cfg Config
	if err := loader.Load(env, &cfg, "APP"); err != nil {
		return nil, fmt.Errorf("failed to load config from %s for env %s: %w", configPath, env, err)
	}

	C = &cfg
	return &cfg, nil
}

// GetConnectionString returns the PostgreSQL connection string
func (c *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// GetMaxConnLifetime returns the max connection lifetime as a duration
func (c *DatabaseConfig) GetMaxConnLifetime() time.Duration {
	d, err := time.ParseDuration(c.MaxConnLifetime)
	if err != nil {
		return time.Hour
	}
	return d
}

// GetMaxConnIdleTime returns the max connection idle time as a duration
func (c *DatabaseConfig) GetMaxConnIdleTime() time.Duration {
	d, err := time.ParseDuration(c.MaxConnIdleTime)
	if err != nil {
		return 30 * time.Minute
	}
	return d
}
