package payment_test

import (
	"context"
	"testing"

	"crypto-checkout/internal/domain/payment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentStatusFSM_NewPaymentStatusFSM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		initialStatus payment.PaymentStatus
		expectError   bool
	}{
		{
			name:          "valid initial status - detected",
			initialStatus: payment.StatusDetected,
			expectError:   false,
		},
		{
			name:          "valid initial status - confirming",
			initialStatus: payment.StatusConfirming,
			expectError:   false,
		},
		{
			name:          "valid initial status - confirmed",
			initialStatus: payment.StatusConfirmed,
			expectError:   false,
		},
		{
			name:          "valid initial status - failed",
			initialStatus: payment.StatusFailed,
			expectError:   false,
		},
		{
			name:          "valid initial status - orphaned",
			initialStatus: payment.StatusOrphaned,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fsm, err := payment.NewPaymentStatusFSM(tt.initialStatus)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, fsm)
			} else {
				require.NoError(t, err)
				require.NotNil(t, fsm)
				assert.Equal(t, tt.initialStatus, fsm.CurrentStatus())
			}
		})
	}
}

func TestPaymentStatusFSM_CurrentStatus(t *testing.T) {
	t.Parallel()

	fsm := createFSM(t, payment.StatusDetected)
	assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())

	fsm = createFSM(t, payment.StatusConfirming)
	assert.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())
}

func TestPaymentStatusFSM_CanTransitionTo(t *testing.T) {
	t.Parallel()

	t.Run("from detected", testCanTransitionFromDetected)
	t.Run("from confirming", testCanTransitionFromConfirming)
	t.Run("from orphaned", testCanTransitionFromOrphaned)
	t.Run("from terminal states", testCanTransitionFromTerminal)
}

func testCanTransitionFromDetected(t *testing.T) {
	t.Parallel()
	fsm := createFSM(t, payment.StatusDetected)

	assert.True(t, fsm.CanTransitionTo(payment.StatusConfirming))
	assert.True(t, fsm.CanTransitionTo(payment.StatusFailed))
	assert.False(t, fsm.CanTransitionTo(payment.StatusConfirmed))
	assert.False(t, fsm.CanTransitionTo(payment.StatusOrphaned))
}

func testCanTransitionFromConfirming(t *testing.T) {
	t.Parallel()
	fsm := createFSM(t, payment.StatusConfirming)

	assert.True(t, fsm.CanTransitionTo(payment.StatusConfirmed))
	assert.True(t, fsm.CanTransitionTo(payment.StatusOrphaned))
	assert.True(t, fsm.CanTransitionTo(payment.StatusFailed))
	assert.False(t, fsm.CanTransitionTo(payment.StatusDetected))
}

func testCanTransitionFromOrphaned(t *testing.T) {
	t.Parallel()
	fsm := createFSM(t, payment.StatusOrphaned)

	assert.True(t, fsm.CanTransitionTo(payment.StatusDetected))
	assert.True(t, fsm.CanTransitionTo(payment.StatusFailed))
	assert.False(t, fsm.CanTransitionTo(payment.StatusConfirming))
}

func testCanTransitionFromTerminal(t *testing.T) {
	t.Parallel()

	confirmedFSM := createFSM(t, payment.StatusConfirmed)
	failedFSM := createFSM(t, payment.StatusFailed)

	assert.False(t, confirmedFSM.CanTransitionTo(payment.StatusDetected))
	assert.False(t, failedFSM.CanTransitionTo(payment.StatusDetected))
}

func TestPaymentStatusFSM_Fire(t *testing.T) {
	t.Parallel()

	t.Run("valid transitions", testValidFireTransitions)
	t.Run("invalid transitions", testInvalidFireTransitions)
}

func testValidFireTransitions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("detected to confirming", func(t *testing.T) {
		t.Parallel()
		fsm := createFSM(t, payment.StatusDetected)
		err := fsm.Fire(ctx, payment.TriggerIncluded)
		require.NoError(t, err)
		assert.Equal(t, payment.StatusConfirming, fsm.CurrentStatus())
	})

	t.Run("detected to failed", func(t *testing.T) {
		t.Parallel()
		fsm := createFSM(t, payment.StatusDetected)
		err := fsm.Fire(ctx, payment.TriggerFailed)
		require.NoError(t, err)
		assert.Equal(t, payment.StatusFailed, fsm.CurrentStatus())
	})

	t.Run("confirming to confirmed", func(t *testing.T) {
		t.Parallel()
		fsm := createFSM(t, payment.StatusConfirming)
		err := fsm.Fire(ctx, payment.TriggerConfirmed)
		require.NoError(t, err)
		assert.Equal(t, payment.StatusConfirmed, fsm.CurrentStatus())
	})

	t.Run("confirming to orphaned", func(t *testing.T) {
		t.Parallel()
		fsm := createFSM(t, payment.StatusConfirming)
		err := fsm.Fire(ctx, payment.TriggerOrphaned)
		require.NoError(t, err)
		assert.Equal(t, payment.StatusOrphaned, fsm.CurrentStatus())
	})

	t.Run("orphaned to detected", func(t *testing.T) {
		t.Parallel()
		fsm := createFSM(t, payment.StatusOrphaned)
		err := fsm.Fire(ctx, payment.TriggerBackToMempool)
		require.NoError(t, err)
		assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
	})
}

