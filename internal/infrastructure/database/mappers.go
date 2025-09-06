package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/shared"
)

// InvoiceMapper handles conversion between domain entities and database models.
type InvoiceMapper struct{}

// NewInvoiceMapper creates a new invoice mapper.
func NewInvoiceMapper() *InvoiceMapper {
	return &InvoiceMapper{}
}

// ToDomain converts a database model to a domain entity.
func (m *InvoiceMapper) ToDomain(model *InvoiceModel) (*invoice.Invoice, error) {
	if model == nil {
		return nil, errors.New("invoice model cannot be nil")
	}

	// Parse items from JSONB
	var items []*invoice.InvoiceItem
	if model.Items != "" {
		var itemData []map[string]interface{}
		if err := json.Unmarshal([]byte(model.Items), &itemData); err != nil {
			return nil, fmt.Errorf("failed to parse items JSON: %w", err)
		}

		items = make([]*invoice.InvoiceItem, len(itemData))
		for i, itemMap := range itemData {
			name, _ := itemMap["name"].(string)
			description, _ := itemMap["description"].(string)
			quantity, _ := itemMap["quantity"].(string)
			unitPriceStr, _ := itemMap["unit_price"].(string)

			unitPrice, err := shared.NewMoney(unitPriceStr, shared.CurrencyUSD)
			if err != nil {
				return nil, fmt.Errorf("failed to create unit price: %w", err)
			}

			invoiceItem, err := invoice.NewInvoiceItem(name, description, quantity, unitPrice)
			if err != nil {
				return nil, fmt.Errorf("failed to create invoice item: %w", err)
			}

			items[i] = invoiceItem
		}
	}

	// Create pricing
	subtotal, err := shared.NewMoney(model.Subtotal, shared.CurrencyUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to create subtotal: %w", err)
	}

	tax, err := shared.NewMoney(model.Tax, shared.CurrencyUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to create tax: %w", err)
	}

	total, err := shared.NewMoney(model.Total, shared.CurrencyUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to create total: %w", err)
	}

	pricing, err := invoice.NewInvoicePricing(subtotal, tax, total)
	if err != nil {
		return nil, fmt.Errorf("failed to create pricing: %w", err)
	}

	// Create payment address if present
	var paymentAddress *shared.PaymentAddress
	if model.PaymentAddress != nil {
		paymentAddress, err = shared.NewPaymentAddress(*model.PaymentAddress, shared.NetworkTron)
		if err != nil {
			return nil, fmt.Errorf("failed to create payment address: %w", err)
		}
	}

	// Create exchange rate (simplified - in real implementation, parse from JSONB)
	exchangeRate, err := shared.NewExchangeRate("1.0", shared.CurrencyUSD, shared.CryptoCurrency(model.CryptoCurrency), "default_provider", 30*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to create exchange rate: %w", err)
	}

	// Create payment tolerance (simplified - in real implementation, parse from JSONB)
	paymentTolerance, err := invoice.NewPaymentTolerance("0.01", "1.0", invoice.OverpaymentActionCredit)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment tolerance: %w", err)
	}

	// Create expiration
	var expiration *invoice.InvoiceExpiration
	if model.ExpiresAt != nil {
		expiration, err = invoice.NewInvoiceExpirationWithTime(*model.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to create expiration: %w", err)
		}
	} else {
		expiration = invoice.NewInvoiceExpiration(30 * time.Minute)
	}

	// Create invoice with correct constructor signature
	inv, err := invoice.NewInvoice(
		model.ID,
		model.MerchantID,
		model.Title,
		model.Description,
		items,
		pricing,
		shared.CryptoCurrency(model.CryptoCurrency),
		paymentAddress,
		exchangeRate,
		paymentTolerance,
		expiration,
		nil, // metadata
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Set customer ID if present
	if model.CustomerID != nil {
		inv.SetCustomerID(*model.CustomerID)
	}

	// Set status from database
	status := invoice.InvoiceStatus(model.Status)
	inv.SetStatus(status)

	// Set paid at if present
	if model.PaidAt != nil {
		// Note: This would require a method to set paidAt, which might not exist
		// For now, we'll skip this as the domain model handles it internally
	}

	return inv, nil
}

// ToModel converts a domain entity to a database model.
func (m *InvoiceMapper) ToModel(inv *invoice.Invoice) *InvoiceModel {
	if inv == nil {
		return nil
	}

	// Convert items to JSONB
	var itemsJSON string
	if len(inv.Items()) > 0 {
		itemData := make([]map[string]interface{}, len(inv.Items()))
		for i, item := range inv.Items() {
			itemData[i] = map[string]interface{}{
				"name":        item.Name(),
				"description": item.Description(),
				"quantity":    item.Quantity().String(),
				"unit_price":  item.UnitPrice().Amount().String(),
			}
		}
		if jsonBytes, err := json.Marshal(itemData); err == nil {
			itemsJSON = string(jsonBytes)
		}
	}

	// Get crypto amount
	cryptoAmount := "0"
	if cryptoAmountMoney, err := inv.GetCryptoAmount(); err == nil {
		cryptoAmount = cryptoAmountMoney.Amount().String()
	}

	model := &InvoiceModel{
		ID:             inv.ID(),
		MerchantID:     inv.MerchantID(),
		CustomerID:     inv.CustomerID(), // This is already *string
		Title:          inv.Title(),
		Description:    inv.Description(),
		Items:          itemsJSON,
		Subtotal:       inv.Pricing().Subtotal().Amount().String(),
		Tax:            inv.Pricing().Tax().Amount().String(),
		Total:          inv.Pricing().Total().Amount().String(),
		Currency:       inv.Pricing().Subtotal().Currency(),
		CryptoCurrency: inv.CryptoCurrency().String(),
		CryptoAmount:   cryptoAmount,
		Status:         inv.Status().String(),
		CreatedAt:      inv.CreatedAt(),
		UpdatedAt:      inv.UpdatedAt(),
		PaidAt:         inv.PaidAt(),
	}

	// Set payment address if present
	if inv.PaymentAddress() != nil {
		address := inv.PaymentAddress().String()
		model.PaymentAddress = &address
	}

	// Set expiration if present
	if inv.Expiration() != nil {
		expiresAt := inv.Expiration().ExpiresAt()
		model.ExpiresAt = &expiresAt
	}

	// TODO: Serialize exchange rate and payment tolerance to JSONB
	// For now, using placeholder values
	model.ExchangeRate = `{"rate": "1.0", "from": "USD", "to": "USDT", "source": "default", "expires_at": "2025-01-01T00:00:00Z"}`
	model.PaymentTolerance = `{"underpayment_threshold": "0.01", "overpayment_threshold": "1.0", "overpayment_action": "credit"}`

	return model
}

// ToDomainSlice converts multiple database models to domain entities.
func (m *InvoiceMapper) ToDomainSlice(models []InvoiceModel) ([]*invoice.Invoice, error) {
	invoices := make([]*invoice.Invoice, len(models))
	for i, model := range models {
		inv, err := m.ToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model %d: %w", i, err)
		}
		invoices[i] = inv
	}
	return invoices, nil
}
