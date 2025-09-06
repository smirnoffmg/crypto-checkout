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
	// DefaultLogDir is the default log directory.
	DefaultLogDir = "logs"
	// DefaultPostgresPort is the default PostgreSQL port.
	DefaultPostgresPort = 5432
)

// Config represents the application configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// LogConfig represents logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"`
	Dir   string `mapstructure:"dir"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	URL      string `mapstructure:"url"`
}

// Load loads configuration using Viper with support for multiple sources.
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", DefaultServerPort)
	v.SetDefault("server.host", DefaultServerHost)
	v.SetDefault("log.level", DefaultLogLevel)
	v.SetDefault("log.dir", DefaultLogDir)
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", DefaultPostgresPort)
	v.SetDefault("database.user", "crypto_user")
	v.SetDefault("database.password", "crypto_password")
	v.SetDefault("database.dbname", "crypto_checkout")
	v.SetDefault("database.sslmode", "disable")

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
			Dir:   DefaultLogDir,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     DefaultPostgresPort,
			User:     "crypto_user",
			Password: "crypto_password",
			DBName:   "crypto_checkout",
			SSLMode:  "disable",
		},
	}
}

// NewConfigProvider creates a new configuration provider for Fx.
func NewConfigProvider() (*Config, error) {
	return Load()
}
