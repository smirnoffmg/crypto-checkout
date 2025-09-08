package payment

import (
	"context"
	"crypto-checkout/internal/domain/shared"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// PaymentServiceImpl implements the PaymentService interface.
type PaymentServiceImpl struct {
	repository Repository
	eventBus   shared.EventBus
	logger     *zap.Logger
}

// NewPaymentService creates a new payment service.
func NewPaymentService(repository Repository, eventBus shared.EventBus, logger *zap.Logger) PaymentService {
	logger.Info("Creating PaymentService",
		zap.Bool("eventBus_provided", eventBus != nil),
		zap.Bool("repository_provided", repository != nil))

	return &PaymentServiceImpl{
		repository: repository,
		eventBus:   eventBus,
		logger:     logger,
	}
}

// CreatePayment creates a new payment record.
func (s *PaymentServiceImpl) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error) {
	if req == nil {
		return nil, NewPaymentError(shared.ErrCodeValidationFailed, "create payment request cannot be nil", nil)
	}

	// Check if payment with same transaction hash already exists
	existingPayment, err := s.repository.FindByTransactionHash(ctx, req.TransactionHash)
	if err != nil && err != ErrPaymentNotFound {
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}

	if existingPayment != nil {
		return nil, NewPaymentAlreadyExistsError(req.TransactionHash.String())
	}

	// Create new payment
	payment, err := NewPayment(
		req.ID,
		req.InvoiceID,
		req.Amount,
		req.FromAddress,
		req.ToAddress,
		req.TransactionHash,
		req.RequiredConfirmations,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Save to repository
	if err := s.repository.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Publish payment detected event
	if s.eventBus != nil {
		eventData := createPaymentEventData(payment)
		eventData["timestamp"] = time.Now().UTC()
		event := shared.CreateDomainEvent(
			shared.EventTypePaymentDetected,
			string(payment.ID()),
			"Payment",
			eventData,
			nil,
		)
		if err := s.eventBus.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
			if s.logger != nil {
				s.logger.Error("Failed to publish domain event",
					zap.String("event_type", shared.EventTypePaymentDetected),
					zap.String("aggregate_id", string(payment.ID())),
					zap.Error(err),
				)
			}
		}
	}

	return payment, nil
}

