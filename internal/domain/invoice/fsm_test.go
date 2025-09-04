package invoice_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/domain/invoice"
)

func TestInvoiceStatusFSM_InitialState(t *testing.T) {
	t.Parallel()

	fsm := invoice.NewInvoiceStatusFSM(invoice.StatusCreated)
	assert.Equal(t, invoice.StatusCreated, fsm.CurrentStatus())
	assert.True(t, fsm.IsActive())
	assert.False(t, fsm.IsTerminal())
	assert.False(t, fsm.IsPaid())
}

func TestInvoiceStatusFSM_ValidTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialStatus  invoice.InvoiceStatus
		trigger        invoice.InvoiceTrigger
		expectedStatus invoice.InvoiceStatus
	}{
		// Created state transitions
		{"Created -> Pending (viewed)", invoice.StatusCreated, invoice.TriggerViewed, invoice.StatusPending},
		{"Created -> Expired", invoice.StatusCreated, invoice.TriggerExpired, invoice.StatusExpired},
		{"Created -> Cancelled", invoice.StatusCreated, invoice.TriggerCancelled, invoice.StatusCancelled},

		// Pending state transitions
		{"Pending -> Partial", invoice.StatusPending, invoice.TriggerPartial, invoice.StatusPartial},
		{
			"Pending -> Confirming (completed)",
			invoice.StatusPending,
			invoice.TriggerCompleted,
			invoice.StatusConfirming,
		},
		{"Pending -> Expired", invoice.StatusPending, invoice.TriggerExpired, invoice.StatusExpired},
		{"Pending -> Cancelled", invoice.StatusPending, invoice.TriggerCancelled, invoice.StatusCancelled},

		// Partial state transitions
		{
			"Partial -> Confirming (completed)",
			invoice.StatusPartial,
			invoice.TriggerCompleted,
			invoice.StatusConfirming,
		},
		{"Partial -> Cancelled", invoice.StatusPartial, invoice.TriggerCancelled, invoice.StatusCancelled},

		// Confirming state transitions
		{"Confirming -> Paid (confirmed)", invoice.StatusConfirming, invoice.TriggerConfirmed, invoice.StatusPaid},
		{"Confirming -> Pending (reorg)", invoice.StatusConfirming, invoice.TriggerReorg, invoice.StatusPending},

		// Paid state transitions
		{"Paid -> Refunded", invoice.StatusPaid, invoice.TriggerRefunded, invoice.StatusRefunded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsm := invoice.NewInvoiceStatusFSM(tt.initialStatus)
			err := fsm.Transition(context.Background(), tt.trigger)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, fsm.CurrentStatus())
		})
	}
}

func TestInvoiceStatusFSM_InvalidTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		initialStatus invoice.InvoiceStatus
		trigger       invoice.InvoiceTrigger
	}{
		// Invalid transitions from Created
		{"Created -> Partial (invalid)", invoice.StatusCreated, invoice.TriggerPartial},
		{"Created -> Completed (invalid)", invoice.StatusCreated, invoice.TriggerCompleted},
		{"Created -> Confirmed (invalid)", invoice.StatusCreated, invoice.TriggerConfirmed},
		{"Created -> Refunded (invalid)", invoice.StatusCreated, invoice.TriggerRefunded},
		{"Created -> Reorg (invalid)", invoice.StatusCreated, invoice.TriggerReorg},

		// Invalid transitions from Pending
		{"Pending -> Confirmed (invalid)", invoice.StatusPending, invoice.TriggerConfirmed},
		{"Pending -> Refunded (invalid)", invoice.StatusPending, invoice.TriggerRefunded},
		{"Pending -> Reorg (invalid)", invoice.StatusPending, invoice.TriggerReorg},

		// Invalid transitions from Partial
		{"Partial -> Viewed (invalid)", invoice.StatusPartial, invoice.TriggerViewed},
		{"Partial -> Confirmed (invalid)", invoice.StatusPartial, invoice.TriggerConfirmed},
		{"Partial -> Expired (invalid)", invoice.StatusPartial, invoice.TriggerExpired},
		{"Partial -> Refunded (invalid)", invoice.StatusPartial, invoice.TriggerRefunded},
		{"Partial -> Reorg (invalid)", invoice.StatusPartial, invoice.TriggerReorg},

		// Invalid transitions from Confirming
		{"Confirming -> Viewed (invalid)", invoice.StatusConfirming, invoice.TriggerViewed},
		{"Confirming -> Partial (invalid)", invoice.StatusConfirming, invoice.TriggerPartial},
		{"Confirming -> Completed (invalid)", invoice.StatusConfirming, invoice.TriggerCompleted},
		{"Confirming -> Expired (invalid)", invoice.StatusConfirming, invoice.TriggerExpired},
		{"Confirming -> Cancelled (invalid)", invoice.StatusConfirming, invoice.TriggerCancelled},
		{"Confirming -> Refunded (invalid)", invoice.StatusConfirming, invoice.TriggerRefunded},

		// Invalid transitions from Paid
		{"Paid -> Viewed (invalid)", invoice.StatusPaid, invoice.TriggerViewed},
		{"Paid -> Partial (invalid)", invoice.StatusPaid, invoice.TriggerPartial},
		{"Paid -> Completed (invalid)", invoice.StatusPaid, invoice.TriggerCompleted},
		{"Paid -> Confirmed (invalid)", invoice.StatusPaid, invoice.TriggerConfirmed},
		{"Paid -> Expired (invalid)", invoice.StatusPaid, invoice.TriggerExpired},
		{"Paid -> Cancelled (invalid)", invoice.StatusPaid, invoice.TriggerCancelled},
		{"Paid -> Reorg (invalid)", invoice.StatusPaid, invoice.TriggerReorg},

		// Terminal states should ignore all triggers
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsm := invoice.NewInvoiceStatusFSM(tt.initialStatus)
			err := fsm.Transition(context.Background(), tt.trigger)
			require.Error(t, err)
			assert.Equal(t, tt.initialStatus, fsm.CurrentStatus()) // Status should remain unchanged
		})
	}
}

