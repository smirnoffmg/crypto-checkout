// Package api provides HTTP handlers and DTOs for the crypto-checkout API.
package web

import (
	"crypto-checkout/internal/domain/invoice"
	"time"
)

const (
	// TaxRateDecimals is the number of decimal places for tax rate display.
	TaxRateDecimals = 2
)

// CreateInvoiceRequest represents the request payload for creating an invoice.
type CreateInvoiceRequest struct {
	Items   []InvoiceItemRequest `binding:"required,min=1" json:"items"`
	TaxRate string               `binding:"required"       json:"tax_rate"`
}

// InvoiceItemRequest represents an invoice item in the request.
type InvoiceItemRequest struct {
	Description string `binding:"required" json:"description"`
	UnitPrice   string `binding:"required" json:"unit_price"`
	Quantity    string `binding:"required" json:"quantity"`
}

// CreateInvoiceResponse represents the response payload for creating an invoice.
type CreateInvoiceResponse struct {
	ID             string                `json:"id"`
	Items          []InvoiceItemResponse `json:"items"`
	Subtotal       string                `json:"subtotal"`
	TaxAmount      string                `json:"tax_amount"`
	Total          string                `json:"total"`
	TaxRate        string                `json:"tax_rate"`
	Status         string                `json:"status"`
	PaymentAddress *string               `json:"payment_address,omitempty"`
	InvoiceURL     string                `json:"invoice_url"`
	CreatedAt      time.Time             `json:"created_at"`
	// API.md required fields
	USDTAmount  string    `json:"usdt_amount"`
	Address     string    `json:"address"`
	CustomerURL string    `json:"customer_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// InvoiceItemResponse represents an invoice item in the response.
type InvoiceItemResponse struct {
	Description string `json:"description"`
	UnitPrice   string `json:"unit_price"`
	Quantity    string `json:"quantity"`
	Total       string `json:"total"`
}

// ErrorResponse represents an error response payload.
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Message   string                 `json:"message,omitempty"`
	Code      string                 `json:"code,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// ToCreateInvoiceResponse converts a domain invoice to a create invoice response.
func ToCreateInvoiceResponse(inv *invoice.Invoice) CreateInvoiceResponse {
	items := make([]InvoiceItemResponse, len(inv.Items()))
	for i, item := range inv.Items() {
		items[i] = InvoiceItemResponse{
			Description: item.Description(),
			UnitPrice:   item.UnitPrice().String(),
			Quantity:    item.Quantity().String(),
			Total:       item.CalculateTotal().String(),
		}
	}

	var paymentAddress *string
	if addr := inv.PaymentAddress(); addr != nil {
		addrStr := addr.String()
		paymentAddress = &addrStr
	}

	// Generate payment address if not set (for API.md compliance)
	address := "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN" // TODO: Generate real crypto address
	if paymentAddress != nil {
		address = *paymentAddress
	}

	// Construct customer URL
	customerURL := "https://checkout.crypto-checkout.com/invoice/" + inv.ID()

	return CreateInvoiceResponse{
		ID:             inv.ID(),
		Items:          items,
		Subtotal:       inv.CalculateSubtotal().String(),
		TaxAmount:      inv.CalculateTax().String(),
		Total:          inv.CalculateTotal().String(),
		TaxRate:        inv.TaxRate().StringFixed(TaxRateDecimals),
		Status:         inv.Status().String(),
		PaymentAddress: paymentAddress,
		InvoiceURL:     "/api/v1/invoices/" + inv.ID(),
		CreatedAt:      inv.CreatedAt(),
		// API.md required fields
		USDTAmount:  inv.CalculateTotal().String(), // 1:1 USD to USDT for now
		Address:     address,
		CustomerURL: customerURL,
		ExpiresAt:   inv.ExpiresAt(),
	}
}
