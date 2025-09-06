package invoice_test

import (
	"context"
	"testing"
	"time"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/assert"
)

func TestInvoiceFSM(t *testing.T) {
	t.Run("NewInvoiceFSM", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)

		assert.Equal(t, invoice.StatusCreated, fsm.CurrentStatus())
		assert.True(t, fsm.IsActive())
		assert.False(t, fsm.IsTerminal())
	})

	t.Run("CanTransitionTo - valid transitions from created", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)

		// From created state
		assert.True(t, fsm.CanTransitionTo(invoice.StatusPending))
		assert.True(t, fsm.CanTransitionTo(invoice.StatusExpired))
		assert.True(t, fsm.CanTransitionTo(invoice.StatusCancelled))

		// Invalid transitions
		assert.False(t, fsm.CanTransitionTo(invoice.StatusPartial))
		assert.False(t, fsm.CanTransitionTo(invoice.StatusConfirming))
		assert.False(t, fsm.CanTransitionTo(invoice.StatusPaid))
		assert.False(t, fsm.CanTransitionTo(invoice.StatusRefunded))
	})

	t.Run("CanTransitionTo - invalid status", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)

		assert.False(t, fsm.CanTransitionTo(invoice.InvoiceStatus("invalid")))
	})

	t.Run("TransitionTo - valid transition", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)

		err := fsm.TransitionTo(invoice.StatusPending)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())
	})

	t.Run("TransitionTo - invalid transition", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)

		err := fsm.TransitionTo(invoice.StatusPaid)
		assert.Error(t, err)
		assert.Equal(t, invoice.StatusCreated, fsm.CurrentStatus())
	})

	t.Run("Event - valid event", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)

		ctx := context.Background()
		err := fsm.Event(ctx, "view")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())
	})

	t.Run("Event - invalid event", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)

		ctx := context.Background()
		err := fsm.Event(ctx, "invalid_event")
		assert.Error(t, err)
		assert.Equal(t, invoice.StatusCreated, fsm.CurrentStatus())
	})

	t.Run("GetValidTransitions", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)

		transitions := fsm.GetValidTransitions()
		expected := []invoice.InvoiceStatus{invoice.StatusPending, invoice.StatusExpired, invoice.StatusCancelled}

		assert.ElementsMatch(t, expected, transitions)
	})

	t.Run("IsTerminal", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)

		// Created is not terminal
		assert.False(t, fsm.IsTerminal())

		// Test terminal states by creating invoices in those states
		paidInvoice := createTestInvoice()
		paidInvoice.SetStatus(invoice.StatusPaid)
		fsmPaid := invoice.NewInvoiceFSM(paidInvoice)
		assert.True(t, fsmPaid.IsTerminal())

		expiredInvoice := createTestInvoice()
		expiredInvoice.SetStatus(invoice.StatusExpired)
		fsmExpired := invoice.NewInvoiceFSM(expiredInvoice)
		assert.True(t, fsmExpired.IsTerminal())

		cancelledInvoice := createTestInvoice()
		cancelledInvoice.SetStatus(invoice.StatusCancelled)
		fsmCancelled := invoice.NewInvoiceFSM(cancelledInvoice)
		assert.True(t, fsmCancelled.IsTerminal())

		refundedInvoice := createTestInvoice()
		refundedInvoice.SetStatus(invoice.StatusRefunded)
		fsmRefunded := invoice.NewInvoiceFSM(refundedInvoice)
		assert.True(t, fsmRefunded.IsTerminal())
	})

	t.Run("IsActive", func(t *testing.T) {
		// Test active states
		createdInvoice := createTestInvoice()
		fsmCreated := invoice.NewInvoiceFSM(createdInvoice)
		assert.True(t, fsmCreated.IsActive())

		pendingInvoice := createTestInvoice()
		pendingInvoice.SetStatus(invoice.StatusPending)
		fsmPending := invoice.NewInvoiceFSM(pendingInvoice)
		assert.True(t, fsmPending.IsActive())

		partialInvoice := createTestInvoice()
		partialInvoice.SetStatus(invoice.StatusPartial)
		fsmPartial := invoice.NewInvoiceFSM(partialInvoice)
		assert.True(t, fsmPartial.IsActive())

		confirmingInvoice := createTestInvoice()
		confirmingInvoice.SetStatus(invoice.StatusConfirming)
		fsmConfirming := invoice.NewInvoiceFSM(confirmingInvoice)
		assert.True(t, fsmConfirming.IsActive())

		// Test terminal states
		paidInvoice := createTestInvoice()
		paidInvoice.SetStatus(invoice.StatusPaid)
		fsmPaid := invoice.NewInvoiceFSM(paidInvoice)
		assert.False(t, fsmPaid.IsActive())

		expiredInvoice := createTestInvoice()
		expiredInvoice.SetStatus(invoice.StatusExpired)
		fsmExpired := invoice.NewInvoiceFSM(expiredInvoice)
		assert.False(t, fsmExpired.IsActive())

		cancelledInvoice := createTestInvoice()
		cancelledInvoice.SetStatus(invoice.StatusCancelled)
		fsmCancelled := invoice.NewInvoiceFSM(cancelledInvoice)
		assert.False(t, fsmCancelled.IsActive())

		refundedInvoice := createTestInvoice()
		refundedInvoice.SetStatus(invoice.StatusRefunded)
		fsmRefunded := invoice.NewInvoiceFSM(refundedInvoice)
		assert.False(t, fsmRefunded.IsActive())
	})

	t.Run("Valid business flow - complete payment cycle", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)
		ctx := context.Background()

		// Created -> Pending (customer views invoice)
		err := fsm.Event(ctx, "view")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())
		assert.Equal(t, invoice.StatusPending, testInvoice.Status())

		// Pending -> Confirming (full payment received)
		err = fsm.Event(ctx, "full_payment")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusConfirming, fsm.CurrentStatus())
		assert.Equal(t, invoice.StatusConfirming, testInvoice.Status())

		// Confirming -> Paid (payment confirmed)
		err = fsm.Event(ctx, "confirm")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPaid, fsm.CurrentStatus())
		assert.Equal(t, invoice.StatusPaid, testInvoice.Status())

		// Paid -> Refunded (refund processed)
		err = fsm.Event(ctx, "refund")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusRefunded, fsm.CurrentStatus())
		assert.Equal(t, invoice.StatusRefunded, testInvoice.Status())
	})

	t.Run("Valid business flow - partial payment", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)
		ctx := context.Background()

		// Created -> Pending (customer views invoice)
		err := fsm.Event(ctx, "view")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())

		// Pending -> Partial (partial payment received)
		err = fsm.Event(ctx, "partial_payment")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPartial, fsm.CurrentStatus())

		// Partial -> Confirming (remaining payment received)
		err = fsm.Event(ctx, "full_payment")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusConfirming, fsm.CurrentStatus())

		// Confirming -> Paid (payment confirmed)
		err = fsm.Event(ctx, "confirm")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPaid, fsm.CurrentStatus())
	})

	t.Run("Invalid transitions - FSM prevents invalid events", func(t *testing.T) {
		testInvoice := createTestInvoice()
		fsm := invoice.NewInvoiceFSM(testInvoice)
		ctx := context.Background()

		// Try to confirm payment from created state (should fail - event not allowed)
		err := fsm.Event(ctx, "confirm")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event confirm inappropriate in current state created")
		assert.Equal(t, invoice.StatusCreated, fsm.CurrentStatus())

		// Try to refund from created state (should fail - event not allowed)
		err = fsm.Event(ctx, "refund")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event refund inappropriate in current state created")
		assert.Equal(t, invoice.StatusCreated, fsm.CurrentStatus())

		// Move to pending state
		err = fsm.Event(ctx, "view")
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())

		// Try to confirm payment from pending state (should fail - event not allowed)
		err = fsm.Event(ctx, "confirm")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event confirm inappropriate in current state pending")
		assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())

		// Try to refund from pending state (should fail - event not allowed)
		err = fsm.Event(ctx, "refund")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event refund inappropriate in current state pending")
		assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())
	})

	t.Run("Business rule enforcement - expiration", func(t *testing.T) {
		// Create an invoice with partial payment
		partialInvoice := createTestInvoice()
		partialInvoice.SetStatus(invoice.StatusPartial)
		fsm := invoice.NewInvoiceFSM(partialInvoice)
		ctx := context.Background()

		// Try to expire invoice with partial payment (should fail - event not allowed)
		err := fsm.Event(ctx, "expire")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event expire inappropriate in current state partial")
		assert.Equal(t, invoice.StatusPartial, fsm.CurrentStatus())
	})
}