func testInvalidFireTransitions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("detected to confirmed", func(t *testing.T) {
		t.Parallel()
		fsm := createFSM(t, payment.StatusDetected)
		err := fsm.Fire(ctx, payment.TriggerConfirmed)
		require.Error(t, err)
		assert.Equal(t, payment.StatusDetected, fsm.CurrentStatus())
	})

	t.Run("confirmed to any", func(t *testing.T) {
		t.Parallel()
		fsm := createFSM(t, payment.StatusConfirmed)
		err := fsm.Fire(ctx, payment.TriggerDetected)
		require.Error(t, err)
		assert.Equal(t, payment.StatusConfirmed, fsm.CurrentStatus())
	})

	t.Run("failed to any", func(t *testing.T) {
		t.Parallel()
		fsm := createFSM(t, payment.StatusFailed)
		err := fsm.Fire(ctx, payment.TriggerDetected)
		require.Error(t, err)
		assert.Equal(t, payment.StatusFailed, fsm.CurrentStatus())
	})
}

func TestPaymentStatusFSM_GetPermittedTriggers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		status           payment.PaymentStatus
		expectedTriggers []payment.Trigger
	}{
		{
			name:             "detected status",
			status:           payment.StatusDetected,
			expectedTriggers: []payment.Trigger{payment.TriggerIncluded, payment.TriggerFailed},
		},
		{
			name:   "confirming status",
			status: payment.StatusConfirming,
			expectedTriggers: []payment.Trigger{
				payment.TriggerConfirmed,
				payment.TriggerOrphaned,
				payment.TriggerFailed,
			},
		},
		{
			name:             "orphaned status",
			status:           payment.StatusOrphaned,
			expectedTriggers: []payment.Trigger{payment.TriggerBackToMempool, payment.TriggerDropped},
		},
		{
			name:             "confirmed status (terminal)",
			status:           payment.StatusConfirmed,
			expectedTriggers: []payment.Trigger{},
		},
		{
			name:             "failed status (terminal)",
			status:           payment.StatusFailed,
			expectedTriggers: []payment.Trigger{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fsm := createFSM(t, tt.status)
			triggers := fsm.GetPermittedTriggers()

			assert.ElementsMatch(t, tt.expectedTriggers, triggers)
		})
	}
}

func TestPaymentStatusFSM_IsInState(t *testing.T) {
	t.Parallel()

	fsm := createFSM(t, payment.StatusDetected)

	assert.True(t, fsm.IsInState(payment.StatusDetected))
	assert.False(t, fsm.IsInState(payment.StatusConfirming))
	assert.False(t, fsm.IsInState(payment.StatusConfirmed))
}

func TestPaymentStatusFSM_ToGraph(t *testing.T) {
	t.Parallel()

	fsm := createFSM(t, payment.StatusDetected)
	graph := fsm.ToGraph()

	assert.NotEmpty(t, graph)
	assert.Contains(t, graph, "digraph")
}

func TestPaymentTrigger_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		trigger  payment.Trigger
		expected string
	}{
		{
			name:     "detected trigger",
			trigger:  payment.TriggerDetected,
			expected: "detected",
		},
		{
			name:     "included trigger",
			trigger:  payment.TriggerIncluded,
			expected: "included",
		},
		{
			name:     "confirmed trigger",
			trigger:  payment.TriggerConfirmed,
			expected: "confirmed",
		},
		{
			name:     "failed trigger",
			trigger:  payment.TriggerFailed,
			expected: "failed",
		},
		{
			name:     "orphaned trigger",
			trigger:  payment.TriggerOrphaned,
			expected: "orphaned",
		},
		{
			name:     "back to mempool trigger",
			trigger:  payment.TriggerBackToMempool,
			expected: "back_to_mempool",
		},
		{
			name:     "dropped trigger",
			trigger:  payment.TriggerDropped,
			expected: "dropped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.trigger.String())
		})
	}
}

// Helper function to create FSM for testing.
func createFSM(t *testing.T, status payment.PaymentStatus) *payment.StatusFSM {
	t.Helper()
	fsm, err := payment.NewPaymentStatusFSM(status)
	require.NoError(t, err)
	return fsm
}
