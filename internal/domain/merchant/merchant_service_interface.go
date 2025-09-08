package merchant

import (
	"context"
)

// MerchantService defines the interface for merchant business operations.
type MerchantService interface {
	// CreateMerchant creates a new merchant account.
	CreateMerchant(ctx context.Context, req *CreateMerchantRequest) (*CreateMerchantResponse, error)

	// GetMerchant retrieves a merchant by ID.
	GetMerchant(ctx context.Context, req *GetMerchantRequest) (*GetMerchantResponse, error)

	// UpdateMerchant updates an existing merchant.
	UpdateMerchant(ctx context.Context, req *UpdateMerchantRequest) (*UpdateMerchantResponse, error)

	// ChangeMerchantStatus changes the status of a merchant.
	ChangeMerchantStatus(ctx context.Context, req *ChangeMerchantStatusRequest) (*ChangeMerchantStatusResponse, error)

	// ListMerchants lists merchants with filtering and pagination.
	ListMerchants(ctx context.Context, req *ListMerchantsRequest) (*ListMerchantsResponse, error)
}

// APIKeyService defines the interface for API key business operations.
type APIKeyService interface {
	// CreateAPIKey creates a new API key for a merchant.
	CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error)

	// GetAPIKey retrieves an API key by ID.
	GetAPIKey(ctx context.Context, req *GetAPIKeyRequest) (*GetAPIKeyResponse, error)

	// ListAPIKeys lists API keys for a merchant.
	ListAPIKeys(ctx context.Context, req *ListAPIKeysRequest) (*ListAPIKeysResponse, error)

	// UpdateAPIKey updates an existing API key.
	UpdateAPIKey(ctx context.Context, req *UpdateAPIKeyRequest) (*UpdateAPIKeyResponse, error)

	// RevokeAPIKey revokes an API key.
	RevokeAPIKey(ctx context.Context, req *RevokeAPIKeyRequest) (*RevokeAPIKeyResponse, error)

	// ValidateAPIKey validates an API key and returns merchant information.
	ValidateAPIKey(ctx context.Context, req *ValidateAPIKeyRequest) (*ValidateAPIKeyResponse, error)
}

// WebhookEndpointService defines the interface for webhook endpoint business operations.
type WebhookEndpointService interface {
	// CreateWebhookEndpoint creates a new webhook endpoint for a merchant.
	CreateWebhookEndpoint(
		ctx context.Context,
		req *CreateWebhookEndpointRequest,
	) (*CreateWebhookEndpointResponse, error)

	// GetWebhookEndpoint retrieves a webhook endpoint by ID.
	GetWebhookEndpoint(ctx context.Context, req *GetWebhookEndpointRequest) (*GetWebhookEndpointResponse, error)

	// ListWebhookEndpoints lists webhook endpoints for a merchant.
	ListWebhookEndpoints(ctx context.Context, req *ListWebhookEndpointsRequest) (*ListWebhookEndpointsResponse, error)

	// UpdateWebhookEndpoint updates an existing webhook endpoint.
	UpdateWebhookEndpoint(
		ctx context.Context,
		req *UpdateWebhookEndpointRequest,
	) (*UpdateWebhookEndpointResponse, error)

	// DeleteWebhookEndpoint deletes a webhook endpoint.
	DeleteWebhookEndpoint(
		ctx context.Context,
		req *DeleteWebhookEndpointRequest,
	) (*DeleteWebhookEndpointResponse, error)

	// TestWebhookEndpoint tests a webhook endpoint.
	TestWebhookEndpoint(ctx context.Context, req *TestWebhookEndpointRequest) (*TestWebhookEndpointResponse, error)
}

// Request/Response DTOs for Merchant operations

// CreateMerchantRequest represents the request to create a merchant.
type CreateMerchantRequest struct {
	BusinessName string            `json:"business_name" validate:"required,min=2,max=255"`
	ContactEmail string            `json:"contact_email" validate:"required,email"`
	Settings     *MerchantSettings `json:"settings"      validate:"required"`
}

// CreateMerchantResponse represents the response from creating a merchant.
type CreateMerchantResponse struct {
	Merchant *Merchant `json:"merchant"`
}

