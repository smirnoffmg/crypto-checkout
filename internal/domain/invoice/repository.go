package invoice

import (
	"context"
)

// Repository defines the interface for invoice data persistence.
type Repository interface {
	// Save persists an invoice to the data store.
	Save(ctx context.Context, invoice *Invoice) error

	// FindByID retrieves an invoice by its ID.
	FindByID(ctx context.Context, id string) (*Invoice, error)

	// FindByPaymentAddress retrieves an invoice by its payment address.
	FindByPaymentAddress(ctx context.Context, address *PaymentAddress) (*Invoice, error)

	// FindByStatus retrieves all invoices with the given status.
	FindByStatus(ctx context.Context, status InvoiceStatus) ([]*Invoice, error)

	// FindActive retrieves all active (non-terminal) invoices.
	FindActive(ctx context.Context) ([]*Invoice, error)

	// FindExpired retrieves all expired invoices.
	FindExpired(ctx context.Context) ([]*Invoice, error)

	// Update updates an existing invoice in the data store.
	Update(ctx context.Context, invoice *Invoice) error

	// Delete removes an invoice from the data store.
	Delete(ctx context.Context, id string) error

	// Exists checks if an invoice with the given ID exists.
	Exists(ctx context.Context, id string) (bool, error)
}
