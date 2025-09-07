package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/shared"

	"gorm.io/gorm"
)

// InvoiceRepository implements the invoice.Repository interface using GORM.
type InvoiceRepository struct {
	db     *gorm.DB
	mapper *InvoiceMapper
}

// NewInvoiceRepository creates a new invoice repository.
func NewInvoiceRepository(db *gorm.DB) invoice.Repository {
	return &InvoiceRepository{
		db:     db,
		mapper: NewInvoiceMapper(),
	}
}

// Save persists an invoice to the database.
func (r *InvoiceRepository) Save(ctx context.Context, inv *invoice.Invoice) error {
	if inv == nil {
		return shared.ErrInvalidInput
	}

	// Convert domain model to database model
	model := r.mapper.ToModel(inv)

	// Retry logic for database locking issues
	const maxRetries = 5
	const baseRetryDelay = 5 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Save invoice (GORM will handle insert/update automatically)
			if err := tx.Save(model).Error; err != nil {
				return fmt.Errorf("failed to save invoice: %w", err)
			}
			return nil
		})

		if err == nil {
			return nil
		}

		// Check if it's a locking error that we should retry
		if isRetryableError(err) && attempt < maxRetries-1 {
			// Exponential backoff with jitter
			delay := baseRetryDelay * time.Duration(1<<attempt) // 5ms, 10ms, 20ms, 40ms
			time.Sleep(delay)
			continue
		}

		return err
	}

	return fmt.Errorf("failed to save invoice after %d attempts", maxRetries)
}

// isRetryableError checks if an error is retryable (e.g., database locked)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "database is locked") ||
		strings.Contains(errStr, "table is locked") ||
		strings.Contains(errStr, "busy") ||
		strings.Contains(errStr, "timeout")
}

// FindByID retrieves an invoice by its ID.
func (r *InvoiceRepository) FindByID(ctx context.Context, id string) (*invoice.Invoice, error) {
	if id == "" {
		return nil, shared.ErrInvalidInput
	}

	var model InvoiceModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	return r.mapper.ToDomain(&model)
}

// FindByPaymentAddress retrieves an invoice by its payment address.
func (r *InvoiceRepository) FindByPaymentAddress(
	ctx context.Context,
	address *shared.PaymentAddress,
) (*invoice.Invoice, error) {
	if address == nil {
		return nil, shared.ErrInvalidInput
	}

	var model InvoiceModel
	err := r.db.WithContext(ctx).
		Where("payment_address = ?", address.String()).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find invoice by payment address: %w", err)
	}

	return r.mapper.ToDomain(&model)
}

// FindByStatus retrieves all invoices with the given status.
func (r *InvoiceRepository) FindByStatus(
	ctx context.Context,
	status invoice.InvoiceStatus,
) ([]*invoice.Invoice, error) {
	var models []InvoiceModel
	err := r.db.WithContext(ctx).
		Where("status = ?", status.String()).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find invoices by status: %w", err)
	}

	return r.mapper.ToDomainSlice(models)
}

// FindActive retrieves all active (non-terminal) invoices.
func (r *InvoiceRepository) FindActive(ctx context.Context) ([]*invoice.Invoice, error) {
	activeStatuses := []string{
		invoice.StatusCreated.String(),
		invoice.StatusPending.String(),
		invoice.StatusPartial.String(),
		invoice.StatusConfirming.String(),
	}

	var models []InvoiceModel
	err := r.db.WithContext(ctx).
		Where("status IN ?", activeStatuses).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find active invoices: %w", err)
	}

	return r.mapper.ToDomainSlice(models)
}

// FindExpired retrieves all invoices that should be expired (have passed expiration time but are still active).
func (r *InvoiceRepository) FindExpired(ctx context.Context) ([]*invoice.Invoice, error) {
	// Find active invoices that have passed their expiration time
	activeStatuses := []string{
		invoice.StatusCreated.String(),
		invoice.StatusPending.String(),
		invoice.StatusPartial.String(),
		invoice.StatusConfirming.String(),
	}

	var models []InvoiceModel
	err := r.db.WithContext(ctx).
		Where("status IN ? AND expires_at < ?", activeStatuses, time.Now().UTC()).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find expired invoices: %w", err)
	}

	return r.mapper.ToDomainSlice(models)
}

// Update updates an existing invoice in the database.
func (r *InvoiceRepository) Update(ctx context.Context, inv *invoice.Invoice) error {
	if inv == nil {
		return shared.ErrInvalidInput
	}

	// Convert domain model to database model
	model := r.mapper.ToModel(inv)

	// Update invoice (items are now stored as JSONB in the main table)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update invoice in transaction: %w", err)
	}

	return nil
}

// Delete removes an invoice from the database.
func (r *InvoiceRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return shared.ErrInvalidInput
	}

	// Check if invoice exists first
	exists, err := r.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if invoice exists: %w", err)
	}
	if !exists {
		return shared.ErrNotFound
	}

	// Use soft delete - GORM will handle this automatically
	err = r.db.WithContext(ctx).Delete(&InvoiceModel{}, "id = ?", id).Error
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	return nil
}

// Exists checks if an invoice with the given ID exists.
func (r *InvoiceRepository) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, shared.ErrInvalidInput
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&InvoiceModel{}).
		Where("id = ?", id).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check invoice existence: %w", err)
	}

	return count > 0, nil
}
