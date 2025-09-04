package payment

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
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

// LessThan returns true if this amount is less than the other amount.
func (a *USDTAmount) LessThan(other *USDTAmount) bool {
	return a.value.LessThan(other.value)
}

// GreaterThanOrEqual returns true if this amount is greater than or equal to the other amount.
func (a *USDTAmount) GreaterThanOrEqual(other *USDTAmount) bool {
	return a.value.GreaterThanOrEqual(other.value)
}

// Address represents a Tron USDT payment address.
type Address struct {
	address string
}

// NewPaymentAddress creates a new PaymentAddress from a string.
func NewPaymentAddress(address string) (*Address, error) {
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

	return &Address{address: address}, nil
}

// String returns the string representation of the address.
func (a *Address) String() string {
	return a.address
}

// Equals checks if two PaymentAddresses are equal.
func (a *Address) Equals(other *Address) bool {
	return a.address == other.address
}

// TransactionHash represents a blockchain transaction hash.
type TransactionHash struct {
	hash string
}

// NewTransactionHash creates a new TransactionHash from a string.
func NewTransactionHash(hash string) (*TransactionHash, error) {
	const minHashLength = 32

	if hash == "" {
		return nil, errors.New("transaction hash cannot be empty")
	}

	if len(hash) < minHashLength {
		return nil, errors.New("transaction hash too short")
	}

	// Basic validation for hex format
	matched, err := regexp.MatchString(`^[a-fA-F0-9]+$`, hash)
	if err != nil || !matched {
		return nil, errors.New("invalid transaction hash format")
	}

	return &TransactionHash{hash: hash}, nil
}

// String returns the string representation of the transaction hash.
func (h *TransactionHash) String() string {
	return h.hash
}

// Equals checks if two TransactionHashes are equal.
func (h *TransactionHash) Equals(other *TransactionHash) bool {
	return h.hash == other.hash
}

// ConfirmationCount represents the number of blockchain confirmations.
type ConfirmationCount struct {
	count int
}

// NewConfirmationCount creates a new ConfirmationCount.
func NewConfirmationCount(count int) (*ConfirmationCount, error) {
	if count < 0 {
		return nil, errors.New("confirmation count cannot be negative")
	}

	return &ConfirmationCount{count: count}, nil
}

// Int returns the integer value of the confirmation count.
func (c *ConfirmationCount) Int() int {
	return c.count
}

// Add adds one to the confirmation count.
func (c *ConfirmationCount) Add() *ConfirmationCount {
	return &ConfirmationCount{count: c.count + 1}
}

// GreaterThanOrEqual returns true if this count is greater than or equal to the other count.
func (c *ConfirmationCount) GreaterThanOrEqual(other *ConfirmationCount) bool {
	return c.count >= other.count
}

// String returns the string representation of the confirmation count.
func (c *ConfirmationCount) String() string {
	return strconv.Itoa(c.count)
}
