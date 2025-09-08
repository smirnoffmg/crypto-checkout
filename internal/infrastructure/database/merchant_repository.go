package database

import (
	"context"
	"crypto-checkout/internal/domain/merchant"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MerchantRepository implements the merchant.Repository interface using GORM.
type MerchantRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMerchantRepository creates a new merchant repository.
func NewMerchantRepository(db *gorm.DB, logger *zap.Logger) merchant.MerchantRepository {
	return &MerchantRepository{
		db:     db,
		logger: logger,
	}
}

// Save saves a merchant to the database.
func (r *MerchantRepository) Save(ctx context.Context, m *merchant.Merchant) error {
	model, err := r.toModel(m)
	if err != nil {
		return fmt.Errorf("failed to convert merchant to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save merchant: %w", err)
	}

	r.logger.Debug("Merchant saved successfully",
		zap.String("merchant_id", m.ID()),
		zap.String("business_name", m.BusinessName()),
	)

	return nil
}

// FindByID finds a merchant by its ID.
func (r *MerchantRepository) FindByID(ctx context.Context, id string) (*merchant.Merchant, error) {
	var model MerchantModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("merchant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find merchant: %w", err)
	}

	return r.toDomain(&model)
}

// FindByEmail finds a merchant by its contact email.
func (r *MerchantRepository) FindByEmail(ctx context.Context, email string) (*merchant.Merchant, error) {
	var model MerchantModel
	if err := r.db.WithContext(ctx).Where("contact_email = ?", email).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("merchant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find merchant by email: %w", err)
	}

	return r.toDomain(&model)
}

// Update updates an existing merchant.
func (r *MerchantRepository) Update(ctx context.Context, m *merchant.Merchant) error {
	model, err := r.toModel(m)
	if err != nil {
		return fmt.Errorf("failed to convert merchant to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}

	r.logger.Debug("Merchant updated successfully",
		zap.String("merchant_id", m.ID()),
	)

	return nil
}

// Delete deletes a merchant by its ID.
func (r *MerchantRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&MerchantModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete merchant: %w", err)
	}

	r.logger.Debug("Merchant deleted successfully",
		zap.String("merchant_id", id),
	)

	return nil
}

// List lists merchants with pagination and filtering.
func (r *MerchantRepository) List(
	ctx context.Context,
	req *merchant.ListMerchantsRequest,
) (*merchant.ListMerchantsResponse, error) {
	var models []MerchantModel
	var total int64

	query := r.db.WithContext(ctx).Model(&MerchantModel{})

	// Apply filters
	if req.Status != nil {
		query = query.Where("status = ?", string(*req.Status))
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count merchants: %w", err)
	}

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Execute query
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list merchants: %w", err)
	}

	// Convert to domain objects
	merchants := make([]*merchant.Merchant, len(models))
	for i := range models {
		merchantObj, err := r.toDomain(&models[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert merchant model to domain: %w", err)
		}
		merchants[i] = merchantObj
	}

	return &merchant.ListMerchantsResponse{
		Merchants: merchants,
		Total:     int(total),
		Limit:     req.Limit,
		Offset:    req.Offset,
	}, nil
}

// toModel converts a domain merchant to a database model.
func (r *MerchantRepository) toModel(m *merchant.Merchant) (*MerchantModel, error) {
	settingsJSON, err := json.Marshal(m.Settings())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	return &MerchantModel{
		ID:           m.ID(),
		BusinessName: m.BusinessName(),
		ContactEmail: m.ContactEmail(),
		Status:       string(m.Status()),
		Settings:     string(settingsJSON),
		CreatedAt:    m.CreatedAt(),
		UpdatedAt:    m.UpdatedAt(),
	}, nil
}

// toDomain converts a database model to a domain merchant.
func (r *MerchantRepository) toDomain(model *MerchantModel) (*merchant.Merchant, error) {
	var settings merchant.MerchantSettings
	if err := json.Unmarshal([]byte(model.Settings), &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	// Create merchant with default status - the status will be set correctly by the domain
	m, err := merchant.NewMerchant(
		model.ID,
		model.BusinessName,
		model.ContactEmail,
		&settings,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create merchant: %w", err)
	}

	// Set the status from the database
	status := merchant.MerchantStatus(model.Status)
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status from database: %s", model.Status)
	}

	// Update the merchant's status to match the database
	if err := m.ChangeStatus(status); err != nil {
		return nil, fmt.Errorf("failed to set merchant status: %w", err)
	}

	return m, nil
}
