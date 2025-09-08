package database

import (
	"context"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/domain/shared"
	"errors"
	"fmt"

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

	// Save payment (GORM will handle insert/update automatically)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
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

	return r.modelToDomain(ctx, &model)
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
		Where("tx_hash = ?", hash.String()).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, payment.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment by transaction hash: %w", err)
	}

	return r.modelToDomain(ctx, &model)
}

// FindByAddress retrieves all payments for a given address.
func (r *PaymentRepository) FindByAddress(
	ctx context.Context,
	address *payment.PaymentAddress,
) ([]*payment.Payment, error) {
	if address == nil {
		return nil, payment.ErrInvalidPayment
	}

	var models []PaymentModel
	err := r.db.WithContext(ctx).
		Where("to_address = ?", address.String()).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find payments by address: %w", err)
	}

	return r.modelsToDomain(ctx, models)
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

	return r.modelsToDomain(ctx, models)
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

	return r.modelsToDomain(ctx, models)
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

	return r.modelsToDomain(ctx, models)
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

	return r.modelsToDomain(ctx, models)
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

	return r.modelsToDomain(ctx, models)
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

	// Check if payment exists first
	exists, err := r.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if payment exists: %w", err)
	}
	if !exists {
		return payment.ErrPaymentNotFound
	}

	// Use soft delete - GORM will handle this automatically
	err = r.db.WithContext(ctx).Delete(&PaymentModel{}, "id = ?", id).Error
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
	model := &PaymentModel{
		ID:                    string(p.ID()),
		InvoiceID:             string(p.InvoiceID()),
		Amount:                p.Amount().Amount().String(),
		FromAddress:           p.FromAddress(),
		ToAddress:             p.ToAddress().String(),
		TxHash:                p.TransactionHash().String(),
		Status:                p.Status().String(),
		Confirmations:         p.Confirmations().Int(),
		RequiredConfirmations: p.RequiredConfirmations(),
		DetectedAt:            p.DetectedAt(),
		CreatedAt:             p.CreatedAt(),
	}

	// Set optional fields
	if p.ConfirmedAt() != nil {
		confirmedAt := *p.ConfirmedAt()
		model.ConfirmedAt = &confirmedAt
	}
	if p.BlockInfo() != nil {
		blockNumber := p.BlockInfo().Number()
		blockHash := p.BlockInfo().Hash()
		model.BlockNumber = &blockNumber
		model.BlockHash = &blockHash
	}
	if p.NetworkFee() != nil {
		fee := p.NetworkFee().Fee().String()
		model.NetworkFee = &fee
	}

	return model
}

// modelToDomain converts a database model to a domain payment.
func (r *PaymentRepository) modelToDomain(ctx context.Context, model *PaymentModel) (*payment.Payment, error) {
	// Create payment amount
	amount, err := shared.NewMoneyWithCrypto(model.Amount, shared.CryptoCurrencyUSDT)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}
	paymentAmount, err := payment.NewPaymentAmount(amount, shared.CryptoCurrencyUSDT)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment amount: %w", err)
	}

	// Parse to address
	toAddress, err := payment.NewPaymentAddress(model.ToAddress, shared.NetworkTron)
	if err != nil {
		return nil, fmt.Errorf("failed to parse to address: %w", err)
	}

	// Parse transaction hash
	transactionHash, err := payment.NewTransactionHash(model.TxHash)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction hash: %w", err)
	}

	// Create payment
	p, err := payment.NewPayment(
		shared.PaymentID(model.ID),
		shared.InvoiceID(model.InvoiceID),
		paymentAmount,
		model.FromAddress,
		toAddress,
		transactionHash,
		model.RequiredConfirmations,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Set status from database
	p.SetStatus(payment.PaymentStatus(model.Status))

	// Update confirmations
	if updateErr := p.UpdateConfirmations(ctx, model.Confirmations); updateErr != nil {
		return nil, fmt.Errorf("failed to update confirmations: %w", updateErr)
	}

	// Set confirmed at if present
	if model.ConfirmedAt != nil {
		p.SetConfirmedAt(*model.ConfirmedAt)
	}

	// Set block info if present
	if model.BlockNumber != nil && model.BlockHash != nil {
		blockNumber := *model.BlockNumber
		blockHash := *model.BlockHash
		if blockErr := p.UpdateBlockInfo(blockNumber, blockHash); blockErr != nil {
			return nil, fmt.Errorf("failed to update block info: %w", blockErr)
		}
	}

	// Set network fee if present
	if model.NetworkFee != nil {
		fee, feeErr := shared.NewMoneyWithCrypto(*model.NetworkFee, shared.CryptoCurrencyUSDT)
		if feeErr != nil {
			return nil, fmt.Errorf("failed to parse network fee: %w", feeErr)
		}
		if networkFeeErr := p.UpdateNetworkFee(fee, shared.CryptoCurrencyUSDT); networkFeeErr != nil {
			return nil, fmt.Errorf("failed to update network fee: %w", networkFeeErr)
		}
	}

	return p, nil
}

// modelsToDomain converts multiple database models to domain payments.
func (r *PaymentRepository) modelsToDomain(ctx context.Context, models []PaymentModel) ([]*payment.Payment, error) {
	payments := make([]*payment.Payment, len(models))
	for i := range models {
		p, err := r.modelToDomain(ctx, &models[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert model %d: %w", i, err)
		}
		payments[i] = p
	}
	return payments, nil
}
