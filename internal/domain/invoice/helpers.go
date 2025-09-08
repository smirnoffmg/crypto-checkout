package invoice

import (
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/domain/shared"
	"time"
)

// InvoiceHelpers provides utility functions for invoice operations.

// CalculateRequiredConfirmations calculates the required number of confirmations based on the payment amount.
func CalculateRequiredConfirmations(amount *shared.Money) (*shared.ConfirmationCount, error) {
	if amount == nil {
		return nil, ErrInvalidAmount
	}

	// Simple confirmation logic based on amount
	// In a real implementation, this would be more sophisticated
	amountValue := amount.Amount()

	threshold100, _ := shared.NewMoney("100", shared.CurrencyUSD)
	threshold1000, _ := shared.NewMoney("1000", shared.CurrencyUSD)
	threshold10000, _ := shared.NewMoney("10000", shared.CurrencyUSD)

	var requiredConfirmations int
	switch {
	case amountValue.LessThan(threshold100.Amount()):
		requiredConfirmations = 1
	case amountValue.LessThan(threshold1000.Amount()):
		requiredConfirmations = 3
	case amountValue.LessThan(threshold10000.Amount()):
		requiredConfirmations = 6
	default:
		requiredConfirmations = 12
	}

	return shared.NewConfirmationCount(requiredConfirmations)
}

// CalculateInvoiceExpiration calculates the expiration time for an invoice.
func CalculateInvoiceExpiration(duration time.Duration) time.Time {
	return time.Now().UTC().Add(duration)
}

// IsInvoiceExpired checks if an invoice has expired.
func IsInvoiceExpired(invoice *Invoice) bool {
	if invoice == nil {
		return false
	}
	return invoice.Expiration().IsExpired()
}

// GetInvoiceTimeRemaining returns the time remaining until invoice expiration.
func GetInvoiceTimeRemaining(invoice *Invoice) time.Duration {
	if invoice == nil {
		return 0
	}
	return invoice.Expiration().TimeRemaining()
}

// FormatInvoiceAmount formats an invoice amount for display.
func FormatInvoiceAmount(amount *shared.Money) string {
	if amount == nil {
		return "0.00"
	}
	return amount.String()
}

// GetInvoiceDisplayStatus returns a user-friendly status message.
func GetInvoiceDisplayStatus(status InvoiceStatus) string {
	switch status {
	case StatusCreated:
		return "Processing..."
	case StatusPending:
		return "Pending Payment"
	case StatusPartial:
		return "Partial Payment Received"
	case StatusConfirming:
		return "Confirming Payment..."
	case StatusPaid:
		return "Payment Successful"
	case StatusExpired:
		return "Expired"
	case StatusCancelled:
		return "Cancelled"
	case StatusRefunded:
		return "Refunded"
	default:
		return "Unknown"
	}
}

// GetInvoiceStatusColor returns a color code for the invoice status.
func GetInvoiceStatusColor(status InvoiceStatus) string {
	switch status {
	case StatusCreated, StatusPending:
		return "#3B82F6" // Blue
	case StatusPartial, StatusConfirming:
		return "#F59E0B" // Yellow
	case StatusPaid:
		return "#10B981" // Green
	case StatusExpired, StatusCancelled, StatusRefunded:
		return "#EF4444" // Red
	default:
		return "#6B7280" // Gray
	}
}

// ValidateInvoiceMetadata validates invoice metadata.
func ValidateInvoiceMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		return nil
	}

	// Check for reserved keys
	reservedKeys := map[string]bool{
		"id":          true,
		"merchant_id": true,
		"customer_id": true,
		"status":      true,
		"created_at":  true,
		"updated_at":  true,
		"paid_at":     true,
		"expires_at":  true,
	}

	for key := range metadata {
		if reservedKeys[key] {
			return ErrInvalidInvoiceID
		}
	}

	return nil
}

// SanitizeInvoiceMetadata sanitizes invoice metadata by removing invalid entries.
func SanitizeInvoiceMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return make(map[string]interface{})
	}

	sanitized := make(map[string]interface{})

	// Reserved keys that should not be in metadata
	reservedKeys := map[string]bool{
		"id":          true,
		"merchant_id": true,
		"customer_id": true,
		"status":      true,
		"created_at":  true,
		"updated_at":  true,
		"paid_at":     true,
		"expires_at":  true,
	}

	for key, value := range metadata {
		if !reservedKeys[key] {
			sanitized[key] = value
		}
	}

	return sanitized
}

