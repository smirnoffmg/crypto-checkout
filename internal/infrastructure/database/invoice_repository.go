package database

import (
	"context"
	"errors"
	"fmt"

	"crypto-checkout/internal/domain/invoice"

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
		return invoice.ErrInvalidInvoice
	}

	// Convert domain model to database model
	model := r.mapper.ToModel(inv)

	// Save invoice and items in a transaction
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Save invoice (GORM will automatically save associated items)
		if err := tx.Create(model).Error; err != nil {
			return fmt.Errorf("failed to save invoice: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to save invoice in transaction: %w", err)
	}

	return nil
}

// FindByID retrieves an invoice by its ID.
func (r *InvoiceRepository) FindByID(ctx context.Context, id string) (*invoice.Invoice, error) {
	if id == "" {
		return nil, invoice.ErrInvalidInvoice
	}

	var model InvoiceModel
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, invoice.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	return r.mapper.ToDomain(&model)
}

// FindByPaymentAddress retrieves an invoice by its payment address.
func (r *InvoiceRepository) FindByPaymentAddress(
	ctx context.Context,
	address *invoice.PaymentAddress,
) (*invoice.Invoice, error) {
	if address == nil {
		return nil, invoice.ErrInvalidInvoice
	}

	var model InvoiceModel
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("payment_address = ?", address.String()).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, invoice.ErrInvoiceNotFound
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
		Preload("Items").
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
		Preload("Items").
		Where("status IN ?", activeStatuses).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find active invoices: %w", err)
	}

	return r.mapper.ToDomainSlice(models)
}

// FindExpired retrieves all expired invoices.
func (r *InvoiceRepository) FindExpired(ctx context.Context) ([]*invoice.Invoice, error) {
	var models []InvoiceModel
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("status = ?", invoice.StatusExpired.String()).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find expired invoices: %w", err)
	}

	return r.mapper.ToDomainSlice(models)
}

// Update updates an existing invoice in the database.
func (r *InvoiceRepository) Update(ctx context.Context, inv *invoice.Invoice) error {
	if inv == nil {
		return invoice.ErrInvalidInvoice
	}

	// Convert domain model to database model
	model := r.mapper.ToModel(inv)

	// Update invoice and items in a transaction
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing items first
		if err := tx.Where("invoice_id = ?", model.ID).Delete(&InvoiceItemModel{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing invoice items: %w", err)
		}

		// Update invoice (GORM will automatically save associated items)
		if err := tx.Save(model).Error; err != nil {
			return fmt.Errorf("failed to update invoice: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to update invoice in transaction: %w", err)
	}

	return nil
}

// Delete removes an invoice from the database.
func (r *InvoiceRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return invoice.ErrInvalidInvoice
	}

	// Use soft delete - GORM will handle this automatically
	err := r.db.WithContext(ctx).Delete(&InvoiceModel{}, "id = ?", id).Error
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	return nil
}

// Exists checks if an invoice with the given ID exists.
func (r *InvoiceRepository) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, invoice.ErrInvalidInvoice
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
