package payment_test

import (
	"testing"

	"crypto-checkout/internal/domain/payment"

	"github.com/stretchr/testify/require"
)

func TestPaymentStatus(t *testing.T) {
	t.Run("String - valid statuses", func(t *testing.T) {
		require.Equal(t, "detected", payment.StatusDetected.String())
		require.Equal(t, "confirming", payment.StatusConfirming.String())
		require.Equal(t, "confirmed", payment.StatusConfirmed.String())
		require.Equal(t, "orphaned", payment.StatusOrphaned.String())
		require.Equal(t, "failed", payment.StatusFailed.String())
	})

	t.Run("IsValid - valid statuses", func(t *testing.T) {
		require.True(t, payment.StatusDetected.IsValid())
		require.True(t, payment.StatusConfirming.IsValid())
		require.True(t, payment.StatusConfirmed.IsValid())
		require.True(t, payment.StatusOrphaned.IsValid())
		require.True(t, payment.StatusFailed.IsValid())
	})

	t.Run("IsValid - invalid status", func(t *testing.T) {
		invalidStatus := payment.PaymentStatus("invalid")
		require.False(t, invalidStatus.IsValid())
	})

	t.Run("IsTerminal - terminal statuses", func(t *testing.T) {
		require.True(t, payment.StatusConfirmed.IsTerminal())
		require.True(t, payment.StatusFailed.IsTerminal())
	})

	t.Run("IsTerminal - non-terminal statuses", func(t *testing.T) {
		require.False(t, payment.StatusDetected.IsTerminal())
		require.False(t, payment.StatusConfirming.IsTerminal())
		require.False(t, payment.StatusOrphaned.IsTerminal())
	})

	t.Run("IsActive - active statuses", func(t *testing.T) {
		require.True(t, payment.StatusDetected.IsActive())
		require.True(t, payment.StatusConfirming.IsActive())
		require.True(t, payment.StatusOrphaned.IsActive())
	})

	t.Run("IsActive - terminal statuses", func(t *testing.T) {
		require.False(t, payment.StatusConfirmed.IsActive())
		require.False(t, payment.StatusFailed.IsActive())
	})

	t.Run("CanTransitionTo - valid transitions from detected", func(t *testing.T) {
		require.True(t, payment.StatusDetected.CanTransitionTo(payment.StatusConfirming))
		require.True(t, payment.StatusDetected.CanTransitionTo(payment.StatusFailed))
	})

	t.Run("CanTransitionTo - invalid transitions from detected", func(t *testing.T) {
		require.False(t, payment.StatusDetected.CanTransitionTo(payment.StatusConfirmed))
		require.False(t, payment.StatusDetected.CanTransitionTo(payment.StatusOrphaned))
	})

	t.Run("CanTransitionTo - valid transitions from confirming", func(t *testing.T) {
		require.True(t, payment.StatusConfirming.CanTransitionTo(payment.StatusConfirmed))
		require.True(t, payment.StatusConfirming.CanTransitionTo(payment.StatusOrphaned))
		require.True(t, payment.StatusConfirming.CanTransitionTo(payment.StatusFailed))
	})

	t.Run("CanTransitionTo - invalid transitions from confirming", func(t *testing.T) {
		require.False(t, payment.StatusConfirming.CanTransitionTo(payment.StatusDetected))
	})

	t.Run("CanTransitionTo - valid transitions from orphaned", func(t *testing.T) {
		require.True(t, payment.StatusOrphaned.CanTransitionTo(payment.StatusDetected))
		require.True(t, payment.StatusOrphaned.CanTransitionTo(payment.StatusFailed))
	})

	t.Run("CanTransitionTo - invalid transitions from orphaned", func(t *testing.T) {
		require.False(t, payment.StatusOrphaned.CanTransitionTo(payment.StatusConfirming))
		require.False(t, payment.StatusOrphaned.CanTransitionTo(payment.StatusConfirmed))
	})

	t.Run("CanTransitionTo - terminal states", func(t *testing.T) {
		require.False(t, payment.StatusConfirmed.CanTransitionTo(payment.StatusDetected))
		require.False(t, payment.StatusConfirmed.CanTransitionTo(payment.StatusConfirming))
		require.False(t, payment.StatusConfirmed.CanTransitionTo(payment.StatusOrphaned))
		require.False(t, payment.StatusConfirmed.CanTransitionTo(payment.StatusFailed))

		require.False(t, payment.StatusFailed.CanTransitionTo(payment.StatusDetected))
		require.False(t, payment.StatusFailed.CanTransitionTo(payment.StatusConfirming))
		require.False(t, payment.StatusFailed.CanTransitionTo(payment.StatusConfirmed))
		require.False(t, payment.StatusFailed.CanTransitionTo(payment.StatusOrphaned))
	})

	t.Run("CanTransitionTo - invalid status", func(t *testing.T) {
		invalidStatus := payment.PaymentStatus("invalid")
		require.False(t, payment.StatusDetected.CanTransitionTo(invalidStatus))
	})

	t.Run("ParsePaymentStatus - valid status", func(t *testing.T) {
		status, err := payment.ParsePaymentStatus("detected")
		require.NoError(t, err)
		require.Equal(t, payment.StatusDetected, status)
	})

	t.Run("ParsePaymentStatus - invalid status", func(t *testing.T) {
		_, err := payment.ParsePaymentStatus("invalid")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid payment status")
	})
}
