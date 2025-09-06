package invoice

// InvoiceStatus represents the current status of an invoice.
type InvoiceStatus string

const (
	// StatusCreated - Invoice just created, not yet viewed
	StatusCreated InvoiceStatus = "created"
	// StatusPending - Waiting for payment
	StatusPending InvoiceStatus = "pending"
	// StatusPartial - Partial payment received
	StatusPartial InvoiceStatus = "partial"
	// StatusConfirming - Full payment received, awaiting blockchain confirmation
	StatusConfirming InvoiceStatus = "confirming"
	// StatusPaid - Payment fully confirmed
	StatusPaid InvoiceStatus = "paid"
	// StatusExpired - Invoice expired without payment
	StatusExpired InvoiceStatus = "expired"
	// StatusCancelled - Manually cancelled
	StatusCancelled InvoiceStatus = "cancelled"
	// StatusRefunded - Payment refunded after completion
	StatusRefunded InvoiceStatus = "refunded"
)

// String returns the string representation of the invoice status.
func (s InvoiceStatus) String() string {
	return string(s)
}

// IsValid returns true if the invoice status is valid.
func (s InvoiceStatus) IsValid() bool {
	switch s {
	case StatusCreated, StatusPending, StatusPartial, StatusConfirming, StatusPaid, StatusExpired, StatusCancelled, StatusRefunded:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the status is a terminal state.
func (s InvoiceStatus) IsTerminal() bool {
	switch s {
	case StatusPaid, StatusExpired, StatusCancelled, StatusRefunded:
		return true
	default:
		return false
	}
}

// IsActive returns true if the invoice is in an active (non-terminal) state.
func (s InvoiceStatus) IsActive() bool {
	return !s.IsTerminal()
}

// CanTransitionTo returns true if the status can transition to the target status.
func (s InvoiceStatus) CanTransitionTo(target InvoiceStatus) bool {
	if !target.IsValid() {
		return false
	}

	// Terminal states cannot transition to other states (except paid -> refunded)
	if s.IsTerminal() {
		return s == StatusPaid && target == StatusRefunded
	}

	// Define valid transitions based on the state machine
	validTransitions := map[InvoiceStatus][]InvoiceStatus{
		StatusCreated:    {StatusPending, StatusExpired, StatusCancelled},
		StatusPending:    {StatusPartial, StatusConfirming, StatusExpired, StatusCancelled},
		StatusPartial:    {StatusConfirming, StatusCancelled},
		StatusConfirming: {StatusPaid, StatusPending}, // pending for blockchain reorg
	}

	if transitions, exists := validTransitions[s]; exists {
		for _, validTarget := range transitions {
			if validTarget == target {
				return true
			}
		}
	}

	return false
}

// OverpaymentAction represents how to handle overpayments.
type OverpaymentAction string

const (
	// OverpaymentActionCredit - Credit the overpayment to merchant account
	OverpaymentActionCredit OverpaymentAction = "credit_account"
	// OverpaymentActionRefund - Refund the overpayment to customer
	OverpaymentActionRefund OverpaymentAction = "refund"
	// OverpaymentActionDonate - Donate the overpayment to charity
	OverpaymentActionDonate OverpaymentAction = "donate"
)

// String returns the string representation of the overpayment action.
func (a OverpaymentAction) String() string {
	return string(a)
}

// IsValid returns true if the overpayment action is valid.
func (a OverpaymentAction) IsValid() bool {
	switch a {
	case OverpaymentActionCredit, OverpaymentActionRefund, OverpaymentActionDonate:
		return true
	default:
		return false
	}
}

// AuditEvent represents the type of audit event.
type AuditEvent string

const (
	// AuditEventCreated - Invoice was created
	AuditEventCreated AuditEvent = "created"
	// AuditEventViewed - Invoice was viewed by customer
	AuditEventViewed AuditEvent = "viewed"
	// AuditEventPaymentDetected - Payment was detected
	AuditEventPaymentDetected AuditEvent = "payment_detected"
	// AuditEventPaymentConfirmed - Payment was confirmed
	AuditEventPaymentConfirmed AuditEvent = "payment_confirmed"
	// AuditEventPaid - Invoice was marked as paid
	AuditEventPaid AuditEvent = "paid"
	// AuditEventExpired - Invoice expired
	AuditEventExpired AuditEvent = "expired"
	// AuditEventCancelled - Invoice was cancelled
	AuditEventCancelled AuditEvent = "cancelled"
	// AuditEventRefunded - Invoice was refunded
	AuditEventRefunded AuditEvent = "refunded"
	// AuditEventStatusChanged - Invoice status changed
	AuditEventStatusChanged AuditEvent = "status_changed"
)

// String returns the string representation of the audit event.
func (e AuditEvent) String() string {
	return string(e)
}

// IsValid returns true if the audit event is valid.
func (e AuditEvent) IsValid() bool {
	switch e {
	case AuditEventCreated, AuditEventViewed, AuditEventPaymentDetected, AuditEventPaymentConfirmed, AuditEventPaid, AuditEventExpired, AuditEventCancelled, AuditEventRefunded, AuditEventStatusChanged:
		return true
	default:
		return false
	}
}

// Actor represents who triggered an audit event.
type Actor string

const (
	// ActorAPIKey - Action performed via API key
	ActorAPIKey Actor = "api_key"
	// ActorSystem - Action performed by system
	ActorSystem Actor = "system"
	// ActorCustomer - Action performed by customer
	ActorCustomer Actor = "customer"
	// ActorAdmin - Action performed by admin
	ActorAdmin Actor = "admin"
)

// String returns the string representation of the actor.
func (a Actor) String() string {
	return string(a)
}

// IsValid returns true if the actor is valid.
func (a Actor) IsValid() bool {
	switch a {
	case ActorAPIKey, ActorSystem, ActorCustomer, ActorAdmin:
		return true
	default:
		return false
	}
}
