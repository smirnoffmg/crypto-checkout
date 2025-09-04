package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"crypto-checkout/internal/domain/payment"
)

// paymentServiceImpl implements the PaymentService interface.
type paymentServiceImpl struct {
	repo payment.Repository
}

// NewPaymentService creates a new PaymentService instance.
func NewPaymentService(repo payment.Repository) PaymentService {
	return &paymentServiceImpl{
		repo: repo,
	}
}

// CreatePayment creates a new payment.
func (s *paymentServiceImpl) CreatePayment(
	ctx context.Context,
	req CreatePaymentRequest,
) (*payment.Payment, error) {
	if req.Amount == "" {
		return nil, fmt.Errorf("%w: amount cannot be empty", ErrInvalidRequest)
	}
	if req.Address == "" {
		return nil, fmt.Errorf("%w: address cannot be empty", ErrInvalidRequest)
	}
	if req.TransactionHash == "" {
		return nil, fmt.Errorf("%w: transaction hash cannot be empty", ErrInvalidRequest)
	}

	// Parse and validate amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid amount format: %w", ErrInvalidAmount, err)
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidAmount)
	}

	usdtAmount, err := payment.NewUSDTAmount(amount.String())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidAmount, err)
	}

	// Parse and validate address
	paymentAddress, err := payment.NewPaymentAddress(req.Address)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPaymentAddress, err)
	}

	// Parse and validate transaction hash
	transactionHash, err := payment.NewTransactionHash(req.TransactionHash)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidTransactionHash, err)
	}

	// Generate payment ID (in real implementation, this would be generated)
	paymentID := fmt.Sprintf("pay_%d", len(req.TransactionHash))

	// Create payment
	paymentEntity, err := payment.NewPayment(
		paymentID,
		usdtAmount,
		paymentAddress,
		transactionHash,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create payment: %w", ErrPaymentServiceError, err)
	}

	// Save payment
	if saveErr := s.repo.Save(ctx, paymentEntity); saveErr != nil {
		return nil, fmt.Errorf("%w: failed to save payment: %w", ErrPaymentServiceError, saveErr)
	}

	return paymentEntity, nil
}

// GetPayment retrieves a payment by its ID.
func (s *paymentServiceImpl) GetPayment(ctx context.Context, id string) (*payment.Payment, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: payment ID cannot be empty", ErrInvalidRequest)
	}

	paymentEntity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrPaymentNotFound, err)
		}
		return nil, fmt.Errorf("%w: failed to retrieve payment: %w", ErrPaymentServiceError, err)
	}

	return paymentEntity, nil
}

// GetPaymentByTransactionHash retrieves a payment by its transaction hash.
func (s *paymentServiceImpl) GetPaymentByTransactionHash(
	ctx context.Context,
	hash string,
) (*payment.Payment, error) {
	if hash == "" {
		return nil, fmt.Errorf("%w: transaction hash cannot be empty", ErrInvalidRequest)
	}

	transactionHash, err := payment.NewTransactionHash(hash)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidTransactionHash, err)
	}

	paymentEntity, err := s.repo.FindByTransactionHash(ctx, transactionHash)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrPaymentNotFound, err)
		}
		return nil, fmt.Errorf("%w: failed to retrieve payment by transaction hash: %w", ErrPaymentServiceError, err)
	}

	return paymentEntity, nil
}

// ListPaymentsByAddress retrieves all payments for a given address.
func (s *paymentServiceImpl) ListPaymentsByAddress(
	ctx context.Context,
	address string,
) ([]*payment.Payment, error) {
	if address == "" {
		return nil, fmt.Errorf("%w: address cannot be empty", ErrInvalidRequest)
	}

	paymentAddress, err := payment.NewPaymentAddress(address)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPaymentAddress, err)
	}

	payments, err := s.repo.FindByAddress(ctx, paymentAddress)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve payments by address: %w", ErrPaymentServiceError, err)
	}

	return payments, nil
}

// ListPaymentsByStatus retrieves all payments with the given status.
func (s *paymentServiceImpl) ListPaymentsByStatus(
	ctx context.Context,
	status payment.PaymentStatus,
) ([]*payment.Payment, error) {
	payments, err := s.repo.FindByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve payments by status: %w", ErrPaymentServiceError, err)
	}

	return payments, nil
}

// ListPendingPayments retrieves all pending payments (detected or confirming).
func (s *paymentServiceImpl) ListPendingPayments(ctx context.Context) ([]*payment.Payment, error) {
	payments, err := s.repo.FindPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve pending payments: %w", ErrPaymentServiceError, err)
	}

	return payments, nil
}

// ListConfirmedPayments retrieves all confirmed payments.
func (s *paymentServiceImpl) ListConfirmedPayments(ctx context.Context) ([]*payment.Payment, error) {
	payments, err := s.repo.FindConfirmed(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve confirmed payments: %w", ErrPaymentServiceError, err)
	}

	return payments, nil
}

// ListFailedPayments retrieves all failed payments.
func (s *paymentServiceImpl) ListFailedPayments(ctx context.Context) ([]*payment.Payment, error) {
	payments, err := s.repo.FindFailed(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve failed payments: %w", ErrPaymentServiceError, err)
	}

	return payments, nil
}

