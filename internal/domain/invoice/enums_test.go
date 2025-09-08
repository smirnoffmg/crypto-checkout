package invoice_test

import (
	"crypto-checkout/internal/domain/invoice"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInvoiceStatus(t *testing.T) {
	t.Run("String - valid statuses", func(t *testing.T) {
		require.Equal(t, "created", invoice.StatusCreated.String())
		require.Equal(t, "pending", invoice.StatusPending.String())
		require.Equal(t, "partial", invoice.StatusPartial.String())
		require.Equal(t, "confirming", invoice.StatusConfirming.String())
		require.Equal(t, "paid", invoice.StatusPaid.String())
		require.Equal(t, "expired", invoice.StatusExpired.String())
		require.Equal(t, "cancelled", invoice.StatusCancelled.String())
		require.Equal(t, "refunded", invoice.StatusRefunded.String())
	})

	t.Run("IsValid - valid statuses", func(t *testing.T) {
		require.True(t, invoice.StatusCreated.IsValid())
		require.True(t, invoice.StatusPending.IsValid())
		require.True(t, invoice.StatusPartial.IsValid())
		require.True(t, invoice.StatusConfirming.IsValid())
		require.True(t, invoice.StatusPaid.IsValid())
		require.True(t, invoice.StatusExpired.IsValid())
		require.True(t, invoice.StatusCancelled.IsValid())
		require.True(t, invoice.StatusRefunded.IsValid())
	})

	t.Run("IsValid - invalid status", func(t *testing.T) {
		invalidStatus := invoice.InvoiceStatus("invalid")
		require.False(t, invalidStatus.IsValid())
	})

	t.Run("IsTerminal - terminal statuses", func(t *testing.T) {
		require.True(t, invoice.StatusPaid.IsTerminal())
		require.True(t, invoice.StatusExpired.IsTerminal())
		require.True(t, invoice.StatusCancelled.IsTerminal())
		require.True(t, invoice.StatusRefunded.IsTerminal())
	})

	t.Run("IsTerminal - non-terminal statuses", func(t *testing.T) {
		require.False(t, invoice.StatusCreated.IsTerminal())
		require.False(t, invoice.StatusPending.IsTerminal())
		require.False(t, invoice.StatusPartial.IsTerminal())
		require.False(t, invoice.StatusConfirming.IsTerminal())
	})

	t.Run("IsActive - active statuses", func(t *testing.T) {
		require.True(t, invoice.StatusCreated.IsActive())
		require.True(t, invoice.StatusPending.IsActive())
		require.True(t, invoice.StatusPartial.IsActive())
		require.True(t, invoice.StatusConfirming.IsActive())
	})

	t.Run("IsActive - non-active statuses", func(t *testing.T) {
		require.False(t, invoice.StatusPaid.IsActive())
		require.False(t, invoice.StatusExpired.IsActive())
		require.False(t, invoice.StatusCancelled.IsActive())
		require.False(t, invoice.StatusRefunded.IsActive())
	})

	t.Run("CanTransitionTo - valid transitions", func(t *testing.T) {
		// Created -> Pending, Expired, Cancelled
		require.True(t, invoice.StatusCreated.CanTransitionTo(invoice.StatusPending))
		require.True(t, invoice.StatusCreated.CanTransitionTo(invoice.StatusExpired))
		require.True(t, invoice.StatusCreated.CanTransitionTo(invoice.StatusCancelled))

		// Pending -> Partial, Confirming, Expired, Cancelled
		require.True(t, invoice.StatusPending.CanTransitionTo(invoice.StatusPartial))
		require.True(t, invoice.StatusPending.CanTransitionTo(invoice.StatusConfirming))
		require.True(t, invoice.StatusPending.CanTransitionTo(invoice.StatusExpired))
		require.True(t, invoice.StatusPending.CanTransitionTo(invoice.StatusCancelled))

		// Partial -> Confirming, Cancelled
		require.True(t, invoice.StatusPartial.CanTransitionTo(invoice.StatusConfirming))
		require.True(t, invoice.StatusPartial.CanTransitionTo(invoice.StatusCancelled))

		// Confirming -> Paid, Pending (for blockchain reorg)
		require.True(t, invoice.StatusConfirming.CanTransitionTo(invoice.StatusPaid))
		require.True(t, invoice.StatusConfirming.CanTransitionTo(invoice.StatusPending))

		// Paid -> Refunded
		require.True(t, invoice.StatusPaid.CanTransitionTo(invoice.StatusRefunded))
	})

	t.Run("CanTransitionTo - invalid transitions", func(t *testing.T) {
		// Created cannot go directly to Paid
		require.False(t, invoice.StatusCreated.CanTransitionTo(invoice.StatusPaid))

		// Partial cannot go to Expired
		require.False(t, invoice.StatusPartial.CanTransitionTo(invoice.StatusExpired))

		// Terminal states cannot transition (except paid -> refunded)
		require.False(t, invoice.StatusExpired.CanTransitionTo(invoice.StatusPaid))
		require.False(t, invoice.StatusCancelled.CanTransitionTo(invoice.StatusPaid))
		require.False(t, invoice.StatusRefunded.CanTransitionTo(invoice.StatusPaid))

		// Invalid target status
		invalidStatus := invoice.InvoiceStatus("invalid")
		require.False(t, invoice.StatusPending.CanTransitionTo(invalidStatus))
	})
}

func TestOverpaymentAction(t *testing.T) {
	t.Run("String - valid actions", func(t *testing.T) {
		require.Equal(t, "credit_account", invoice.OverpaymentActionCredit.String())
		require.Equal(t, "refund", invoice.OverpaymentActionRefund.String())
		require.Equal(t, "donate", invoice.OverpaymentActionDonate.String())
	})

	t.Run("IsValid - valid actions", func(t *testing.T) {
		require.True(t, invoice.OverpaymentActionCredit.IsValid())
		require.True(t, invoice.OverpaymentActionRefund.IsValid())
		require.True(t, invoice.OverpaymentActionDonate.IsValid())
	})

	t.Run("IsValid - invalid action", func(t *testing.T) {
		invalidAction := invoice.OverpaymentAction("invalid")
		require.False(t, invalidAction.IsValid())
	})
}

func TestAuditEvent(t *testing.T) {
	t.Run("String - valid events", func(t *testing.T) {
		require.Equal(t, "created", invoice.AuditEventCreated.String())
		require.Equal(t, "viewed", invoice.AuditEventViewed.String())
		require.Equal(t, "payment_detected", invoice.AuditEventPaymentDetected.String())
		require.Equal(t, "payment_confirmed", invoice.AuditEventPaymentConfirmed.String())
		require.Equal(t, "paid", invoice.AuditEventPaid.String())
		require.Equal(t, "expired", invoice.AuditEventExpired.String())
		require.Equal(t, "cancelled", invoice.AuditEventCancelled.String())
		require.Equal(t, "refunded", invoice.AuditEventRefunded.String())
		require.Equal(t, "status_changed", invoice.AuditEventStatusChanged.String())
	})

	t.Run("IsValid - valid events", func(t *testing.T) {
		require.True(t, invoice.AuditEventCreated.IsValid())
		require.True(t, invoice.AuditEventViewed.IsValid())
		require.True(t, invoice.AuditEventPaymentDetected.IsValid())
		require.True(t, invoice.AuditEventPaymentConfirmed.IsValid())
		require.True(t, invoice.AuditEventPaid.IsValid())
		require.True(t, invoice.AuditEventExpired.IsValid())
		require.True(t, invoice.AuditEventCancelled.IsValid())
		require.True(t, invoice.AuditEventRefunded.IsValid())
		require.True(t, invoice.AuditEventStatusChanged.IsValid())
	})

	t.Run("IsValid - invalid event", func(t *testing.T) {
		invalidEvent := invoice.AuditEvent("invalid")
		require.False(t, invalidEvent.IsValid())
	})
}

func TestActor(t *testing.T) {
	t.Run("String - valid actors", func(t *testing.T) {
		require.Equal(t, "api_key", invoice.ActorAPIKey.String())
		require.Equal(t, "system", invoice.ActorSystem.String())
		require.Equal(t, "customer", invoice.ActorCustomer.String())
		require.Equal(t, "admin", invoice.ActorAdmin.String())
	})

	t.Run("IsValid - valid actors", func(t *testing.T) {
		require.True(t, invoice.ActorAPIKey.IsValid())
		require.True(t, invoice.ActorSystem.IsValid())
		require.True(t, invoice.ActorCustomer.IsValid())
		require.True(t, invoice.ActorAdmin.IsValid())
	})

	t.Run("IsValid - invalid actor", func(t *testing.T) {
		invalidActor := invoice.Actor("invalid")
		require.False(t, invalidActor.IsValid())
	})
}
