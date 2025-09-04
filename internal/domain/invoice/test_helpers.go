package invoice

import "github.com/shopspring/decimal"

// MustNewUSDTAmount is a helper function for tests that panics on error.
func MustNewUSDTAmount(value string) *USDTAmount {
	amount, err := NewUSDTAmount(value)
	if err != nil {
		panic(err)
	}
	return amount
}

// MustNewInvoiceItem is a helper function for tests that panics on error.
func MustNewInvoiceItem(description string, unitPrice *USDTAmount, quantity decimal.Decimal) *InvoiceItem {
	item, err := NewInvoiceItem(description, unitPrice, quantity)
	if err != nil {
		panic(err)
	}
	return item
}

// MustNewPaymentAddress is a helper function for tests that panics on error.
func MustNewPaymentAddress(address string) *PaymentAddress {
	addr, err := NewPaymentAddress(address)
	if err != nil {
		panic(err)
	}
	return addr
}

// MustNewDecimal is a helper function for tests that panics on error.
func MustNewDecimal(value string) decimal.Decimal {
	d, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return d
}
