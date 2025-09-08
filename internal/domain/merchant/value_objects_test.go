package merchant

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyHash(t *testing.T) {
	t.Run("NewAPIKeyHash", func(t *testing.T) {
		t.Run("ValidKey", func(t *testing.T) {
			hash, err := NewAPIKeyHash("sk_test_abc123")
			require.NoError(t, err)
			assert.NotEmpty(t, hash.String())
		})

		t.Run("EmptyKey", func(t *testing.T) {
			hash, err := NewAPIKeyHash("")
			assert.Error(t, err)
			assert.Nil(t, hash)
		})
	})

	t.Run("Equals", func(t *testing.T) {
		hash1, _ := NewAPIKeyHash("sk_test_abc123")
		hash2, _ := NewAPIKeyHash("sk_test_abc123")
		hash3, _ := NewAPIKeyHash("sk_test_def456")

		assert.True(t, hash1.Equals(hash2))
		assert.False(t, hash1.Equals(hash3))
		assert.False(t, hash1.Equals(nil))
	})
}

func TestWebhookEndpointConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    WebhookEndpointConfig
		expectErr bool
	}{
		{
			name: "Valid config",
			config: WebhookEndpointConfig{
				URL:          "https://example.com/webhook",
				Events:       []string{"invoice.paid"},
				Secret:       "secret123",
				MaxRetries:   5,
				RetryBackoff: BackoffStrategyExponential,
				Timeout:      30,
			},
			expectErr: false,
		},
		{
			name: "Empty URL",
			config: WebhookEndpointConfig{
				URL:    "",
				Events: []string{"invoice.paid"},
			},
			expectErr: true,
		},
		{
			name: "No events",
			config: WebhookEndpointConfig{
				URL:    "https://example.com/webhook",
				Events: []string{},
			},
			expectErr: true,
		},
		{
			name: "Invalid max retries",
			config: WebhookEndpointConfig{
				URL:        "https://example.com/webhook",
				Events:     []string{"invoice.paid"},
				MaxRetries: 15,
			},
			expectErr: true,
		},
		{
			name: "Invalid timeout",
			config: WebhookEndpointConfig{
				URL:     "https://example.com/webhook",
				Events:  []string{"invoice.paid"},
				Timeout: 100,
			},
			expectErr: true,
		},
		{
			name: "Invalid backoff strategy",
			config: WebhookEndpointConfig{
				URL:          "https://example.com/webhook",
				Events:       []string{"invoice.paid"},
				RetryBackoff: BackoffStrategy("invalid"),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
