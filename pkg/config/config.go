// Package config provides configuration loading with environment-specific overrides.
//
// Usage:
//   - Application should have a directory holding default.toml and environment
//     specific files (dev.toml, test.toml, prod.toml).
//   - Use NewDefaultConfig().Load("dev", &config) to load config for dev environment.
//   - The loader first loads default.toml, then merges environment-specific overrides.
package config

import (
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Default options for configuration loading.
const (
	DefaultConfigType     = "toml"
	DefaultConfigDir      = "./config"
	DefaultConfigFileName = "default"
	WorkDirEnv            = "WORKDIR"
	AppEnvKey             = "APP_ENV"
	DefaultEnv            = "dev"
)

// Options holds configuration loading options.
type Options struct {
	configType            string
	configPath            string
	defaultConfigFileName string
}

// Config is a wrapper over viper for configuration loading.
type Config struct {
	opts  Options
	viper *viper.Viper
}

// NewDefaultOptions returns default configuration options.
// It uses WORKDIR env var if set, otherwise uses relative path from this file.
func NewDefaultOptions() Options {
	var configPath string
	workDir := os.Getenv(WorkDirEnv)
	if workDir != "" {
		configPath = path.Join(workDir, DefaultConfigDir)
	} else {
		_, thisFile, _, ok := runtime.Caller(1)
		if ok {
			configPath = path.Join(path.Dir(thisFile), "../../"+DefaultConfigDir)
		} else {
			configPath = DefaultConfigDir
		}
	}
	return NewOptions(DefaultConfigType, configPath, DefaultConfigFileName)
}

// NewOptions creates new Options with specified values.
func NewOptions(configType string, configPath string, defaultConfigFileName string) Options {
	return Options{
		configType:            configType,
		configPath:            configPath,
		defaultConfigFileName: defaultConfigFileName,
	}
}

// NewDefaultConfig creates a new Config with default options.
func NewDefaultConfig() *Config {
	return NewConfig(NewDefaultOptions())
}

// NewConfig creates a new Config with specified options.
func NewConfig(opts Options) *Config {
	return &Config{
		opts:  opts,
		viper: viper.New(),
	}
}

// Load reads default configuration and environment-specific overrides,
// then unmarshals into the provided config struct.
func (c *Config) Load(env string, config interface{}, prefix string) error {
	// First load the default configuration
	if err := c.loadByConfigName(c.opts.defaultConfigFileName, config, prefix); err != nil {
		return err
	}
	// Then load environment-specific overrides
	return c.loadByConfigName(env, config, prefix)
}

// loadByConfigName reads configuration from file, expands environment variable
// templates, and unmarshals into config.
func (c *Config) loadByConfigName(configName string, config interface{}, prefix string) error {
	c.viper.SetEnvPrefix(strings.ToUpper(prefix))
	c.viper.SetConfigName(configName)
	c.viper.SetConfigType(c.opts.configType)
	c.viper.AddConfigPath(c.opts.configPath)
	c.viper.AutomaticEnv()
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := c.viper.ReadInConfig(); err != nil {
		return err
	}

	configFile := c.viper.ConfigFileUsed()

	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	// Expand environment variables in the config content
	expandedContent := expandEnvVars(content)

	if err := c.viper.ReadConfig(strings.NewReader(string(expandedContent))); err != nil {
		return err
	}

	return c.viper.Unmarshal(config)
}

// GetEnv returns the current environment from APP_ENV or default.
func GetEnv() string {
	env := os.Getenv(AppEnvKey)
	if env == "" {
		return DefaultEnv
	}
	return env
}
