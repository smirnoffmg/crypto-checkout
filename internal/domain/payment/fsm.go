package payment

import (
	"context"
	"time"

	"crypto-checkout/internal/domain/shared"

	"github.com/looplab/fsm"
)

// PaymentFSM wraps the looplab/fsm for payment state management.
type PaymentFSM struct {
	*fsm.FSM
	payment *Payment
}

// NewPaymentFSM creates a new payment FSM.
func NewPaymentFSM(payment *Payment) *PaymentFSM {
	fsmInstance := fsm.NewFSM(
		string(payment.Status()),
		fsm.Events{
			// From detected state
			{Name: "include_in_block", Src: []string{string(StatusDetected)}, Dst: string(StatusConfirming)},
			{Name: "fail", Src: []string{string(StatusDetected)}, Dst: string(StatusFailed)},

			// From confirming state
			{Name: "confirm", Src: []string{string(StatusConfirming)}, Dst: string(StatusConfirmed)},
			{Name: "orphan", Src: []string{string(StatusConfirming)}, Dst: string(StatusOrphaned)},
			{Name: "fail", Src: []string{string(StatusConfirming)}, Dst: string(StatusFailed)},

			// From orphaned state
			{Name: "detect", Src: []string{string(StatusOrphaned)}, Dst: string(StatusDetected)},
			{Name: "fail", Src: []string{string(StatusOrphaned)}, Dst: string(StatusFailed)},

			// Terminal states have no outgoing transitions
		},
		fsm.Callbacks{
			// Guard conditions
			"before_include_in_block": func(ctx context.Context, e *fsm.Event) {
				if len(e.Args) > 0 {
					payment := e.Args[0].(*Payment)
					if err := CanIncludeInBlock(payment); err != nil {
						e.Cancel(err)
					}
				}
			},
			"before_confirm": func(ctx context.Context, e *fsm.Event) {
				if len(e.Args) > 0 {
					payment := e.Args[0].(*Payment)
					if err := CanConfirm(payment); err != nil {
						e.Cancel(err)
					}
				}
			},
			"before_orphan": func(ctx context.Context, e *fsm.Event) {
				if len(e.Args) > 0 {
					payment := e.Args[0].(*Payment)
					if err := CanOrphan(payment); err != nil {
						e.Cancel(err)
					}
				}
			},
			"before_detect": func(ctx context.Context, e *fsm.Event) {
				if len(e.Args) > 0 {
					payment := e.Args[0].(*Payment)
					if err := CanDetect(payment); err != nil {
						e.Cancel(err)
					}
				}
			},
			"before_fail": func(ctx context.Context, e *fsm.Event) {
				if len(e.Args) > 0 {
					payment := e.Args[0].(*Payment)
					if err := CanFail(payment); err != nil {
						e.Cancel(err)
					}
				}
			},

			// State entry callbacks
			"enter_confirmed": func(ctx context.Context, e *fsm.Event) {
				if len(e.Args) > 0 {
					payment := e.Args[0].(*Payment)
					now := time.Now().UTC()
					if payment.ConfirmedAt() == nil {
						payment.SetConfirmedAt(now)
					}
					// Update payment status to match FSM state
					payment.status = StatusConfirmed
					payment.timestamps.SetUpdatedAt(now)
				}
			},
			"enter_state": func(ctx context.Context, e *fsm.Event) {
				if len(e.Args) > 0 {
					payment := e.Args[0].(*Payment)
					// Update payment status to match FSM state
					payment.status = PaymentStatus(e.Dst)
					payment.timestamps.SetUpdatedAt(time.Now().UTC())
				}
			},
		},
	)

	return &PaymentFSM{
		FSM:     fsmInstance,
		payment: payment,
	}
}

// CurrentStatus returns the current payment status.
func (pfsm *PaymentFSM) CurrentStatus() PaymentStatus {
	return PaymentStatus(pfsm.FSM.Current())
}

// CanTransitionTo checks if the payment can transition to the target status.
func (pfsm *PaymentFSM) CanTransitionTo(target PaymentStatus) bool {
	return pfsm.payment.Status().CanTransitionTo(target)
}

