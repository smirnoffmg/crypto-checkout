package shared

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// ExchangeRate represents a currency exchange rate.
type ExchangeRate struct {
	rate         decimal.Decimal
	fromCurrency Currency
	toCurrency   CryptoCurrency
	source       string
	lockedAt     time.Time
	expiresAt    time.Time
}

// NewExchangeRate creates a new ExchangeRate.
func NewExchangeRate(rate string, fromCurrency Currency, toCurrency CryptoCurrency, source string, validDuration time.Duration) (*ExchangeRate, error) {
	if rate == "" {
		return nil, errors.New("exchange rate cannot be empty")
	}

	if !fromCurrency.IsValid() {
		return nil, errors.New("invalid from currency")
	}

	if !toCurrency.IsValid() {
		return nil, errors.New("invalid to currency")
	}

	if source == "" {
		return nil, errors.New("rate source cannot be empty")
	}

	// Parse the rate
	rateDecimal, err := decimal.NewFromString(rate)
	if err != nil {
		return nil, errors.New("invalid exchange rate format")
	}

	if rateDecimal.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("exchange rate must be positive")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(validDuration)

	return &ExchangeRate{
		rate:         rateDecimal,
		fromCurrency: fromCurrency,
		toCurrency:   toCurrency,
		source:       source,
		lockedAt:     now,
		expiresAt:    expiresAt,
	}, nil
}

// Rate returns the exchange rate.
func (er *ExchangeRate) Rate() decimal.Decimal {
	return er.rate
}

// FromCurrency returns the source currency.
func (er *ExchangeRate) FromCurrency() Currency {
	return er.fromCurrency
}

// ToCurrency returns the target currency.
func (er *ExchangeRate) ToCurrency() CryptoCurrency {
	return er.toCurrency
}

// Source returns the rate provider.
func (er *ExchangeRate) Source() string {
	return er.source
}

// LockedAt returns when the rate was locked.
func (er *ExchangeRate) LockedAt() time.Time {
	return er.lockedAt
}

// ExpiresAt returns when the rate expires.
func (er *ExchangeRate) ExpiresAt() time.Time {
	return er.expiresAt
}

// IsExpired returns true if the exchange rate has expired.
func (er *ExchangeRate) IsExpired() bool {
	return time.Now().UTC().After(er.expiresAt)
}

// Convert converts an amount using this exchange rate.
func (er *ExchangeRate) Convert(amount *Money) (*Money, error) {
	if amount == nil {
		return nil, errors.New("amount cannot be nil")
	}

	if amount.Currency() != string(er.fromCurrency) {
		return nil, errors.New("currency mismatch for conversion")
	}

	if er.IsExpired() {
		return nil, errors.New("exchange rate has expired")
	}

	convertedAmount := amount.Amount().Mul(er.rate)
	return NewMoneyWithCrypto(convertedAmount.String(), er.toCurrency)
}

// String returns the string representation of the exchange rate.
func (er *ExchangeRate) String() string {
	return er.rate.String()
}

// Equals returns true if this exchange rate equals the other.
func (er *ExchangeRate) Equals(other *ExchangeRate) bool {
	if other == nil {
		return false
	}
	return er.rate.Equal(other.rate) &&
		er.fromCurrency == other.fromCurrency &&
		er.toCurrency == other.toCurrency &&
		er.source == other.source
}
