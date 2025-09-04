package payment_test

import (
	"testing"

	"crypto-checkout/internal/domain/payment"

	"github.com/stretchr/testify/assert"
)

func TestPaymentStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   payment.PaymentStatus
		expected string
	}{
		{
			name:     "detected status",
			status:   payment.StatusDetected,
			expected: "detected",
		},
		{
			name:     "confirming status",
			status:   payment.StatusConfirming,
			expected: "confirming",
		},
		{
			name:     "confirmed status",
			status:   payment.StatusConfirmed,
			expected: "confirmed",
		},
		{
			name:     "failed status",
			status:   payment.StatusFailed,
			expected: "failed",
		},
		{
			name:     "orphaned status",
			status:   payment.StatusOrphaned,
			expected: "orphaned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestPaymentStatus_IsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   payment.PaymentStatus
		expected bool
	}{
		{
			name:     "detected is not terminal",
			status:   payment.StatusDetected,
			expected: false,
		},
		{
			name:     "confirming is not terminal",
			status:   payment.StatusConfirming,
			expected: false,
		},
		{
			name:     "confirmed is terminal",
			status:   payment.StatusConfirmed,
			expected: true,
		},
		{
			name:     "failed is terminal",
			status:   payment.StatusFailed,
			expected: true,
		},
		{
			name:     "orphaned is not terminal",
			status:   payment.StatusOrphaned,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.status.IsTerminal())
		})
	}
}

func TestPaymentStatus_IsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   payment.PaymentStatus
		expected bool
	}{
		{
			name:     "detected is active",
			status:   payment.StatusDetected,
			expected: true,
		},
		{
			name:     "confirming is active",
			status:   payment.StatusConfirming,
			expected: true,
		},
		{
			name:     "confirmed is not active",
			status:   payment.StatusConfirmed,
			expected: false,
		},
		{
			name:     "failed is not active",
			status:   payment.StatusFailed,
			expected: false,
		},
		{
			name:     "orphaned is active",
			status:   payment.StatusOrphaned,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.status.IsActive())
		})
	}
}

func TestPaymentStatus_IsSuccessful(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   payment.PaymentStatus
		expected bool
	}{
		{
			name:     "detected is not successful",
			status:   payment.StatusDetected,
			expected: false,
		},
		{
			name:     "confirming is not successful",
			status:   payment.StatusConfirming,
			expected: false,
		},
		{
			name:     "confirmed is successful",
			status:   payment.StatusConfirmed,
			expected: true,
		},
		{
			name:     "failed is not successful",
			status:   payment.StatusFailed,
			expected: false,
		},
		{
			name:     "orphaned is not successful",
			status:   payment.StatusOrphaned,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.status.IsSuccessful())
		})
	}
}

func TestPaymentStatus_IsFailed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   payment.PaymentStatus
		expected bool
	}{
		{
			name:     "detected is not failed",
			status:   payment.StatusDetected,
			expected: false,
		},
		{
			name:     "confirming is not failed",
			status:   payment.StatusConfirming,
			expected: false,
		},
		{
			name:     "confirmed is not failed",
			status:   payment.StatusConfirmed,
			expected: false,
		},
		{
			name:     "failed is failed",
			status:   payment.StatusFailed,
			expected: true,
		},
		{
			name:     "orphaned is not failed",
			status:   payment.StatusOrphaned,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.status.IsFailed())
		})
	}
}

func TestPaymentStatus_IsTemporary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   payment.PaymentStatus
		expected bool
	}{
		{
			name:     "detected is not temporary",
			status:   payment.StatusDetected,
			expected: false,
		},
		{
			name:     "confirming is not temporary",
			status:   payment.StatusConfirming,
			expected: false,
		},
		{
			name:     "confirmed is not temporary",
			status:   payment.StatusConfirmed,
			expected: false,
		},
		{
			name:     "failed is not temporary",
			status:   payment.StatusFailed,
			expected: false,
		},
		{
			name:     "orphaned is temporary",
			status:   payment.StatusOrphaned,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.status.IsTemporary())
		})
	}
}
