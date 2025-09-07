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
	Title             string                   `binding:"required" json:"title"`
	Description       string                   `json:"description"`
	Items             []InvoiceItemRequest     `binding:"required,min=1" json:"items"`
	Tax               *string                  `json:"tax,omitempty"` // Fixed tax amount (deprecated, use tax_rate)
	TaxRate           string                   `json:"tax_rate"`      // Tax rate as decimal (e.g., "0.10" for 10%)
	Currency          string                   `json:"currency,omitempty"`
	CryptoCurrency    string                   `json:"crypto_currency,omitempty"`
	PriceLockDuration *int                     `json:"price_lock_duration,omitempty"`
	ExpiresIn         *int                     `json:"expires_in,omitempty"`
	PaymentTolerance  *PaymentToleranceRequest `json:"payment_tolerance,omitempty"`
	WebhookURL        *string                  `json:"webhook_url,omitempty"`
	ReturnURL         *string                  `json:"return_url,omitempty"`
	CancelURL         *string                  `json:"cancel_url,omitempty"`
	Metadata          map[string]interface{}   `json:"metadata,omitempty"`
}

// InvoiceItemRequest represents an invoice item in the request.
type InvoiceItemRequest struct {
	Name        string `binding:"required" json:"name"`
	Description string `json:"description"`
	Quantity    string `binding:"required" json:"quantity"`
	UnitPrice   string `binding:"required" json:"unit_price"`
}

// PaymentToleranceRequest represents payment tolerance settings.
type PaymentToleranceRequest struct {
	UnderpaymentThreshold string `json:"underpayment_threshold"`
	OverpaymentThreshold  string `json:"overpayment_threshold"`
	OverpaymentAction     string `json:"overpayment_action"`
}

// PaymentToleranceResponse represents payment tolerance settings in responses.
type PaymentToleranceResponse struct {
	UnderpaymentThreshold string `json:"underpayment_threshold"`
	OverpaymentThreshold  string `json:"overpayment_threshold"`
	OverpaymentAction     string `json:"overpayment_action"`
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
	// Payment tolerance settings
	PaymentTolerance *PaymentToleranceResponse `json:"payment_tolerance,omitempty"`
}

// InvoiceItemResponse represents an invoice item in the response.
type InvoiceItemResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UnitPrice   string `json:"unit_price"`
	Quantity    string `json:"quantity"`
	Total       string `json:"total"`
}

// TokenRequest represents the request payload for generating JWT tokens.
type TokenRequest struct {
	GrantType string   `binding:"required" json:"grant_type"`
	APIKey    string   `binding:"required" json:"api_key"`
	Scope     []string `binding:"required,min=1" json:"scope"`
	ExpiresIn int64    `binding:"required,min=1,max=86400" json:"expires_in"` // 1 second to 24 hours
}

// TokenResponse represents the response payload for JWT token generation.
type TokenResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int64    `json:"expires_in"`
	Scope       []string `json:"scope"`
}

// PublicInvoiceResponse represents the public invoice data for customers.
type PublicInvoiceResponse struct {
	ID              string                   `json:"id"`
	Title           string                   `json:"title"`
	Description     string                   `json:"description"`
	Items           []InvoiceItemResponse    `json:"items"`
	Subtotal        string                   `json:"subtotal"`
	TaxAmount       string                   `json:"tax_amount"`
	Total           string                   `json:"total"`
	Currency        string                   `json:"currency"`
	CryptoCurrency  string                   `json:"crypto_currency"`
	USDTAmount      string                   `json:"usdt_amount"`
	Address         string                   `json:"address"`
	Status          string                   `json:"status"`
	ExpiresAt       time.Time                `json:"expires_at"`
	CreatedAt       time.Time                `json:"created_at"`
	PaidAt          *time.Time               `json:"paid_at,omitempty"`
	Payments        []PublicPaymentResponse  `json:"payments,omitempty"`
	PaymentProgress *PaymentProgressResponse `json:"payment_progress,omitempty"`
	ReturnURL       *string                  `json:"return_url,omitempty"`
	CancelURL       *string                  `json:"cancel_url,omitempty"`
	TimeRemaining   int64                    `json:"time_remaining,omitempty"`
}

