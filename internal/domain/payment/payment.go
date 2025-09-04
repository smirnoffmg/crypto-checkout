package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const (
	// MinConfirmationsSmall is the minimum confirmations for small amounts (<$100).
	MinConfirmationsSmall = 1
	// MinConfirmationsStandard is the minimum confirmations for standard amounts ($100-$10k).
	MinConfirmationsStandard = 12
	// MinConfirmationsLarge is the minimum confirmations for large amounts (>$10k).
	MinConfirmationsLarge = 19

	// Amount thresholds for confirmation requirements.
	smallAmountThreshold    = 100
	standardAmountThreshold = 10000
)

// Payment represents a USDT payment transaction with FSM-based status management.
type Payment struct {
	id              string
	amount          *USDTAmount
	address         *Address
	transactionHash *TransactionHash
	confirmations   *ConfirmationCount
	status          *StatusFSM
	createdAt       time.Time
	updatedAt       time.Time
}

// NewPayment creates a new Payment with the given parameters.
func NewPayment(
	id string,
	amount *USDTAmount,
	address *Address,
	transactionHash *TransactionHash,
) (*Payment, error) {
	if id == "" {
		return nil, errors.New("payment ID cannot be empty")
	}
	if amount == nil {
		return nil, errors.New("amount cannot be nil")
	}
	if address == nil {
		return nil, errors.New("address cannot be nil")
	}
	if transactionHash == nil {
		return nil, errors.New("transaction hash cannot be nil")
	}

	// Initialize with detected status
	fsm, err := NewPaymentStatusFSM(StatusDetected)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment FSM: %w", err)
	}

	confirmations, err := NewConfirmationCount(0)
	if err != nil {
		return nil, fmt.Errorf("failed to create confirmation count: %w", err)
	}

	now := time.Now()
	return &Payment{
		id:              id,
		amount:          amount,
		address:         address,
		transactionHash: transactionHash,
		confirmations:   confirmations,
		status:          fsm,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// ID returns the payment ID.
func (p *Payment) ID() string {
	return p.id
}

// Amount returns the payment amount.
func (p *Payment) Amount() *USDTAmount {
	return p.amount
}

// Address returns the payment address.
func (p *Payment) Address() *Address {
	return p.address
}

// TransactionHash returns the transaction hash.
func (p *Payment) TransactionHash() *TransactionHash {
	return p.transactionHash
}

// Confirmations returns the current confirmation count.
func (p *Payment) Confirmations() *ConfirmationCount {
	return p.confirmations
}

// Status returns the current payment status.
func (p *Payment) Status() PaymentStatus {
	return p.status.CurrentStatus()
}

// CreatedAt returns the creation timestamp.
func (p *Payment) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns the last update timestamp.
func (p *Payment) UpdatedAt() time.Time {
	return p.updatedAt
}

// GetRequiredConfirmations returns the required confirmations based on the payment amount.
func (p *Payment) GetRequiredConfirmations() int {
	amount := p.amount.value

	// Small amounts (<$100)
	if amount.LessThan(decimal.NewFromInt(smallAmountThreshold)) {
		return MinConfirmationsSmall
	}

	// Standard amounts ($100-$10k)
	if amount.LessThan(decimal.NewFromInt(standardAmountThreshold)) {
		return MinConfirmationsStandard
	}

	// Large amounts (>$10k)
	return MinConfirmationsLarge
}

// IsConfirmed returns true if the payment has sufficient confirmations.
func (p *Payment) IsConfirmed() bool {
	required := p.GetRequiredConfirmations()
	return p.confirmations.count >= required
}

// CanTransitionTo checks if a transition to the given status is possible.
func (p *Payment) CanTransitionTo(status PaymentStatus) bool {
	return p.status.CanTransitionTo(status)
}

// TransitionToDetected transitions the payment to detected status.
func (p *Payment) TransitionToDetected(ctx context.Context) error {
	return p.status.Fire(ctx, TriggerDetected)
}

// TransitionToConfirming transitions the payment to confirming status.
func (p *Payment) TransitionToConfirming(ctx context.Context) error {
	return p.status.Fire(ctx, TriggerIncluded)
}

// TransitionToConfirmed transitions the payment to confirmed status.
func (p *Payment) TransitionToConfirmed(ctx context.Context) error {
	return p.status.Fire(ctx, TriggerConfirmed)
}

// TransitionToFailed transitions the payment to failed status.
func (p *Payment) TransitionToFailed(ctx context.Context) error {
	return p.status.Fire(ctx, TriggerFailed)
}

// TransitionToOrphaned transitions the payment to orphaned status.
func (p *Payment) TransitionToOrphaned(ctx context.Context) error {
	return p.status.Fire(ctx, TriggerOrphaned)
}

// TransitionBackToDetected transitions the payment back to detected status from orphaned.
func (p *Payment) TransitionBackToDetected(ctx context.Context) error {
	return p.status.Fire(ctx, TriggerBackToMempool)
}

// TransitionToDropped transitions the payment to failed status from orphaned.
func (p *Payment) TransitionToDropped(ctx context.Context) error {
	return p.status.Fire(ctx, TriggerDropped)
}

// UpdateConfirmations updates the confirmation count and transitions status if needed.
func (p *Payment) UpdateConfirmations(ctx context.Context, count int) error {
	confirmations, err := NewConfirmationCount(count)
	if err != nil {
		return fmt.Errorf("invalid confirmation count: %w", err)
	}

	p.confirmations = confirmations
	p.updatedAt = time.Now()

	// Auto-transition based on confirmations and current status
	currentStatus := p.Status()
	switch currentStatus {
	case StatusDetected:
		if count > 0 {
			return p.TransitionToConfirming(ctx)
		}
	case StatusConfirming:
		if p.IsConfirmed() {
			return p.TransitionToConfirmed(ctx)
		}
	case StatusOrphaned:
		// Orphaned payments need manual intervention
		return nil
	case StatusConfirmed, StatusFailed:
		// Terminal states - no auto-transitions
		return nil
	}

	return nil
}

// GetPermittedTriggers returns all triggers that are currently permitted.
func (p *Payment) GetPermittedTriggers() []Trigger {
	return p.status.GetPermittedTriggers()
}

// IsInStatus checks if the payment is in the given status.
func (p *Payment) IsInStatus(status PaymentStatus) bool {
	return p.status.IsInState(status)
}

// ToGraph returns a DOT graph representation of the payment's state machine.
func (p *Payment) ToGraph() string {
	return p.status.ToGraph()
}
