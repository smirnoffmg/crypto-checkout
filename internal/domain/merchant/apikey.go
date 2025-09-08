package merchant

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// APIKey represents an API key entity within the Merchant aggregate.
type APIKey struct {
	id          string
	merchantID  string
	keyHash     *APIKeyHash
	keyType     KeyType
	permissions []string
	status      KeyStatus
	name        string
	lastUsedAt  *time.Time
	expiresAt   *time.Time
	createdAt   time.Time
}

// APIKeyValidation represents the validation structure for APIKey creation.
type APIKeyValidation struct {
	ID          string    `validate:"required,min=1"    json:"id"`
	MerchantID  string    `validate:"required,min=1"    json:"merchant_id"`
	KeyType     KeyType   `validate:"required"          json:"key_type"`
	Permissions []string  `validate:"required,min=1"    json:"permissions"`
	Status      KeyStatus `validate:"required"          json:"status"`
	Name        string    `validate:"omitempty,max=100" json:"name"`
}

// NewAPIKey creates a new APIKey with validation.
func NewAPIKey(
	id, merchantID, rawKey string,
	keyType KeyType,
	permissions []string,
	name string,
	expiresAt *time.Time,
) (*APIKey, error) {
	if id == "" {
		return nil, errors.New("API key ID is required")
	}
	if merchantID == "" {
		return nil, errors.New("merchant ID is required")
	}
	if rawKey == "" {
		return nil, errors.New("API key value is required")
	}
	if !keyType.IsValid() {
		return nil, fmt.Errorf("invalid key type: %s", keyType)
	}
	if len(permissions) == 0 {
		return nil, errors.New("at least one permission is required")
	}
	if len(name) > 100 {
		return nil, errors.New("name cannot exceed 100 characters")
	}

	keyHash, err := NewAPIKeyHash(rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	now := time.Now()
	apiKey := &APIKey{
		id:          id,
		merchantID:  merchantID,
		keyHash:     keyHash,
		keyType:     keyType,
		permissions: permissions,
		status:      KeyStatusActive,
		name:        name,
		expiresAt:   expiresAt,
		createdAt:   now,
	}

	// Validate using go-playground/validator
	validation := APIKeyValidation{
		ID:          apiKey.id,
		MerchantID:  apiKey.merchantID,
		KeyType:     apiKey.keyType,
		Permissions: apiKey.permissions,
		Status:      apiKey.status,
		Name:        apiKey.name,
	}

	validate := validator.New()
	if err := validate.Struct(validation); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return apiKey, nil
}

// NewAPIKeyFromHash creates a new APIKey from existing hash (for repository use).
func NewAPIKeyFromHash(
	id, merchantID string,
	keyHash *APIKeyHash,
	keyType KeyType,
	permissions []string,
	status KeyStatus,
	name string,
	lastUsedAt, expiresAt *time.Time,
) (*APIKey, error) {
	if id == "" {
		return nil, errors.New("API key ID is required")
	}
	if merchantID == "" {
		return nil, errors.New("merchant ID is required")
	}
	if keyHash == nil {
		return nil, errors.New("key hash is required")
	}
	if !keyType.IsValid() {
		return nil, fmt.Errorf("invalid key type: %s", keyType)
	}
	if len(permissions) == 0 {
		return nil, errors.New("at least one permission is required")
	}
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", status)
	}
	if len(name) > 100 {
		return nil, errors.New("name cannot exceed 100 characters")
	}

	now := time.Now()
	apiKey := &APIKey{
		id:          id,
		merchantID:  merchantID,
		keyHash:     keyHash,
		keyType:     keyType,
		permissions: permissions,
		status:      status,
		name:        name,
		lastUsedAt:  lastUsedAt,
		expiresAt:   expiresAt,
		createdAt:   now,
	}

	// Validate using go-playground/validator
	validation := APIKeyValidation{
		ID:          apiKey.id,
		MerchantID:  apiKey.merchantID,
		KeyType:     apiKey.keyType,
		Permissions: apiKey.permissions,
		Status:      apiKey.status,
		Name:        apiKey.name,
	}

	validate := validator.New()
	if err := validate.Struct(validation); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return apiKey, nil
}

// ID returns the API key ID.
func (k *APIKey) ID() string {
	return k.id
}

// MerchantID returns the merchant ID.
func (k *APIKey) MerchantID() string {
	return k.merchantID
}

// KeyHash returns the key hash.
func (k *APIKey) KeyHash() *APIKeyHash {
	return k.keyHash
}

// KeyType returns the key type.
func (k *APIKey) KeyType() KeyType {
	return k.keyType
}

// Permissions returns the permissions.
func (k *APIKey) Permissions() []string {
	return k.permissions
}

// Status returns the key status.
func (k *APIKey) Status() KeyStatus {
	return k.status
}

// Name returns the key name.
func (k *APIKey) Name() string {
	return k.name
}

// LastUsedAt returns the last used timestamp.
func (k *APIKey) LastUsedAt() *time.Time {
	return k.lastUsedAt
}

// ExpiresAt returns the expiration timestamp.
func (k *APIKey) ExpiresAt() *time.Time {
	return k.expiresAt
}

// CreatedAt returns the creation timestamp.
func (k *APIKey) CreatedAt() time.Time {
	return k.createdAt
}

// UpdateName updates the API key name.
func (k *APIKey) UpdateName(name string) error {
	if len(name) > 100 {
		return errors.New("name cannot exceed 100 characters")
	}

	k.name = name
	return nil
}

// UpdatePermissions updates the API key permissions.
func (k *APIKey) UpdatePermissions(permissions []string) error {
	if len(permissions) == 0 {
		return errors.New("at least one permission is required")
	}

	k.permissions = permissions
	return nil
}

// Revoke revokes the API key.
func (k *APIKey) Revoke() error {
	if k.status == KeyStatusRevoked {
		return errors.New("API key is already revoked")
	}

	k.status = KeyStatusRevoked
	return nil
}

// Activate activates the API key.
func (k *APIKey) Activate() error {
	if k.status == KeyStatusActive {
		return errors.New("API key is already active")
	}

	if k.status == KeyStatusExpired {
		return errors.New("cannot activate expired API key")
	}

	k.status = KeyStatusActive
	return nil
}

// MarkAsUsed marks the API key as used at the current time.
func (k *APIKey) MarkAsUsed() {
	now := time.Now()
	k.lastUsedAt = &now
}

// IsActive checks if the API key is active.
func (k *APIKey) IsActive() bool {
	if k.status != KeyStatusActive {
		return false
	}

	// Check if expired
	if k.expiresAt != nil && time.Now().After(*k.expiresAt) {
		k.status = KeyStatusExpired
		return false
	}

	return true
}

// IsExpired checks if the API key is expired.
func (k *APIKey) IsExpired() bool {
	if k.expiresAt == nil {
		return false
	}
	return time.Now().After(*k.expiresAt)
}

// HasPermission checks if the API key has a specific permission.
func (k *APIKey) HasPermission(permission string) bool {
	for _, p := range k.permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// ValidatePermission checks if the API key has the required permission.
func (k *APIKey) ValidatePermission(requiredPermission string) error {
	if !k.IsActive() {
		return errors.New("API key is not active")
	}

	if !k.HasPermission(requiredPermission) {
		return fmt.Errorf("API key does not have required permission: %s", requiredPermission)
	}

	return nil
}

// GetKeyPrefix returns a masked version of the key for display purposes.
func (k *APIKey) GetKeyPrefix() string {
	// In a real implementation, this would return the first few characters
	// of the original key, not the hash
	return fmt.Sprintf("sk_%s_***", k.keyType)
}
