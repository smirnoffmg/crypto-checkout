package shared_test

import (
	"crypto-checkout/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/require"
)

// mockInvoiceItem implements the InvoiceItem interface for testing
type mockInvoiceItem struct {
	unitPrice  *shared.Money
	totalPrice *shared.Money
}

func (m *mockInvoiceItem) UnitPrice() *shared.Money {
	return m.unitPrice
}

func (m *mockInvoiceItem) TotalPrice() *shared.Money {
	return m.totalPrice
}

func TestTaxCalculator(t *testing.T) {
	t.Run("CalculateTaxFromRate - valid rate", func(t *testing.T) {
		calculator := shared.NewTaxCalculator()
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)

		tax, err := calculator.CalculateTaxFromRate("0.10", subtotal)
		require.NoError(t, err)
		require.Equal(t, "10.00", tax.String())
		require.Equal(t, string(shared.CurrencyUSD), tax.Currency())
	})

	t.Run("CalculateTaxFromRate - empty rate", func(t *testing.T) {
		calculator := shared.NewTaxCalculator()
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)

		tax, err := calculator.CalculateTaxFromRate("", subtotal)
		require.NoError(t, err)
		require.Equal(t, "0.00", tax.String())
		require.Equal(t, string(shared.CurrencyUSD), tax.Currency())
	})

	t.Run("CalculateTaxFromRate - invalid rate format", func(t *testing.T) {
		calculator := shared.NewTaxCalculator()
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)

		_, err := calculator.CalculateTaxFromRate("invalid", subtotal)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid tax rate format")
	})

	t.Run("CalculateTaxFromRate - zero rate", func(t *testing.T) {
		calculator := shared.NewTaxCalculator()
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)

		tax, err := calculator.CalculateTaxFromRate("0", subtotal)
		require.NoError(t, err)
		require.Equal(t, "0.00", tax.String())
	})

	t.Run("CalculateTaxFromRate - decimal rate", func(t *testing.T) {
		calculator := shared.NewTaxCalculator()
		subtotal, _ := shared.NewMoney("119.98", shared.CurrencyUSD)

		tax, err := calculator.CalculateTaxFromRate("0.10", subtotal)
		require.NoError(t, err)
		require.Equal(t, "12.00", tax.String())
	})

	t.Run("CalculateSubtotal - valid items", func(t *testing.T) {
		calculator := shared.NewTaxCalculator()

		item1 := &mockInvoiceItem{
			unitPrice:  &shared.Money{},
			totalPrice: &shared.Money{},
		}
		// Set up mock items with proper Money objects
		unitPrice1, _ := shared.NewMoney("50.00", shared.CurrencyUSD)
		totalPrice1, _ := shared.NewMoney("50.00", shared.CurrencyUSD)
		item1.unitPrice = unitPrice1
		item1.totalPrice = totalPrice1

		item2 := &mockInvoiceItem{
			unitPrice:  &shared.Money{},
			totalPrice: &shared.Money{},
		}
		unitPrice2, _ := shared.NewMoney("69.98", shared.CurrencyUSD)
		totalPrice2, _ := shared.NewMoney("69.98", shared.CurrencyUSD)
		item2.unitPrice = unitPrice2
		item2.totalPrice = totalPrice2

		items := []shared.InvoiceItem{item1, item2}

		subtotal, err := calculator.CalculateSubtotal(items)
		require.NoError(t, err)
		require.Equal(t, "119.98", subtotal.String())
		require.Equal(t, string(shared.CurrencyUSD), subtotal.Currency())
	})

	t.Run("CalculateSubtotal - empty items", func(t *testing.T) {
		calculator := shared.NewTaxCalculator()
		items := []shared.InvoiceItem{}

		_, err := calculator.CalculateSubtotal(items)
		require.Error(t, err)
		require.Contains(t, err.Error(), "items list cannot be empty")
	})

	t.Run("CalculateSubtotal - currency mismatch", func(t *testing.T) {
		calculator := shared.NewTaxCalculator()

		item1 := &mockInvoiceItem{
			unitPrice:  &shared.Money{},
			totalPrice: &shared.Money{},
		}
		unitPrice1, _ := shared.NewMoney("50.00", shared.CurrencyUSD)
		totalPrice1, _ := shared.NewMoney("50.00", shared.CurrencyUSD)
		item1.unitPrice = unitPrice1
		item1.totalPrice = totalPrice1

		item2 := &mockInvoiceItem{
			unitPrice:  &shared.Money{},
			totalPrice: &shared.Money{},
		}
		unitPrice2, _ := shared.NewMoney("69.98", shared.CurrencyEUR)
		totalPrice2, _ := shared.NewMoney("69.98", shared.CurrencyEUR)
		item2.unitPrice = unitPrice2
		item2.totalPrice = totalPrice2

		items := []shared.InvoiceItem{item1, item2}

		_, err := calculator.CalculateSubtotal(items)
		require.Error(t, err)
		require.Contains(t, err.Error(), "all items must have the same currency")
	})
}
