package invoice

import (
	"context"
	"errors"

	"github.com/qmuntal/stateless"
)

type InvoiceTrigger string

const (
	TriggerViewed    InvoiceTrigger = "viewed"
	TriggerPartial   InvoiceTrigger = "partial"
	TriggerCompleted InvoiceTrigger = "completed"
	TriggerConfirmed InvoiceTrigger = "confirmed"
	TriggerExpired   InvoiceTrigger = "expired"
	TriggerCancelled InvoiceTrigger = "cancelled"
	TriggerRefunded  InvoiceTrigger = "refunded"
	TriggerReorg     InvoiceTrigger = "reorg"
)

// InvoiceStatusFSM manages invoice status transitions using a finite state machine.
type InvoiceStatusFSM struct {
	machine *stateless.StateMachine
}

// NewInvoiceStatusFSM creates a new invoice status FSM starting from the given status.
func NewInvoiceStatusFSM(initialStatus InvoiceStatus) *InvoiceStatusFSM {
	fsm := &InvoiceStatusFSM{
		machine: stateless.NewStateMachine(stateless.State(initialStatus)),
	}
	fsm.configureTransitions()
	return fsm
}

// configureTransitions sets up all valid state transitions based on the mermaid diagram.
func (fsm *InvoiceStatusFSM) configureTransitions() {
	// Created state transitions
	fsm.machine.Configure(stateless.State(StatusCreated)).
		Permit(stateless.Trigger(TriggerViewed), stateless.State(StatusPending)).
		Permit(stateless.Trigger(TriggerExpired), stateless.State(StatusExpired)).
		Permit(stateless.Trigger(TriggerCancelled), stateless.State(StatusCancelled))

	// Pending state transitions
	fsm.machine.Configure(stateless.State(StatusPending)).
		Permit(stateless.Trigger(TriggerPartial), stateless.State(StatusPartial)).
		Permit(stateless.Trigger(TriggerCompleted), stateless.State(StatusConfirming)).
		Permit(stateless.Trigger(TriggerExpired), stateless.State(StatusExpired)).
		Permit(stateless.Trigger(TriggerCancelled), stateless.State(StatusCancelled))

	// Partial state transitions
	fsm.machine.Configure(stateless.State(StatusPartial)).
		Permit(stateless.Trigger(TriggerCompleted), stateless.State(StatusConfirming)).
		Permit(stateless.Trigger(TriggerCancelled), stateless.State(StatusCancelled))

	// Confirming state transitions
	fsm.machine.Configure(stateless.State(StatusConfirming)).
		Permit(stateless.Trigger(TriggerConfirmed), stateless.State(StatusPaid)).
		Permit(stateless.Trigger(TriggerReorg), stateless.State(StatusPending))

	// Paid state transitions
	fsm.machine.Configure(stateless.State(StatusPaid)).
		Permit(stateless.Trigger(TriggerRefunded), stateless.State(StatusRefunded))

	// Terminal states (Expired, Cancelled, Refunded) have no transitions
	fsm.machine.Configure(stateless.State(StatusExpired)).
		Ignore(stateless.Trigger(TriggerViewed)).
		Ignore(stateless.Trigger(TriggerPartial)).
		Ignore(stateless.Trigger(TriggerCompleted)).
		Ignore(stateless.Trigger(TriggerConfirmed)).
		Ignore(stateless.Trigger(TriggerExpired)).
		Ignore(stateless.Trigger(TriggerCancelled)).
		Ignore(stateless.Trigger(TriggerRefunded)).
		Ignore(stateless.Trigger(TriggerReorg))

	fsm.machine.Configure(stateless.State(StatusCancelled)).
		Ignore(stateless.Trigger(TriggerViewed)).
		Ignore(stateless.Trigger(TriggerPartial)).
		Ignore(stateless.Trigger(TriggerCompleted)).
		Ignore(stateless.Trigger(TriggerConfirmed)).
		Ignore(stateless.Trigger(TriggerExpired)).
		Ignore(stateless.Trigger(TriggerCancelled)).
		Ignore(stateless.Trigger(TriggerRefunded)).
		Ignore(stateless.Trigger(TriggerReorg))

	fsm.machine.Configure(stateless.State(StatusRefunded)).
		Ignore(stateless.Trigger(TriggerViewed)).
		Ignore(stateless.Trigger(TriggerPartial)).
		Ignore(stateless.Trigger(TriggerCompleted)).
		Ignore(stateless.Trigger(TriggerConfirmed)).
		Ignore(stateless.Trigger(TriggerExpired)).
		Ignore(stateless.Trigger(TriggerCancelled)).
		Ignore(stateless.Trigger(TriggerRefunded)).
		Ignore(stateless.Trigger(TriggerReorg))
}

// CurrentStatus returns the current status of the FSM.
func (fsm *InvoiceStatusFSM) CurrentStatus() InvoiceStatus {
	state, err := fsm.machine.State(context.Background())
	if err != nil {
		// This should never happen in normal operation
		panic("failed to get current state: " + err.Error())
	}
	return state.(InvoiceStatus) //nolint:errcheck // Type assertion is safe here
}

// CanTransition checks if a transition is possible from the current state.
func (fsm *InvoiceStatusFSM) CanTransition(trigger InvoiceTrigger) bool {
	canFire, _ := fsm.machine.CanFire(stateless.Trigger(trigger))
	return canFire
}

// Transition attempts to transition to a new state using the given trigger.
func (fsm *InvoiceStatusFSM) Transition(_ context.Context, trigger InvoiceTrigger) error {
	if !fsm.CanTransition(trigger) {
		return errors.New("invalid transition: " + string(trigger) + " from state " + string(fsm.CurrentStatus()))
	}

	return fsm.machine.Fire(stateless.Trigger(trigger)) //nolint:wrapcheck // Error is already descriptive
}

// GetPermittedTriggers returns all triggers that are valid from the current state.
// Note: This excludes ignored triggers (terminal states ignore all triggers).
func (fsm *InvoiceStatusFSM) GetPermittedTriggers() []InvoiceTrigger {
	triggers, err := fsm.machine.PermittedTriggers()
	if err != nil {
		// This should never happen in normal operation
		panic("failed to get permitted triggers: " + err.Error())
	}

	// Filter out ignored triggers for terminal states
	currentStatus := fsm.CurrentStatus()
	if currentStatus.IsTerminal() {
		// Terminal states ignore all triggers, so return empty slice
		return []InvoiceTrigger{}
	}

	result := make([]InvoiceTrigger, len(triggers))
	for i, trigger := range triggers {
		result[i] = trigger.(InvoiceTrigger) //nolint:errcheck // Type assertion is safe here
	}
	return result
}

// IsInState checks if the FSM is in the given state.
func (fsm *InvoiceStatusFSM) IsInState(status InvoiceStatus) bool {
	isInState, _ := fsm.machine.IsInState(stateless.State(status))
	return isInState
}

// IsTerminal checks if the current state is terminal.
func (fsm *InvoiceStatusFSM) IsTerminal() bool {
	return fsm.CurrentStatus().IsTerminal()
}

// IsActive checks if the current state allows further processing.
func (fsm *InvoiceStatusFSM) IsActive() bool {
	return fsm.CurrentStatus().IsActive()
}

// IsPaid checks if the current state indicates payment completion.
func (fsm *InvoiceStatusFSM) IsPaid() bool {
	return fsm.CurrentStatus().IsPaid()
}
