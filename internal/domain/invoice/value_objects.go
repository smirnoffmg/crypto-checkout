package invoice

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	// USDTDecimalPlaces is the number of decimal places for USDT amounts.
	USDTDecimalPlaces = 2
)

// USDTAmount represents a USDT amount with precision handling.
type USDTAmount struct {
	value decimal.Decimal
}

// NewUSDTAmount creates a new USDTAmount from a string value.
func NewUSDTAmount(value string) (*USDTAmount, error) {
	if value == "" {
		return nil, errors.New("amount cannot be empty")
	}

	// Parse the value using decimal library for precision
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	if parsed.IsNegative() {
		return nil, errors.New("amount cannot be negative")
	}

	return &USDTAmount{value: parsed}, nil
}

// String returns the string representation of the amount.
func (a *USDTAmount) String() string {
	return a.value.StringFixed(USDTDecimalPlaces)
}

// Add adds another USDTAmount to this one.
func (a *USDTAmount) Add(other *USDTAmount) (*USDTAmount, error) {
	result := a.value.Add(other.value).Round(USDTDecimalPlaces)
	return &USDTAmount{value: result}, nil
}

// Multiply multiplies this amount by a decimal multiplier.
func (a *USDTAmount) Multiply(multiplier decimal.Decimal) (*USDTAmount, error) {
	result := a.value.Mul(multiplier).Round(USDTDecimalPlaces)
	return &USDTAmount{value: result}, nil
}

// PaymentAddress represents a Tron USDT payment address.
type PaymentAddress struct {
	address string
}

// NewPaymentAddress creates a new PaymentAddress from a string.
func NewPaymentAddress(address string) (*PaymentAddress, error) {
	const minAddressLength = 10

	if address == "" {
		return nil, errors.New("address cannot be empty")
	}

	if len(address) < minAddressLength {
		return nil, errors.New("address too short")
	}

	// Basic validation for Tron address format
	// Tron addresses start with 'T'
	if !strings.HasPrefix(address, "T") {
		return nil, errors.New("invalid Tron address format")
	}

	// Check for valid characters (alphanumeric)
	matched, err := regexp.MatchString(`^T[A-Za-z0-9]+$`, address)
	if err != nil || !matched {
		return nil, errors.New("invalid address format")
	}

	return &PaymentAddress{address: address}, nil
}

// String returns the string representation of the address.
func (a *PaymentAddress) String() string {
	return a.address
}

// Equals checks if two PaymentAddresses are equal.
func (a *PaymentAddress) Equals(other *PaymentAddress) bool {
	return a.address == other.address
}