func TestInvoiceStatusFSM_CanTransition(t *testing.T) {
	t.Parallel()

	fsm := invoice.NewInvoiceStatusFSM(invoice.StatusCreated)

	// Valid transitions
	assert.True(t, fsm.CanTransition(invoice.TriggerViewed))
	assert.True(t, fsm.CanTransition(invoice.TriggerExpired))
	assert.True(t, fsm.CanTransition(invoice.TriggerCancelled))

	// Invalid transitions
	assert.False(t, fsm.CanTransition(invoice.TriggerPartial))
	assert.False(t, fsm.CanTransition(invoice.TriggerCompleted))
	assert.False(t, fsm.CanTransition(invoice.TriggerConfirmed))
	assert.False(t, fsm.CanTransition(invoice.TriggerRefunded))
	assert.False(t, fsm.CanTransition(invoice.TriggerReorg))
}

func getPermittedTriggersTestCases() []struct {
	name               string
	status             invoice.InvoiceStatus
	expectedTriggers   []invoice.InvoiceTrigger
	unexpectedTriggers []invoice.InvoiceTrigger
} {
	return []struct {
		name               string
		status             invoice.InvoiceStatus
		expectedTriggers   []invoice.InvoiceTrigger
		unexpectedTriggers []invoice.InvoiceTrigger
	}{
		{
			name:   "Created state",
			status: invoice.StatusCreated,
			expectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerViewed, invoice.TriggerExpired, invoice.TriggerCancelled,
			},
			unexpectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerPartial, invoice.TriggerCompleted, invoice.TriggerConfirmed,
				invoice.TriggerRefunded, invoice.TriggerReorg,
			},
		},
		{
			name:   "Pending state",
			status: invoice.StatusPending,
			expectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerPartial, invoice.TriggerCompleted, invoice.TriggerExpired, invoice.TriggerCancelled,
			},
			unexpectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerViewed, invoice.TriggerConfirmed, invoice.TriggerRefunded, invoice.TriggerReorg,
			},
		},
		{
			name:   "Partial state",
			status: invoice.StatusPartial,
			expectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerCompleted, invoice.TriggerCancelled,
			},
			unexpectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerViewed, invoice.TriggerPartial, invoice.TriggerConfirmed,
				invoice.TriggerExpired, invoice.TriggerRefunded, invoice.TriggerReorg,
			},
		},
		{
			name:   "Confirming state",
			status: invoice.StatusConfirming,
			expectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerConfirmed, invoice.TriggerReorg,
			},
			unexpectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerViewed, invoice.TriggerPartial, invoice.TriggerCompleted,
				invoice.TriggerExpired, invoice.TriggerCancelled, invoice.TriggerRefunded,
			},
		},
		{
			name:   "Paid state",
			status: invoice.StatusPaid,
			expectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerRefunded,
			},
			unexpectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerViewed, invoice.TriggerPartial, invoice.TriggerCompleted,
				invoice.TriggerConfirmed, invoice.TriggerExpired, invoice.TriggerCancelled, invoice.TriggerReorg,
			},
		},
		{
			name:             "Terminal states",
			status:           invoice.StatusExpired,
			expectedTriggers: []invoice.InvoiceTrigger{}, // Terminal states ignore all triggers
			unexpectedTriggers: []invoice.InvoiceTrigger{
				invoice.TriggerViewed, invoice.TriggerPartial, invoice.TriggerCompleted,
				invoice.TriggerConfirmed, invoice.TriggerExpired, invoice.TriggerCancelled,
				invoice.TriggerRefunded, invoice.TriggerReorg,
			},
		},
	}
}

