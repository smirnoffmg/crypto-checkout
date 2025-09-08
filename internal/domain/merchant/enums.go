package merchant

// MerchantStatus represents the current status of a merchant account.
type MerchantStatus string

const (
	StatusActive              MerchantStatus = "active"
	StatusSuspended           MerchantStatus = "suspended"
	StatusPendingVerification MerchantStatus = "pending_verification"
	StatusClosed              MerchantStatus = "closed"
)

// KeyType represents the environment designation for API keys.
type KeyType string

const (
	KeyTypeLive KeyType = "live"
	KeyTypeTest KeyType = "test"
)

// KeyStatus represents the current status of an API key.
type KeyStatus string

const (
	KeyStatusActive  KeyStatus = "active"
	KeyStatusRevoked KeyStatus = "revoked"
	KeyStatusExpired KeyStatus = "expired"
)

// EndpointStatus represents the delivery status of a webhook endpoint.
type EndpointStatus string

const (
	EndpointStatusActive   EndpointStatus = "active"
	EndpointStatusDisabled EndpointStatus = "disabled"
	EndpointStatusFailed   EndpointStatus = "failed"
)

// BackoffStrategy represents the retry timing strategy for webhook delivery.
type BackoffStrategy string

const (
	BackoffStrategyLinear      BackoffStrategy = "linear"
	BackoffStrategyExponential BackoffStrategy = "exponential"
)

// IsValid validates if the merchant status is valid.
func (s MerchantStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusSuspended, StatusPendingVerification, StatusClosed:
		return true
	default:
		return false
	}
}

// IsValid validates if the key type is valid.
func (k KeyType) IsValid() bool {
	switch k {
	case KeyTypeLive, KeyTypeTest:
		return true
	default:
		return false
	}
}

// IsValid validates if the key status is valid.
func (k KeyStatus) IsValid() bool {
	switch k {
	case KeyStatusActive, KeyStatusRevoked, KeyStatusExpired:
		return true
	default:
		return false
	}
}

// IsValid validates if the endpoint status is valid.
func (e EndpointStatus) IsValid() bool {
	switch e {
	case EndpointStatusActive, EndpointStatusDisabled, EndpointStatusFailed:
		return true
	default:
		return false
	}
}

// IsValid validates if the backoff strategy is valid.
func (b BackoffStrategy) IsValid() bool {
	switch b {
	case BackoffStrategyLinear, BackoffStrategyExponential:
		return true
	default:
		return false
	}
}
