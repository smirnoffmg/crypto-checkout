package merchant

import "errors"

// Domain errors for merchant operations
var (
	// Merchant creation errors
	ErrInvalidMerchantID       = errors.New("invalid merchant ID")
	ErrInvalidBusinessName     = errors.New("invalid business name")
	ErrInvalidContactEmail     = errors.New("invalid contact email")
	ErrInvalidMerchantSettings = errors.New("invalid merchant settings")
	ErrMerchantAlreadyExists   = errors.New("merchant already exists")
	ErrMerchantNotFound        = errors.New("merchant not found")

	// API key errors
	ErrInvalidAPIKeyID     = errors.New("invalid API key ID")
	ErrInvalidAPIKeyType   = errors.New("invalid API key type")
	ErrInvalidPermissions  = errors.New("invalid permissions")
	ErrAPIKeyNotFound      = errors.New("API key not found")
	ErrAPIKeyAlreadyExists = errors.New("API key already exists")
	ErrAPIKeyLimitExceeded = errors.New("API key limit exceeded")
	ErrAPIKeyNotActive     = errors.New("API key is not active")
	ErrAPIKeyExpired       = errors.New("API key has expired")

	// Webhook endpoint errors
	ErrInvalidWebhookEndpointID     = errors.New("invalid webhook endpoint ID")
	ErrInvalidWebhookURL            = errors.New("invalid webhook URL")
	ErrInvalidWebhookSecret         = errors.New("invalid webhook secret")
	ErrInvalidWebhookEvents         = errors.New("invalid webhook events")
	ErrWebhookEndpointNotFound      = errors.New("webhook endpoint not found")
	ErrWebhookEndpointLimitExceeded = errors.New("webhook endpoint limit exceeded")

	// Business rule errors
	ErrMerchantNotActive           = errors.New("merchant is not active")
	ErrMerchantSuspended           = errors.New("merchant is suspended")
	ErrMerchantPendingVerification = errors.New("merchant is pending verification")
	ErrPlanLimitExceeded           = errors.New("plan limit exceeded")
	ErrInvalidStatusTransition     = errors.New("invalid status transition")

	// Validation errors
	ErrValidationFailed = errors.New("validation failed")
	ErrRequiredField    = errors.New("required field is missing")
	ErrInvalidFormat    = errors.New("invalid format")
)

// Error codes for API responses
const (
	ErrCodeInvalidMerchantID       = "INVALID_MERCHANT_ID"
	ErrCodeInvalidBusinessName     = "INVALID_BUSINESS_NAME"
	ErrCodeInvalidContactEmail     = "INVALID_CONTACT_EMAIL"
	ErrCodeInvalidMerchantSettings = "INVALID_MERCHANT_SETTINGS"
	ErrCodeMerchantAlreadyExists   = "MERCHANT_ALREADY_EXISTS"
	ErrCodeMerchantNotFound        = "MERCHANT_NOT_FOUND"

	ErrCodeInvalidAPIKeyID     = "INVALID_API_KEY_ID"   //nolint:gosec // This is an error code constant, not a credential
	ErrCodeInvalidAPIKeyType   = "INVALID_API_KEY_TYPE" //nolint:gosec // This is an error code constant, not a credential
	ErrCodeInvalidPermissions  = "INVALID_PERMISSIONS"
	ErrCodeAPIKeyNotFound      = "API_KEY_NOT_FOUND"      //nolint:gosec // This is an error code constant, not a credential
	ErrCodeAPIKeyAlreadyExists = "API_KEY_ALREADY_EXISTS" //nolint:gosec // This is an error code constant, not a credential
	ErrCodeAPIKeyLimitExceeded = "API_KEY_LIMIT_EXCEEDED" //nolint:gosec // This is an error code constant, not a credential
	ErrCodeAPIKeyNotActive     = "API_KEY_NOT_ACTIVE"     //nolint:gosec // This is an error code constant, not a credential
	ErrCodeAPIKeyExpired       = "API_KEY_EXPIRED"        //nolint:gosec // This is an error code constant, not a credential

	ErrCodeInvalidWebhookEndpointID     = "INVALID_WEBHOOK_ENDPOINT_ID"
	ErrCodeInvalidWebhookURL            = "INVALID_WEBHOOK_URL"
	ErrCodeInvalidWebhookSecret         = "INVALID_WEBHOOK_SECRET" //nolint:gosec // This is an error code constant, not a credential
	ErrCodeInvalidWebhookEvents         = "INVALID_WEBHOOK_EVENTS"
	ErrCodeWebhookEndpointNotFound      = "WEBHOOK_ENDPOINT_NOT_FOUND"
	ErrCodeWebhookEndpointLimitExceeded = "WEBHOOK_ENDPOINT_LIMIT_EXCEEDED"

	ErrCodeMerchantNotActive           = "MERCHANT_NOT_ACTIVE"
	ErrCodeMerchantSuspended           = "MERCHANT_SUSPENDED"
	ErrCodeMerchantPendingVerification = "MERCHANT_PENDING_VERIFICATION"
	ErrCodePlanLimitExceeded           = "PLAN_LIMIT_EXCEEDED"
	ErrCodeInvalidStatusTransition     = "INVALID_STATUS_TRANSITION"

	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeRequiredField    = "REQUIRED_FIELD"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
)
