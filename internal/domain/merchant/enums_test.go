package merchant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerchantStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   MerchantStatus
		expected bool
	}{
		{"active", StatusActive, true},
		{"suspended", StatusSuspended, true},
		{"pending_verification", StatusPendingVerification, true},
		{"closed", StatusClosed, true},
		{"invalid", MerchantStatus("invalid"), false},
		{"empty", MerchantStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKeyType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		keyType  KeyType
		expected bool
	}{
		{"live", KeyTypeLive, true},
		{"test", KeyTypeTest, true},
		{"invalid", KeyType("invalid"), false},
		{"empty", KeyType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.keyType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKeyStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   KeyStatus
		expected bool
	}{
		{"active", KeyStatusActive, true},
		{"revoked", KeyStatusRevoked, true},
		{"expired", KeyStatusExpired, true},
		{"invalid", KeyStatus("invalid"), false},
		{"empty", KeyStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEndpointStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   EndpointStatus
		expected bool
	}{
		{"active", EndpointStatusActive, true},
		{"disabled", EndpointStatusDisabled, true},
		{"failed", EndpointStatusFailed, true},
		{"invalid", EndpointStatus("invalid"), false},
		{"empty", EndpointStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBackoffStrategy_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		strategy BackoffStrategy
		expected bool
	}{
		{"linear", BackoffStrategyLinear, true},
		{"exponential", BackoffStrategyExponential, true},
		{"invalid", BackoffStrategy("invalid"), false},
		{"empty", BackoffStrategy(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}
