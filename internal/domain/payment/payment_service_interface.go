package payment

import (
	"context"
	"crypto-checkout/internal/domain/shared"
)

// PaymentService defines the interface for payment operations.
type PaymentService interface {
	// CreatePayment creates a new payment record.
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error)

	// GetPayment retrieves a payment by ID.
	GetPayment(ctx context.Context, id shared.PaymentID) (*Payment, error)

	// GetPaymentByTransactionHash retrieves a payment by transaction hash.
	GetPaymentByTransactionHash(ctx context.Context, txHash *TransactionHash) (*Payment, error)

	// UpdatePaymentStatus updates the payment status using the FSM.
	UpdatePaymentStatus(ctx context.Context, id shared.PaymentID, event string) error

	// UpdateConfirmations updates the confirmation count for a payment.
	UpdateConfirmations(ctx context.Context, id shared.PaymentID, count int) error

	// UpdateBlockInfo updates the block information for a payment.
	UpdateBlockInfo(ctx context.Context, id shared.PaymentID, blockNumber int64, blockHash string) error

	// UpdateNetworkFee updates the network fee for a payment.
	UpdateNetworkFee(ctx context.Context, id shared.PaymentID, fee *shared.Money, currency shared.CryptoCurrency) error

	// ListPaymentsByInvoice retrieves all payments for an invoice.
	ListPaymentsByInvoice(ctx context.Context, invoiceID shared.InvoiceID) ([]*Payment, error)

	// ListPaymentsByStatus retrieves all payments with the given status.
	ListPaymentsByStatus(ctx context.Context, status PaymentStatus) ([]*Payment, error)

	// ListPendingPayments retrieves all pending payments.
	ListPendingPayments(ctx context.Context) ([]*Payment, error)

	// ListConfirmedPayments retrieves all confirmed payments.
	ListConfirmedPayments(ctx context.Context) ([]*Payment, error)

	// ListFailedPayments retrieves all failed payments.
	ListFailedPayments(ctx context.Context) ([]*Payment, error)

	// ListOrphanedPayments retrieves all orphaned payments.
	ListOrphanedPayments(ctx context.Context) ([]*Payment, error)

	// GetPaymentStatistics returns payment statistics.
	GetPaymentStatistics(ctx context.Context) (*PaymentStatistics, error)
}

// CreatePaymentRequest represents a request to create a new payment.
type CreatePaymentRequest struct {
	ID                    shared.PaymentID
	InvoiceID             shared.InvoiceID
	Amount                *PaymentAmount
	FromAddress           string
	ToAddress             *PaymentAddress
	TransactionHash       *TransactionHash
	RequiredConfirmations int
}

// PaymentStatistics represents payment statistics.
type PaymentStatistics struct {
	TotalPayments           int
	ConfirmedPayments       int
	PendingPayments         int
	FailedPayments          int
	OrphanedPayments        int
	TotalAmount             *shared.Money
	AverageConfirmationTime int64 // in seconds
}

// PaymentListResponse represents a response containing multiple payments.
type PaymentListResponse struct {
	Payments []*Payment
	Total    int
	Page     int
	PageSize int
}
