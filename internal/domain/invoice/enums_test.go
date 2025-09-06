package invoice_test

import (
	"testing"

	"crypto-checkout/internal/domain/invoice"

	"github.com/stretchr/testify/assert"
)

func TestInvoiceStatus(t *testing.T) {
	t.Run("String - valid statuses", func(t *testing.T) {
		assert.Equal(t, "created", invoice.StatusCreated.String())
		assert.Equal(t, "pending", invoice.StatusPending.String())
		assert.Equal(t, "partial", invoice.StatusPartial.String())
		assert.Equal(t, "confirming", invoice.StatusConfirming.String())
		assert.Equal(t, "paid", invoice.StatusPaid.String())
		assert.Equal(t, "expired", invoice.StatusExpired.String())
		assert.Equal(t, "cancelled", invoice.StatusCancelled.String())
		assert.Equal(t, "refunded", invoice.StatusRefunded.String())
	})

	t.Run("IsValid - valid statuses", func(t *testing.T) {
		assert.True(t, invoice.StatusCreated.IsValid())
		assert.True(t, invoice.StatusPending.IsValid())
		assert.True(t, invoice.StatusPartial.IsValid())
		assert.True(t, invoice.StatusConfirming.IsValid())
		assert.True(t, invoice.StatusPaid.IsValid())
		assert.True(t, invoice.StatusExpired.IsValid())
		assert.True(t, invoice.StatusCancelled.IsValid())
		assert.True(t, invoice.StatusRefunded.IsValid())
	})

	t.Run("IsValid - invalid status", func(t *testing.T) {
		invalidStatus := invoice.InvoiceStatus("invalid")
		assert.False(t, invalidStatus.IsValid())
	})

	t.Run("IsTerminal - terminal statuses", func(t *testing.T) {
		assert.True(t, invoice.StatusPaid.IsTerminal())
		assert.True(t, invoice.StatusExpired.IsTerminal())
		assert.True(t, invoice.StatusCancelled.IsTerminal())
		assert.True(t, invoice.StatusRefunded.IsTerminal())
	})

	t.Run("IsTerminal - non-terminal statuses", func(t *testing.T) {
		assert.False(t, invoice.StatusCreated.IsTerminal())
		assert.False(t, invoice.StatusPending.IsTerminal())
		assert.False(t, invoice.StatusPartial.IsTerminal())
		assert.False(t, invoice.StatusConfirming.IsTerminal())
	})

	t.Run("IsActive - active statuses", func(t *testing.T) {
		assert.True(t, invoice.StatusCreated.IsActive())
		assert.True(t, invoice.StatusPending.IsActive())
		assert.True(t, invoice.StatusPartial.IsActive())
		assert.True(t, invoice.StatusConfirming.IsActive())
	})

	t.Run("IsActive - non-active statuses", func(t *testing.T) {
		assert.False(t, invoice.StatusPaid.IsActive())
		assert.False(t, invoice.StatusExpired.IsActive())
		assert.False(t, invoice.StatusCancelled.IsActive())
		assert.False(t, invoice.StatusRefunded.IsActive())
	})

	t.Run("CanTransitionTo - valid transitions", func(t *testing.T) {
		// Created -> Pending, Expired, Cancelled
		assert.True(t, invoice.StatusCreated.CanTransitionTo(invoice.StatusPending))
		assert.True(t, invoice.StatusCreated.CanTransitionTo(invoice.StatusExpired))
		assert.True(t, invoice.StatusCreated.CanTransitionTo(invoice.StatusCancelled))

		// Pending -> Partial, Confirming, Expired, Cancelled
		assert.True(t, invoice.StatusPending.CanTransitionTo(invoice.StatusPartial))
		assert.True(t, invoice.StatusPending.CanTransitionTo(invoice.StatusConfirming))
		assert.True(t, invoice.StatusPending.CanTransitionTo(invoice.StatusExpired))
		assert.True(t, invoice.StatusPending.CanTransitionTo(invoice.StatusCancelled))

		// Partial -> Confirming, Cancelled
		assert.True(t, invoice.StatusPartial.CanTransitionTo(invoice.StatusConfirming))
		assert.True(t, invoice.StatusPartial.CanTransitionTo(invoice.StatusCancelled))

		// Confirming -> Paid, Pending (for blockchain reorg)
		assert.True(t, invoice.StatusConfirming.CanTransitionTo(invoice.StatusPaid))
		assert.True(t, invoice.StatusConfirming.CanTransitionTo(invoice.StatusPending))

		// Paid -> Refunded
		assert.True(t, invoice.StatusPaid.CanTransitionTo(invoice.StatusRefunded))
	})

	t.Run("CanTransitionTo - invalid transitions", func(t *testing.T) {
		// Created cannot go directly to Paid
		assert.False(t, invoice.StatusCreated.CanTransitionTo(invoice.StatusPaid))

		// Partial cannot go to Expired
		assert.False(t, invoice.StatusPartial.CanTransitionTo(invoice.StatusExpired))

		// Terminal states cannot transition (except paid -> refunded)
		assert.False(t, invoice.StatusExpired.CanTransitionTo(invoice.StatusPaid))
		assert.False(t, invoice.StatusCancelled.CanTransitionTo(invoice.StatusPaid))
		assert.False(t, invoice.StatusRefunded.CanTransitionTo(invoice.StatusPaid))

		// Invalid target status
		invalidStatus := invoice.InvoiceStatus("invalid")
		assert.False(t, invoice.StatusPending.CanTransitionTo(invalidStatus))
	})
}

