// Package payment provides domain models for payment management in the crypto-checkout system.
package payment

// PaymentStatus represents the status of a payment using FSM.
//
//nolint:revive // Domain-specific naming convention
type PaymentStatus string

// Payment status constants based on PAYMENT_STATUSES.md specification.
const (
	// StatusDetected represents a transaction found in mempool or block (0 confirmations).
	StatusDetected PaymentStatus = "detected"
	// StatusConfirming represents a transaction included in block (1-11 confirmations).
	StatusConfirming PaymentStatus = "confirming"
	// StatusConfirmed represents sufficient confirmations received (12+ confirmations).
	StatusConfirmed PaymentStatus = "confirmed"
	// StatusFailed represents a transaction failed or reverted.
	StatusFailed PaymentStatus = "failed"
	// StatusOrphaned represents a block containing tx was orphaned (temporary state).
	StatusOrphaned PaymentStatus = "orphaned"
)

// String returns the string representation of the status.
func (s PaymentStatus) String() string {
	return string(s)
}

// IsTerminal returns true if the status is a terminal state (no further transitions possible).
func (s PaymentStatus) IsTerminal() bool {
	return s == StatusConfirmed || s == StatusFailed
}

// IsActive returns true if the status allows further processing.
func (s PaymentStatus) IsActive() bool {
	return s == StatusDetected || s == StatusConfirming || s == StatusOrphaned
}

// IsSuccessful returns true if the status indicates successful payment completion.
func (s PaymentStatus) IsSuccessful() bool {
	return s == StatusConfirmed
}

// IsFailed returns true if the status indicates payment failure.
func (s PaymentStatus) IsFailed() bool {
	return s == StatusFailed
}

// IsTemporary returns true if the status is temporary and must transition.
func (s PaymentStatus) IsTemporary() bool {
	return s == StatusOrphaned
}
