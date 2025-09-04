package payment

import (
	"context"
	"fmt"

	"github.com/qmuntal/stateless"
)

// Trigger represents the triggers for payment status transitions.
type Trigger string

// Payment trigger constants based on PAYMENT_STATUSES.md specification.
const (
	// TriggerDetected represents a transaction found in mempool or block.
	TriggerDetected Trigger = "detected"
	// TriggerIncluded represents a transaction included in block.
	TriggerIncluded Trigger = "included"
	// TriggerConfirmed represents sufficient confirmations received.
	TriggerConfirmed Trigger = "confirmed"
	// TriggerFailed represents a transaction failed or reverted.
	TriggerFailed Trigger = "failed"
	// TriggerOrphaned represents a block containing tx was orphaned.
	TriggerOrphaned Trigger = "orphaned"
	// TriggerBackToMempool represents a transaction back in mempool after reorg.
	TriggerBackToMempool Trigger = "back_to_mempool"
	// TriggerDropped represents a transaction dropped permanently.
	TriggerDropped Trigger = "dropped"
)

// String returns the string representation of the trigger.
func (t Trigger) String() string {
	return string(t)
}

// StatusFSM manages payment status transitions using a finite state machine.
type StatusFSM struct {
	stateMachine *stateless.StateMachine
}

// NewPaymentStatusFSM creates a new PaymentStatusFSM with the given initial status.
func NewPaymentStatusFSM(initialStatus PaymentStatus) (*StatusFSM, error) {
	sm := stateless.NewStateMachine(stateless.State(initialStatus))
	fsm := &StatusFSM{stateMachine: sm}

	fsm.configureTransitions()

	return fsm, nil
}

// configureTransitions configures all valid state transitions based on PAYMENT_STATUSES.md.
func (fsm *StatusFSM) configureTransitions() {
	// Configure detected state transitions
	fsm.stateMachine.Configure(StatusDetected).
		Permit(TriggerIncluded, StatusConfirming).
		Permit(TriggerFailed, StatusFailed)

	// Configure confirming state transitions
	fsm.stateMachine.Configure(StatusConfirming).
		Permit(TriggerConfirmed, StatusConfirmed).
		Permit(TriggerOrphaned, StatusOrphaned).
		Permit(TriggerFailed, StatusFailed)

	// Configure orphaned state transitions
	fsm.stateMachine.Configure(StatusOrphaned).
		Permit(TriggerBackToMempool, StatusDetected).
		Permit(TriggerDropped, StatusFailed)

	// Configure terminal states (no transitions allowed)
	fsm.stateMachine.Configure(StatusConfirmed).
		Ignore(TriggerDetected).
		Ignore(TriggerIncluded).
		Ignore(TriggerConfirmed).
		Ignore(TriggerFailed).
		Ignore(TriggerOrphaned).
		Ignore(TriggerBackToMempool).
		Ignore(TriggerDropped)

	fsm.stateMachine.Configure(StatusFailed).
		Ignore(TriggerDetected).
		Ignore(TriggerIncluded).
		Ignore(TriggerConfirmed).
		Ignore(TriggerFailed).
		Ignore(TriggerOrphaned).
		Ignore(TriggerBackToMempool).
		Ignore(TriggerDropped)
}

// CurrentStatus returns the current payment status.
func (fsm *StatusFSM) CurrentStatus() PaymentStatus {
	state, err := fsm.stateMachine.State(context.Background())
	if err != nil {
		// This should never happen in normal operation
		panic("failed to get current state: " + err.Error())
	}
	return state.(PaymentStatus) //nolint:errcheck // Type assertion is safe here
}

// CanTransitionTo checks if a transition to the given status is possible.
func (fsm *StatusFSM) CanTransitionTo(status PaymentStatus) bool {
	currentStatus := fsm.CurrentStatus()

	// Check valid transitions based on PAYMENT_STATUSES.md
	switch currentStatus {
	case StatusDetected:
		return status == StatusConfirming || status == StatusFailed
	case StatusConfirming:
		return status == StatusConfirmed || status == StatusOrphaned || status == StatusFailed
	case StatusOrphaned:
		return status == StatusDetected || status == StatusFailed
	case StatusConfirmed, StatusFailed:
		return false // Terminal states
	default:
		return false
	}
}

// CanTransition checks if a transition is possible from the current state.
func (fsm *StatusFSM) CanTransition(trigger Trigger) bool {
	canFire, _ := fsm.stateMachine.CanFire(stateless.Trigger(trigger))
	return canFire
}

// Fire triggers a state transition with the given trigger.
func (fsm *StatusFSM) Fire(_ context.Context, trigger Trigger) error {
	// Check if this is a valid transition by checking if the trigger is permitted
	permittedTriggers := fsm.GetPermittedTriggers()
	isPermitted := false
	for _, permitted := range permittedTriggers {
		if permitted == trigger {
			isPermitted = true
			break
		}
	}

	if !isPermitted {
		return fmt.Errorf("invalid transition: %s from state %s", string(trigger), string(fsm.CurrentStatus()))
	}

	return fsm.stateMachine.Fire(stateless.Trigger(trigger)) //nolint:wrapcheck // Error is already descriptive
}

// FireWithParameters triggers a state transition with the given trigger and parameters.
func (fsm *StatusFSM) FireWithParameters(_ context.Context, trigger Trigger, args ...any) error {
	return fsm.stateMachine.Fire(stateless.Trigger(trigger), args...) //nolint:wrapcheck // Error is already descriptive
}

// GetPermittedTriggers returns all triggers that are currently permitted.
// Note: This excludes ignored triggers (terminal states ignore all triggers).
func (fsm *StatusFSM) GetPermittedTriggers() []Trigger {
	triggers, err := fsm.stateMachine.PermittedTriggers()
	if err != nil {
		// This should never happen in normal operation
		panic("failed to get permitted triggers: " + err.Error())
	}

	// Filter out ignored triggers for terminal states
	currentStatus := fsm.CurrentStatus()
	if currentStatus.IsTerminal() {
		// Terminal states ignore all triggers, so return empty slice
		return []Trigger{}
	}

	result := make([]Trigger, len(triggers))
	for i, trigger := range triggers {
		result[i] = trigger.(Trigger) //nolint:errcheck // Type assertion is safe here
	}
	return result
}

// IsInState checks if the FSM is in the given status.
func (fsm *StatusFSM) IsInState(status PaymentStatus) bool {
	isInState, _ := fsm.stateMachine.IsInState(stateless.State(status))
	return isInState
}

// ToGraph returns a DOT graph representation of the state machine.
func (fsm *StatusFSM) ToGraph() string {
	return fsm.stateMachine.ToGraph()
}
