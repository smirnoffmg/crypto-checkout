package invoice

import (
	"context"
	"errors"
	"time"

	"github.com/looplab/fsm"
)

// InvoiceFSM represents the finite state machine for invoice status transitions.
type InvoiceFSM struct {
	fsm     *fsm.FSM
	invoice *Invoice
}

// NewInvoiceFSM creates a new invoice finite state machine.
func NewInvoiceFSM(invoice *Invoice) *InvoiceFSM {
	// Define events and their transitions
	events := fsm.Events{
		// From created state
		{Name: "view", Src: []string{"created"}, Dst: "pending"},
		{Name: "expire", Src: []string{"created"}, Dst: "expired"},
		{Name: "cancel", Src: []string{"created"}, Dst: "cancelled"},

		// From pending state
		{Name: "partial_payment", Src: []string{"pending"}, Dst: "partial"},
		{Name: "full_payment", Src: []string{"pending"}, Dst: "confirming"},
		{Name: "expire", Src: []string{"pending"}, Dst: "expired"},
		{Name: "cancel", Src: []string{"pending"}, Dst: "cancelled"},

		// From partial state
		{Name: "full_payment", Src: []string{"partial"}, Dst: "confirming"},
		{Name: "cancel", Src: []string{"partial"}, Dst: "cancelled"},

		// From confirming state
		{Name: "confirm", Src: []string{"confirming"}, Dst: "paid"},
		{Name: "reorg", Src: []string{"confirming"}, Dst: "pending"}, // blockchain reorganization

		// From paid state
		{Name: "refund", Src: []string{"paid"}, Dst: "refunded"},
	}

	// Define callbacks for guard conditions and side effects
	callbacks := fsm.Callbacks{
		"before_expire": func(ctx context.Context, e *fsm.Event) {
			if len(e.Args) > 0 {
				if err := canExpire(e.Args[0].(*Invoice)); err != nil {
					e.Cancel(err)
				}
			}
		},
		"before_cancel": func(ctx context.Context, e *fsm.Event) {
			if len(e.Args) > 0 {
				if err := canCancel(e.Args[0].(*Invoice)); err != nil {
					e.Cancel(err)
				}
			}
		},
		"before_confirm": func(ctx context.Context, e *fsm.Event) {
			if len(e.Args) > 0 {
				if err := canMarkPaid(e.Args[0].(*Invoice)); err != nil {
					e.Cancel(err)
				}
			}
		},
		"before_refund": func(ctx context.Context, e *fsm.Event) {
			if len(e.Args) > 0 {
				if err := canRefund(e.Args[0].(*Invoice)); err != nil {
					e.Cancel(err)
				}
			}
		},
		"enter_paid": func(ctx context.Context, e *fsm.Event) {
			if len(e.Args) > 0 {
				invoice := e.Args[0].(*Invoice)
				now := time.Now().UTC()
				if invoice.paidAt == nil {
					invoice.paidAt = &now
				}
				// Update invoice status to match FSM state
				invoice.status = StatusPaid
				invoice.updatedAt = now
			}
		},
		"enter_state": func(ctx context.Context, e *fsm.Event) {
			if len(e.Args) > 0 {
				invoice := e.Args[0].(*Invoice)
				// Update invoice status to match FSM state
				invoice.status = InvoiceStatus(e.Dst)
				invoice.updatedAt = time.Now().UTC()
			}
		},
	}

	// Create the FSM with the initial state
	initialState := invoice.Status().String()
	fsmInstance := fsm.NewFSM(initialState, events, callbacks)

	return &InvoiceFSM{
		fsm:     fsmInstance,
		invoice: invoice,
	}
}

// Event triggers a state transition event.
func (ifs *InvoiceFSM) Event(ctx context.Context, event string) error {
	return ifs.fsm.Event(ctx, event, ifs.invoice)
}

// CanTransitionTo checks if a transition from current status to target status is valid.
func (ifs *InvoiceFSM) CanTransitionTo(target InvoiceStatus) bool {
	if !target.IsValid() {
		return false
	}

	// Get the event name that would lead to the target state
	eventName := ifs.getEventForTarget(target)
	if eventName == "" {
		return false
	}

	// Check if the event is available from current state
	return ifs.fsm.Can(eventName)
}

// TransitionTo attempts to transition to the target status.
func (ifs *InvoiceFSM) TransitionTo(target InvoiceStatus) error {
	eventName := ifs.getEventForTarget(target)
	if eventName == "" {
		return errors.New("invalid transition to " + target.String())
	}

	ctx := context.Background()
	return ifs.fsm.Event(ctx, eventName, ifs.invoice)
}

