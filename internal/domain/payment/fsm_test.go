package payment_test

import (
	"context"
	"crypto-checkout/internal/domain/payment"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPaymentFSM(t *testing.T) {
	t.Run("NewPaymentFSM", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		require.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
		require.True(t, fsm.IsActive())
		require.False(t, fsm.IsTerminal())
	})

	t.Run("CanTransitionTo - valid transitions from detected", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		require.True(t, fsm.CanTransitionTo(payment.StatusConfirming))
		require.True(t, fsm.CanTransitionTo(payment.StatusFailed))
		require.False(t, fsm.CanTransitionTo(payment.StatusConfirmed))
		require.False(t, fsm.CanTransitionTo(payment.StatusOrphaned))
	})

	t.Run("CanTransitionTo - invalid status", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		require.False(t, fsm.CanTransitionTo(payment.PaymentStatus("invalid")))
	})

	t.Run("TransitionTo - valid transition", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		err := fsm.TransitionTo(payment.StatusFailed)
		require.NoError(t, err)
		require.Equal(t, payment.StatusFailed, fsm.CurrentStatus())
		require.Equal(t, payment.StatusFailed, testPayment.Status())
	})

	t.Run("TransitionTo - invalid transition", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		err := fsm.TransitionTo(payment.StatusConfirmed)
		require.Error(t, err)
		require.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
	})

	t.Run("Event - valid event", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		ctx := context.Background()
		err := fsm.Event(ctx, "fail")
		require.NoError(t, err)
		require.Equal(t, payment.StatusFailed, fsm.CurrentStatus())
		require.Equal(t, payment.StatusFailed, testPayment.Status())
	})

	t.Run("Event - invalid event", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		ctx := context.Background()
		err := fsm.Event(ctx, "invalid_event")
		require.Error(t, err)
		require.Contains(t, err.Error(), "event invalid_event does not exist")
		require.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
	})

	t.Run("GetValidTransitions", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		transitions := fsm.GetValidTransitions()
		expected := []payment.PaymentStatus{payment.StatusConfirming, payment.StatusFailed}

		require.ElementsMatch(t, expected, transitions)
	})

	t.Run("IsTerminal", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)

		// Detected is not terminal
		require.False(t, fsm.IsTerminal())

		// Test terminal states by creating payments in those states
		confirmedPayment := createTestPayment()
		confirmedPayment.SetStatus(payment.StatusConfirmed)
		fsmConfirmed := payment.NewPaymentFSM(confirmedPayment)
		require.True(t, fsmConfirmed.IsTerminal())

		failedPayment := createTestPayment()
		failedPayment.SetStatus(payment.StatusFailed)
		fsmFailed := payment.NewPaymentFSM(failedPayment)
		require.True(t, fsmFailed.IsTerminal())
	})

	t.Run("IsActive", func(t *testing.T) {
		// Test active states
		detectedPayment := createTestPayment()
		fsmDetected := payment.NewPaymentFSM(detectedPayment)
		require.True(t, fsmDetected.IsActive())

		confirmingPayment := createTestPayment()
		confirmingPayment.SetStatus(payment.StatusConfirming)
		fsmConfirming := payment.NewPaymentFSM(confirmingPayment)
		require.True(t, fsmConfirming.IsActive())

		orphanedPayment := createTestPayment()
		orphanedPayment.SetStatus(payment.StatusOrphaned)
		fsmOrphaned := payment.NewPaymentFSM(orphanedPayment)
		require.True(t, fsmOrphaned.IsActive())

		// Test terminal states
		confirmedPayment := createTestPayment()
		confirmedPayment.SetStatus(payment.StatusConfirmed)
		fsmConfirmed := payment.NewPaymentFSM(confirmedPayment)
		require.False(t, fsmConfirmed.IsActive())

		failedPayment := createTestPayment()
		failedPayment.SetStatus(payment.StatusFailed)
		fsmFailed := payment.NewPaymentFSM(failedPayment)
		require.False(t, fsmFailed.IsActive())
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
		require.NoError(t, err)
		require.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())
		require.Equal(t, payment.StatusConfirming, testPayment.Status())

		// Set sufficient confirmations so we can transition to confirmed
		err = testPayment.SetConfirmations(6)
		require.NoError(t, err)

		// Confirming -> Confirmed (sufficient confirmations)
		err = fsm.Event(ctx, "confirm")
		require.NoError(t, err)
		require.Equal(t, payment.StatusConfirmed, fsm.CurrentStatus())
		require.Equal(t, payment.StatusConfirmed, testPayment.Status())

		// Check that confirmedAt was set
		require.NotNil(t, testPayment.ConfirmedAt())
		require.WithinDuration(t, time.Now().UTC(), *testPayment.ConfirmedAt(), time.Second)
	})

	t.Run("Valid business flow - payment failure", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// Detected -> Failed (transaction failed)
		err := fsm.Event(ctx, "fail")
		require.NoError(t, err)
		require.Equal(t, payment.StatusFailed, fsm.CurrentStatus())
		require.Equal(t, payment.StatusFailed, testPayment.Status())
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
		require.NoError(t, err)
		require.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())

		// Confirming -> Orphaned (block orphaned)
		err = fsm.Event(ctx, "orphan")
		require.NoError(t, err)
		require.Equal(t, payment.StatusOrphaned, fsm.CurrentStatus())
		require.Equal(t, payment.StatusOrphaned, testPayment.Status())

		// Orphaned -> Detected (back to mempool)
		err = fsm.Event(ctx, "detect")
		require.NoError(t, err)
		require.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
		require.Equal(t, payment.StatusDetected, testPayment.Status())
	})

	t.Run("Invalid transitions - FSM prevents invalid events", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// Try to confirm payment from detected state (should fail - event not allowed)
		err := fsm.Event(ctx, "confirm")
		require.Error(t, err)
		require.Contains(t, err.Error(), "event confirm inappropriate in current state detected")
		require.Equal(t, payment.StatusDetected, fsm.CurrentStatus())

		// Try to orphan payment from detected state (should fail - event not allowed)
		err = fsm.Event(ctx, "orphan")
		require.Error(t, err)
		require.Contains(t, err.Error(), "event orphan inappropriate in current state detected")
		require.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
	})

	t.Run("Business rule enforcement - include in block", func(t *testing.T) {
		testPayment := createTestPayment()
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// Try to include in block without block info (should fail - guard condition)
		err := fsm.Event(ctx, "include_in_block")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid block information")
		require.Equal(t, payment.StatusDetected, fsm.CurrentStatus())

		// Add block info and try again
		err = testPayment.UpdateBlockInfo(12345, "blockhash123")
		require.NoError(t, err)

		err = fsm.Event(ctx, "include_in_block")
		require.NoError(t, err)
		require.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())
	})

	t.Run("Business rule enforcement - confirm payment", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)
		fsm := payment.NewPaymentFSM(testPayment)
		ctx := context.Background()

		// Try to confirm without sufficient confirmations (should fail - guard condition)
		err := fsm.Event(ctx, "confirm")
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient confirmations")
		require.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())

		// Set sufficient confirmations and try again
		err = testPayment.SetConfirmations(6)
		require.NoError(t, err)

		err = fsm.Event(ctx, "confirm")
		require.NoError(t, err)
		require.Equal(t, payment.StatusConfirmed, fsm.CurrentStatus())
	})
}

