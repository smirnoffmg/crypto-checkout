package merchant

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// MerchantSettings represents configuration preferences for a merchant.
type MerchantSettings struct {
	DefaultCurrency       string                 `json:"default_currency"`
	DefaultCryptoCurrency string                 `json:"default_crypto_currency"`
	InvoiceExpiryMinutes  int                    `json:"invoice_expiry_minutes"`
	FeePercentage         float64                `json:"fee_percentage"` // 0.0-10.0% as per domain model
	PaymentTolerance      *PaymentTolerance      `json:"payment_tolerance"`
	WebhookSettings       *WebhookSettings       `json:"webhook_settings"`
	CustomFields          map[string]interface{} `json:"custom_fields"`
}

// PaymentTolerance represents under/overpayment handling configuration.
type PaymentTolerance struct {
	UnderpaymentThreshold float64 `json:"underpayment_threshold"`
	OverpaymentThreshold  float64 `json:"overpayment_threshold"`
	OverpaymentAction     string  `json:"overpayment_action"`
}

// WebhookSettings represents webhook delivery configuration.
type WebhookSettings struct {
	DefaultTimeout    int    `json:"default_timeout"`
	DefaultMaxRetries int    `json:"default_max_retries"`
	DefaultBackoff    string `json:"default_backoff"`
	DefaultSecret     string `json:"default_secret"`
}

// APIKeyHash represents a hashed API key value.
type APIKeyHash struct {
	value string
}

// NewAPIKeyHash creates a new APIKeyHash from a raw API key.
func NewAPIKeyHash(rawKey string) (*APIKeyHash, error) {
	if rawKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	hash := sha256.Sum256([]byte(rawKey))
	return &APIKeyHash{
		value: hex.EncodeToString(hash[:]),
	}, nil
}

// NewAPIKeyHashFromString creates a new APIKeyHash from an existing hash string.
func NewAPIKeyHashFromString(hashString string) (*APIKeyHash, error) {
	if hashString == "" {
		return nil, fmt.Errorf("hash string cannot be empty")
	}

	return &APIKeyHash{
		value: hashString,
	}, nil
}

// String returns the string representation of the hash.
func (h *APIKeyHash) String() string {
	return h.value
}

// Equals compares two APIKeyHash instances.
func (h *APIKeyHash) Equals(other *APIKeyHash) bool {
	if other == nil {
		return false
	}
	return h.value == other.value
}

// Usage represents current usage statistics for a merchant.
type Usage struct {
	APIKeysCount          int       `json:"api_keys_count"`
	WebhookEndpointsCount int       `json:"webhook_endpoints_count"`
	MonthlyInvoices       int       `json:"monthly_invoices"`
	LastInvoiceAt         time.Time `json:"last_invoice_at"`
	TotalInvoices         int       `json:"total_invoices"`
	TotalRevenue          float64   `json:"total_revenue"`
}

// WebhookEndpointConfig represents configuration for webhook delivery.
type WebhookEndpointConfig struct {
	URL          string            `json:"url"`
	Events       []string          `json:"events"`
	Secret       string            `json:"secret"`
	MaxRetries   int               `json:"max_retries"`
	RetryBackoff BackoffStrategy   `json:"retry_backoff"`
	Timeout      int               `json:"timeout_seconds"`
	AllowedIPs   []string          `json:"allowed_ips"`
	Headers      map[string]string `json:"headers"`
}

// Validate validates the webhook endpoint configuration.
func (c *WebhookEndpointConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if len(c.Events) == 0 {
		return fmt.Errorf("at least one event type must be specified")
	}
	if c.MaxRetries < 0 || c.MaxRetries > 10 {
		return fmt.Errorf("max retries must be between 0 and 10")
	}
	if c.Timeout < 5 || c.Timeout > 60 {
		return fmt.Errorf("timeout must be between 5 and 60 seconds")
	}
	if !c.RetryBackoff.IsValid() {
		return fmt.Errorf("invalid retry backoff strategy")
	}
	return nil
}
