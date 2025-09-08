package merchant

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// Merchant represents the main merchant aggregate root.
type Merchant struct {
	id           string
	businessName string
	contactEmail string
	status       MerchantStatus
	settings     *MerchantSettings
	createdAt    time.Time
	updatedAt    time.Time
}

// MerchantValidation represents the validation structure for Merchant creation.
type MerchantValidation struct {
	ID           string            `validate:"required,min=1"         json:"id"`
	BusinessName string            `validate:"required,min=2,max=255" json:"business_name"`
	ContactEmail string            `validate:"required,email"         json:"contact_email"`
	Status       MerchantStatus    `validate:"required"               json:"status"`
	Settings     *MerchantSettings `validate:"required"               json:"settings"`
}

// NewMerchant creates a new Merchant with validation.
func NewMerchant(
	id, businessName, contactEmail string,
	settings *MerchantSettings,
) (*Merchant, error) {
	if id == "" {
		return nil, errors.New("merchant ID is required")
	}
	if businessName == "" {
		return nil, errors.New("business name is required")
	}
	if contactEmail == "" {
		return nil, errors.New("contact email is required")
	}
	if settings == nil {
		return nil, errors.New("merchant settings are required")
	}

	now := time.Now()
	merchant := &Merchant{
		id:           id,
		businessName: businessName,
		contactEmail: contactEmail,
		status:       StatusPendingVerification,
		settings:     settings,
		createdAt:    now,
		updatedAt:    now,
	}

	// Validate using go-playground/validator
	validation := MerchantValidation{
		ID:           merchant.id,
		BusinessName: merchant.businessName,
		ContactEmail: merchant.contactEmail,
		Status:       merchant.status,
		Settings:     merchant.settings,
	}

	validate := validator.New()
	if err := validate.Struct(validation); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return merchant, nil
}

// ID returns the merchant ID.
func (m *Merchant) ID() string {
	return m.id
}

// BusinessName returns the business name.
func (m *Merchant) BusinessName() string {
	return m.businessName
}

// ContactEmail returns the contact email.
func (m *Merchant) ContactEmail() string {
	return m.contactEmail
}

// Status returns the merchant status.
func (m *Merchant) Status() MerchantStatus {
	return m.status
}

// Settings returns the merchant settings.
func (m *Merchant) Settings() *MerchantSettings {
	return m.settings
}

// CreatedAt returns the creation timestamp.
func (m *Merchant) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns the last update timestamp.
func (m *Merchant) UpdatedAt() time.Time {
	return m.updatedAt
}

// UpdateBusinessName updates the business name.
func (m *Merchant) UpdateBusinessName(name string) error {
	if name == "" {
		return errors.New("business name cannot be empty")
	}
	if len(name) < 2 || len(name) > 255 {
		return errors.New("business name must be between 2 and 255 characters")
	}

	m.businessName = name
	m.updatedAt = time.Now()
	return nil
}

// UpdateContactEmail updates the contact email.
func (m *Merchant) UpdateContactEmail(email string) error {
	if email == "" {
		return errors.New("contact email cannot be empty")
	}

	// Basic email validation
	if !isValidEmail(email) {
		return errors.New("invalid email format")
	}

	m.contactEmail = email
	m.updatedAt = time.Now()
	return nil
}

// UpdateSettings updates the merchant settings.
func (m *Merchant) UpdateSettings(settings *MerchantSettings) error {
	if settings == nil {
		return errors.New("settings cannot be nil")
	}

	m.settings = settings
	m.updatedAt = time.Now()
	return nil
}

// ChangeStatus changes the merchant status.
func (m *Merchant) ChangeStatus(newStatus MerchantStatus) error {
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	// Business rule: cannot change to active without verification
	if newStatus == StatusActive && m.status != StatusPendingVerification {
		return errors.New("merchant must be in pending verification status to be activated")
	}

	m.status = newStatus
	m.updatedAt = time.Now()
	return nil
}

// IsActive checks if the merchant is active.
func (m *Merchant) IsActive() bool {
	return m.status == StatusActive
}

// CanCreateAPIKey checks if the merchant can create a new API key.
func (m *Merchant) CanCreateAPIKey(currentCount int) error {
	if !m.IsActive() {
		return errors.New("merchant must be active to create API keys")
	}

	// For now, allow unlimited API keys - this can be configured via settings later
	// TODO: Implement API key limits via merchant settings
	return nil
}

// CanCreateWebhookEndpoint checks if the merchant can create a new webhook endpoint.
func (m *Merchant) CanCreateWebhookEndpoint(currentCount int) error {
	if !m.IsActive() {
		return errors.New("merchant must be active to create webhook endpoints")
	}

	// For now, allow unlimited webhook endpoints - this can be configured via settings later
	// TODO: Implement webhook endpoint limits via merchant settings
	return nil
}

// isValidEmail performs basic email validation.
func isValidEmail(email string) bool {
	// Simple email validation - in production, use a proper email validation library
	return email != "" &&
		len(email) <= 254 &&
		contains(email, "@") &&
		contains(email, ".")
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
