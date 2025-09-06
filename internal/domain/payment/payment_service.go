// Package service provides application services for business logic orchestration.
package payment

import (
	"context"
)

// CreatePaymentRequest represents a request to create a new
type CreatePaymentRequest struct {
	Amount          string `json:"amount"`
	Address         string `json:"address"`
	TransactionHash string `json:"transactionHash"`
}

// UpdatePaymentConfirmationsRequest represents a request to update payment confirmations.
type UpdatePaymentConfirmationsRequest struct {
	Confirmations int `json:"confirmations"`
}

// PaymentService defines the interface for payment-related business operations.
type PaymentService interface {
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (*Payment, error)

	// GetPayment retrieves a payment by its ID.
	GetPayment(ctx context.Context, id string) (*Payment, error)

	// GetPaymentByTransactionHash retrieves a payment by its transaction hash.
	GetPaymentByTransactionHash(ctx context.Context, hash string) (*Payment, error)

	// ListPaymentsByAddress retrieves all payments for a given address.
	ListPaymentsByAddress(ctx context.Context, address string) ([]*Payment, error)

	// ListPaymentsByStatus retrieves all payments with the given status.
	ListPaymentsByStatus(ctx context.Context, status PaymentStatus) ([]*Payment, error)

	// ListPendingPayments retrieves all pending payments (detected or confirming).
	ListPendingPayments(ctx context.Context) ([]*Payment, error)

	// ListConfirmedPayments retrieves all confirmed payments.
	ListConfirmedPayments(ctx context.Context) ([]*Payment, error)

	// ListFailedPayments retrieves all failed payments.
	ListFailedPayments(ctx context.Context) ([]*Payment, error)

	// ListOrphanedPayments retrieves all orphaned payments.
	ListOrphanedPayments(ctx context.Context) ([]*Payment, error)

	// UpdatePaymentConfirmations updates the confirmation count for a
	UpdatePaymentConfirmations(ctx context.Context, id string, req UpdatePaymentConfirmationsRequest) error

	// MarkPaymentAsDetected marks a payment as detected.
	MarkPaymentAsDetected(ctx context.Context, id string) error

	// MarkPaymentAsIncluded marks a payment as included in a block.
	MarkPaymentAsIncluded(ctx context.Context, id string) error

	// MarkPaymentAsConfirmed marks a payment as confirmed.
	MarkPaymentAsConfirmed(ctx context.Context, id string) error

	// MarkPaymentAsFailed marks a payment as failed.
	MarkPaymentAsFailed(ctx context.Context, id string) error

	// MarkPaymentAsOrphaned marks a payment as orphaned.
	MarkPaymentAsOrphaned(ctx context.Context, id string) error

	// MarkPaymentAsBackToMempool marks a payment as back to mempool.
	MarkPaymentAsBackToMempool(ctx context.Context, id string) error

	// MarkPaymentAsDropped marks a payment as dropped.
	MarkPaymentAsDropped(ctx context.Context, id string) error

	// GetPaymentStatistics returns payment statistics by status.
	GetPaymentStatistics(ctx context.Context) (map[PaymentStatus]int, error)
}
