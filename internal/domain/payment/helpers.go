package payment

import (
	"time"

	"crypto-checkout/internal/domain/shared"
)

// CalculateRequiredConfirmations calculates the required confirmations based on amount and network.
func CalculateRequiredConfirmations(amount *shared.Money, network shared.BlockchainNetwork) int {
	// Basic confirmation requirements based on amount and network
	// This is a simplified implementation - in reality, this would be more complex

	switch network {
	case shared.NetworkBitcoin:
		// Bitcoin: 1 confirmation for small amounts, 6 for large amounts
		threshold, _ := shared.NewMoney("1000.00", shared.CurrencyUSD)
		if amount.Amount().LessThan(threshold.Amount()) {
			return 1
		}
		return 6

	case shared.NetworkEthereum:
		// Ethereum: 12 confirmations for small amounts, 30 for large amounts
		threshold, _ := shared.NewMoney("10000.00", shared.CurrencyUSD)
		if amount.Amount().LessThan(threshold.Amount()) {
			return 12
		}
		return 30

	case shared.NetworkTron:
		// Tron: 1 confirmation for small amounts, 3 for large amounts
		threshold, _ := shared.NewMoney("5000.00", shared.CurrencyUSD)
		if amount.Amount().LessThan(threshold.Amount()) {
			return 1
		}
		return 3

	default:
		return 1
	}
}

// IsPaymentExpired checks if a payment has expired based on detection time and network.
func IsPaymentExpired(payment *Payment, maxAge time.Duration) bool {
	if payment == nil {
		return false
	}

	detectedAt := payment.DetectedAt()
	expiresAt := detectedAt.Add(maxAge)

	return time.Now().UTC().After(expiresAt)
}

// CalculateNetworkFee estimates the network fee based on network and transaction size.
func CalculateNetworkFee(network shared.BlockchainNetwork, estimatedSize int) (*shared.Money, shared.CryptoCurrency, error) {
	// This is a simplified implementation - in reality, this would query current network conditions

	switch network {
	case shared.NetworkBitcoin:
		// Bitcoin: ~$5-50 depending on network congestion
		fee, err := shared.NewMoney("10.00", shared.CurrencyUSD)
		if err != nil {
			return nil, "", err
		}
		return fee, shared.CryptoCurrencyBTC, nil

	case shared.NetworkEthereum:
		// Ethereum: ~$2-100 depending on gas price
		fee, err := shared.NewMoney("20.00", shared.CurrencyUSD)
		if err != nil {
			return nil, "", err
		}
		return fee, shared.CryptoCurrencyETH, nil

	case shared.NetworkTron:
		// Tron: very low fees, ~$0.01
		fee, err := shared.NewMoney("0.01", shared.CurrencyUSD)
		if err != nil {
			return nil, "", err
		}
		return fee, shared.CryptoCurrencyUSDT, nil

	default:
		return nil, "", shared.ErrInvalidNetwork
	}
}

// ValidatePaymentAddress validates a payment address against the network.
func ValidatePaymentAddress(address string, network shared.BlockchainNetwork) error {
	if address == "" {
		return shared.ErrInvalidPaymentAddress
	}

	// Basic validation - in reality, this would use network-specific validation
	switch network {
	case shared.NetworkBitcoin:
		// Bitcoin addresses start with 1, 3, or bc1
		if len(address) < 26 || len(address) > 62 {
			return shared.ErrInvalidPaymentAddress
		}

	case shared.NetworkEthereum:
		// Ethereum addresses are 40 hex characters (20 bytes)
		if len(address) != 42 || address[:2] != "0x" {
			return shared.ErrInvalidPaymentAddress
		}

	case shared.NetworkTron:
		// Tron addresses start with T
		if len(address) != 34 || address[0] != 'T' {
			return shared.ErrInvalidPaymentAddress
		}

	default:
		return shared.ErrInvalidNetwork
	}

	return nil
}

// EstimateConfirmationTime estimates the time to confirmation based on network.
func EstimateConfirmationTime(network shared.BlockchainNetwork, confirmations int) time.Duration {
	// This is a simplified implementation - in reality, this would consider current network conditions

	switch network {
	case shared.NetworkBitcoin:
		// Bitcoin: ~10 minutes per block
		return time.Duration(confirmations) * 10 * time.Minute

	case shared.NetworkEthereum:
		// Ethereum: ~13 seconds per block
		return time.Duration(confirmations) * 13 * time.Second

	case shared.NetworkTron:
		// Tron: ~3 seconds per block
		return time.Duration(confirmations) * 3 * time.Second

	default:
		return time.Duration(confirmations) * 10 * time.Minute
	}
}

// IsHighValuePayment checks if a payment is considered high value.
func IsHighValuePayment(amount *shared.Money) bool {
	if amount == nil {
		return false
	}

	// Consider payments over $10,000 as high value
	threshold, err := shared.NewMoney("10000.00", shared.CurrencyUSD)
	if err != nil {
		return false
	}

	return amount.Amount().GreaterThanOrEqual(threshold.Amount())
}

// GetPaymentRiskLevel determines the risk level of a payment.
func GetPaymentRiskLevel(payment *Payment) string {
	if payment == nil {
		return "unknown"
	}

	// Simple risk assessment based on amount and network
	if IsHighValuePayment(payment.Amount().Amount()) {
		return "high"
	}

	// Additional risk factors could be considered here
	// - Network reputation
	// - Address history
	// - Time of day
	// - Geographic factors

	return "low"
}
