package payment_test

import (
	"context"
	"testing"
	"time"

	"crypto-checkout/internal/domain/payment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentFSM(t *testing.T) {
	t.Run("NewPaymentFSM", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
		assert.True(t, fsm.IsActive())
		assert.False(t, fsm.IsTerminal())
	})

	t.Run("CanTransitionTo - valid transitions from detected", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		assert.True(t, fsm.CanTransitionTo(payment.StatusConfirming))
		assert.True(t, fsm.CanTransitionTo(payment.StatusFailed))
		assert.False(t, fsm.CanTransitionTo(payment.StatusConfirmed))
		assert.False(t, fsm.CanTransitionTo(payment.StatusOrphaned))
	})

	t.Run("CanTransitionTo - invalid status", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		assert.False(t, fsm.CanTransitionTo(payment.PaymentStatus("invalid")))
	})

	t.Run("TransitionTo - valid transition", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		err := fsm.TransitionTo(payment.StatusFailed)
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusFailed, fsm.CurrentStatus())
		assert.Equal(t, payment.StatusFailed, testPayment.Status())
	})

	t.Run("TransitionTo - invalid transition", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		err := fsm.TransitionTo(payment.StatusConfirmed)
		assert.Error(t, err)
		assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
	})

	t.Run("Event - valid event", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		ctx := context.Background()
		err := fsm.Event(ctx, "fail")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusFailed, fsm.CurrentStatus())
		assert.Equal(t, payment.StatusFailed, testPayment.Status())
	})

	t.Run("Event - invalid event", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		ctx := context.Background()
		err := fsm.Event(ctx, "invalid_event")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event invalid_event does not exist")
		assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
	})

	t.Run("GetValidTransitions", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		transitions := fsm.GetValidTransitions()
		expected := []payment.PaymentStatus{payment.StatusConfirming, payment.StatusFailed}

		assert.ElementsMatch(t, expected, transitions)
	})

	t.Run("IsTerminal", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		// Detected is not terminal
		assert.False(t, fsm.IsTerminal())

		// Test terminal states by creating payments in those states
		confirmedPayment := createTestPayment()
		confirmedPayment.SetStatus(payment.StatusConfirmed)
		fsmConfirmed := payment.NewPaymentFSM(confirmedPayment)
		assert.True(t, fsmConfirmed.IsTerminal())

		failedPayment := createTestPayment()
		failedPayment.SetStatus(payment.StatusFailed)
		fsmFailed := payment.NewPaymentFSM(failedPayment)
		assert.True(t, fsmFailed.IsTerminal())
	})

	t.Run("IsActive", func(t *testing.T) {
		// Test active states
		detectedPayment := createTestPayment()
		fsmDetected := payment.NewPaymentFSM(detectedPayment)
		assert.True(t, fsmDetected.IsActive())

		confirmingPayment := createTestPayment()
		confirmingPayment.SetStatus(payment.StatusConfirming)
		fsmConfirming := payment.NewPaymentFSM(confirmingPayment)
		assert.True(t, fsmConfirming.IsActive())

		orphanedPayment := createTestPayment()
		orphanedPayment.SetStatus(payment.StatusOrphaned)
		fsmOrphaned := payment.NewPaymentFSM(orphanedPayment)
		assert.True(t, fsmOrphaned.IsActive())

		// Test terminal states
		confirmedPayment := createTestPayment()
		confirmedPayment.SetStatus(payment.StatusConfirmed)
		fsmConfirmed := payment.NewPaymentFSM(confirmedPayment)
		assert.False(t, fsmConfirmed.IsActive())

		failedPayment := createTestPayment()
		failedPayment.SetStatus(payment.StatusFailed)
		fsmFailed := payment.NewPaymentFSM(failedPayment)
		assert.False(t, fsmFailed.IsActive())
	})

	t.Run("Valid business flow - complete payment cycle", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// First, add block info so we can transition to confirming
		err := testPayment.UpdateBlockInfo(12345, "blockhash123")
		require.NoError(t, err)

		// Detected -> Confirming (included in block)
		err = fsm.Event(ctx, "include_in_block")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())
		assert.Equal(t, payment.StatusConfirming, testPayment.Status())

		// Set sufficient confirmations so we can transition to confirmed
		err = testPayment.SetConfirmations(6)
		require.NoError(t, err)

		// Confirming -> Confirmed (sufficient confirmations)
		err = fsm.Event(ctx, "confirm")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusConfirmed, fsm.CurrentStatus())
		assert.Equal(t, payment.StatusConfirmed, testPayment.Status())

		// Check that confirmedAt was set
		assert.NotNil(t, testPayment.ConfirmedAt())
		assert.WithinDuration(t, time.Now().UTC(), *testPayment.ConfirmedAt(), time.Second)
	})

	t.Run("Valid business flow - payment failure", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// Detected -> Failed (transaction failed)
		err := fsm.Event(ctx, "fail")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusFailed, fsm.CurrentStatus())
		assert.Equal(t, payment.StatusFailed, testPayment.Status())
	})

	t.Run("Valid business flow - orphaned payment", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// First, add block info so we can transition to confirming
		err := testPayment.UpdateBlockInfo(12345, "blockhash123")
		require.NoError(t, err)

		// Detected -> Confirming (included in block)
		err = fsm.Event(ctx, "include_in_block")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())

		// Confirming -> Orphaned (block orphaned)
		err = fsm.Event(ctx, "orphan")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusOrphaned, fsm.CurrentStatus())
		assert.Equal(t, payment.StatusOrphaned, testPayment.Status())

		// Orphaned -> Detected (back to mempool)
		err = fsm.Event(ctx, "detect")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
		assert.Equal(t, payment.StatusDetected, testPayment.Status())
	})

	t.Run("Invalid transitions - FSM prevents invalid events", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// Try to confirm payment from detected state (should fail - event not allowed)
		err := fsm.Event(ctx, "confirm")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event confirm inappropriate in current state detected")
		assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())

		// Try to orphan payment from detected state (should fail - event not allowed)
		err = fsm.Event(ctx, "orphan")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event orphan inappropriate in current state detected")
		assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
	})

	t.Run("Business rule enforcement - include in block", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// Try to include in block without block info (should fail - guard condition)
		err := fsm.Event(ctx, "include_in_block")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid block information")
		assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())

		// Add block info and try again
		err = testPayment.UpdateBlockInfo(12345, "blockhash123")
		require.NoError(t, err)

		err = fsm.Event(ctx, "include_in_block")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())
	})

	t.Run("Business rule enforcement - confirm payment", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// Try to confirm without sufficient confirmations (should fail - guard condition)
		err := fsm.Event(ctx, "confirm")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient confirmations")
		assert.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())

		// Set sufficient confirmations and try again
		err = testPayment.SetConfirmations(6)
		require.NoError(t, err)

		err = fsm.Event(ctx, "confirm")
		assert.NoError(t, err)
		assert.Equal(t, payment.StatusConfirmed, fsm.CurrentStatus())
	})
}