func TestInvoiceStatusFSM_GetPermittedTriggers(t *testing.T) {
	t.Parallel()

	tests := getPermittedTriggersTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsm := invoice.NewInvoiceStatusFSM(tt.status)
			permittedTriggers := fsm.GetPermittedTriggers()

			// Check expected triggers are present
			for _, expected := range tt.expectedTriggers {
				assert.Contains(t, permittedTriggers, expected)
			}

			// Check unexpected triggers are not present
			for _, unexpected := range tt.unexpectedTriggers {
				assert.NotContains(t, permittedTriggers, unexpected)
			}
		})
	}
}

func TestInvoiceStatusFSM_StateProperties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     invoice.InvoiceStatus
		isActive   bool
		isTerminal bool
		isPaid     bool
	}{
		{"Created", invoice.StatusCreated, true, false, false},
		{"Pending", invoice.StatusPending, true, false, false},
		{"Partial", invoice.StatusPartial, true, false, false},
		{"Confirming", invoice.StatusConfirming, true, false, false},
		{"Paid", invoice.StatusPaid, false, false, true},
		{"Expired", invoice.StatusExpired, false, true, false},
		{"Cancelled", invoice.StatusCancelled, false, true, false},
		{"Refunded", invoice.StatusRefunded, false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsm := invoice.NewInvoiceStatusFSM(tt.status)
			assert.Equal(t, tt.isActive, fsm.IsActive(), "IsActive() mismatch for %s", tt.name)
			assert.Equal(t, tt.isTerminal, fsm.IsTerminal(), "IsTerminal() mismatch for %s", tt.name)
			assert.Equal(t, tt.isPaid, fsm.IsPaid(), "IsPaid() mismatch for %s", tt.name)
		})
	}
}

func TestInvoiceStatusFSM_IsInState(t *testing.T) {
	t.Parallel()

	fsm := invoice.NewInvoiceStatusFSM(invoice.StatusCreated)
	assert.True(t, fsm.IsInState(invoice.StatusCreated))
	assert.False(t, fsm.IsInState(invoice.StatusPending))
	assert.False(t, fsm.IsInState(invoice.StatusPaid))

	// Transition to pending
	err := fsm.Transition(context.Background(), invoice.TriggerViewed)
	require.NoError(t, err)
	assert.False(t, fsm.IsInState(invoice.StatusCreated))
	assert.True(t, fsm.IsInState(invoice.StatusPending))
	assert.False(t, fsm.IsInState(invoice.StatusPaid))
}

func TestInvoiceStatusFSM_ComplexWorkflow(t *testing.T) {
	t.Parallel()

	// Test a complete workflow: Created -> Pending -> Partial -> Confirming -> Paid -> Refunded
	fsm := invoice.NewInvoiceStatusFSM(invoice.StatusCreated)

	// Created -> Pending
	err := fsm.Transition(context.Background(), invoice.TriggerViewed)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())

	// Pending -> Partial
	err = fsm.Transition(context.Background(), invoice.TriggerPartial)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPartial, fsm.CurrentStatus())

	// Partial -> Confirming
	err = fsm.Transition(context.Background(), invoice.TriggerCompleted)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusConfirming, fsm.CurrentStatus())

	// Confirming -> Paid
	err = fsm.Transition(context.Background(), invoice.TriggerConfirmed)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPaid, fsm.CurrentStatus())

	// Paid -> Refunded
	err = fsm.Transition(context.Background(), invoice.TriggerRefunded)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusRefunded, fsm.CurrentStatus())

	// Verify final state properties
	assert.False(t, fsm.IsActive())
	assert.True(t, fsm.IsTerminal())
	assert.True(t, fsm.IsPaid())
}

func TestInvoiceStatusFSM_ReorgWorkflow(t *testing.T) {
	t.Parallel()

	// Test reorg workflow: Created -> Pending -> Confirming -> Pending (reorg) -> Confirming -> Paid
	fsm := invoice.NewInvoiceStatusFSM(invoice.StatusCreated)

	// Created -> Pending
	err := fsm.Transition(context.Background(), invoice.TriggerViewed)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())

	// Pending -> Confirming
	err = fsm.Transition(context.Background(), invoice.TriggerCompleted)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusConfirming, fsm.CurrentStatus())

	// Confirming -> Pending (reorg)
	err = fsm.Transition(context.Background(), invoice.TriggerReorg)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPending, fsm.CurrentStatus())

	// Pending -> Confirming (again)
	err = fsm.Transition(context.Background(), invoice.TriggerCompleted)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusConfirming, fsm.CurrentStatus())

	// Confirming -> Paid
	err = fsm.Transition(context.Background(), invoice.TriggerConfirmed)
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPaid, fsm.CurrentStatus())
}
