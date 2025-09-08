package database

import (
	"context"
	"crypto-checkout/internal/domain/merchant"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebhookEndpointRepository implements the merchant.WebhookEndpointRepository interface using GORM.
type WebhookEndpointRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewWebhookEndpointRepository creates a new webhook endpoint repository.
func NewWebhookEndpointRepository(db *gorm.DB, logger *zap.Logger) merchant.WebhookEndpointRepository {
	return &WebhookEndpointRepository{
		db:     db,
		logger: logger,
	}
}

// Save saves a webhook endpoint to the database.
func (r *WebhookEndpointRepository) Save(ctx context.Context, endpoint *merchant.WebhookEndpoint) error {
	model, err := r.toModel(endpoint)
	if err != nil {
		return fmt.Errorf("failed to convert webhook endpoint to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save webhook endpoint: %w", err)
	}

	r.logger.Debug("Webhook endpoint saved successfully",
		zap.String("endpoint_id", endpoint.ID()),
		zap.String("merchant_id", endpoint.MerchantID()),
	)

	return nil
}

// FindByID finds a webhook endpoint by its ID.
func (r *WebhookEndpointRepository) FindByID(ctx context.Context, id string) (*merchant.WebhookEndpoint, error) {
	var model WebhookEndpointModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("webhook endpoint not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find webhook endpoint: %w", err)
	}

	return r.toDomain(&model)
}

// FindByMerchantID finds all webhook endpoints for a merchant.
func (r *WebhookEndpointRepository) FindByMerchantID(
	ctx context.Context,
	merchantID string,
) ([]*merchant.WebhookEndpoint, error) {
	var models []WebhookEndpointModel
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find webhook endpoints for merchant: %w", err)
	}

	endpoints := make([]*merchant.WebhookEndpoint, len(models))
	for i := range models {
		endpoint, err := r.toDomain(&models[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert webhook endpoint model to domain: %w", err)
		}
		endpoints[i] = endpoint
	}

	return endpoints, nil
}

// FindActiveByMerchantID finds all active webhook endpoints for a merchant.
func (r *WebhookEndpointRepository) FindActiveByMerchantID(
	ctx context.Context,
	merchantID string,
) ([]*merchant.WebhookEndpoint, error) {
	var models []WebhookEndpointModel
	if err := r.db.WithContext(ctx).Where("merchant_id = ? AND status = ?", merchantID, "active").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find active webhook endpoints for merchant: %w", err)
	}

	endpoints := make([]*merchant.WebhookEndpoint, len(models))
	for i := range models {
		endpoint, err := r.toDomain(&models[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert webhook endpoint model to domain: %w", err)
		}
		endpoints[i] = endpoint
	}

	return endpoints, nil
}

// Update updates an existing webhook endpoint.
func (r *WebhookEndpointRepository) Update(ctx context.Context, endpoint *merchant.WebhookEndpoint) error {
	model, err := r.toModel(endpoint)
	if err != nil {
		return fmt.Errorf("failed to convert webhook endpoint to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update webhook endpoint: %w", err)
	}

	r.logger.Debug("Webhook endpoint updated successfully",
		zap.String("endpoint_id", endpoint.ID()),
	)

	return nil
}

// Delete deletes a webhook endpoint by its ID.
func (r *WebhookEndpointRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&WebhookEndpointModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete webhook endpoint: %w", err)
	}

	r.logger.Debug("Webhook endpoint deleted successfully",
		zap.String("endpoint_id", id),
	)

	return nil
}

// CountByMerchantID counts webhook endpoints for a merchant.
func (r *WebhookEndpointRepository) CountByMerchantID(ctx context.Context, merchantID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&WebhookEndpointModel{}).Where("merchant_id = ?", merchantID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count webhook endpoints for merchant: %w", err)
	}

	return int(count), nil
}

// toModel converts a domain webhook endpoint to a database model.
func (r *WebhookEndpointRepository) toModel(endpoint *merchant.WebhookEndpoint) (*WebhookEndpointModel, error) {
	eventsJSON, err := json.Marshal(endpoint.Events())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	allowedIPsJSON, err := json.Marshal(endpoint.AllowedIPs())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allowed IPs: %w", err)
	}

	headersJSON, err := json.Marshal(endpoint.Headers())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal headers: %w", err)
	}

	return &WebhookEndpointModel{
		ID:           endpoint.ID(),
		MerchantID:   endpoint.MerchantID(),
		URL:          endpoint.URL(),
		Events:       string(eventsJSON),
		Secret:       endpoint.Secret(),
		Status:       string(endpoint.Status()),
		MaxRetries:   endpoint.MaxRetries(),
		RetryBackoff: string(endpoint.RetryBackoff()),
		Timeout:      endpoint.Timeout(),
		AllowedIPs:   string(allowedIPsJSON),
		Headers:      string(headersJSON),
		CreatedAt:    endpoint.CreatedAt(),
		UpdatedAt:    endpoint.UpdatedAt(),
	}, nil
}

// toDomain converts a database model to a domain webhook endpoint.
func (r *WebhookEndpointRepository) toDomain(model *WebhookEndpointModel) (*merchant.WebhookEndpoint, error) {
	var events []string
	if err := json.Unmarshal([]byte(model.Events), &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	var allowedIPs []string
	if model.AllowedIPs != "" {
		if err := json.Unmarshal([]byte(model.AllowedIPs), &allowedIPs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed IPs: %w", err)
		}
	}

	var headers map[string]string
	if model.Headers != "" {
		if err := json.Unmarshal([]byte(model.Headers), &headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	}

	retryBackoff := merchant.BackoffStrategy(model.RetryBackoff)
	if !retryBackoff.IsValid() {
		return nil, fmt.Errorf("invalid retry backoff strategy from database: %s", model.RetryBackoff)
	}

	endpoint, err := merchant.NewWebhookEndpoint(
		model.ID,
		model.MerchantID,
		model.URL,
		events,
		model.Secret,
		model.MaxRetries,
		retryBackoff,
		model.Timeout,
		allowedIPs,
		headers,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook endpoint: %w", err)
	}

	// Set the status from the database
	status := merchant.EndpointStatus(model.Status)
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status from database: %s", model.Status)
	}

	if err := endpoint.ChangeStatus(status); err != nil {
		return nil, fmt.Errorf("failed to set webhook endpoint status: %w", err)
	}

	return endpoint, nil
}
