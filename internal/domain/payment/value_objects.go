package payment

import (
	"crypto-checkout/internal/domain/shared"
	"errors"
	"fmt"
	"time"
)

// Use shared value objects
type (
	TransactionHash   = shared.TransactionHash
	PaymentAddress    = shared.PaymentAddress
	ConfirmationCount = shared.ConfirmationCount
)

// NewTransactionHash creates a new transaction hash using shared implementation.
func NewTransactionHash(hash string) (*TransactionHash, error) {
	return shared.NewTransactionHash(hash)
}

// NewPaymentAddress creates a new payment address using shared implementation.
func NewPaymentAddress(address string, network shared.BlockchainNetwork) (*PaymentAddress, error) {
	return shared.NewPaymentAddress(address, network)
}

// NewConfirmationCount creates a new confirmation count using shared implementation.
func NewConfirmationCount(count int) (*ConfirmationCount, error) {
	return shared.NewConfirmationCount(count)
}

// BlockInfo represents information about a blockchain block.
type BlockInfo struct {
	number int64
	hash   string
}

// NewBlockInfo creates a new block info.
func NewBlockInfo(number int64, hash string) (*BlockInfo, error) {
	if number < 0 {
		return nil, errors.New("block number cannot be negative")
	}

	if hash == "" {
		return nil, errors.New("block hash cannot be empty")
	}

	return &BlockInfo{
		number: number,
		hash:   hash,
	}, nil
}

// Number returns the block number.
func (bi *BlockInfo) Number() int64 {
	return bi.number
}

// Hash returns the block hash.
func (bi *BlockInfo) Hash() string {
	return bi.hash
}

// Equals returns true if this block info equals the other.
func (bi *BlockInfo) Equals(other *BlockInfo) bool {
	if other == nil {
		return false
	}
	return bi.number == other.number && bi.hash == other.hash
}

// PaymentAmount represents a payment amount with currency.
// This is a domain-specific wrapper around shared.Money for payment context.
type PaymentAmount struct {
	amount   *shared.Money
	currency shared.CryptoCurrency
}

// NewPaymentAmount creates a new payment amount.
func NewPaymentAmount(amount *shared.Money, currency shared.CryptoCurrency) (*PaymentAmount, error) {
	if amount == nil {
		return nil, errors.New("amount cannot be nil")
	}

	if !currency.IsValid() {
		return nil, errors.New("invalid cryptocurrency")
	}

	return &PaymentAmount{
		amount:   amount,
		currency: currency,
	}, nil
}

// Amount returns the money amount.
func (pa *PaymentAmount) Amount() *shared.Money {
	return pa.amount
}

// Currency returns the cryptocurrency.
func (pa *PaymentAmount) Currency() shared.CryptoCurrency {
	return pa.currency
}

// String returns the string representation of the payment amount.
func (pa *PaymentAmount) String() string {
	return fmt.Sprintf("%s %s", pa.amount.String(), pa.currency.String())
}

// Equals returns true if this payment amount equals the other.
func (pa *PaymentAmount) Equals(other *PaymentAmount) bool {
	if other == nil {
		return false
	}
	return pa.amount.Equals(other.amount) && pa.currency == other.currency
}

// NetworkFee represents a network fee for a transaction.
type NetworkFee struct {
	fee      *shared.Money
	currency shared.CryptoCurrency
}

// NewNetworkFee creates a new network fee.
func NewNetworkFee(fee *shared.Money, currency shared.CryptoCurrency) (*NetworkFee, error) {
	if fee == nil {
		return nil, errors.New("fee cannot be nil")
	}

	if !currency.IsValid() {
		return nil, errors.New("invalid cryptocurrency")
	}

	// Network fees should be positive
	if fee.Amount().Sign() <= 0 {
		return nil, errors.New("network fee must be positive")
	}

	return &NetworkFee{
		fee:      fee,
		currency: currency,
	}, nil
}

// Fee returns the fee amount.
func (nf *NetworkFee) Fee() *shared.Money {
	return nf.fee
}

// Currency returns the cryptocurrency.
func (nf *NetworkFee) Currency() shared.CryptoCurrency {
	return nf.currency
}

// String returns the string representation of the network fee.
func (nf *NetworkFee) String() string {
	return fmt.Sprintf("%s %s", nf.fee.String(), nf.currency.String())
}

// Equals returns true if this network fee equals the other.
func (nf *NetworkFee) Equals(other *NetworkFee) bool {
	if other == nil {
		return false
	}
	return nf.fee.Equals(other.fee) && nf.currency == other.currency
}

// PaymentTimestamps represents the various timestamps for a payment.
type PaymentTimestamps struct {
	detectedAt  time.Time
	confirmedAt *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

// NewPaymentTimestamps creates new payment timestamps.
func NewPaymentTimestamps(detectedAt time.Time) *PaymentTimestamps {
	now := time.Now().UTC()
	return &PaymentTimestamps{
		detectedAt: detectedAt,
		createdAt:  now,
		updatedAt:  now,
	}
}

// DetectedAt returns when the payment was first detected.
func (pt *PaymentTimestamps) DetectedAt() time.Time {
	return pt.detectedAt
}

// ConfirmedAt returns when the payment was confirmed (if confirmed).
func (pt *PaymentTimestamps) ConfirmedAt() *time.Time {
	return pt.confirmedAt
}

// CreatedAt returns when the payment record was created.
func (pt *PaymentTimestamps) CreatedAt() time.Time {
	return pt.createdAt
}

// UpdatedAt returns when the payment record was last updated.
func (pt *PaymentTimestamps) UpdatedAt() time.Time {
	return pt.updatedAt
}

// SetConfirmedAt sets the confirmation timestamp.
func (pt *PaymentTimestamps) SetConfirmedAt(confirmedAt time.Time) {
	pt.confirmedAt = &confirmedAt
	pt.updatedAt = time.Now().UTC()
}

// SetUpdatedAt updates the last updated timestamp.
func (pt *PaymentTimestamps) SetUpdatedAt(updatedAt time.Time) {
	pt.updatedAt = updatedAt
}

// Equals returns true if this payment timestamps equals the other.
func (pt *PaymentTimestamps) Equals(other *PaymentTimestamps) bool {
	if other == nil {
		return false
	}

	// Compare confirmedAt pointers
	var confirmedAtEqual bool
	switch {
	case pt.confirmedAt == nil && other.confirmedAt == nil:
		confirmedAtEqual = true
	case pt.confirmedAt != nil && other.confirmedAt != nil:
		confirmedAtEqual = pt.confirmedAt.Equal(*other.confirmedAt)
	default:
		confirmedAtEqual = false
	}

	return pt.detectedAt.Equal(other.detectedAt) &&
		confirmedAtEqual &&
		pt.createdAt.Equal(other.createdAt) &&
		pt.updatedAt.Equal(other.updatedAt)
}
