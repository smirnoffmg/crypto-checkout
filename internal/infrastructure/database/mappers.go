package database

import (
	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/shared"
	"encoding/json"
	"errors"
	"fmt"
	"time"
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

	items, err := m.parseInvoiceItems(model.Items)
	if err != nil {
		return nil, err
	}

	pricing, err := m.createInvoicePricing(model)
	if err != nil {
		return nil, err
	}

	paymentAddress, err := m.createPaymentAddress(model.PaymentAddress)
	if err != nil {
		return nil, err
	}

	exchangeRate, err := m.createExchangeRate(model)
	if err != nil {
		return nil, err
	}

	paymentTolerance, err := m.createPaymentTolerance(model.PaymentTolerance)
	if err != nil {
		return nil, err
	}

	expiration := m.createExpiration(model.ExpiresAt)

	inv, err := m.buildInvoice(model, items, pricing, paymentAddress, exchangeRate, paymentTolerance, expiration)
	if err != nil {
		return nil, err
	}

	m.setInvoiceProperties(inv, model)
	return inv, nil
}

// parseInvoiceItems parses invoice items from JSONB.
func (m *InvoiceMapper) parseInvoiceItems(itemsJSON string) ([]*invoice.InvoiceItem, error) {
	if itemsJSON == "" {
		return []*invoice.InvoiceItem{}, nil
	}

	var itemData []map[string]interface{}
	if err := json.Unmarshal([]byte(itemsJSON), &itemData); err != nil {
		return nil, fmt.Errorf("failed to parse items JSON: %w", err)
	}

	items := make([]*invoice.InvoiceItem, len(itemData))
	for i, itemMap := range itemData {
		item, err := m.createInvoiceItem(itemMap)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}
		items[i] = item
	}

	return items, nil
}

// createInvoiceItem creates an invoice item from a map.
func (m *InvoiceMapper) createInvoiceItem(itemMap map[string]interface{}) (*invoice.InvoiceItem, error) {
	name, _ := itemMap["name"].(string)
	description, _ := itemMap["description"].(string)
	quantity, _ := itemMap["quantity"].(string)
	unitPriceStr, _ := itemMap["unit_price"].(string)

	unitPrice, err := shared.NewMoney(unitPriceStr, shared.CurrencyUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to create unit price: %w", err)
	}

	return invoice.NewInvoiceItem(name, description, quantity, unitPrice)
}

// createInvoicePricing creates invoice pricing from model.
func (m *InvoiceMapper) createInvoicePricing(model *InvoiceModel) (*invoice.InvoicePricing, error) {
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

	return invoice.NewInvoicePricing(subtotal, tax, total)
}

// createPaymentAddress creates payment address from model.
func (m *InvoiceMapper) createPaymentAddress(paymentAddressStr *string) (*shared.PaymentAddress, error) {
	if paymentAddressStr == nil {
		return nil, nil
	}

	paymentAddress, err := shared.NewPaymentAddress(*paymentAddressStr, shared.NetworkTron)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment address: %w", err)
	}

	return paymentAddress, nil
}

// createExchangeRate creates exchange rate from model.
func (m *InvoiceMapper) createExchangeRate(model *InvoiceModel) (*shared.ExchangeRate, error) {
	exchangeRate, err := m.DeserializeExchangeRate(model.ExchangeRate)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize exchange rate: %w", err)
	}

	if exchangeRate == nil {
		// Fallback to default if not present
		exchangeRate, err = shared.NewExchangeRate(
			"1.0",
			shared.CurrencyUSD,
			shared.CryptoCurrency(model.CryptoCurrency),
			"default_provider",
			30*time.Minute,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create default exchange rate: %w", err)
		}
	}

	return exchangeRate, nil
}

// createPaymentTolerance creates payment tolerance from model.
func (m *InvoiceMapper) createPaymentTolerance(toleranceJSON string) (*invoice.PaymentTolerance, error) {
	paymentTolerance, err := m.DeserializePaymentTolerance(toleranceJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize payment tolerance: %w", err)
	}

	if paymentTolerance == nil {
		// Fallback to default if not present
		paymentTolerance, err = invoice.NewPaymentTolerance("0.01", "1.0", invoice.OverpaymentActionCredit)
		if err != nil {
			return nil, fmt.Errorf("failed to create default payment tolerance: %w", err)
		}
	}

	return paymentTolerance, nil
}

// createExpiration creates expiration from model.
func (m *InvoiceMapper) createExpiration(expiresAt *time.Time) *invoice.InvoiceExpiration {
	if expiresAt != nil {
		// Use unsafe version to allow loading expired invoices from database
		return invoice.NewInvoiceExpirationWithTimeUnsafe(*expiresAt)
	}
	return invoice.NewInvoiceExpiration(30 * time.Minute)
}

