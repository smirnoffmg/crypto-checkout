package payment

// PaymentStatus represents the current status of a payment in the blockchain confirmation process.
type PaymentStatus string

const (
	// StatusDetected indicates the payment transaction has been detected in the mempool
	// but not yet included in a block.
	StatusDetected PaymentStatus = "detected"

	// StatusConfirming indicates the payment transaction has been included in a block
	// and is gaining confirmations.
	StatusConfirming PaymentStatus = "confirming"

	// StatusConfirmed indicates the payment has received sufficient confirmations
	// and is considered final.
	StatusConfirmed PaymentStatus = "confirmed"

	// StatusOrphaned indicates the payment transaction was included in a block
	// that was later orphaned by the blockchain.
	StatusOrphaned PaymentStatus = "orphaned"

	// StatusFailed indicates the payment transaction failed or was reverted.
	StatusFailed PaymentStatus = "failed"
)

// String returns the string representation of the payment status.
func (ps PaymentStatus) String() string {
	return string(ps)
}

// IsValid checks if the payment status is valid.
func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case StatusDetected, StatusConfirming, StatusConfirmed, StatusOrphaned, StatusFailed:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the payment status is a terminal state.
func (ps PaymentStatus) IsTerminal() bool {
	switch ps {
	case StatusConfirmed, StatusFailed:
		return true
	default:
		return false
	}
}

// IsActive returns true if the payment status is an active (non-terminal) state.
func (ps PaymentStatus) IsActive() bool {
	return !ps.IsTerminal()
}

// CanTransitionTo checks if the payment can transition to the target status.
func (ps PaymentStatus) CanTransitionTo(target PaymentStatus) bool {
	if !target.IsValid() {
		return false
	}

	// Define valid transitions based on the state machine
	validTransitions := map[PaymentStatus][]PaymentStatus{
		StatusDetected:   {StatusConfirming, StatusFailed},
		StatusConfirming: {StatusConfirmed, StatusOrphaned, StatusFailed},
		StatusOrphaned:   {StatusDetected, StatusFailed},
		// Terminal states cannot transition
		StatusConfirmed: {},
		StatusFailed:    {},
	}

	allowedTransitions, exists := validTransitions[ps]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == target {
			return true
		}
	}

	return false
}

// ParsePaymentStatus parses a string into a PaymentStatus.
func ParsePaymentStatus(s string) (PaymentStatus, error) {
	status := PaymentStatus(s)
	if !status.IsValid() {
		return "", NewInvalidPaymentStatusError(s)
	}
	return status, nil
}