// GetPayment retrieves a payment by ID.
func (s *PaymentServiceImpl) GetPayment(ctx context.Context, id shared.PaymentID) (*Payment, error) {
	if id == "" {
		return nil, NewPaymentError(shared.ErrCodeValidationFailed, "payment ID cannot be empty", nil)
	}

	payment, err := s.repository.FindByID(ctx, string(id))
	if err != nil {
		if err == ErrPaymentNotFound {
			return nil, NewPaymentNotFoundError(string(id))
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// GetPaymentByTransactionHash retrieves a payment by transaction hash.
func (s *PaymentServiceImpl) GetPaymentByTransactionHash(
	ctx context.Context,
	txHash *TransactionHash,
) (*Payment, error) {
	if txHash == nil {
		return nil, NewPaymentError(shared.ErrCodeValidationFailed, "transaction hash cannot be nil", nil)
	}

	payment, err := s.repository.FindByTransactionHash(ctx, txHash)
	if err != nil {
		if err == ErrPaymentNotFound {
			return nil, NewPaymentNotFoundError(txHash.String())
		}
		return nil, fmt.Errorf("failed to get payment by transaction hash: %w", err)
	}

	return payment, nil
}

// UpdatePaymentStatus updates the payment status using the FSM.
func (s *PaymentServiceImpl) UpdatePaymentStatus(ctx context.Context, id shared.PaymentID, event string) error {
	if id == "" {
		return NewPaymentError(shared.ErrCodeValidationFailed, "payment ID cannot be empty", nil)
	}

	if event == "" {
		return NewPaymentError(shared.ErrCodeValidationFailed, "event cannot be empty", nil)
	}

	// Get the payment
	payment, err := s.GetPayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Create FSM and trigger event
	fsm := NewPaymentFSM(payment)
	if err := fsm.Event(ctx, event); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Save updated payment
	if err := s.repository.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to save updated payment: %w", err)
	}

	// Publish payment status changed event
	if s.eventBus != nil {
		eventData := createPaymentEventData(payment)
		eventData["event_triggered"] = event
		eventData["status_changed_at"] = time.Now().UTC()

		eventData["timestamp"] = time.Now().UTC()
		event := shared.CreateDomainEvent(
			shared.EventTypePaymentStatusChanged,
			string(payment.ID()),
			"Payment",
			eventData,
			nil,
		)
		if err := s.eventBus.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
			if s.logger != nil {
				s.logger.Error("Failed to publish domain event",
					zap.String("event_type", shared.EventTypePaymentStatusChanged),
					zap.String("aggregate_id", string(payment.ID())),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// UpdateConfirmations updates the confirmation count for a payment.
func (s *PaymentServiceImpl) UpdateConfirmations(ctx context.Context, id shared.PaymentID, count int) error {
	if id == "" {
		return NewPaymentError(shared.ErrCodeValidationFailed, "payment ID cannot be empty", nil)
	}

	// Get the payment
	payment, err := s.GetPayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Update confirmations
	if err := payment.UpdateConfirmations(ctx, count); err != nil {
		return fmt.Errorf("failed to update confirmations: %w", err)
	}

	// Check if payment should be confirmed
	if payment.IsConfirmed() && payment.Status() == StatusConfirming {
		if err := s.UpdatePaymentStatus(ctx, id, "confirm"); err != nil {
			return fmt.Errorf("failed to confirm payment: %w", err)
		}
		return nil
	}

	// Save updated payment
	if err := s.repository.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to save updated payment: %w", err)
	}

	return nil
}

// UpdateBlockInfo updates the block information for a payment.
func (s *PaymentServiceImpl) UpdateBlockInfo(
	ctx context.Context,
	id shared.PaymentID,
	blockNumber int64,
	blockHash string,
) error {
	if id == "" {
		return NewPaymentError(shared.ErrCodeValidationFailed, "payment ID cannot be empty", nil)
	}

	// Get the payment
	payment, err := s.GetPayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Update block info
	if err := payment.UpdateBlockInfo(blockNumber, blockHash); err != nil {
		return fmt.Errorf("failed to update block info: %w", err)
	}

	// If payment is detected, transition to confirming
	if payment.Status() == StatusDetected {
		if err := s.UpdatePaymentStatus(ctx, id, "include_in_block"); err != nil {
			return fmt.Errorf("failed to transition payment to confirming: %w", err)
		}
		return nil
	}

	// Save updated payment
	if err := s.repository.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to save updated payment: %w", err)
	}

	return nil
}

// UpdateNetworkFee updates the network fee for a payment.
func (s *PaymentServiceImpl) UpdateNetworkFee(
	ctx context.Context,
	id shared.PaymentID,
	fee *shared.Money,
	currency shared.CryptoCurrency,
) error {
	if id == "" {
		return NewPaymentError(shared.ErrCodeValidationFailed, "payment ID cannot be empty", nil)
	}

	// Get the payment
	payment, err := s.GetPayment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Update network fee
	if err := payment.UpdateNetworkFee(fee, currency); err != nil {
		return fmt.Errorf("failed to update network fee: %w", err)
	}

	// Save updated payment
	if err := s.repository.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to save updated payment: %w", err)
	}

	return nil
}

// ListPaymentsByInvoice retrieves all payments for an invoice.
func (s *PaymentServiceImpl) ListPaymentsByInvoice(
	ctx context.Context,
	invoiceID shared.InvoiceID,
) ([]*Payment, error) {
	if invoiceID == "" {
		return nil, NewPaymentError(shared.ErrCodeValidationFailed, "invoice ID cannot be empty", nil)
	}

	// This would need to be implemented in the repository
	// For now, we'll return an error indicating it's not implemented
	return nil, fmt.Errorf("list payments by invoice not implemented")
}

// ListPaymentsByStatus retrieves all payments with the given status.
func (s *PaymentServiceImpl) ListPaymentsByStatus(ctx context.Context, status PaymentStatus) ([]*Payment, error) {
	payments, err := s.repository.FindByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments by status: %w", err)
	}

	return payments, nil
}

// ListPendingPayments retrieves all pending payments.
func (s *PaymentServiceImpl) ListPendingPayments(ctx context.Context) ([]*Payment, error) {
	payments, err := s.repository.FindPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending payments: %w", err)
	}

	return payments, nil
}

// ListConfirmedPayments retrieves all confirmed payments.
func (s *PaymentServiceImpl) ListConfirmedPayments(ctx context.Context) ([]*Payment, error) {
	payments, err := s.repository.FindConfirmed(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list confirmed payments: %w", err)
	}

	return payments, nil
}

// ListFailedPayments retrieves all failed payments.
func (s *PaymentServiceImpl) ListFailedPayments(ctx context.Context) ([]*Payment, error) {
	payments, err := s.repository.FindFailed(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list failed payments: %w", err)
	}

	return payments, nil
}

// ListOrphanedPayments retrieves all orphaned payments.
func (s *PaymentServiceImpl) ListOrphanedPayments(ctx context.Context) ([]*Payment, error) {
	payments, err := s.repository.FindOrphaned(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list orphaned payments: %w", err)
	}

	return payments, nil
}

// GetPaymentStatistics returns payment statistics.
func (s *PaymentServiceImpl) GetPaymentStatistics(ctx context.Context) (*PaymentStatistics, error) {
	// Get counts by status
	counts, err := s.repository.CountByStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment counts: %w", err)
	}

	stats := &PaymentStatistics{
		TotalPayments:     0,
		ConfirmedPayments: counts[StatusConfirmed],
		PendingPayments:   counts[StatusDetected] + counts[StatusConfirming],
		FailedPayments:    counts[StatusFailed],
		OrphanedPayments:  counts[StatusOrphaned],
	}

	// Calculate total
	for _, count := range counts {
		stats.TotalPayments += count
	}

	// TODO: Calculate total amount and average confirmation time
	// This would require additional repository methods

	return stats, nil
}
