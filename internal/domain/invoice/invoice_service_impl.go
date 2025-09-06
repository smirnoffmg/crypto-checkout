package invoice

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// invoiceServiceImpl implements the InvoiceService interface.
type invoiceServiceImpl struct {
	repo Repository
}

// NewInvoiceService creates a new InvoiceService instance.
func NewInvoiceService(repo Repository) InvoiceService {
	return &invoiceServiceImpl{
		repo: repo,
	}
}

// CreateInvoice creates a new invoice with the given items and tax rate.
func (s *invoiceServiceImpl) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: invoice must have at least one item", ErrInvalidRequest)
	}

	// Parse tax rate
	taxRate, err := decimal.NewFromString(req.TaxRate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid tax rate format: %w", ErrInvalidRequest, err)
	}

	// Validate tax rate range
	if taxRate.IsNegative() || taxRate.GreaterThan(decimal.NewFromInt(1)) {
		return nil, fmt.Errorf("%w: tax rate must be between 0 and 1", ErrInvalidRequest)
	}

	// Create invoice items
	items := make([]*InvoiceItem, 0, len(req.Items))
	for i, itemReq := range req.Items {
		item, itemErr := s.createInvoiceItem(itemReq)
		if itemErr != nil {
			return nil, fmt.Errorf("%w: invalid item at index %d: %w", ErrInvalidRequest, i, itemErr)
		}
		items = append(items, item)
	}

	// Create the invoice
	inv, err := NewInvoice(items, taxRate)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create invoice: %w", ErrInvoiceServiceError, err)
	}

	// Save to repository
	if saveErr := s.repo.Save(ctx, inv); saveErr != nil {
		return nil, fmt.Errorf("%w: failed to save invoice: %w", ErrInvoiceServiceError, saveErr)
	}

	return inv, nil
}

// GetInvoice retrieves an invoice by its ID.
func (s *invoiceServiceImpl) GetInvoice(ctx context.Context, id string) (*Invoice, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: invoice ID cannot be empty", ErrInvalidRequest)
	}

	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrInvoiceNotFound, err)
		}
		return nil, fmt.Errorf("%w: failed to retrieve invoice: %w", ErrInvoiceServiceError, err)
	}

	return inv, nil
}

// GetInvoiceByPaymentAddress retrieves an invoice by its payment address.
func (s *invoiceServiceImpl) GetInvoiceByPaymentAddress(ctx context.Context, address string) (*Invoice, error) {
	if address == "" {
		return nil, fmt.Errorf("%w: payment address cannot be empty", ErrInvalidRequest)
	}

	paymentAddr, err := NewPaymentAddress(address)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPaymentAddress, err)
	}

	inv, err := s.repo.FindByPaymentAddress(ctx, paymentAddr)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrInvoiceNotFound, err)
		}
		return nil, fmt.Errorf("%w: failed to retrieve invoice by payment address: %w", ErrInvoiceServiceError, err)
	}

	return inv, nil
}

// ListInvoicesByStatus retrieves all invoices with the given status.
func (s *invoiceServiceImpl) ListInvoicesByStatus(
	ctx context.Context,
	status InvoiceStatus,
) ([]*Invoice, error) {
	invoices, err := s.repo.FindByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve invoices by status: %w", ErrInvoiceServiceError, err)
	}

	return invoices, nil
}

// ListActiveInvoices retrieves all active (non-terminal) invoices.
func (s *invoiceServiceImpl) ListActiveInvoices(ctx context.Context) ([]*Invoice, error) {
	invoices, err := s.repo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve active invoices: %w", ErrInvoiceServiceError, err)
	}

	return invoices, nil
}

// AssignPaymentAddress assigns a payment address to an
func (s *invoiceServiceImpl) AssignPaymentAddress(ctx context.Context, id string, address string) error {
	if id == "" {
		return fmt.Errorf("%w: invoice ID cannot be empty", ErrInvalidRequest)
	}

	if address == "" {
		return fmt.Errorf("%w: payment address cannot be empty", ErrInvalidRequest)
	}

	// Get the invoice
	inv, err := s.GetInvoice(ctx, id)
	if err != nil {
		return err
	}

	// Create payment address
	paymentAddr, err := NewPaymentAddress(address)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPaymentAddress, err)
	}

	// Assign the address
	if assignErr := inv.AssignPaymentAddress(paymentAddr); assignErr != nil {
		return fmt.Errorf("%w: failed to assign payment address: %w", ErrInvoiceServiceError, assignErr)
	}

	// Save the updated invoice
	if updateErr := s.repo.Update(ctx, inv); updateErr != nil {
		return fmt.Errorf("%w: failed to update invoice: %w", ErrInvoiceServiceError, updateErr)
	}

	return nil
}

