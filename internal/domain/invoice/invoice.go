package invoice

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

const (
	// DefaultInvoiceExpirationDuration is the default time after which an invoice expires.
	DefaultInvoiceExpirationDuration = 30 * time.Minute
)

type Invoice struct {
	id             string
	items          []*InvoiceItem
	taxRate        decimal.Decimal
	statusFSM      *InvoiceStatusFSM
	paymentAddress *PaymentAddress
	createdAt      time.Time
	paidAt         *time.Time
}

// NewInvoice creates a new Invoice.
func NewInvoice(items []*InvoiceItem, taxRate decimal.Decimal) (*Invoice, error) {
	if len(items) == 0 {
		return nil, errors.New("invoice must have at least one item")
	}

	if taxRate.IsNegative() || taxRate.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errors.New("tax rate must be between 0 and 1")
	}

	return &Invoice{
		id:             generateID(),
		items:          items,
		taxRate:        taxRate,
		statusFSM:      NewInvoiceStatusFSM(StatusCreated),
		paymentAddress: nil,
		createdAt:      time.Now(),
		paidAt:         nil,
	}, nil
}

// ID returns the invoice ID.
func (i *Invoice) ID() string {
	return i.id
}

// SetID sets the invoice ID (used for reconstruction from database).
func (i *Invoice) SetID(id string) error {
	if id == "" {
		return errors.New("invoice ID cannot be empty")
	}
	i.id = id
	return nil
}

// SetStatusFSM sets the status FSM (used for reconstruction from database).
func (i *Invoice) SetStatusFSM(fsm *InvoiceStatusFSM) {
	i.statusFSM = fsm
}

// Items returns the invoice items.
func (i *Invoice) Items() []*InvoiceItem {
	return i.items
}

// TaxRate returns the tax rate.
func (i *Invoice) TaxRate() decimal.Decimal {
	return i.taxRate
}

// Status returns the invoice status.
func (i *Invoice) Status() InvoiceStatus {
	return i.statusFSM.CurrentStatus()
}

// PaymentAddress returns the payment address.
func (i *Invoice) PaymentAddress() *PaymentAddress {
	return i.paymentAddress
}

// CreatedAt returns the creation timestamp.
func (i *Invoice) CreatedAt() time.Time {
	return i.createdAt
}

// ExpiresAt returns the expiration timestamp (30 minutes after creation by default).
func (i *Invoice) ExpiresAt() time.Time {
	return i.createdAt.Add(DefaultInvoiceExpirationDuration)
}

// PaidAt returns the payment timestamp.
func (i *Invoice) PaidAt() *time.Time {
	return i.paidAt
}

// CalculateSubtotal calculates the subtotal of all items.
func (i *Invoice) CalculateSubtotal() *USDTAmount {
	subtotal, _ := NewUSDTAmount("0.00")

	for _, item := range i.items {
		itemTotal := item.CalculateTotal()
		subtotal, _ = subtotal.Add(itemTotal)
	}

	return subtotal
}

// CalculateTax calculates the tax amount based on the subtotal and tax rate.
func (i *Invoice) CalculateTax() *USDTAmount {
	subtotal := i.CalculateSubtotal()
	tax, _ := subtotal.Multiply(i.taxRate)
	return tax
}

// CalculateTotal calculates the total amount including tax.
func (i *Invoice) CalculateTotal() *USDTAmount {
	subtotal := i.CalculateSubtotal()
	tax := i.CalculateTax()
	total, _ := subtotal.Add(tax)
	return total
}

// AssignPaymentAddress assigns a payment address to the invoice.
func (i *Invoice) AssignPaymentAddress(address *PaymentAddress) error {
	if address == nil {
		return errors.New("payment address cannot be nil")
	}

	i.paymentAddress = address
	return nil
}

// MarkAsViewed marks the invoice as viewed by the customer.
func (i *Invoice) MarkAsViewed() error {
	return i.statusFSM.Transition(context.Background(), TriggerViewed)
}

// MarkAsPartial marks the invoice as having received partial payment.
func (i *Invoice) MarkAsPartial() error {
	return i.statusFSM.Transition(context.Background(), TriggerPartial)
}

// MarkAsCompleted marks the invoice as having received full payment.
func (i *Invoice) MarkAsCompleted() error {
	return i.statusFSM.Transition(context.Background(), TriggerCompleted)
}

// MarkAsConfirmed marks the invoice as confirmed (payment verified).
func (i *Invoice) MarkAsConfirmed() error {
	err := i.statusFSM.Transition(context.Background(), TriggerConfirmed)
	if err == nil {
		now := time.Now()
		i.paidAt = &now
	}
	return err
}

// MarkAsPaid marks the invoice as paid (legacy method for backward compatibility).
func (i *Invoice) MarkAsPaid() error {
	// First complete the payment, then confirm it
	if err := i.MarkAsCompleted(); err != nil {
		return err
	}
	return i.MarkAsConfirmed()
}

// Expire marks the invoice as expired.
func (i *Invoice) Expire() error {
	return i.statusFSM.Transition(context.Background(), TriggerExpired)
}

// Cancel marks the invoice as cancelled.
func (i *Invoice) Cancel() error {
	return i.statusFSM.Transition(context.Background(), TriggerCancelled)
}

// Refund marks the invoice as refunded.
func (i *Invoice) Refund() error {
	return i.statusFSM.Transition(context.Background(), TriggerRefunded)
}

// HandleReorg handles blockchain reorganization by moving back to pending.
func (i *Invoice) HandleReorg() error {
	return i.statusFSM.Transition(context.Background(), TriggerReorg)
}

// IsPaid returns true if the invoice is paid.
func (i *Invoice) IsPaid() bool {
	return i.statusFSM.IsPaid()
}

// IsPending returns true if the invoice is pending.
func (i *Invoice) IsPending() bool {
	return i.statusFSM.IsInState(StatusPending)
}

// IsActive returns true if the invoice is in an active state.
func (i *Invoice) IsActive() bool {
	return i.statusFSM.IsActive()
}

// IsTerminal returns true if the invoice is in a terminal state.
func (i *Invoice) IsTerminal() bool {
	return i.statusFSM.IsTerminal()
}

// CanTransition checks if a transition is possible from the current state.
func (i *Invoice) CanTransition(trigger InvoiceTrigger) bool {
	return i.statusFSM.CanTransition(trigger)
}

// GetPermittedTriggers returns all triggers that are valid from the current state.
func (i *Invoice) GetPermittedTriggers() []InvoiceTrigger {
	return i.statusFSM.GetPermittedTriggers()
}

// generateID generates a unique ID for the invoice.
func generateID() string {
	const idLength = 16
	bytes := make([]byte, idLength)
	_, _ = rand.Read(bytes) // Ignore error as it's very unlikely to fail
	return hex.EncodeToString(bytes)
}