// PublicPaymentResponse represents payment data visible to customers.
type PublicPaymentResponse struct {
	Amount        string     `json:"amount"`
	Status        string     `json:"status"`
	Confirmations int        `json:"confirmations"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
}

// PaymentProgressResponse represents payment progress information.
type PaymentProgressResponse struct {
	Received  string  `json:"received"`
	Required  string  `json:"required"`
	Remaining string  `json:"remaining"`
	Percent   float64 `json:"percent"`
}

// PublicInvoiceStatusResponse represents a simple status response.
type PublicInvoiceStatusResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ListInvoicesRequest represents the request parameters for listing invoices.
type ListInvoicesRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	Limit    int    `form:"limit,default=20" binding:"min=1,max=100"`
	Status   string `form:"status"`
	Merchant string `form:"merchant"`
}

// ListInvoicesResponse represents the response for listing invoices.
type ListInvoicesResponse struct {
	Invoices []CreateInvoiceResponse `json:"invoices"`
	Total    int                     `json:"total"`
	Page     int                     `json:"page"`
	Limit    int                     `json:"limit"`
	Pages    int                     `json:"pages"`
}

// CancelInvoiceRequest represents the request payload for cancelling an invoice.
type CancelInvoiceRequest struct {
	Reason string `binding:"required" json:"reason"`
}

// CancelInvoiceResponse represents the response payload for cancelling an invoice.
type CancelInvoiceResponse struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// AnalyticsRequest represents the request parameters for analytics.
type AnalyticsRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Merchant  string `form:"merchant"`
}

// AnalyticsResponse represents the response for analytics data.
type AnalyticsResponse struct {
	Summary  AnalyticsSummary  `json:"summary"`
	Revenue  AnalyticsRevenue  `json:"revenue"`
	Invoices AnalyticsInvoices `json:"invoices"`
	Payments AnalyticsPayments `json:"payments"`
}

// AnalyticsSummary represents summary analytics data.
type AnalyticsSummary struct {
	TotalInvoices     int     `json:"total_invoices"`
	TotalRevenue      string  `json:"total_revenue"`
	TotalPayments     int     `json:"total_payments"`
	SuccessRate       float64 `json:"success_rate"`
	AverageAmount     string  `json:"average_amount"`
	PendingInvoices   int     `json:"pending_invoices"`
	CompletedInvoices int     `json:"completed_invoices"`
	CancelledInvoices int     `json:"cancelled_invoices"`
}

// AnalyticsRevenue represents revenue analytics data.
type AnalyticsRevenue struct {
	Total     string `json:"total"`
	ThisMonth string `json:"this_month"`
	LastMonth string `json:"last_month"`
	Growth    string `json:"growth"`
}

// AnalyticsInvoices represents invoice analytics data.
type AnalyticsInvoices struct {
	ByStatus map[string]int `json:"by_status"`
	ByMonth  map[string]int `json:"by_month"`
}

// AnalyticsPayments represents payment analytics data.
type AnalyticsPayments struct {
	ByStatus map[string]int `json:"by_status"`
	ByMonth  map[string]int `json:"by_month"`
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
			Name:        item.Name(),
			Description: item.Description(),
			UnitPrice:   item.UnitPrice().String(),
			Quantity:    item.Quantity().String(),
			Total:       item.TotalPrice().String(),
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

	// Get expiration time
	var expiresAt time.Time
	if exp := inv.Expiration(); exp != nil {
		expiresAt = exp.ExpiresAt()
	}

	// Get payment tolerance settings
	var paymentTolerance *PaymentToleranceResponse
	if pt := inv.PaymentTolerance(); pt != nil {
		paymentTolerance = &PaymentToleranceResponse{
			UnderpaymentThreshold: pt.UnderpaymentThreshold().StringFixed(2),
			OverpaymentThreshold:  pt.OverpaymentThreshold().StringFixed(2),
			OverpaymentAction:     pt.OverpaymentAction().String(),
		}
	}

	return CreateInvoiceResponse{
		ID:             inv.ID(),
		Items:          items,
		Subtotal:       inv.Pricing().Subtotal().String(),
		TaxAmount:      inv.Pricing().Tax().String(),
		Total:          inv.Pricing().Total().String(),
		TaxRate:        inv.Pricing().Tax().Amount().String(),
		Status:         inv.Status().String(),
		PaymentAddress: paymentAddress,
		InvoiceURL:     "/api/v1/invoices/" + inv.ID(),
		CreatedAt:      inv.CreatedAt(),
		// API.md required fields
		USDTAmount:  inv.Pricing().Total().String(), // 1:1 USD to USDT for now
		Address:     address,
		CustomerURL: customerURL,
		ExpiresAt:   expiresAt,
		// Payment tolerance settings
		PaymentTolerance: paymentTolerance,
	}
}