func TestInvoiceStateMachine(t *testing.T) {
	t.Run("NewInvoiceStateMachine", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		assert.Equal(t, invoice.StatusCreated, ism.CurrentStatus())
		assert.True(t, ism.IsActive())
		assert.False(t, ism.IsTerminal())
		assert.Equal(t, testInvoice, ism.GetInvoice())
	})

	t.Run("TransitionTo - valid transition", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		err := ism.TransitionTo(invoice.StatusPending, "customer viewed invoice", invoice.ActorCustomer, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, ism.CurrentStatus())
		assert.Equal(t, invoice.StatusPending, ism.GetInvoice().Status())

		// Check history
		history := ism.GetTransitionHistory()
		assert.Len(t, history, 1)
		assert.Equal(t, invoice.StatusCreated, history[0].FromStatus)
		assert.Equal(t, invoice.StatusPending, history[0].ToStatus)
		assert.Equal(t, "customer viewed invoice", history[0].Reason)
		assert.Equal(t, invoice.ActorCustomer, history[0].Actor)
	})

	t.Run("TransitionTo - invalid transition", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		// Try to transition directly to paid (should fail - business rule violation)
		err := ism.TransitionTo(invoice.StatusPaid, "invalid transition", invoice.ActorSystem, nil)
		assert.Error(t, err)
		assert.Equal(t, invoice.StatusCreated, ism.CurrentStatus())
		assert.Len(t, ism.GetTransitionHistory(), 0)
	})

	t.Run("Event - valid event", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		ctx := context.Background()
		err := ism.Event(ctx, "view", "customer viewed invoice", invoice.ActorCustomer, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, ism.CurrentStatus())

		// Check history
		history := ism.GetTransitionHistory()
		assert.Len(t, history, 1)
		assert.Equal(t, invoice.StatusCreated, history[0].FromStatus)
		assert.Equal(t, invoice.StatusPending, history[0].ToStatus)
	})

	t.Run("Event - invalid event", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		ctx := context.Background()
		err := ism.Event(ctx, "invalid_event", "invalid event", invoice.ActorSystem, nil)
		assert.Error(t, err)
		assert.Equal(t, invoice.StatusCreated, ism.CurrentStatus())
		assert.Len(t, ism.GetTransitionHistory(), 0)
	})

	t.Run("Event - business rule violation", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		ctx := context.Background()
		// Try to confirm payment from created state (should fail - event not allowed)
		err := ism.Event(ctx, "confirm", "trying to confirm without payment", invoice.ActorSystem, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event confirm inappropriate in current state created")
		assert.Equal(t, invoice.StatusCreated, ism.CurrentStatus())
		assert.Len(t, ism.GetTransitionHistory(), 0)
	})

	t.Run("CanTransitionTo", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		assert.True(t, ism.CanTransitionTo(invoice.StatusPending))
		assert.True(t, ism.CanTransitionTo(invoice.StatusExpired))
		assert.True(t, ism.CanTransitionTo(invoice.StatusCancelled))
		assert.False(t, ism.CanTransitionTo(invoice.StatusPaid))
	})

	t.Run("CanEvent", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		assert.True(t, ism.CanEvent("view"))
		assert.True(t, ism.CanEvent("expire"))
		assert.True(t, ism.CanEvent("cancel"))
		assert.False(t, ism.CanEvent("confirm"))
		assert.False(t, ism.CanEvent("refund"))
	})

	t.Run("GetValidTransitions", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		transitions := ism.GetValidTransitions()
		expected := []invoice.InvoiceStatus{invoice.StatusPending, invoice.StatusExpired, invoice.StatusCancelled}

		assert.ElementsMatch(t, expected, transitions)
	})

	t.Run("GetAvailableEvents", func(t *testing.T) {
		testInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(testInvoice)

		events := ism.GetAvailableEvents()
		expected := []string{"view", "expire", "cancel"}

		assert.ElementsMatch(t, expected, events)
	})

	t.Run("Multiple transitions - valid business flow", func(t *testing.T) {
		freshInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(freshInvoice)

		// Created -> Pending
		err := ism.TransitionTo(invoice.StatusPending, "viewed", invoice.ActorCustomer, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, ism.CurrentStatus())

		// Pending -> Confirming
		err = ism.TransitionTo(invoice.StatusConfirming, "payment received", invoice.ActorSystem, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusConfirming, ism.CurrentStatus())

		// Confirming -> Paid
		err = ism.TransitionTo(invoice.StatusPaid, "confirmed", invoice.ActorSystem, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPaid, ism.CurrentStatus())

		// Check history
		history := ism.GetTransitionHistory()
		assert.Len(t, history, 3)
		assert.Equal(t, invoice.StatusCreated, history[0].FromStatus)
		assert.Equal(t, invoice.StatusPending, history[0].ToStatus)
		assert.Equal(t, invoice.StatusPending, history[1].FromStatus)
		assert.Equal(t, invoice.StatusConfirming, history[1].ToStatus)
		assert.Equal(t, invoice.StatusConfirming, history[2].FromStatus)
		assert.Equal(t, invoice.StatusPaid, history[2].ToStatus)
	})

	t.Run("Event-based flow", func(t *testing.T) {
		freshInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(freshInvoice)
		ctx := context.Background()

		// Created -> Pending (view event)
		err := ism.Event(ctx, "view", "customer viewed", invoice.ActorCustomer, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, ism.CurrentStatus())

		// Pending -> Confirming (full_payment event)
		err = ism.Event(ctx, "full_payment", "payment received", invoice.ActorSystem, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusConfirming, ism.CurrentStatus())

		// Confirming -> Paid (confirm event)
		err = ism.Event(ctx, "confirm", "payment confirmed", invoice.ActorSystem, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPaid, ism.CurrentStatus())

		// Check history
		history := ism.GetTransitionHistory()
		assert.Len(t, history, 3)
		assert.Equal(t, "customer viewed", history[0].Reason)
		assert.Equal(t, "payment received", history[1].Reason)
		assert.Equal(t, "payment confirmed", history[2].Reason)
	})

	t.Run("PaidAt timestamp", func(t *testing.T) {
		freshInvoice := createTestInvoice()
		ism := invoice.NewInvoiceStateMachine(freshInvoice)
		ctx := context.Background()

		// Follow proper business flow: created -> pending -> confirming -> paid
		err := ism.Event(ctx, "view", "customer viewed", invoice.ActorCustomer, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPending, ism.CurrentStatus())

		err = ism.Event(ctx, "full_payment", "payment received", invoice.ActorSystem, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusConfirming, ism.CurrentStatus())

		err = ism.Event(ctx, "confirm", "payment confirmed", invoice.ActorSystem, nil)
		assert.NoError(t, err)
		assert.Equal(t, invoice.StatusPaid, ism.CurrentStatus())

		// Check that paidAt was set
		assert.NotNil(t, ism.GetInvoice().PaidAt())
		assert.WithinDuration(t, time.Now().UTC(), *ism.GetInvoice().PaidAt(), time.Second)
	})
}

