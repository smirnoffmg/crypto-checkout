// Package service provides application services for business logic orchestration.
package service

import (
	"context"
	"errors"

	"crypto-checkout/internal/domain/invoice"
)

// InvoiceService defines the interface for invoice business operations.
type InvoiceService interface {
	// CreateInvoice creates a new invoice with the given items and tax rate.
	CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*invoice.Invoice, error)

	// GetInvoice retrieves an invoice by its ID.
	GetInvoice(ctx context.Context, id string) (*invoice.Invoice, error)

	// GetInvoiceByPaymentAddress retrieves an invoice by its payment address.
	GetInvoiceByPaymentAddress(ctx context.Context, address string) (*invoice.Invoice, error)

	// ListInvoicesByStatus retrieves all invoices with the given status.
	ListInvoicesByStatus(ctx context.Context, status invoice.InvoiceStatus) ([]*invoice.Invoice, error)

	// ListActiveInvoices retrieves all active (non-terminal) invoices.
	ListActiveInvoices(ctx context.Context) ([]*invoice.Invoice, error)

	// AssignPaymentAddress assigns a payment address to an invoice.
	AssignPaymentAddress(ctx context.Context, id string, address string) error

	// MarkInvoiceAsViewed marks an invoice as viewed by the customer.
	MarkInvoiceAsViewed(ctx context.Context, id string) error

	// MarkInvoiceAsPartial marks an invoice as having received partial payment.
	MarkInvoiceAsPartial(ctx context.Context, id string) error

	// MarkInvoiceAsCompleted marks an invoice as having received full payment.
	MarkInvoiceAsCompleted(ctx context.Context, id string) error

	// MarkInvoiceAsConfirmed marks an invoice as confirmed (payment verified).
	MarkInvoiceAsConfirmed(ctx context.Context, id string) error

	// ExpireInvoice marks an invoice as expired.
	ExpireInvoice(ctx context.Context, id string) error

	// CancelInvoice marks an invoice as cancelled.
	CancelInvoice(ctx context.Context, id string) error

	// RefundInvoice marks an invoice as refunded.
	RefundInvoice(ctx context.Context, id string) error

	// HandleReorg handles blockchain reorganization for an invoice.
	HandleReorg(ctx context.Context, id string) error
}

// CreateInvoiceRequest represents the request to create a new invoice.
type CreateInvoiceRequest struct {
	Items   []CreateInvoiceItemRequest `json:"items"`
	TaxRate string                     `json:"tax_rate"` // Decimal string like "0.10" for 10%
}

// CreateInvoiceItemRequest represents an item in the create invoice request.
type CreateInvoiceItemRequest struct {
	Description string `json:"description"`
	UnitPrice   string `json:"unit_price"` // USDT amount as decimal string
	Quantity    string `json:"quantity"`   // Decimal string like "2.5"
}

// Common service errors.
var (
	ErrInvoiceServiceError   = errors.New("invoice service error")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrInvoiceNotFound       = errors.New("invoice not found")
	ErrInvalidPaymentAddress = errors.New("invalid payment address")
	ErrInvalidTransition     = errors.New("invalid status transition")
)