// GetInvoiceSummary returns a summary of the invoice for display.
func GetInvoiceSummary(invoice *Invoice) *InvoiceSummary {
	if invoice == nil {
		return nil
	}

	return &InvoiceSummary{
		ID:             invoice.ID(),
		Title:          invoice.Title(),
		Status:         invoice.Status(),
		StatusDisplay:  GetInvoiceDisplayStatus(invoice.Status()),
		StatusColor:    GetInvoiceStatusColor(invoice.Status()),
		Total:          FormatInvoiceAmount(invoice.Pricing().Total()),
		Currency:       invoice.Pricing().Total().Currency(),
		CryptoAmount:   getCryptoAmountString(invoice),
		CryptoCurrency: invoice.CryptoCurrency().String(),
		ExpiresAt:      invoice.Expiration().ExpiresAt(),
		TimeRemaining:  GetInvoiceTimeRemaining(invoice),
		IsExpired:      invoice.Expiration().IsExpired(),
		IsActive:       invoice.Status().IsActive(),
		CreatedAt:      invoice.CreatedAt(),
		UpdatedAt:      invoice.UpdatedAt(),
		PaidAt:         invoice.PaidAt(),
	}
}

// InvoiceSummary represents a summary of an invoice for display purposes.
type InvoiceSummary struct {
	ID             string
	Title          string
	Status         InvoiceStatus
	StatusDisplay  string
	StatusColor    string
	Total          string
	Currency       string
	CryptoAmount   string
	CryptoCurrency string
	ExpiresAt      time.Time
	TimeRemaining  time.Duration
	IsExpired      bool
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	PaidAt         *time.Time
}

// getCryptoAmountString returns the cryptocurrency amount as a string.
func getCryptoAmountString(invoice *Invoice) string {
	cryptoAmount, err := invoice.GetCryptoAmount()
	if err != nil {
		return "0.00"
	}
	return cryptoAmount.String()
}

// GetPaymentProgress returns the payment progress for an invoice.
func GetPaymentProgress(invoice *Invoice, payments []*payment.Payment) *PaymentProgress {
	if invoice == nil {
		return nil
	}

	requiredAmount, err := invoice.GetCryptoAmount()
	if err != nil {
		return &PaymentProgress{
			Required:  "0.00",
			Received:  "0.00",
			Remaining: "0.00",
			Percent:   0,
		}
	}

	var totalReceived string
	if len(payments) == 0 {
		totalReceived = "0.00"
	} else {
		// Calculate total received from confirmed payments
		total, _ := shared.NewMoneyWithCrypto("0.00", invoice.CryptoCurrency())
		for _, payment := range payments {
			if payment.IsConfirmed() {
				total, _ = total.Add(payment.Amount().Amount())
			}
		}
		totalReceived = total.String()
	}

	// Calculate remaining amount
	receivedMoney, _ := shared.NewMoneyWithCrypto(totalReceived, invoice.CryptoCurrency())
	remainingMoney, _ := requiredAmount.Add(receivedMoney)

	// Calculate percentage
	requiredValue := requiredAmount.Amount()
	receivedValue := receivedMoney.Amount()
	var percent float64
	if !requiredValue.IsZero() {
		percent = receivedValue.Div(requiredValue).InexactFloat64() * 100
	}

	return &PaymentProgress{
		Required:  requiredAmount.String(),
		Received:  totalReceived,
		Remaining: remainingMoney.String(),
		Percent:   percent,
	}
}

// PaymentProgress represents the payment progress for an invoice.
type PaymentProgress struct {
	Required  string
	Received  string
	Remaining string
	Percent   float64
}

// GetInvoiceQRData returns the QR code data for an invoice.
func GetInvoiceQRData(invoice *Invoice) string {
	if invoice == nil {
		return ""
	}

	cryptoAmount, err := invoice.GetCryptoAmount()
	if err != nil {
		return ""
	}

	// Generate QR code data based on cryptocurrency
	switch invoice.CryptoCurrency() {
	case shared.CryptoCurrencyUSDT:
		return generateUSDTQRData(invoice.PaymentAddress().Address(), cryptoAmount.String())
	case shared.CryptoCurrencyBTC:
		return generateBTCQRData(invoice.PaymentAddress().Address(), cryptoAmount.String())
	case shared.CryptoCurrencyETH:
		return generateETHQRData(invoice.PaymentAddress().Address(), cryptoAmount.String())
	default:
		return ""
	}
}

// generateUSDTQRData generates QR code data for USDT payments.
func generateUSDTQRData(address, amount string) string {
	// Tron USDT QR format: tron:address?amount=amount
	return "tron:" + address + "?amount=" + amount
}

// generateBTCQRData generates QR code data for BTC payments.
func generateBTCQRData(address, amount string) string {
	// Bitcoin QR format: bitcoin:address?amount=amount
	return "bitcoin:" + address + "?amount=" + amount
}

// generateETHQRData generates QR code data for ETH payments.
func generateETHQRData(address, amount string) string {
	// Ethereum QR format: ethereum:address?amount=amount
	return "ethereum:" + address + "?amount=" + amount
}
