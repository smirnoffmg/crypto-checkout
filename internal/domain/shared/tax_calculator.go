package shared

import (
	"errors"

	"github.com/shopspring/decimal"
)

// TaxCalculator provides tax calculation functionality.
type TaxCalculator struct{}

// NewTaxCalculator creates a new TaxCalculator.
func NewTaxCalculator() *TaxCalculator {
	return &TaxCalculator{}
}

// CalculateTaxFromRate calculates tax amount from a tax rate and subtotal.
func (tc *TaxCalculator) CalculateTaxFromRate(taxRate string, subtotal *Money) (*Money, error) {
	if taxRate == "" {
		// Return zero tax if no rate provided
		return NewMoney("0.00", Currency(subtotal.Currency()))
	}

	// Parse tax rate
	rate, err := decimal.NewFromString(taxRate)
	if err != nil {
		return nil, errors.New("invalid tax rate format")
	}

	// Calculate tax amount
	taxAmount := subtotal.Amount().Mul(rate)

	// Create tax money with same currency as subtotal
	return NewMoney(taxAmount.StringFixed(2), Currency(subtotal.Currency()))
}

// CalculateSubtotal calculates the subtotal from a list of items.
func (tc *TaxCalculator) CalculateSubtotal(items []InvoiceItem) (*Money, error) {
	if len(items) == 0 {
		return nil, errors.New("items list cannot be empty")
	}

	subtotal := decimal.Zero
	currency := ""

	for _, item := range items {
		// Get currency from first item
		if currency == "" {
			currency = item.UnitPrice().Currency()
		}

		// Validate currency consistency
		if item.UnitPrice().Currency() != currency {
			return nil, errors.New("all items must have the same currency")
		}

		// Add item total to subtotal
		subtotal = subtotal.Add(item.TotalPrice().Amount())
	}

	return NewMoney(subtotal.StringFixed(2), Currency(currency))
}

// InvoiceItem represents an item for tax calculation.
// This is a minimal interface to avoid circular dependencies.
type InvoiceItem interface {
	UnitPrice() *Money
	TotalPrice() *Money
}