func TestGuardConditions(t *testing.T) {
	t.Run("CanIncludeInBlock - no block info", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanIncludeInBlock(testPayment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid block information")
	})

	t.Run("CanIncludeInBlock - with block info", func(t *testing.T) {
		testPayment := createTestPayment()
		err := testPayment.UpdateBlockInfo(12345, "blockhash123")
		require.NoError(t, err)

		err = payment.CanIncludeInBlock(testPayment)
		assert.NoError(t, err)
	})

	t.Run("CanIncludeInBlock - wrong status", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)

		err := payment.CanIncludeInBlock(testPayment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payment transition")
	})

	t.Run("CanConfirm - insufficient confirmations", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)

		err := payment.CanConfirm(testPayment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient confirmations")
	})

	t.Run("CanConfirm - sufficient confirmations", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)
		err := testPayment.SetConfirmations(6)
		require.NoError(t, err)

		err = payment.CanConfirm(testPayment)
		assert.NoError(t, err)
	})

	t.Run("CanConfirm - wrong status", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanConfirm(testPayment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payment transition")
	})

	t.Run("CanOrphan - wrong status", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanOrphan(testPayment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payment transition")
	})

	t.Run("CanOrphan - correct status", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)

		err := payment.CanOrphan(testPayment)
		assert.NoError(t, err)
	})

	t.Run("CanDetect - wrong status", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanDetect(testPayment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payment transition")
	})

	t.Run("CanDetect - correct status", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusOrphaned)

		err := payment.CanDetect(testPayment)
		assert.NoError(t, err)
	})

	t.Run("CanFail - terminal state", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirmed)

		err := payment.CanFail(testPayment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot perform action in terminal state")
	})

	t.Run("CanFail - non-terminal state", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanFail(testPayment)
		assert.NoError(t, err)
	})
}