// MarkInvoiceAsViewed marks an invoice as viewed by the customer.
func (s *invoiceServiceImpl) MarkInvoiceAsViewed(ctx context.Context, id string) error {
	return s.transitionInvoice(ctx, id, func(inv *Invoice) error {
		return inv.MarkAsViewed()
	})
}

// MarkInvoiceAsPartial marks an invoice as having received partial payment.
func (s *invoiceServiceImpl) MarkInvoiceAsPartial(ctx context.Context, id string) error {
	return s.transitionInvoice(ctx, id, func(inv *Invoice) error {
		return inv.MarkAsPartial()
	})
}

// MarkInvoiceAsCompleted marks an invoice as having received full payment.
func (s *invoiceServiceImpl) MarkInvoiceAsCompleted(ctx context.Context, id string) error {
	return s.transitionInvoice(ctx, id, func(inv *Invoice) error {
		return inv.MarkAsCompleted()
	})
}

// MarkInvoiceAsConfirmed marks an invoice as confirmed (payment verified).
func (s *invoiceServiceImpl) MarkInvoiceAsConfirmed(ctx context.Context, id string) error {
	return s.transitionInvoice(ctx, id, func(inv *Invoice) error {
		return inv.MarkAsConfirmed()
	})
}

// ExpireInvoice marks an invoice as expired.
func (s *invoiceServiceImpl) ExpireInvoice(ctx context.Context, id string) error {
	return s.transitionInvoice(ctx, id, func(inv *Invoice) error {
		return inv.Expire()
	})
}

// CancelInvoice marks an invoice as cancelled.
func (s *invoiceServiceImpl) CancelInvoice(ctx context.Context, id string) error {
	return s.transitionInvoice(ctx, id, func(inv *Invoice) error {
		return inv.Cancel()
	})
}

// RefundInvoice marks an invoice as refunded.
func (s *invoiceServiceImpl) RefundInvoice(ctx context.Context, id string) error {
	return s.transitionInvoice(ctx, id, func(inv *Invoice) error {
		return inv.Refund()
	})
}

// HandleReorg handles blockchain reorganization for an
func (s *invoiceServiceImpl) HandleReorg(ctx context.Context, id string) error {
	return s.transitionInvoice(ctx, id, func(inv *Invoice) error {
		return inv.HandleReorg()
	})
}

// transitionInvoice is a helper method to perform status transitions on invoices.
func (s *invoiceServiceImpl) transitionInvoice(
	ctx context.Context,
	id string,
	transition func(*Invoice) error,
) error {
	if id == "" {
		return fmt.Errorf("%w: invoice ID cannot be empty", ErrInvalidRequest)
	}

	// Get the invoice
	inv, err := s.GetInvoice(ctx, id)
	if err != nil {
		return err
	}

	// Perform the transition
	if transitionErr := transition(inv); transitionErr != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTransition, transitionErr)
	}

	// Save the updated invoice
	if updateErr := s.repo.Update(ctx, inv); updateErr != nil {
		return fmt.Errorf("%w: failed to update invoice: %w", ErrInvoiceServiceError, updateErr)
	}

	return nil
}

// createInvoiceItem creates an InvoiceItem from a CreateInvoiceItemRequest.
func (s *invoiceServiceImpl) createInvoiceItem(req CreateInvoiceItemRequest) (*InvoiceItem, error) {
	if req.Description == "" {
		return nil, errors.New("description cannot be empty")
	}

	// Parse unit price
	unitPrice, err := NewUSDTAmount(req.UnitPrice)
	if err != nil {
		return nil, fmt.Errorf("invalid unit price: %w", err)
	}

	// Parse quantity
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	item, err := NewInvoiceItem(req.Description, unitPrice, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice item: %w", err)
	}

	return item, nil
}
