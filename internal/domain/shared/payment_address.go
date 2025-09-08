package shared

import (
	"errors"
	"time"
)

// PaymentAddress represents a blockchain payment address.
type PaymentAddress struct {
	address     string
	network     BlockchainNetwork
	generatedAt time.Time
	expiresAt   *time.Time
}

// NewPaymentAddress creates a new PaymentAddress.
func NewPaymentAddress(address string, network BlockchainNetwork) (*PaymentAddress, error) {
	if address == "" {
		return nil, errors.New("address cannot be empty")
	}

	if !network.IsValid() {
		return nil, errors.New("invalid blockchain network")
	}

	// Basic address format validation
	if len(address) < 10 {
		return nil, errors.New("address format is too short")
	}

	return &PaymentAddress{
		address:     address,
		network:     network,
		generatedAt: time.Now().UTC(),
	}, nil
}

// NewPaymentAddressWithExpiry creates a new PaymentAddress with expiration.
func NewPaymentAddressWithExpiry(
	address string,
	network BlockchainNetwork,
	expiresAt time.Time,
) (*PaymentAddress, error) {
	addr, err := NewPaymentAddress(address, network)
	if err != nil {
		return nil, err
	}

	if expiresAt.Before(time.Now().UTC()) {
		return nil, errors.New("expiration time must be in the future")
	}

	addr.expiresAt = &expiresAt
	return addr, nil
}

// Address returns the blockchain address.
func (pa *PaymentAddress) Address() string {
	return pa.address
}

// Network returns the blockchain network.
func (pa *PaymentAddress) Network() BlockchainNetwork {
	return pa.network
}

// GeneratedAt returns when the address was generated.
func (pa *PaymentAddress) GeneratedAt() time.Time {
	return pa.generatedAt
}

// ExpiresAt returns the expiration time if set.
func (pa *PaymentAddress) ExpiresAt() *time.Time {
	return pa.expiresAt
}

// IsExpired returns true if the address has expired.
func (pa *PaymentAddress) IsExpired() bool {
	if pa.expiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*pa.expiresAt)
}

// String returns the string representation of the address.
func (pa *PaymentAddress) String() string {
	return pa.address
}

// Equals returns true if this address equals the other address.
func (pa *PaymentAddress) Equals(other *PaymentAddress) bool {
	if other == nil {
		return false
	}
	return pa.address == other.address && pa.network == other.network
}