// CurrentStatus returns the current status.
func (ifs *InvoiceFSM) CurrentStatus() InvoiceStatus {
	statusStr := ifs.fsm.Current()
	return InvoiceStatus(statusStr)
}

// IsTerminal returns true if the current status is terminal.
func (ifs *InvoiceFSM) IsTerminal() bool {
	return ifs.CurrentStatus().IsTerminal()
}

// IsActive returns true if the current status is active (non-terminal).
func (ifs *InvoiceFSM) IsActive() bool {
	return ifs.CurrentStatus().IsActive()
}

// GetValidTransitions returns all valid transitions from the current status.
func (ifs *InvoiceFSM) GetValidTransitions() []InvoiceStatus {
	availableEvents := ifs.fsm.AvailableTransitions()
	var validTransitions []InvoiceStatus

	for _, event := range availableEvents {
		if target := ifs.getTargetForEvent(event); target != "" {
			validTransitions = append(validTransitions, InvoiceStatus(target))
		}
	}

	return validTransitions
}

// GetTransitionHistory returns the history of transitions (if tracked).
func (ifs *InvoiceFSM) GetTransitionHistory() []StatusTransition {
	// The looplab/fsm library doesn't provide built-in history tracking
	// This would need to be implemented separately if needed
	return []StatusTransition{}
}

// getEventForTarget returns the event name that leads to the target state.
func (ifs *InvoiceFSM) getEventForTarget(target InvoiceStatus) string {
	currentState := ifs.fsm.Current()

	// Map state transitions to events
	transitionMap := map[string]map[string]string{
		"created": {
			"pending":   "view",
			"expired":   "expire",
			"cancelled": "cancel",
		},
		"pending": {
			"partial":    "partial_payment",
			"confirming": "full_payment",
			"expired":    "expire",
			"cancelled":  "cancel",
		},
		"partial": {
			"confirming": "full_payment",
			"cancelled":  "cancel",
		},
		"confirming": {
			"paid":    "confirm",
			"pending": "reorg",
		},
		"paid": {
			"refunded": "refund",
		},
	}

	if stateMap, exists := transitionMap[currentState]; exists {
		if event, exists := stateMap[target.String()]; exists {
			return event
		}
	}

	return ""
}

// getTargetForEvent returns the target state for a given event.
func (ifs *InvoiceFSM) getTargetForEvent(event string) string {
	currentState := ifs.fsm.Current()

	// Map events to target states
	eventMap := map[string]map[string]string{
		"created": {
			"view":   "pending",
			"expire": "expired",
			"cancel": "cancelled",
		},
		"pending": {
			"partial_payment": "partial",
			"full_payment":    "confirming",
			"expire":          "expired",
			"cancel":          "cancelled",
		},
		"partial": {
			"full_payment": "confirming",
			"cancel":       "cancelled",
		},
		"confirming": {
			"confirm": "paid",
			"reorg":   "pending",
		},
		"paid": {
			"refund": "refunded",
		},
	}

	if stateMap, exists := eventMap[currentState]; exists {
		if target, exists := stateMap[event]; exists {
			return target
		}
	}

	return ""
}

// Guard condition implementations (now standalone functions)

// CanExpire checks if an invoice can be expired.
func CanExpire(invoice *Invoice) error {
	// Cannot expire invoices with partial payments
	if invoice.Status() == StatusPartial {
		return errors.New("cannot auto-expire invoices with partial payments")
	}

	// Check if invoice has actually expired
	if !invoice.Expiration().IsExpired() {
		return errors.New("invoice has not expired yet")
	}

	return nil
}

// CanCancel checks if an invoice can be cancelled.
func CanCancel(invoice *Invoice) error {
	// Cannot cancel invoices in terminal states
	if invoice.Status().IsTerminal() {
		return errors.New("cannot cancel invoice in terminal state")
	}

	return nil
}

// CanMarkPaid checks if an invoice can be marked as paid.
func CanMarkPaid(invoice *Invoice) error {
	// Can only mark confirming invoices as paid
	if invoice.Status() != StatusConfirming {
		return errors.New("can only mark confirming invoices as paid")
	}

	return nil
}

// CanRefund checks if an invoice can be refunded.
func CanRefund(invoice *Invoice) error {
	// Can only refund paid invoices
	if invoice.Status() != StatusPaid {
		return errors.New("can only refund paid invoices")
	}

	return nil
}

// Private versions for internal use
func canExpire(invoice *Invoice) error {
	return CanExpire(invoice)
}

func canCancel(invoice *Invoice) error {
	return CanCancel(invoice)
}

func canMarkPaid(invoice *Invoice) error {
	return CanMarkPaid(invoice)
}

func canRefund(invoice *Invoice) error {
	return CanRefund(invoice)
}

