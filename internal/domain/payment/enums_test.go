package payment_test

import (
	"testing"

	"crypto-checkout/internal/domain/payment"

	"github.com/stretchr/testify/assert"
)

func TestPaymentStatus(t *testing.T) {
	t.Run("String - valid statuses", func(t *testing.T) {
		assert.Equal(t, "detected", payment.StatusDetected.String())
		assert.Equal(t, "confirming", payment.StatusConfirming.String())
		assert.Equal(t, "confirmed", payment.StatusConfirmed.String())
		assert.Equal(t, "orphaned", payment.StatusOrphaned.String())
		assert.Equal(t, "failed", payment.StatusFailed.String())
	})

	t.Run("IsValid - valid statuses", func(t *testing.T) {
		assert.True(t, payment.StatusDetected.IsValid())
		assert.True(t, payment.StatusConfirming.IsValid())
		assert.True(t, payment.StatusConfirmed.IsValid())
		assert.True(t, payment.StatusOrphaned.IsValid())
		assert.True(t, payment.StatusFailed.IsValid())
	})

	t.Run("IsValid - invalid status", func(t *testing.T) {
		invalidStatus := payment.PaymentStatus("invalid")
		assert.False(t, invalidStatus.IsValid())
	})

	t.Run("IsTerminal - terminal statuses", func(t *testing.T) {
		assert.True(t, payment.StatusConfirmed.IsTerminal())
		assert.True(t, payment.StatusFailed.IsTerminal())
	})

	t.Run("IsTerminal - non-terminal statuses", func(t *testing.T) {
		assert.False(t, payment.StatusDetected.IsTerminal())
		assert.False(t, payment.StatusConfirming.IsTerminal())
		assert.False(t, payment.StatusOrphaned.IsTerminal())
	})

	t.Run("IsActive - active statuses", func(t *testing.T) {
		assert.True(t, payment.StatusDetected.IsActive())
		assert.True(t, payment.StatusConfirming.IsActive())
		assert.True(t, payment.StatusOrphaned.IsActive())
	})

	t.Run("IsActive - terminal statuses", func(t *testing.T) {
		assert.False(t, payment.StatusConfirmed.IsActive())
		assert.False(t, payment.StatusFailed.IsActive())
	})

	t.Run("CanTransitionTo - valid transitions from detected", func(t *testing.T) {
		assert.True(t, payment.StatusDetected.CanTransitionTo(payment.StatusConfirming))
		assert.True(t, payment.StatusDetected.CanTransitionTo(payment.StatusFailed))
	})

	t.Run("CanTransitionTo - invalid transitions from detected", func(t *testing.T) {
		assert.False(t, payment.StatusDetected.CanTransitionTo(payment.StatusConfirmed))
		assert.False(t, payment.StatusDetected.CanTransitionTo(payment.StatusOrphaned))
	})

	t.Run("CanTransitionTo - valid transitions from confirming", func(t *testing.T) {
		assert.True(t, payment.StatusConfirming.CanTransitionTo(payment.StatusConfirmed))
		assert.True(t, payment.StatusConfirming.CanTransitionTo(payment.StatusOrphaned))
		assert.True(t, payment.StatusConfirming.CanTransitionTo(payment.StatusFailed))
	})

	t.Run("CanTransitionTo - invalid transitions from confirming", func(t *testing.T) {
		assert.False(t, payment.StatusConfirming.CanTransitionTo(payment.StatusDetected))
	})

	t.Run("CanTransitionTo - valid transitions from orphaned", func(t *testing.T) {
		assert.True(t, payment.StatusOrphaned.CanTransitionTo(payment.StatusDetected))
		assert.True(t, payment.StatusOrphaned.CanTransitionTo(payment.StatusFailed))
	})

	t.Run("CanTransitionTo - invalid transitions from orphaned", func(t *testing.T) {
		assert.False(t, payment.StatusOrphaned.CanTransitionTo(payment.StatusConfirming))
		assert.False(t, payment.StatusOrphaned.CanTransitionTo(payment.StatusConfirmed))
	})

	t.Run("CanTransitionTo - terminal states", func(t *testing.T) {
		assert.False(t, payment.StatusConfirmed.CanTransitionTo(payment.StatusDetected))
		assert.False(t, payment.StatusConfirmed.CanTransitionTo(payment.StatusConfirming))
		assert.False(t, payment.StatusConfirmed.CanTransitionTo(payment.StatusOrphaned))
		assert.False(t, payment.StatusConfirmed.CanTransitionTo(payment.StatusFailed))

		assert.False(t, payment.StatusFailed.CanTransitionTo(payment.StatusDetected))
		assert.False(t, payment.StatusFailed.CanTransitionTo(payment.StatusConfirming))
		assert.False(t, payment.StatusFailed.CanTransitionTo(payment.StatusConfirmed))
		assert.False(t, payment.StatusFailed.CanTransitionTo(payment.StatusOrphaned))
	})

	t.Run("CanTransitionTo - invalid status", func(t *testing.T) {
		invalidStatus := payment.PaymentStatus("invalid")
		assert.False(t, payment.StatusDetected.CanTransitionTo(invalidStatus))
	})

	t.Run("ParsePaymentStatus - valid status", func(t *testing.T) {
		status, err := payment.ParsePaymentStatus("detected")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusDetected, status)
	})

	t.Run("ParsePaymentStatus - invalid status", func(t *testing.T) {
		_, err := payment.ParsePaymentStatus("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payment status")
	})
}