func TestStatusTransition(t *testing.T) {
	t.Run("NewStatusTransition", func(t *testing.T) {
		transition := invoice.NewStatusTransition(invoice.StatusCreated, invoice.StatusPending, "test reason", invoice.ActorCustomer, nil)

		assert.Equal(t, invoice.StatusCreated, transition.FromStatus)
		assert.Equal(t, invoice.StatusPending, transition.ToStatus)
		assert.Equal(t, "test reason", transition.Reason)
		assert.Equal(t, invoice.ActorCustomer, transition.Actor)
		assert.NotNil(t, transition.Timestamp)
		assert.Nil(t, transition.Metadata) // Can be nil when passed as nil
	})

	t.Run("String", func(t *testing.T) {
		transition := invoice.NewStatusTransition(invoice.StatusCreated, invoice.StatusPending, "test reason", invoice.ActorCustomer, nil)

		expected := "created -> pending (test reason)"
		assert.Equal(t, expected, transition.String())
	})

	t.Run("Equals", func(t *testing.T) {
		transition1 := invoice.NewStatusTransition(invoice.StatusCreated, invoice.StatusPending, "test reason", invoice.ActorCustomer, nil)
		transition2 := invoice.NewStatusTransition(invoice.StatusCreated, invoice.StatusPending, "test reason", invoice.ActorCustomer, nil)
		transition3 := invoice.NewStatusTransition(invoice.StatusPending, invoice.StatusPaid, "different reason", invoice.ActorSystem, nil)

		assert.True(t, transition1.Equals(transition2))
		assert.False(t, transition1.Equals(transition3))
		assert.False(t, transition1.Equals(nil))
	})
}