// StatusTransition represents a transition in the invoice state machine.
type StatusTransition struct {
	FromStatus InvoiceStatus
	ToStatus   InvoiceStatus
	Timestamp  time.Time
	Reason     string
	Actor      Actor
	Metadata   map[string]interface{}
}

// NewStatusTransition creates a new status transition record.
func NewStatusTransition(from, to InvoiceStatus, reason string, actor Actor, metadata map[string]interface{}) *StatusTransition {
	return &StatusTransition{
		FromStatus: from,
		ToStatus:   to,
		Timestamp:  time.Now().UTC(),
		Reason:     reason,
		Actor:      actor,
		Metadata:   metadata,
	}
}

// String returns the string representation of the status transition.
func (st *StatusTransition) String() string {
	return st.FromStatus.String() + " -> " + st.ToStatus.String() + " (" + st.Reason + ")"
}

// Equals returns true if this transition equals the other.
func (st *StatusTransition) Equals(other *StatusTransition) bool {
	if other == nil {
		return false
	}
	return st.FromStatus == other.FromStatus &&
		st.ToStatus == other.ToStatus &&
		st.Timestamp.Equal(other.Timestamp) &&
		st.Reason == other.Reason &&
		st.Actor == other.Actor
}

// InvoiceStateMachine provides a higher-level interface for managing invoice state transitions.
type InvoiceStateMachine struct {
	fsm     *InvoiceFSM
	invoice *Invoice
	history []StatusTransition
}

// NewInvoiceStateMachine creates a new invoice state machine for a specific invoice.
func NewInvoiceStateMachine(invoice *Invoice) *InvoiceStateMachine {
	fsm := NewInvoiceFSM(invoice)

	return &InvoiceStateMachine{
		fsm:     fsm,
		invoice: invoice,
		history: make([]StatusTransition, 0),
	}
}

// TransitionTo attempts to transition the invoice to a new status.
func (ism *InvoiceStateMachine) TransitionTo(target InvoiceStatus, reason string, actor Actor, metadata map[string]interface{}) error {
	fromStatus := ism.fsm.CurrentStatus()

	if err := ism.fsm.TransitionTo(target); err != nil {
		return err
	}

	// Record the transition
	transition := NewStatusTransition(fromStatus, target, reason, actor, metadata)
	ism.history = append(ism.history, *transition)

	// Update the invoice status
	ism.invoice.status = target
	ism.invoice.updatedAt = time.Now().UTC()

	return nil
}

// Event triggers a specific event in the state machine.
func (ism *InvoiceStateMachine) Event(ctx context.Context, event string, reason string, actor Actor, metadata map[string]interface{}) error {
	fromStatus := ism.fsm.CurrentStatus()

	if err := ism.fsm.Event(ctx, event); err != nil {
		return err
	}

	// Record the transition
	toStatus := ism.fsm.CurrentStatus()
	transition := NewStatusTransition(fromStatus, toStatus, reason, actor, metadata)
	ism.history = append(ism.history, *transition)

	// Update the invoice status
	ism.invoice.status = toStatus
	ism.invoice.updatedAt = time.Now().UTC()

	return nil
}

// CanTransitionTo checks if a transition to the target status is valid.
func (ism *InvoiceStateMachine) CanTransitionTo(target InvoiceStatus) bool {
	return ism.fsm.CanTransitionTo(target)
}

// CanEvent checks if a specific event can be triggered.
func (ism *InvoiceStateMachine) CanEvent(event string) bool {
	return ism.fsm.fsm.Can(event)
}

// CurrentStatus returns the current status.
func (ism *InvoiceStateMachine) CurrentStatus() InvoiceStatus {
	return ism.fsm.CurrentStatus()
}

// GetValidTransitions returns all valid transitions from the current status.
func (ism *InvoiceStateMachine) GetValidTransitions() []InvoiceStatus {
	return ism.fsm.GetValidTransitions()
}

// GetAvailableEvents returns all available events from the current state.
func (ism *InvoiceStateMachine) GetAvailableEvents() []string {
	return ism.fsm.fsm.AvailableTransitions()
}

// GetTransitionHistory returns the history of transitions.
func (ism *InvoiceStateMachine) GetTransitionHistory() []StatusTransition {
	return ism.history
}

// IsTerminal returns true if the current status is terminal.
func (ism *InvoiceStateMachine) IsTerminal() bool {
	return ism.fsm.IsTerminal()
}

// IsActive returns true if the current status is active.
func (ism *InvoiceStateMachine) IsActive() bool {
	return ism.fsm.IsActive()
}

// GetInvoice returns the associated invoice.
func (ism *InvoiceStateMachine) GetInvoice() *Invoice {
	return ism.invoice
}