func TestGuardConditions(t *testing.T) {
	t.Run("CanIncludeInBlock - no block info", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanIncludeInBlock(testPayment)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid block information")
	})

	t.Run("CanIncludeInBlock - with block info", func(t *testing.T) {
		testPayment := createTestPayment()
		err := testPayment.UpdateBlockInfo(12345, "blockhash123")
		require.NoError(t, err)

		err = payment.CanIncludeInBlock(testPayment)
		require.NoError(t, err)
	})

	t.Run("CanIncludeInBlock - wrong status", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)

		err := payment.CanIncludeInBlock(testPayment)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid payment transition")
	})

	t.Run("CanConfirm - insufficient confirmations", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)

		err := payment.CanConfirm(testPayment)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient confirmations")
	})

	t.Run("CanConfirm - sufficient confirmations", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)
		err := testPayment.SetConfirmations(6)
		require.NoError(t, err)

		err = payment.CanConfirm(testPayment)
		require.NoError(t, err)
	})

	t.Run("CanConfirm - wrong status", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanConfirm(testPayment)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid payment transition")
	})

	t.Run("CanOrphan - wrong status", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanOrphan(testPayment)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid payment transition")
	})

	t.Run("CanOrphan - correct status", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirming)

		err := payment.CanOrphan(testPayment)
		require.NoError(t, err)
	})

	t.Run("CanDetect - wrong status", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanDetect(testPayment)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid payment transition")
	})

	t.Run("CanDetect - correct status", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusOrphaned)

		err := payment.CanDetect(testPayment)
		require.NoError(t, err)
	})

	t.Run("CanFail - terminal state", func(t *testing.T) {
		testPayment := createTestPayment()
		testPayment.SetStatus(payment.StatusConfirmed)

		err := payment.CanFail(testPayment)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot perform action in terminal state")
	})

	t.Run("CanFail - non-terminal state", func(t *testing.T) {
		testPayment := createTestPayment()

		err := payment.CanFail(testPayment)
		require.NoError(t, err)
	})
}