func TestGuardConditions(t *testing.T) {
	t.Run("canExpire - partial payment", func(t *testing.T) {
		testInvoice := createTestInvoice()
		// Set invoice to partial status
		testInvoice.SetStatus(invoice.StatusPartial)

		err := invoice.CanExpire(testInvoice)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot auto-expire invoices with partial payments")
	})

	t.Run("canExpire - not expired", func(t *testing.T) {
		testInvoice := createTestInvoice()
		// Set future expiration
		futureTime := time.Now().UTC().Add(24 * time.Hour)
		expiration, _ := invoice.NewInvoiceExpirationWithTime(futureTime)
		testInvoice.SetExpiration(expiration)

		err := invoice.CanExpire(testInvoice)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invoice has not expired yet")
	})

	t.Run("canCancel - terminal state", func(t *testing.T) {
		testInvoice := createTestInvoice()
		// Set invoice to terminal status
		testInvoice.SetStatus(invoice.StatusPaid)

		err := invoice.CanCancel(testInvoice)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot cancel invoice in terminal state")
	})

	t.Run("canMarkPaid - not confirming", func(t *testing.T) {
		testInvoice := createTestInvoice()
		// Set invoice to pending status
		testInvoice.SetStatus(invoice.StatusPending)

		err := invoice.CanMarkPaid(testInvoice)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only mark confirming invoices as paid")
	})

	t.Run("canRefund - not paid", func(t *testing.T) {
		testInvoice := createTestInvoice()
		// Set invoice to pending status
		testInvoice.SetStatus(invoice.StatusPending)

		err := invoice.CanRefund(testInvoice)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only refund paid invoices")
	})
}