// GetMerchantRequest represents the request to get a merchant.
type GetMerchantRequest struct {
	MerchantID string `json:"merchant_id" validate:"required"`
}

// GetMerchantResponse represents the response from getting a merchant.
type GetMerchantResponse struct {
	Merchant *Merchant `json:"merchant"`
}

// UpdateMerchantRequest represents the request to update a merchant.
type UpdateMerchantRequest struct {
	MerchantID   string            `json:"merchant_id"             validate:"required"`
	BusinessName *string           `json:"business_name,omitempty" validate:"omitempty,min=2,max=255"`
	ContactEmail *string           `json:"contact_email,omitempty" validate:"omitempty,email"`
	Settings     *MerchantSettings `json:"settings,omitempty"`
}

// UpdateMerchantResponse represents the response from updating a merchant.
type UpdateMerchantResponse struct {
	Merchant *Merchant `json:"merchant"`
}

// ChangeMerchantStatusRequest represents the request to change merchant status.
type ChangeMerchantStatusRequest struct {
	MerchantID string         `json:"merchant_id"      validate:"required"`
	Status     MerchantStatus `json:"status"           validate:"required"`
	Reason     string         `json:"reason,omitempty"`
}

// ChangeMerchantStatusResponse represents the response from changing merchant status.
type ChangeMerchantStatusResponse struct {
	Merchant *Merchant `json:"merchant"`
}

// Request/Response DTOs for API Key operations

