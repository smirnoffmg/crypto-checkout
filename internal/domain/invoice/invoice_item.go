package invoice

import (
	"errors"

	"github.com/shopspring/decimal"
)

// InvoiceItem represents an item in an invoice.
//
//nolint:revive // InvoiceItem is a clear domain model name
type InvoiceItem struct {
	description string
	unitPrice   *USDTAmount
	quantity    decimal.Decimal
}

// NewInvoiceItem creates a new InvoiceItem.
func NewInvoiceItem(description string, unitPrice *USDTAmount, quantity decimal.Decimal) (*InvoiceItem, error) {
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}

	if quantity.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("quantity must be positive")
	}

	if unitPrice == nil {
		return nil, errors.New("unit price cannot be nil")
	}

	return &InvoiceItem{
		description: description,
		unitPrice:   unitPrice,
		quantity:    quantity,
	}, nil
}

// Description returns the item description.
func (i *InvoiceItem) Description() string {
	return i.description
}

// UnitPrice returns the unit price.
func (i *InvoiceItem) UnitPrice() *USDTAmount {
	return i.unitPrice
}

// Quantity returns the quantity.
func (i *InvoiceItem) Quantity() decimal.Decimal {
	return i.quantity
}

// CalculateTotal calculates the total price for this item (unit price * quantity).
func (i *InvoiceItem) CalculateTotal() *USDTAmount {
	total, _ := i.unitPrice.Multiply(i.quantity)
	return total
}
