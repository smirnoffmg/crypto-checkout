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
	db *gorm.DB
}

// NewInvoiceRepository creates a new invoice repository.
func NewInvoiceRepository(db *gorm.DB) invoice.Repository {
	return &InvoiceRepository{db: db}
}

// Save persists an invoice to the database.
func (r *InvoiceRepository) Save(ctx context.Context, inv *invoice.Invoice) error {
	if inv == nil {
		return invoice.ErrInvalidInvoice
	}

	// Convert domain model to database model
	model := r.domainToModel(inv)

	// Save invoice and items in a transaction
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Save invoice
		if err := tx.Create(model).Error; err != nil {
			return fmt.Errorf("failed to save invoice: %w", err)
		}

		// Save invoice items
		for _, item := range model.Items {
			item.InvoiceID = model.ID
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("failed to save invoice item: %w", err)
			}
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

	return r.modelToDomain(&model)
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

	return r.modelToDomain(&model)
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

	return r.modelsToDomain(models)
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

	return r.modelsToDomain(models)
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

	return r.modelsToDomain(models)
}

// Update updates an existing invoice in the database.
func (r *InvoiceRepository) Update(ctx context.Context, inv *invoice.Invoice) error {
	if inv == nil {
		return invoice.ErrInvalidInvoice
	}

	// Convert domain model to database model
	model := r.domainToModel(inv)

	// Update invoice and items in a transaction
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update invoice
		if err := tx.Save(model).Error; err != nil {
			return fmt.Errorf("failed to update invoice: %w", err)
		}

		// Delete existing items and create new ones
		if err := tx.Where("invoice_id = ?", model.ID).Delete(&InvoiceItemModel{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing invoice items: %w", err)
		}

		// Create new items
		for _, item := range model.Items {
			item.InvoiceID = model.ID
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("failed to create invoice item: %w", err)
			}
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

// domainToModel converts a domain invoice to a database model.
func (r *InvoiceRepository) domainToModel(inv *invoice.Invoice) *InvoiceModel {
	model := &InvoiceModel{
		ID:        inv.ID(),
		Status:    inv.Status().String(),
		TaxRate:   inv.TaxRate().String(),
		CreatedAt: inv.CreatedAt(),
		PaidAt:    inv.PaidAt(),
	}

	// Set payment address if present
	if inv.PaymentAddress() != nil {
		address := inv.PaymentAddress().String()
		model.PaymentAddress = &address
	}

	// Convert items
	items := make([]InvoiceItemModel, len(inv.Items()))
	for i, item := range inv.Items() {
		items[i] = InvoiceItemModel{
			Description: item.Description(),
			UnitPrice:   item.UnitPrice().String(),
			Quantity:    item.Quantity().String(),
		}
	}
	model.Items = items

	return model
}

// modelToDomain converts a database model to a domain invoice.
func (r *InvoiceRepository) modelToDomain(model *InvoiceModel) (*invoice.Invoice, error) {
	// Parse tax rate
	taxRate := invoice.MustNewDecimal(model.TaxRate)

	// Convert items
	items := make([]*invoice.InvoiceItem, len(model.Items))
	for i, item := range model.Items {
		unitPrice := invoice.MustNewUSDTAmount(item.UnitPrice)
		quantity := invoice.MustNewDecimal(item.Quantity)

		invoiceItem, err := invoice.NewInvoiceItem(item.Description, unitPrice, quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}

		items[i] = invoiceItem
	}

	// Create invoice
	inv, err := invoice.NewInvoice(items, taxRate)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Set payment address if present
	if model.PaymentAddress != nil {
		address, addrErr := invoice.NewPaymentAddress(*model.PaymentAddress)
		if addrErr != nil {
			return nil, fmt.Errorf("failed to create payment address: %w", addrErr)
		}
		if assignErr := inv.AssignPaymentAddress(address); assignErr != nil {
			return nil, fmt.Errorf("failed to assign payment address: %w", assignErr)
		}
	}

	return inv, nil
}

// modelsToDomain converts multiple database models to domain invoices.
func (r *InvoiceRepository) modelsToDomain(models []InvoiceModel) ([]*invoice.Invoice, error) {
	invoices := make([]*invoice.Invoice, len(models))
	for i, model := range models {
		inv, err := r.modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model %d: %w", i, err)
		}
		invoices[i] = inv
	}
	return invoices, nil
}