// Helper function to create a test invoice
func createTestInvoice() *invoice.Invoice {
	// Create test money amounts
	subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
	tax, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
	total, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

	// Create test pricing
	pricing, _ := invoice.NewInvoicePricing(subtotal, tax, total)

	// Create test item
	item, _ := invoice.NewInvoiceItem("Test Item", "A test item", "1", subtotal)

	// Create test payment address
	paymentAddress, _ := shared.NewPaymentAddress("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", shared.NetworkBitcoin)

	// Create test exchange rate
	exchangeRate, _ := shared.NewExchangeRate("50000.00", shared.CurrencyUSD, shared.CryptoCurrencyBTC, "test-source", 1*time.Hour)

	// Create test payment tolerance
	tolerance, _ := invoice.NewPaymentTolerance("0.95", "1.05", invoice.OverpaymentActionRefund)

	// Create test expiration
	expiration := invoice.NewInvoiceExpiration(24 * time.Hour)

	// Create invoice
	testInvoice, _ := invoice.NewInvoice(
		"test-invoice-id",
		"test-merchant-id",
		"Test Invoice",
		"A test invoice",
		[]*invoice.InvoiceItem{item},
		pricing,
		shared.CryptoCurrencyBTC,
		paymentAddress,
		exchangeRate,
		tolerance,
		expiration,
		nil,
	)

	return testInvoice
}
