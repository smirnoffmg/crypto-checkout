package database

import (
	"errors"
	"fmt"

	"crypto-checkout/internal/domain/invoice"
)

// InvoiceMapper handles conversion between domain entities and database models.
type InvoiceMapper struct{}

// NewInvoiceMapper creates a new invoice mapper.
func NewInvoiceMapper() *InvoiceMapper {
	return &InvoiceMapper{}
}

// ToDomain converts a database model to a domain entity.
func (m *InvoiceMapper) ToDomain(model *InvoiceModel) (*invoice.Invoice, error) {
	if model == nil {
		return nil, errors.New("invoice model cannot be nil")
	}

	// Parse tax rate
	taxRate := invoice.MustNewDecimal(model.TaxRate)

	// Convert items
	items := make([]*invoice.InvoiceItem, len(model.Items))
	for i, item := range model.Items {
		unitPrice := invoice.MustNewUSDTAmount(item.UnitPrice)
		quantity := invoice.MustNewDecimal(item.Quantity)

		invoiceItem, err := invoice.NewInvoiceItem(item.Description, unitPrice, quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}

		items[i] = invoiceItem
	}

	// Create invoice
	inv, err := invoice.NewInvoice(items, taxRate)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Set the ID from the database model
	if setErr := inv.SetID(model.ID); setErr != nil {
		return nil, fmt.Errorf("failed to set invoice ID: %w", setErr)
	}

	// Set payment address if present
	if model.PaymentAddress != nil {
		address, addrErr := invoice.NewPaymentAddress(*model.PaymentAddress)
		if addrErr != nil {
			return nil, fmt.Errorf("failed to create payment address: %w", addrErr)
		}
		if assignErr := inv.AssignPaymentAddress(address); assignErr != nil {
			return nil, fmt.Errorf("failed to assign payment address: %w", assignErr)
		}
	}

	// Set status from database - create FSM with the correct status
	status := invoice.InvoiceStatus(model.Status)
	inv.SetStatusFSM(invoice.NewInvoiceStatusFSM(status))

	return inv, nil
}

// ToModel converts a domain entity to a database model.
func (m *InvoiceMapper) ToModel(inv *invoice.Invoice) *InvoiceModel {
	if inv == nil {
		return nil
	}

	model := &InvoiceModel{
		ID:        inv.ID(),
		Status:    inv.Status().String(),
		TaxRate:   inv.TaxRate().String(),
		CreatedAt: inv.CreatedAt(),
		PaidAt:    inv.PaidAt(),
	}

	// Set payment address if present
	if inv.PaymentAddress() != nil {
		address := inv.PaymentAddress().String()
		model.PaymentAddress = &address
	}

	// Convert items
	items := make([]InvoiceItemModel, len(inv.Items()))
	for i, item := range inv.Items() {
		items[i] = InvoiceItemModel{
			Description: item.Description(),
			UnitPrice:   item.UnitPrice().String(),
			Quantity:    item.Quantity().String(),
		}
	}
	model.Items = items

	return model
}

// ToDomainSlice converts multiple database models to domain entities.
func (m *InvoiceMapper) ToDomainSlice(models []InvoiceModel) ([]*invoice.Invoice, error) {
	invoices := make([]*invoice.Invoice, len(models))
	for i, model := range models {
		inv, err := m.ToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model %d: %w", i, err)
		}
		invoices[i] = inv
	}
	return invoices, nil
}
