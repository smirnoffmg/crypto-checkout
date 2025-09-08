package merchant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// APIKeyServiceImpl implements the APIKeyService interface.
type APIKeyServiceImpl struct {
	apiKeyRepo APIKeyRepository
	logger     *zap.Logger
}

// NewAPIKeyService creates a new API key service.
func NewAPIKeyService(apiKeyRepo APIKeyRepository, logger *zap.Logger) APIKeyService {
	return &APIKeyServiceImpl{
		apiKeyRepo: apiKeyRepo,
		logger:     logger,
	}
}

// CreateAPIKey creates a new API key for a merchant.
func (s *APIKeyServiceImpl) CreateAPIKey(
	ctx context.Context,
	req *CreateAPIKeyRequest,
) (*CreateAPIKeyResponse, error) {
	if req == nil {
		return nil, errors.New("create API key request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate API key ID
	apiKeyID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key ID: %w", err)
	}

	// Generate raw API key
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Parse expires at
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format: %w", err)
		}
		expiresAt = &parsedTime
	}

	// Create API key
	apiKey, err := NewAPIKey(
		apiKeyID,
		req.MerchantID,
		rawKey,
		req.KeyType,
		req.Permissions,
		req.Name,
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Save to repository
	if err := s.apiKeyRepo.Save(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to save API key: %w", err)
	}

	s.logger.Info("API key created successfully",
		zap.String("api_key_id", apiKey.ID()),
		zap.String("merchant_id", apiKey.MerchantID()),
		zap.String("key_type", string(apiKey.KeyType())),
	)

	return &CreateAPIKeyResponse{
		APIKey: apiKey,
		RawKey: rawKey,
	}, nil
}

// GetAPIKey retrieves an API key by ID.
func (s *APIKeyServiceImpl) GetAPIKey(
	ctx context.Context,
	req *GetAPIKeyRequest,
) (*GetAPIKeyResponse, error) {
	if req == nil {
		return nil, errors.New("get API key request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Find API key
	apiKey, err := s.apiKeyRepo.FindByID(ctx, req.APIKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find API key: %w", err)
	}

	if apiKey == nil {
		return nil, errors.New("API key not found")
	}

	return &GetAPIKeyResponse{
		APIKey: apiKey,
	}, nil
}

// ListAPIKeys lists API keys for a merchant.
func (s *APIKeyServiceImpl) ListAPIKeys(
	ctx context.Context,
	req *ListAPIKeysRequest,
) (*ListAPIKeysResponse, error) {
	if req == nil {
		return nil, errors.New("list API keys request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Set default pagination
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Find API keys
	apiKeys, err := s.apiKeyRepo.FindByMerchantID(ctx, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find API keys: %w", err)
	}

	// Apply filters
	var filteredKeys []*APIKey
	for _, key := range apiKeys {
		// Filter by status
		if req.Status != nil && key.Status() != *req.Status {
			continue
		}
		// Filter by key type
		if req.KeyType != nil && key.KeyType() != *req.KeyType {
			continue
		}
		filteredKeys = append(filteredKeys, key)
	}

	// Apply pagination
	total := len(filteredKeys)
	start := req.Offset
	end := start + req.Limit
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	var paginatedKeys []*APIKey
	if start < end {
		paginatedKeys = filteredKeys[start:end]
	}

	return &ListAPIKeysResponse{
		APIKeys: paginatedKeys,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}

// UpdateAPIKey updates an existing API key.
func (s *APIKeyServiceImpl) UpdateAPIKey(
	ctx context.Context,
	req *UpdateAPIKeyRequest,
) (*UpdateAPIKeyResponse, error) {
	if req == nil {
		return nil, errors.New("update API key request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Find existing API key
	apiKey, err := s.apiKeyRepo.FindByID(ctx, req.APIKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find API key: %w", err)
	}

	if apiKey == nil {
		return nil, errors.New("API key not found")
	}

	// Update fields
	if req.Name != nil {
		if err := apiKey.UpdateName(*req.Name); err != nil {
			return nil, fmt.Errorf("failed to update API key name: %w", err)
		}
	}
	if req.Permissions != nil {
		if err := apiKey.UpdatePermissions(req.Permissions); err != nil {
			return nil, fmt.Errorf("failed to update API key permissions: %w", err)
		}
	}

	// Save updated API key
	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to update API key: %w", err)
	}

	s.logger.Info("API key updated successfully",
		zap.String("api_key_id", apiKey.ID()),
		zap.String("merchant_id", apiKey.MerchantID()),
	)

	return &UpdateAPIKeyResponse{
		APIKey: apiKey,
	}, nil
}

// RevokeAPIKey revokes an API key.
func (s *APIKeyServiceImpl) RevokeAPIKey(
	ctx context.Context,
	req *RevokeAPIKeyRequest,
) (*RevokeAPIKeyResponse, error) {
	if req == nil {
		return nil, errors.New("revoke API key request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Find existing API key
	apiKey, err := s.apiKeyRepo.FindByID(ctx, req.APIKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find API key: %w", err)
	}

	if apiKey == nil {
		return nil, errors.New("API key not found")
	}

	// Revoke the API key
	if err := apiKey.Revoke(); err != nil {
		return nil, fmt.Errorf("failed to revoke API key: %w", err)
	}

	// Save updated API key
	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to revoke API key: %w", err)
	}

	s.logger.Info("API key revoked successfully",
		zap.String("api_key_id", apiKey.ID()),
		zap.String("merchant_id", apiKey.MerchantID()),
		zap.String("reason", req.Reason),
	)

	return &RevokeAPIKeyResponse{
		APIKey: apiKey,
	}, nil
}

// ValidateAPIKey validates an API key and returns merchant information.
func (s *APIKeyServiceImpl) ValidateAPIKey(
	ctx context.Context,
	req *ValidateAPIKeyRequest,
) (*ValidateAPIKeyResponse, error) {
	if req == nil {
		return nil, errors.New("validate API key request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create hash from raw key
	hash, err := NewAPIKeyHash(req.RawKey)
	if err != nil {
		return &ValidateAPIKeyResponse{Valid: false}, nil
	}

	// Find API key by hash
	apiKey, err := s.apiKeyRepo.FindByHash(ctx, hash)
	if err != nil {
		return &ValidateAPIKeyResponse{Valid: false}, nil
	}

	if apiKey == nil {
		return &ValidateAPIKeyResponse{Valid: false}, nil
	}

	// Check if API key is active
	if apiKey.Status() != KeyStatusActive {
		return &ValidateAPIKeyResponse{Valid: false}, nil
	}

	// Check if API key is expired
	if apiKey.ExpiresAt() != nil && apiKey.ExpiresAt().Before(time.Now()) {
		return &ValidateAPIKeyResponse{Valid: false}, nil
	}

	return &ValidateAPIKeyResponse{
		Valid:  true,
		APIKey: apiKey,
	}, nil
}

// generateAPIKey generates a new API key.
func generateAPIKey() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to hex and add prefix
	key := "sk_" + hex.EncodeToString(bytes)
	return key, nil
}
