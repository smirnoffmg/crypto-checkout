package shared

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// Currency represents supported fiat currencies.
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
)

// String returns the string representation of the currency.
func (c Currency) String() string {
	return string(c)
}

// IsValid returns true if the currency is valid.
func (c Currency) IsValid() bool {
	switch c {
	case CurrencyUSD, CurrencyEUR, CurrencyGBP:
		return true
	default:
		return false
	}
}

// CryptoCurrency represents supported cryptocurrencies.
type CryptoCurrency string

const (
	CryptoCurrencyUSDT CryptoCurrency = "USDT"
	CryptoCurrencyBTC  CryptoCurrency = "BTC"
	CryptoCurrencyETH  CryptoCurrency = "ETH"
)

// String returns the string representation of the cryptocurrency.
func (c CryptoCurrency) String() string {
	return string(c)
}

// IsValid returns true if the cryptocurrency is valid.
func (c CryptoCurrency) IsValid() bool {
	switch c {
	case CryptoCurrencyUSDT, CryptoCurrencyBTC, CryptoCurrencyETH:
		return true
	default:
		return false
	}
}

// Money represents a monetary amount with currency.
type Money struct {
	amount   decimal.Decimal
	currency string // Can be Currency or CryptoCurrency
}

// NewMoney creates a new Money from a string value and currency.
func NewMoney(amount string, currency Currency) (*Money, error) {
	if amount == "" {
		return nil, errors.New("amount cannot be empty")
	}

	if !currency.IsValid() {
		return nil, errors.New("invalid currency")
	}

	// Parse the value using decimal library for precision
	parsed, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	if parsed.IsNegative() {
		return nil, errors.New("amount cannot be negative")
	}

	return &Money{amount: parsed, currency: string(currency)}, nil
}

// NewMoneyWithCrypto creates a new Money from a string value and cryptocurrency.
func NewMoneyWithCrypto(amount string, currency CryptoCurrency) (*Money, error) {
	if amount == "" {
		return nil, errors.New("amount cannot be empty")
	}

	if !currency.IsValid() {
		return nil, errors.New("invalid cryptocurrency")
	}

	// Parse the value using decimal library for precision
	parsed, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	if parsed.IsNegative() {
		return nil, errors.New("amount cannot be negative")
	}

	return &Money{amount: parsed, currency: string(currency)}, nil
}

// Amount returns the decimal amount.
func (m *Money) Amount() decimal.Decimal {
	return m.amount
}

// Currency returns the currency as a string.
func (m *Money) Currency() string {
	return m.currency
}

// String returns the string representation of the amount.
func (m *Money) String() string {
	return m.amount.StringFixed(2)
}

// Add adds another Money to this one.
func (m *Money) Add(other *Money) (*Money, error) {
	if m.currency != other.currency {
		return nil, errors.New("currency mismatch")
	}
	result := m.amount.Add(other.amount).Round(2)
	return &Money{amount: result, currency: m.currency}, nil
}

// Multiply multiplies this amount by a decimal multiplier.
func (m *Money) Multiply(multiplier decimal.Decimal) (*Money, error) {
	result := m.amount.Mul(multiplier).Round(2)
	return &Money{amount: result, currency: m.currency}, nil
}

// LessThan returns true if this amount is less than the other amount.
func (m *Money) LessThan(other *Money) bool {
	if m.currency != other.currency {
		return false // Cannot compare different currencies
	}
	return m.amount.LessThan(other.amount)
}

// GreaterThanOrEqual returns true if this amount is greater than or equal to the other amount.
func (m *Money) GreaterThanOrEqual(other *Money) bool {
	if m.currency != other.currency {
		return false // Cannot compare different currencies
	}
	return m.amount.GreaterThanOrEqual(other.amount)
}

// Equals returns true if this amount equals the other amount.
func (m *Money) Equals(other *Money) bool {
	if m.currency != other.currency {
		return false
	}
	return m.amount.Equal(other.amount)
}