// TransitionTo transitions the payment to the target status.
func (pfsm *PaymentFSM) TransitionTo(target PaymentStatus) error {
	if !pfsm.CanTransitionTo(target) {
		return NewInvalidPaymentTransitionError(string(pfsm.CurrentStatus()), string(target))
	}

	// Map status to event
	event := statusToEvent(pfsm.CurrentStatus(), target)
	if event == "" {
		return NewInvalidPaymentTransitionError(string(pfsm.CurrentStatus()), string(target))
	}

	ctx := context.Background()
	return pfsm.FSM.Event(ctx, event, pfsm.payment)
}

// Event triggers a payment event.
func (pfsm *PaymentFSM) Event(ctx context.Context, event string) error {
	return pfsm.FSM.Event(ctx, event, pfsm.payment)
}

// IsTerminal returns true if the payment is in a terminal state.
func (pfsm *PaymentFSM) IsTerminal() bool {
	return pfsm.CurrentStatus().IsTerminal()
}

// IsActive returns true if the payment is in an active state.
func (pfsm *PaymentFSM) IsActive() bool {
	return pfsm.CurrentStatus().IsActive()
}

// GetValidTransitions returns all valid transitions from the current state.
func (pfsm *PaymentFSM) GetValidTransitions() []PaymentStatus {
	current := pfsm.CurrentStatus()
	var valid []PaymentStatus

	for _, status := range []PaymentStatus{StatusDetected, StatusConfirming, StatusConfirmed, StatusOrphaned, StatusFailed} {
		if current.CanTransitionTo(status) {
			valid = append(valid, status)
		}
	}

	return valid
}

// Guard condition functions (made public for testing)

// CanIncludeInBlock checks if the payment can be included in a block.
func CanIncludeInBlock(payment *Payment) error {
	if payment.Status() != StatusDetected {
		return NewInvalidPaymentTransitionError(string(payment.Status()), string(StatusConfirming))
	}

	if payment.BlockInfo() == nil {
		return NewInvalidBlockInfoError("block information required for inclusion")
	}

	return nil
}

// CanConfirm checks if the payment can be confirmed.
func CanConfirm(payment *Payment) error {
	if payment.Status() != StatusConfirming {
		return NewInvalidPaymentTransitionError(string(payment.Status()), string(StatusConfirmed))
	}

	if !payment.IsConfirmed() {
		return NewInsufficientConfirmationsError(payment.Confirmations().Int(), payment.RequiredConfirmations())
	}

	return nil
}

// CanOrphan checks if the payment can be orphaned.
func CanOrphan(payment *Payment) error {
	if payment.Status() != StatusConfirming {
		return NewInvalidPaymentTransitionError(string(payment.Status()), string(StatusOrphaned))
	}

	return nil
}

// CanDetect checks if the payment can be detected again.
func CanDetect(payment *Payment) error {
	if payment.Status() != StatusOrphaned {
		return NewInvalidPaymentTransitionError(string(payment.Status()), string(StatusDetected))
	}

	return nil
}

// CanFail checks if the payment can fail.
func CanFail(payment *Payment) error {
	// Payment can fail from any non-terminal state
	if payment.Status().IsTerminal() {
		return shared.NewTerminalStateError(string(payment.Status()), "fail")
	}

	return nil
}

// Helper function to map status transitions to events
func statusToEvent(from, to PaymentStatus) string {
	transitions := map[string]string{
		string(StatusDetected) + "->" + string(StatusConfirming):  "include_in_block",
		string(StatusDetected) + "->" + string(StatusFailed):      "fail",
		string(StatusConfirming) + "->" + string(StatusConfirmed): "confirm",
		string(StatusConfirming) + "->" + string(StatusOrphaned):  "orphan",
		string(StatusConfirming) + "->" + string(StatusFailed):    "fail",
		string(StatusOrphaned) + "->" + string(StatusDetected):    "detect",
		string(StatusOrphaned) + "->" + string(StatusFailed):      "fail",
	}

	return transitions[string(from)+"->"+string(to)]
}
