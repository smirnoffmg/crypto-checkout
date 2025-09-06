package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"crypto-checkout/pkg/config"
)

func TestConfig(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv

	tests := []struct {
		name     string
		envVars  map[string]string
		expected config.Config
	}{
		{
			name:    "should load default configuration",
			envVars: map[string]string{},
			expected: config.Config{
				Server: config.ServerConfig{
					Port: 8080,
					Host: "localhost",
				},
				Log: config.LogConfig{
					Level: "info",
				},
			},
		},
		{
			name: "should load configuration from environment variables",
			envVars: map[string]string{
				"CRYPTO_CHECKOUT_SERVER_PORT": "9090",
				"CRYPTO_CHECKOUT_SERVER_HOST": "0.0.0.0",
				"CRYPTO_CHECKOUT_LOG_LEVEL":   "debug",
			},
			expected: config.Config{
				Server: config.ServerConfig{
					Port: 9090,
					Host: "0.0.0.0",
				},
				Log: config.LogConfig{
					Level: "debug",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cannot use t.Parallel() with t.Setenv

			// Clear all CRYPTO_CHECKOUT environment variables
			_ = os.Unsetenv("CRYPTO_CHECKOUT_SERVER_PORT")
			_ = os.Unsetenv("CRYPTO_CHECKOUT_SERVER_HOST")
			_ = os.Unsetenv("CRYPTO_CHECKOUT_LOG_LEVEL")

			// Set test environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Restore environment after test
			defer func() {
				// Clear test environment variables
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			cfg, err := loadConfigForTest(tt.envVars)
			require.NoError(t, err)

			assert.Equal(t, tt.expected.Server.Port, cfg.Server.Port)
			assert.Equal(t, tt.expected.Server.Host, cfg.Server.Host)
			assert.Equal(t, tt.expected.Log.Level, cfg.Log.Level)
		})
	}
}

// loadConfigForTest loads configuration for testing without reading config files.
func loadConfigForTest(envVars map[string]string) (*config.Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", config.DefaultServerPort)
	v.SetDefault("server.host", config.DefaultServerHost)
	v.SetDefault("log.level", config.DefaultLogLevel)

	// Enable reading from environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("CRYPTO_CHECKOUT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set test environment variables
	for key, value := range envVars {
		v.Set(key, value)
	}

	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func TestNewConfig(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "info", cfg.Log.Level)
}
