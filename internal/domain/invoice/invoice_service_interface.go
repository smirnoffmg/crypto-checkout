package invoice

import (
	"context"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/domain/shared"
	"time"
)

// InvoiceService defines the interface for invoice business operations.
type InvoiceService interface {
	// CreateInvoice creates a new invoice with the given parameters.
	CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*Invoice, error)

	// GetInvoice retrieves an invoice by ID.
	GetInvoice(ctx context.Context, id string) (*Invoice, error)

	// GetInvoiceByPaymentAddress retrieves an invoice by payment address.
	GetInvoiceByPaymentAddress(ctx context.Context, address *shared.PaymentAddress) (*Invoice, error)

	// ListInvoices retrieves invoices with the given filters.
	ListInvoices(ctx context.Context, req *ListInvoicesRequest) (*ListInvoicesResponse, error)

	// MarkInvoiceAsViewed marks an invoice as viewed by the customer.
	MarkInvoiceAsViewed(ctx context.Context, id string) error

	// CancelInvoice cancels an invoice.
	CancelInvoice(ctx context.Context, id string, reason string) error

	// ProcessPayment processes a payment for an invoice.
	ProcessPayment(ctx context.Context, invoiceID string, payment *payment.Payment) error

	// GetExpiredInvoices retrieves invoices that have expired.
	GetExpiredInvoices(ctx context.Context) ([]*Invoice, error)

	// ProcessExpiredInvoices processes expired invoices.
	ProcessExpiredInvoices(ctx context.Context) error

	// GetInvoiceStatus returns the current status of an invoice.
	GetInvoiceStatus(ctx context.Context, id string) (InvoiceStatus, error)

	// UpdateInvoiceStatus updates the status of an invoice.
	UpdateInvoiceStatus(ctx context.Context, id string, newStatus InvoiceStatus, reason string) error
}

// CreateInvoiceRequest represents the request to create a new invoice.
type CreateInvoiceRequest struct {
	MerchantID         string
	CustomerID         *string
	Title              string
	Description        string
	Items              []*CreateInvoiceItemRequest
	Tax                *shared.Money
	Currency           shared.Currency
	CryptoCurrency     shared.CryptoCurrency
	PaymentTolerance   *PaymentTolerance
	ExpirationDuration time.Duration
	Metadata           map[string]interface{}
	WebhookURL         *string
	ReturnURL          *string
	CancelURL          *string
}

// CreateInvoiceItemRequest represents a request to create an invoice item.
type CreateInvoiceItemRequest struct {
	Name        string
	Description string
	Quantity    string
	UnitPrice   *shared.Money
}

// ListInvoicesRequest represents the request to list invoices.
type ListInvoicesRequest struct {
	MerchantID    string
	Status        *InvoiceStatus
	CustomerID    *string
	Limit         int
	Offset        int
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Search        *string
}

// ListInvoicesResponse represents the response to list invoices.
type ListInvoicesResponse struct {
	Invoices []*Invoice
	Total    int
	Limit    int
	Offset   int
}
