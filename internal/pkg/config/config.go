// Package config provides configuration management for the crypto-checkout application.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	// DefaultServerPort is the default server port.
	DefaultServerPort = 8080
	// DefaultServerHost is the default server host.
	DefaultServerHost = "localhost"
	// DefaultLogLevel is the default log level.
	DefaultLogLevel = "info"
)

// Config represents the application configuration.
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Log    LogConfig    `mapstructure:"log"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// LogConfig represents logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// Load loads configuration using Viper with support for multiple sources.
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", DefaultServerPort)
	v.SetDefault("server.host", DefaultServerHost)
	v.SetDefault("log.level", DefaultLogLevel)

	// Set config file name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/crypto-checkout")

	// Enable reading from environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("CRYPTO_CHECKOUT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// NewConfig creates a new configuration with default values.
func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: DefaultServerPort,
			Host: DefaultServerHost,
		},
		Log: LogConfig{
			Level: DefaultLogLevel,
		},
	}
}

// NewConfigProvider creates a new configuration provider for Fx.
func NewConfigProvider() (*Config, error) {
	return Load()
}
