// Package config provides configuration management for the application.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	App         AppConfig      `mapstructure:"app"`
	Database    DatabaseConfig `mapstructure:"database"`
	Logging     LoggingConfig  `mapstructure:"logging"`
	Metrics     MetricsConfig  `mapstructure:"metrics"`
	Idempotency IdempotencyConfig `mapstructure:"idempotency"`
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
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxConnections  int32  `mapstructure:"max_connections"`
	MinConnections  int32  `mapstructure:"min_connections"`
	MaxConnLifetime string `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime string `mapstructure:"max_conn_idle_time"`
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

// C is the global configuration instance
var C *Config

// Load loads configuration from the specified path
func Load(configPath string, configName string) (*Config, error) {
	v := viper.New()

	v.SetConfigName(configName)
	v.SetConfigType("toml")
	v.AddConfigPath(configPath)

	// Set defaults
	setDefaults(v)

	// Enable environment variable overrides
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	C = &cfg
	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.name", "internal-transfers-service")
	v.SetDefault("app.port", ":8080")
	v.SetDefault("app.ops_port", ":8081")
	v.SetDefault("app.shutdown_delay", 5)
	v.SetDefault("app.shutdown_timeout", 30)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "transfers")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_connections", 25)
	v.SetDefault("database.min_connections", 5)
	v.SetDefault("database.max_conn_lifetime", "1h")
	v.SetDefault("database.max_conn_idle_time", "30m")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")

	// Idempotency defaults
	v.SetDefault("idempotency.ttl", "24h")
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
