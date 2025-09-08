package database

import (
	"context"
	"crypto-checkout/internal/domain/merchant"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// APIKeyRepository implements the merchant.APIKeyRepository interface using GORM.
type APIKeyRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAPIKeyRepository creates a new API key repository.
func NewAPIKeyRepository(db *gorm.DB, logger *zap.Logger) merchant.APIKeyRepository {
	return &APIKeyRepository{
		db:     db,
		logger: logger,
	}
}

// Save saves an API key to the database.
func (r *APIKeyRepository) Save(ctx context.Context, apiKey *merchant.APIKey) error {
	model, err := r.toModel(apiKey)
	if err != nil {
		return fmt.Errorf("failed to convert API key to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}

	r.logger.Debug("API key saved successfully",
		zap.String("api_key_id", apiKey.ID()),
		zap.String("merchant_id", apiKey.MerchantID()),
	)

	return nil
}

// FindByID finds an API key by its ID.
func (r *APIKeyRepository) FindByID(ctx context.Context, id string) (*merchant.APIKey, error) {
	var model APIKeyModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("API key not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find API key: %w", err)
	}

	return r.toDomain(&model)
}

// FindByHash finds an API key by its hash.
func (r *APIKeyRepository) FindByHash(ctx context.Context, hash *merchant.APIKeyHash) (*merchant.APIKey, error) {
	var model APIKeyModel
	if err := r.db.WithContext(ctx).Where("key_hash = ?", hash.String()).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("API key not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find API key by hash: %w", err)
	}

	return r.toDomain(&model)
}

// FindByMerchantID finds all API keys for a merchant.
func (r *APIKeyRepository) FindByMerchantID(ctx context.Context, merchantID string) ([]*merchant.APIKey, error) {
	var models []APIKeyModel
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find API keys for merchant: %w", err)
	}

	apiKeys := make([]*merchant.APIKey, len(models))
	for i := range models {
		apiKey, err := r.toDomain(&models[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert API key model to domain: %w", err)
		}
		apiKeys[i] = apiKey
	}

	return apiKeys, nil
}

// Update updates an existing API key.
func (r *APIKeyRepository) Update(ctx context.Context, apiKey *merchant.APIKey) error {
	model, err := r.toModel(apiKey)
	if err != nil {
		return fmt.Errorf("failed to convert API key to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	r.logger.Debug("API key updated successfully",
		zap.String("api_key_id", apiKey.ID()),
	)

	return nil
}

// Delete deletes an API key by its ID.
func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&APIKeyModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	r.logger.Debug("API key deleted successfully",
		zap.String("api_key_id", id),
	)

	return nil
}

// CountByMerchantID counts API keys for a merchant.
func (r *APIKeyRepository) CountByMerchantID(ctx context.Context, merchantID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&APIKeyModel{}).Where("merchant_id = ?", merchantID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count API keys for merchant: %w", err)
	}

	return int(count), nil
}

// toModel converts a domain API key to a database model.
func (r *APIKeyRepository) toModel(apiKey *merchant.APIKey) (*APIKeyModel, error) {
	permissionsJSON, err := json.Marshal(apiKey.Permissions())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	model := &APIKeyModel{
		ID:          apiKey.ID(),
		MerchantID:  apiKey.MerchantID(),
		KeyHash:     apiKey.KeyHash().String(),
		KeyType:     string(apiKey.KeyType()),
		Permissions: string(permissionsJSON),
		Status:      string(apiKey.Status()),
		Name:        apiKey.Name(),
		CreatedAt:   apiKey.CreatedAt(),
	}

	if apiKey.LastUsedAt() != nil {
		model.LastUsedAt = apiKey.LastUsedAt()
	}
	if apiKey.ExpiresAt() != nil {
		model.ExpiresAt = apiKey.ExpiresAt()
	}

	return model, nil
}

// toDomain converts a database model to a domain API key.
func (r *APIKeyRepository) toDomain(model *APIKeyModel) (*merchant.APIKey, error) {
	var permissions []string
	if err := json.Unmarshal([]byte(model.Permissions), &permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	// Create APIKeyHash with the stored hash value
	keyHash, err := merchant.NewAPIKeyHashFromString(model.KeyHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key hash from string: %w", err)
	}

	keyType := merchant.KeyType(model.KeyType)
	if !keyType.IsValid() {
		return nil, fmt.Errorf("invalid key type from database: %s", model.KeyType)
	}

	status := merchant.KeyStatus(model.Status)
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status from database: %s", model.Status)
	}

	apiKey, err := merchant.NewAPIKeyFromHash(
		model.ID,
		model.MerchantID,
		keyHash,
		keyType,
		permissions,
		status,
		model.Name,
		model.LastUsedAt,
		model.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return apiKey, nil
}