func TestOverpaymentAction(t *testing.T) {
	t.Run("String - valid actions", func(t *testing.T) {
		assert.Equal(t, "credit_account", invoice.OverpaymentActionCredit.String())
		assert.Equal(t, "refund", invoice.OverpaymentActionRefund.String())
		assert.Equal(t, "donate", invoice.OverpaymentActionDonate.String())
	})

	t.Run("IsValid - valid actions", func(t *testing.T) {
		assert.True(t, invoice.OverpaymentActionCredit.IsValid())
		assert.True(t, invoice.OverpaymentActionRefund.IsValid())
		assert.True(t, invoice.OverpaymentActionDonate.IsValid())
	})

	t.Run("IsValid - invalid action", func(t *testing.T) {
		invalidAction := invoice.OverpaymentAction("invalid")
		assert.False(t, invalidAction.IsValid())
	})
}

func TestAuditEvent(t *testing.T) {
	t.Run("String - valid events", func(t *testing.T) {
		assert.Equal(t, "created", invoice.AuditEventCreated.String())
		assert.Equal(t, "viewed", invoice.AuditEventViewed.String())
		assert.Equal(t, "payment_detected", invoice.AuditEventPaymentDetected.String())
		assert.Equal(t, "payment_confirmed", invoice.AuditEventPaymentConfirmed.String())
		assert.Equal(t, "paid", invoice.AuditEventPaid.String())
		assert.Equal(t, "expired", invoice.AuditEventExpired.String())
		assert.Equal(t, "cancelled", invoice.AuditEventCancelled.String())
		assert.Equal(t, "refunded", invoice.AuditEventRefunded.String())
		assert.Equal(t, "status_changed", invoice.AuditEventStatusChanged.String())
	})

	t.Run("IsValid - valid events", func(t *testing.T) {
		assert.True(t, invoice.AuditEventCreated.IsValid())
		assert.True(t, invoice.AuditEventViewed.IsValid())
		assert.True(t, invoice.AuditEventPaymentDetected.IsValid())
		assert.True(t, invoice.AuditEventPaymentConfirmed.IsValid())
		assert.True(t, invoice.AuditEventPaid.IsValid())
		assert.True(t, invoice.AuditEventExpired.IsValid())
		assert.True(t, invoice.AuditEventCancelled.IsValid())
		assert.True(t, invoice.AuditEventRefunded.IsValid())
		assert.True(t, invoice.AuditEventStatusChanged.IsValid())
	})

	t.Run("IsValid - invalid event", func(t *testing.T) {
		invalidEvent := invoice.AuditEvent("invalid")
		assert.False(t, invalidEvent.IsValid())
	})
}

func TestActor(t *testing.T) {
	t.Run("String - valid actors", func(t *testing.T) {
		assert.Equal(t, "api_key", invoice.ActorAPIKey.String())
		assert.Equal(t, "system", invoice.ActorSystem.String())
		assert.Equal(t, "customer", invoice.ActorCustomer.String())
		assert.Equal(t, "admin", invoice.ActorAdmin.String())
	})

	t.Run("IsValid - valid actors", func(t *testing.T) {
		assert.True(t, invoice.ActorAPIKey.IsValid())
		assert.True(t, invoice.ActorSystem.IsValid())
		assert.True(t, invoice.ActorCustomer.IsValid())
		assert.True(t, invoice.ActorAdmin.IsValid())
	})

	t.Run("IsValid - invalid actor", func(t *testing.T) {
		invalidActor := invoice.Actor("invalid")
		assert.False(t, invalidActor.IsValid())
	})
}