// ListOrphanedPayments retrieves all orphaned payments.
func (s *paymentServiceImpl) ListOrphanedPayments(ctx context.Context) ([]*payment.Payment, error) {
	payments, err := s.repo.FindOrphaned(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve orphaned payments: %w", ErrPaymentServiceError, err)
	}

	return payments, nil
}

// UpdatePaymentConfirmations updates the confirmation count for a payment.
func (s *paymentServiceImpl) UpdatePaymentConfirmations(
	ctx context.Context,
	id string,
	req UpdatePaymentConfirmationsRequest,
) error {
	if id == "" {
		return fmt.Errorf("%w: payment ID cannot be empty", ErrInvalidRequest)
	}
	if req.Confirmations < 0 {
		return fmt.Errorf("%w: confirmations cannot be negative", ErrInvalidRequest)
	}

	// Get the payment
	paymentEntity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			return fmt.Errorf("%w: %w", ErrPaymentNotFound, err)
		}
		return fmt.Errorf("%w: failed to retrieve payment: %w", ErrPaymentServiceError, err)
	}

	// Update confirmations
	if updateErr := paymentEntity.UpdateConfirmations(ctx, req.Confirmations); updateErr != nil {
		return fmt.Errorf("%w: failed to update confirmations: %w", ErrPaymentServiceError, updateErr)
	}

	// Save updated payment
	if saveErr := s.repo.Update(ctx, paymentEntity); saveErr != nil {
		return fmt.Errorf("%w: failed to save updated payment: %w", ErrPaymentServiceError, saveErr)
	}

	return nil
}

// MarkPaymentAsDetected marks a payment as detected (from orphaned state).
func (s *paymentServiceImpl) MarkPaymentAsDetected(ctx context.Context, id string) error {
	return s.transitionPayment(ctx, id, func(p *payment.Payment) error {
		return p.TransitionBackToDetected(ctx)
	})
}

// MarkPaymentAsIncluded marks a payment as included in a block.
func (s *paymentServiceImpl) MarkPaymentAsIncluded(ctx context.Context, id string) error {
	return s.transitionPayment(ctx, id, func(p *payment.Payment) error {
		return p.TransitionToConfirming(ctx)
	})
}

// MarkPaymentAsConfirmed marks a payment as confirmed.
func (s *paymentServiceImpl) MarkPaymentAsConfirmed(ctx context.Context, id string) error {
	return s.transitionPayment(ctx, id, func(p *payment.Payment) error {
		return p.TransitionToConfirmed(ctx)
	})
}

// MarkPaymentAsFailed marks a payment as failed.
func (s *paymentServiceImpl) MarkPaymentAsFailed(ctx context.Context, id string) error {
	return s.transitionPayment(ctx, id, func(p *payment.Payment) error {
		return p.TransitionToFailed(ctx)
	})
}

// MarkPaymentAsOrphaned marks a payment as orphaned.
func (s *paymentServiceImpl) MarkPaymentAsOrphaned(ctx context.Context, id string) error {
	return s.transitionPayment(ctx, id, func(p *payment.Payment) error {
		return p.TransitionToOrphaned(ctx)
	})
}

// MarkPaymentAsBackToMempool marks a payment as back to mempool.
func (s *paymentServiceImpl) MarkPaymentAsBackToMempool(ctx context.Context, id string) error {
	return s.transitionPayment(ctx, id, func(p *payment.Payment) error {
		return p.TransitionBackToDetected(ctx)
	})
}

// MarkPaymentAsDropped marks a payment as dropped.
func (s *paymentServiceImpl) MarkPaymentAsDropped(ctx context.Context, id string) error {
	return s.transitionPayment(ctx, id, func(p *payment.Payment) error {
		return p.TransitionToDropped(ctx)
	})
}

// GetPaymentStatistics returns payment statistics by status.
func (s *paymentServiceImpl) GetPaymentStatistics(ctx context.Context) (map[payment.PaymentStatus]int, error) {
	stats, err := s.repo.CountByStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve payment statistics: %w", ErrPaymentServiceError, err)
	}

	return stats, nil
}

// transitionPayment is a helper method to perform status transitions on payments.
func (s *paymentServiceImpl) transitionPayment(
	ctx context.Context,
	id string,
	transition func(*payment.Payment) error,
) error {
	if id == "" {
		return fmt.Errorf("%w: payment ID cannot be empty", ErrInvalidRequest)
	}

	// Get the payment
	paymentEntity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			return fmt.Errorf("%w: %w", ErrPaymentNotFound, err)
		}
		return fmt.Errorf("%w: failed to retrieve payment: %w", ErrPaymentServiceError, err)
	}

	// Perform the transition
	if transitionErr := transition(paymentEntity); transitionErr != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTransition, transitionErr)
	}

	// Save the updated payment
	if saveErr := s.repo.Update(ctx, paymentEntity); saveErr != nil {
		return fmt.Errorf("%w: failed to save updated payment: %w", ErrPaymentServiceError, saveErr)
	}

	return nil
}