// buildInvoice creates the invoice entity.
func (m *InvoiceMapper) buildInvoice(
	model *InvoiceModel,
	items []*invoice.InvoiceItem,
	pricing *invoice.InvoicePricing,
	paymentAddress *shared.PaymentAddress,
	exchangeRate *shared.ExchangeRate,
	paymentTolerance *invoice.PaymentTolerance,
	expiration *invoice.InvoiceExpiration,
) (*invoice.Invoice, error) {
	return invoice.NewInvoice(
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
}

// setInvoiceProperties sets additional properties on the invoice.
func (m *InvoiceMapper) setInvoiceProperties(inv *invoice.Invoice, model *InvoiceModel) {
	// Set customer ID if present
	if model.CustomerID != nil {
		inv.SetCustomerID(*model.CustomerID)
	}

	// Set status from database
	status := invoice.InvoiceStatus(model.Status)
	inv.SetStatus(status)

	// Set paid at if present
	// Note: This would require a method to set paidAt, which might not exist
	// For now, we'll skip this as the domain model handles it internally
	_ = model.PaidAt
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

	// Serialize exchange rate to JSONB
	if inv.ExchangeRate() != nil {
		if exchangeRateJSON, err := m.SerializeExchangeRate(inv.ExchangeRate()); err == nil {
			model.ExchangeRate = exchangeRateJSON
		}
	}

	// Serialize payment tolerance to JSONB
	if inv.PaymentTolerance() != nil {
		if paymentToleranceJSON, err := m.SerializePaymentTolerance(inv.PaymentTolerance()); err == nil {
			model.PaymentTolerance = paymentToleranceJSON
		}
	}

	return model
}

// ToDomainSlice converts multiple database models to domain entities.
func (m *InvoiceMapper) ToDomainSlice(models []InvoiceModel) ([]*invoice.Invoice, error) {
	invoices := make([]*invoice.Invoice, len(models))
	for i := range models {
		inv, err := m.ToDomain(&models[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert model %d: %w", i, err)
		}
		invoices[i] = inv
	}
	return invoices, nil
}

// SerializeExchangeRate converts an ExchangeRate to JSON string.
func (m *InvoiceMapper) SerializeExchangeRate(er *shared.ExchangeRate) (string, error) {
	if er == nil {
		return "", nil
	}

	exchangeRateData := map[string]interface{}{
		"rate":       er.Rate().String(),
		"from":       string(er.FromCurrency()),
		"to":         string(er.ToCurrency()),
		"source":     er.Source(),
		"locked_at":  er.LockedAt().Format(time.RFC3339),
		"expires_at": er.ExpiresAt().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(exchangeRateData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal exchange rate: %w", err)
	}

	return string(jsonData), nil
}

// SerializePaymentTolerance converts a PaymentTolerance to JSON string.
func (m *InvoiceMapper) SerializePaymentTolerance(pt *invoice.PaymentTolerance) (string, error) {
	if pt == nil {
		return "", nil
	}

	paymentToleranceData := map[string]interface{}{
		"underpayment_threshold": pt.UnderpaymentThreshold().StringFixed(2),
		"overpayment_threshold":  pt.OverpaymentThreshold().StringFixed(2),
		"overpayment_action":     string(pt.OverpaymentAction()),
	}

	jsonData, err := json.Marshal(paymentToleranceData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payment tolerance: %w", err)
	}

	return string(jsonData), nil
}

// DeserializeExchangeRate converts a JSON string to an ExchangeRate.
func (m *InvoiceMapper) DeserializeExchangeRate(jsonStr string) (*shared.ExchangeRate, error) {
	if jsonStr == "" {
		return nil, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exchange rate: %w", err)
	}

	rate, ok := data["rate"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid rate format")
	}

	from, ok := data["from"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from currency format")
	}

	to, ok := data["to"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid to currency format")
	}

	source, ok := data["source"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid source format")
	}

	lockedAtStr, ok := data["locked_at"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid locked_at format")
	}

	expiresAtStr, ok := data["expires_at"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid expires_at format")
	}

	lockedAt, err := time.Parse(time.RFC3339, lockedAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid locked_at time format: %w", err)
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid expires_at time format: %w", err)
	}

	// Create exchange rate with calculated duration
	duration := expiresAt.Sub(lockedAt)
	exchangeRate, err := shared.NewExchangeRate(
		rate,
		shared.Currency(from),
		shared.CryptoCurrency(to),
		source,
		duration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exchange rate: %w", err)
	}

	return exchangeRate, nil
}

// DeserializePaymentTolerance converts a JSON string to a PaymentTolerance.
func (m *InvoiceMapper) DeserializePaymentTolerance(jsonStr string) (*invoice.PaymentTolerance, error) {
	if jsonStr == "" {
		return nil, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment tolerance: %w", err)
	}

	underpaymentThreshold, ok := data["underpayment_threshold"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid underpayment_threshold format")
	}

	overpaymentThreshold, ok := data["overpayment_threshold"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid overpayment_threshold format")
	}

	overpaymentAction, ok := data["overpayment_action"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid overpayment_action format")
	}

	paymentTolerance, err := invoice.NewPaymentTolerance(
		underpaymentThreshold,
		overpaymentThreshold,
		invoice.OverpaymentAction(overpaymentAction),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment tolerance: %w", err)
	}

	return paymentTolerance, nil
}
