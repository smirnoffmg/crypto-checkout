package database

import (
	"context"
	"errors"
	"fmt"

	"crypto-checkout/internal/domain/payment"

	"gorm.io/gorm"
)

// PaymentRepository implements the payment.Repository interface using GORM.
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new payment repository.
func NewPaymentRepository(db *gorm.DB) payment.Repository {
	return &PaymentRepository{db: db}
}

// Save persists a payment to the database.
func (r *PaymentRepository) Save(ctx context.Context, p *payment.Payment) error {
	if p == nil {
		return payment.ErrInvalidPayment
	}

	// Convert domain model to database model
	model := r.domainToModel(p)

	// Save payment
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	return nil
}

// FindByID retrieves a payment by its ID.
func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*payment.Payment, error) {
	if id == "" {
		return nil, payment.ErrInvalidPayment
	}

	var model PaymentModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, payment.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return r.modelToDomain(&model)
}

// FindByTransactionHash retrieves a payment by its transaction hash.
func (r *PaymentRepository) FindByTransactionHash(
	ctx context.Context,
	hash *payment.TransactionHash,
) (*payment.Payment, error) {
	if hash == nil {
		return nil, payment.ErrInvalidPayment
	}

	var model PaymentModel
	err := r.db.WithContext(ctx).
		Where("transaction_hash = ?", hash.String()).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, payment.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment by transaction hash: %w", err)
	}

	return r.modelToDomain(&model)
}

// FindByAddress retrieves all payments for a given address.
func (r *PaymentRepository) FindByAddress(ctx context.Context, address *payment.Address) ([]*payment.Payment, error) {
	if address == nil {
		return nil, payment.ErrInvalidPayment
	}

	var models []PaymentModel
	err := r.db.WithContext(ctx).
		Where("address = ?", address.String()).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find payments by address: %w", err)
	}

	return r.modelsToDomain(models)
}

// FindByStatus retrieves all payments with the given status.
func (r *PaymentRepository) FindByStatus(
	ctx context.Context,
	status payment.PaymentStatus,
) ([]*payment.Payment, error) {
	var models []PaymentModel
	err := r.db.WithContext(ctx).
		Where("status = ?", status.String()).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find payments by status: %w", err)
	}

	return r.modelsToDomain(models)
}

// FindPending retrieves all pending payments (detected or confirming).
func (r *PaymentRepository) FindPending(ctx context.Context) ([]*payment.Payment, error) {
	pendingStatuses := []string{
		payment.StatusDetected.String(),
		payment.StatusConfirming.String(),
	}

	var models []PaymentModel
	err := r.db.WithContext(ctx).
		Where("status IN ?", pendingStatuses).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find pending payments: %w", err)
	}

	return r.modelsToDomain(models)
}

// FindConfirmed retrieves all confirmed payments.
func (r *PaymentRepository) FindConfirmed(ctx context.Context) ([]*payment.Payment, error) {
	var models []PaymentModel
	err := r.db.WithContext(ctx).
		Where("status = ?", payment.StatusConfirmed.String()).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find confirmed payments: %w", err)
	}

	return r.modelsToDomain(models)
}

// FindFailed retrieves all failed payments.
func (r *PaymentRepository) FindFailed(ctx context.Context) ([]*payment.Payment, error) {
	var models []PaymentModel
	err := r.db.WithContext(ctx).
		Where("status = ?", payment.StatusFailed.String()).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find failed payments: %w", err)
	}

	return r.modelsToDomain(models)
}

// FindOrphaned retrieves all orphaned payments.
func (r *PaymentRepository) FindOrphaned(ctx context.Context) ([]*payment.Payment, error) {
	var models []PaymentModel
	err := r.db.WithContext(ctx).
		Where("status = ?", payment.StatusOrphaned.String()).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find orphaned payments: %w", err)
	}

	return r.modelsToDomain(models)
}

// Update updates an existing payment in the database.
func (r *PaymentRepository) Update(ctx context.Context, p *payment.Payment) error {
	if p == nil {
		return payment.ErrInvalidPayment
	}

	// Convert domain model to database model
	model := r.domainToModel(p)

	// Update payment
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	return nil
}

// Delete removes a payment from the database.
func (r *PaymentRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return payment.ErrInvalidPayment
	}

	// Use soft delete - GORM will handle this automatically
	err := r.db.WithContext(ctx).Delete(&PaymentModel{}, "id = ?", id).Error
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	return nil
}

// Exists checks if a payment with the given ID exists.
func (r *PaymentRepository) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, payment.ErrInvalidPayment
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&PaymentModel{}).
		Where("id = ?", id).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check payment existence: %w", err)
	}

	return count > 0, nil
}

// CountByStatus returns the count of payments for each status.
func (r *PaymentRepository) CountByStatus(ctx context.Context) (map[payment.PaymentStatus]int, error) {
	var results []struct {
		Status string `gorm:"column:status"`
		Count  int    `gorm:"column:count"`
	}

	err := r.db.WithContext(ctx).
		Model(&PaymentModel{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count payments by status: %w", err)
	}

	// Convert to map
	counts := make(map[payment.PaymentStatus]int)
	for _, result := range results {
		status, parseErr := payment.ParsePaymentStatus(result.Status)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse status %s: %w", result.Status, parseErr)
		}
		counts[status] = result.Count
	}

	return counts, nil
}

// domainToModel converts a domain payment to a database model.
func (r *PaymentRepository) domainToModel(p *payment.Payment) *PaymentModel {
	return &PaymentModel{
		ID:              p.ID(),
		Amount:          p.Amount().String(),
		Address:         p.Address().String(),
		TransactionHash: p.TransactionHash().String(),
		Confirmations:   p.Confirmations().Int(),
		Status:          p.Status().String(),
		CreatedAt:       p.CreatedAt(),
		UpdatedAt:       p.UpdatedAt(),
	}
}

// modelToDomain converts a database model to a domain payment.
func (r *PaymentRepository) modelToDomain(model *PaymentModel) (*payment.Payment, error) {
	// Parse amount
	amount, err := payment.NewUSDTAmount(model.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}

	// Parse address
	address, err := payment.NewPaymentAddress(model.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address: %w", err)
	}

	// Parse transaction hash
	transactionHash, err := payment.NewTransactionHash(model.TransactionHash)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction hash: %w", err)
	}

	// Create payment
	p, err := payment.NewPayment(model.ID, amount, address, transactionHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update confirmations
	if updateErr := p.UpdateConfirmations(context.Background(), model.Confirmations); updateErr != nil {
		return nil, fmt.Errorf("failed to update confirmations: %w", updateErr)
	}

	return p, nil
}

// modelsToDomain converts multiple database models to domain payments.
func (r *PaymentRepository) modelsToDomain(models []PaymentModel) ([]*payment.Payment, error) {
	payments := make([]*payment.Payment, len(models))
	for i, model := range models {
		p, err := r.modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model %d: %w", i, err)
		}
		payments[i] = p
	}
	return payments, nil
}
