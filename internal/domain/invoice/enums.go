// Package invoice provides domain models for invoice management in the crypto-checkout system.
package invoice

// InvoiceStatus represents the status of an invoice using FSM.
//
//nolint:revive // Domain-specific naming convention
type InvoiceStatus string

// Invoice status constants.
const (
	// StatusCreated represents a newly created invoice that hasn't been viewed yet.
	StatusCreated InvoiceStatus = "created"
	// StatusPending represents an invoice waiting for payment.
	StatusPending InvoiceStatus = "pending"
	// StatusPartial represents an invoice with partial payment received.
	StatusPartial InvoiceStatus = "partial"
	// StatusConfirming represents an invoice with payment received and being confirmed.
	StatusConfirming InvoiceStatus = "confirming"
	// StatusPaid represents a fully paid invoice.
	StatusPaid InvoiceStatus = "paid"
	// StatusExpired represents an invoice that expired without payment.
	StatusExpired InvoiceStatus = "expired"
	// StatusCancelled represents a manually cancelled invoice.
	StatusCancelled InvoiceStatus = "cancelled"
	// StatusRefunded represents an invoice that was refunded after payment.
	StatusRefunded InvoiceStatus = "refunded"
)

// String returns the string representation of the status.
func (s InvoiceStatus) String() string {
	return string(s)
}

// IsTerminal returns true if the status is a terminal state (no further transitions possible).
func (s InvoiceStatus) IsTerminal() bool {
	return s == StatusExpired || s == StatusCancelled || s == StatusRefunded
}

// IsActive returns true if the status allows further payment processing.
func (s InvoiceStatus) IsActive() bool {
	return s == StatusCreated || s == StatusPending || s == StatusPartial || s == StatusConfirming
}

// IsPaid returns true if the status indicates payment completion.
func (s InvoiceStatus) IsPaid() bool {
	return s == StatusPaid || s == StatusRefunded
}
