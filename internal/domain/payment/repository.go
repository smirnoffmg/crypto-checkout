package payment

import (
	"context"
	"errors"
)

// Repository defines the interface for payment data persistence.
type Repository interface {
	// Save persists a payment to the data store.
	Save(ctx context.Context, payment *Payment) error

	// FindByID retrieves a payment by its ID.
	FindByID(ctx context.Context, id string) (*Payment, error)

	// FindByTransactionHash retrieves a payment by its transaction hash.
	FindByTransactionHash(ctx context.Context, hash *TransactionHash) (*Payment, error)

	// FindByAddress retrieves all payments for a given address.
	FindByAddress(ctx context.Context, address *Address) ([]*Payment, error)

	// FindByStatus retrieves all payments with the given status.
	FindByStatus(ctx context.Context, status PaymentStatus) ([]*Payment, error)

	// FindPending retrieves all pending payments (detected or confirming).
	FindPending(ctx context.Context) ([]*Payment, error)

	// FindConfirmed retrieves all confirmed payments.
	FindConfirmed(ctx context.Context) ([]*Payment, error)

	// FindFailed retrieves all failed payments.
	FindFailed(ctx context.Context) ([]*Payment, error)

	// FindOrphaned retrieves all orphaned payments.
	FindOrphaned(ctx context.Context) ([]*Payment, error)

	// Update updates an existing payment in the data store.
	Update(ctx context.Context, payment *Payment) error

	// Delete removes a payment from the data store.
	Delete(ctx context.Context, id string) error

	// Exists checks if a payment with the given ID exists.
	Exists(ctx context.Context, id string) (bool, error)

	// CountByStatus returns the count of payments for each status.
	CountByStatus(ctx context.Context) (map[PaymentStatus]int, error)
}

// Common repository errors.
var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentAlreadyExists = errors.New("payment already exists")
	ErrInvalidPayment       = errors.New("invalid payment")
	ErrRepositoryError      = errors.New("repository error")
)