// CreateAPIKeyRequest represents the request to create an API key.
type CreateAPIKeyRequest struct {
	MerchantID  string   `json:"merchant_id"          validate:"required"`
	KeyType     KeyType  `json:"key_type"             validate:"required"`
	Permissions []string `json:"permissions"          validate:"required,min=1"`
	Name        string   `json:"name,omitempty"       validate:"omitempty,max=100"`
	ExpiresAt   *string  `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse represents the response from creating an API key.
type CreateAPIKeyResponse struct {
	APIKey *APIKey `json:"api_key"`
	RawKey string  `json:"raw_key"` // Only returned once during creation
}

// GetAPIKeyRequest represents the request to get an API key.
type GetAPIKeyRequest struct {
	APIKeyID string `json:"api_key_id" validate:"required"`
}

// GetAPIKeyResponse represents the response from getting an API key.
type GetAPIKeyResponse struct {
	APIKey *APIKey `json:"api_key"`
}

// ListAPIKeysRequest represents the request to list API keys.
type ListAPIKeysRequest struct {
	MerchantID string     `json:"merchant_id"        validate:"required"`
	Status     *KeyStatus `json:"status,omitempty"`
	KeyType    *KeyType   `json:"key_type,omitempty"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

// ListAPIKeysResponse represents the response from listing API keys.
type ListAPIKeysResponse struct {
	APIKeys []*APIKey `json:"api_keys"`
	Total   int       `json:"total"`
	Limit   int       `json:"limit"`
	Offset  int       `json:"offset"`
}

// UpdateAPIKeyRequest represents the request to update an API key.
type UpdateAPIKeyRequest struct {
	APIKeyID    string   `json:"api_key_id"            validate:"required"`
	Name        *string  `json:"name,omitempty"        validate:"omitempty,max=100"`
	Permissions []string `json:"permissions,omitempty" validate:"omitempty,min=1"`
}

// UpdateAPIKeyResponse represents the response from updating an API key.
type UpdateAPIKeyResponse struct {
	APIKey *APIKey `json:"api_key"`
}

// RevokeAPIKeyRequest represents the request to revoke an API key.
type RevokeAPIKeyRequest struct {
	APIKeyID string `json:"api_key_id"       validate:"required"`
	Reason   string `json:"reason,omitempty"`
}

// RevokeAPIKeyResponse represents the response from revoking an API key.
type RevokeAPIKeyResponse struct {
	APIKey *APIKey `json:"api_key"`
}

// ValidateAPIKeyRequest represents the request to validate an API key.
type ValidateAPIKeyRequest struct {
	RawKey string `json:"raw_key" validate:"required"`
}

// ValidateAPIKeyResponse represents the response from validating an API key.
type ValidateAPIKeyResponse struct {
	Valid    bool      `json:"valid"`
	APIKey   *APIKey   `json:"api_key,omitempty"`
	Merchant *Merchant `json:"merchant,omitempty"`
}

// Request/Response DTOs for Webhook Endpoint operations

// CreateWebhookEndpointRequest represents the request to create a webhook endpoint.
type CreateWebhookEndpointRequest struct {
	MerchantID   string            `json:"merchant_id"           validate:"required"`
	URL          string            `json:"url"                   validate:"required,url"`
	Events       []string          `json:"events"                validate:"required,min=1"`
	Secret       string            `json:"secret"                validate:"required,min=32"`
	MaxRetries   int               `json:"max_retries"           validate:"min=0,max=10"`
	RetryBackoff string            `json:"retry_backoff"         validate:"required"`
	Timeout      int               `json:"timeout"               validate:"min=5,max=60"`
	AllowedIPs   []string          `json:"allowed_ips,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
}

// CreateWebhookEndpointResponse represents the response from creating a webhook endpoint.
type CreateWebhookEndpointResponse struct {
	Endpoint *WebhookEndpoint `json:"endpoint"`
}

// GetWebhookEndpointRequest represents the request to get a webhook endpoint.
type GetWebhookEndpointRequest struct {
	EndpointID string `json:"endpoint_id" validate:"required"`
}

// GetWebhookEndpointResponse represents the response from getting a webhook endpoint.
type GetWebhookEndpointResponse struct {
	Endpoint *WebhookEndpoint `json:"endpoint"`
}

// ListWebhookEndpointsRequest represents the request to list webhook endpoints.
type ListWebhookEndpointsRequest struct {
	MerchantID string          `json:"merchant_id"      validate:"required"`
	Status     *EndpointStatus `json:"status,omitempty"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
}

// ListWebhookEndpointsResponse represents the response from listing webhook endpoints.
type ListWebhookEndpointsResponse struct {
	Endpoints []*WebhookEndpoint `json:"endpoints"`
	Total     int                `json:"total"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
}

// UpdateWebhookEndpointRequest represents the request to update a webhook endpoint.
type UpdateWebhookEndpointRequest struct {
	EndpointID   string            `json:"endpoint_id"             validate:"required"`
	URL          *string           `json:"url,omitempty"           validate:"omitempty,url"`
	Events       []string          `json:"events,omitempty"        validate:"omitempty,min=1"`
	Secret       *string           `json:"secret,omitempty"        validate:"omitempty,min=32"`
	MaxRetries   *int              `json:"max_retries,omitempty"   validate:"omitempty,min=0,max=10"`
	RetryBackoff *string           `json:"retry_backoff,omitempty"`
	Timeout      *int              `json:"timeout,omitempty"       validate:"omitempty,min=5,max=60"`
	AllowedIPs   []string          `json:"allowed_ips,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
}

// UpdateWebhookEndpointResponse represents the response from updating a webhook endpoint.
type UpdateWebhookEndpointResponse struct {
	Endpoint *WebhookEndpoint `json:"endpoint"`
}

// DeleteWebhookEndpointRequest represents the request to delete a webhook endpoint.
type DeleteWebhookEndpointRequest struct {
	EndpointID string `json:"endpoint_id" validate:"required"`
}

// DeleteWebhookEndpointResponse represents the response from deleting a webhook endpoint.
type DeleteWebhookEndpointResponse struct {
	Success bool `json:"success"`
}

// TestWebhookEndpointRequest represents the request to test a webhook endpoint.
type TestWebhookEndpointRequest struct {
	EndpointID string `json:"endpoint_id" validate:"required"`
}

// TestWebhookEndpointResponse represents the response from testing a webhook endpoint.
type TestWebhookEndpointResponse struct {
	Success      bool   `json:"success"`
	ResponseCode int    `json:"response_code"`
	ResponseTime int    `json:"response_time_ms"`
	Error        string `json:"error,omitempty"`
}
