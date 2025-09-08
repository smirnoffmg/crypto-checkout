package merchant

import (
	"context"
)

// MerchantRepository defines the interface for merchant data persistence.
type MerchantRepository interface {
	// Save saves a merchant to the repository.
	Save(ctx context.Context, merchant *Merchant) error

	// FindByID finds a merchant by its ID.
	FindByID(ctx context.Context, id string) (*Merchant, error)

	// FindByEmail finds a merchant by its contact email.
	FindByEmail(ctx context.Context, email string) (*Merchant, error)

	// Update updates an existing merchant.
	Update(ctx context.Context, merchant *Merchant) error

	// Delete deletes a merchant by its ID.
	Delete(ctx context.Context, id string) error

	// List lists merchants with pagination and filtering.
	List(ctx context.Context, req *ListMerchantsRequest) (*ListMerchantsResponse, error)
}

// APIKeyRepository defines the interface for API key data persistence.
type APIKeyRepository interface {
	// Save saves an API key to the repository.
	Save(ctx context.Context, apiKey *APIKey) error

	// FindByID finds an API key by its ID.
	FindByID(ctx context.Context, id string) (*APIKey, error)

	// FindByHash finds an API key by its hash.
	FindByHash(ctx context.Context, hash *APIKeyHash) (*APIKey, error)

	// FindByMerchantID finds all API keys for a merchant.
	FindByMerchantID(ctx context.Context, merchantID string) ([]*APIKey, error)

	// Update updates an existing API key.
	Update(ctx context.Context, apiKey *APIKey) error

	// Delete deletes an API key by its ID.
	Delete(ctx context.Context, id string) error

	// CountByMerchantID counts API keys for a merchant.
	CountByMerchantID(ctx context.Context, merchantID string) (int, error)
}

// WebhookEndpointRepository defines the interface for webhook endpoint data persistence.
type WebhookEndpointRepository interface {
	// Save saves a webhook endpoint to the repository.
	Save(ctx context.Context, endpoint *WebhookEndpoint) error

	// FindByID finds a webhook endpoint by its ID.
	FindByID(ctx context.Context, id string) (*WebhookEndpoint, error)

	// FindByMerchantID finds all webhook endpoints for a merchant.
	FindByMerchantID(ctx context.Context, merchantID string) ([]*WebhookEndpoint, error)

	// FindActiveByMerchantID finds all active webhook endpoints for a merchant.
	FindActiveByMerchantID(ctx context.Context, merchantID string) ([]*WebhookEndpoint, error)

	// Update updates an existing webhook endpoint.
	Update(ctx context.Context, endpoint *WebhookEndpoint) error

	// Delete deletes a webhook endpoint by its ID.
	Delete(ctx context.Context, id string) error

	// CountByMerchantID counts webhook endpoints for a merchant.
	CountByMerchantID(ctx context.Context, merchantID string) (int, error)
}

// ListMerchantsRequest represents the request to list merchants.
type ListMerchantsRequest struct {
	Status *MerchantStatus `json:"status,omitempty"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// ListMerchantsResponse represents the response from listing merchants.
type ListMerchantsResponse struct {
	Merchants []*Merchant `json:"merchants"`
	Total     int         `json:"total"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
}
