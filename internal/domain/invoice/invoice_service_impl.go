package invoice

import (
	"context"
	"errors"
	"strings"
	"time"

	"crypto-checkout/internal/domain/shared"

	"github.com/shopspring/decimal"
)

// InvoiceServiceImpl implements the InvoiceService interface.
type InvoiceServiceImpl struct {
	repository Repository
}

// NewInvoiceService creates a new InvoiceService implementation.
func NewInvoiceService(repository Repository) InvoiceService {
	return &InvoiceServiceImpl{
		repository: repository,
	}
}

// CreateInvoice creates a new invoice with the given parameters.
func (s *InvoiceServiceImpl) CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*Invoice, error) {
	if req == nil {
		return nil, errors.New("create invoice request cannot be nil")
	}

	if req.MerchantID == "" {
		return nil, errors.New("merchant ID is required")
	}

	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	if len(req.Items) == 0 {
		return nil, errors.New("at least one item is required")
	}

	if !req.CryptoCurrency.IsValid() {
		return nil, errors.New("invalid cryptocurrency")
	}

	// Create invoice items
	items := make([]*InvoiceItem, 0, len(req.Items))
	subtotal := decimal.Zero

	for _, itemReq := range req.Items {
		item, err := NewInvoiceItem(
			itemReq.Name,
			itemReq.Description,
			itemReq.Quantity,
			itemReq.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		subtotal = subtotal.Add(item.TotalPrice().Amount())
	}

	// Create subtotal money
	subtotalMoney, err := shared.NewMoney(subtotal.String(), req.Currency)
	if err != nil {
		return nil, err
	}

	// Handle tax
	var taxMoney *shared.Money
	if req.Tax != nil {
		taxMoney = req.Tax
	} else {
		taxMoney, err = shared.NewMoney("0.00", req.Currency)
		if err != nil {
			return nil, err
		}
	}

	// Calculate total
	totalMoney, err := subtotalMoney.Add(taxMoney)
	if err != nil {
		return nil, err
	}

	// Create pricing
	pricing, err := NewInvoicePricing(subtotalMoney, taxMoney, totalMoney)
	if err != nil {
		return nil, err
	}

	// Get exchange rate (this would typically come from an exchange rate service)
	exchangeRate, err := s.getExchangeRate(ctx, req.Currency, req.CryptoCurrency)
	if err != nil {
		return nil, err
	}

	// Generate payment address (this would typically come from a payment address service)
	paymentAddress, err := s.generatePaymentAddress(ctx, req.CryptoCurrency)
	if err != nil {
		return nil, err
	}

	// Set default payment tolerance if not provided
	paymentTolerance := req.PaymentTolerance
	if paymentTolerance == nil {
		paymentTolerance = DefaultPaymentTolerance()
	}

	// Set default expiration duration if not provided
	expirationDuration := req.ExpirationDuration
	if expirationDuration == 0 {
		expirationDuration = 30 * time.Minute
	}

	// Create expiration
	expiration := NewInvoiceExpiration(expirationDuration)

	// Generate invoice ID (this would typically come from an ID generator service)
	invoiceID := s.generateInvoiceID()

	// Create invoice
	invoice, err := NewInvoice(
		invoiceID,
		req.MerchantID,
		req.Title,
		req.Description,
		items,
		pricing,
		req.CryptoCurrency,
		paymentAddress,
		exchangeRate,
		paymentTolerance,
		expiration,
		req.Metadata,
	)
	if err != nil {
		return nil, err
	}

	// Set customer ID if provided
	if req.CustomerID != nil {
		invoice.SetCustomerID(*req.CustomerID)
	}

	// Save invoice
	if err := s.repository.Save(ctx, invoice); err != nil {
		return nil, err
	}

	return invoice, nil
}

// GetInvoice retrieves an invoice by ID.
func (s *InvoiceServiceImpl) GetInvoice(ctx context.Context, id string) (*Invoice, error) {
	if id == "" {
		return nil, errors.New("invoice ID cannot be empty")
	}

	return s.repository.FindByID(ctx, id)
}

// GetInvoiceByPaymentAddress retrieves an invoice by payment address.
func (s *InvoiceServiceImpl) GetInvoiceByPaymentAddress(ctx context.Context, address *shared.PaymentAddress) (*Invoice, error) {
	if address == nil {
		return nil, errors.New("payment address cannot be nil")
	}

	return s.repository.FindByPaymentAddress(ctx, address)
}

// ListInvoices retrieves invoices with the given filters.
func (s *InvoiceServiceImpl) ListInvoices(ctx context.Context, req *ListInvoicesRequest) (*ListInvoicesResponse, error) {
	if req == nil {
		return nil, errors.New("list invoices request cannot be nil")
	}

	if req.MerchantID == "" {
		return nil, errors.New("merchant ID is required")
	}

	// Set default limit
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// This would typically be implemented in the repository with proper filtering
	// For now, we'll use a simple implementation
	var invoices []*Invoice
	var err error

	if req.Status != nil {
		invoices, err = s.repository.FindByStatus(ctx, *req.Status)
	} else {
		invoices, err = s.repository.FindActive(ctx)
	}

	if err != nil {
		return nil, err
	}

	// Apply additional filtering (this would typically be done in the repository)
	filteredInvoices := make([]*Invoice, 0)
	for _, invoice := range invoices {
		// Filter by merchant ID
		if invoice.MerchantID() != req.MerchantID {
			continue
		}

		// Filter by customer ID if provided
		if req.CustomerID != nil {
			if invoice.CustomerID() == nil || *invoice.CustomerID() != *req.CustomerID {
				continue
			}
		}

		// Filter by date range if provided
		if req.CreatedAfter != nil && invoice.CreatedAt().Before(*req.CreatedAfter) {
			continue
		}
		if req.CreatedBefore != nil && invoice.CreatedAt().After(*req.CreatedBefore) {
			continue
		}

		// Filter by search term if provided
		if req.Search != nil {
			searchTerm := *req.Search
			if !s.matchesSearch(invoice, searchTerm) {
				continue
			}
		}

		filteredInvoices = append(filteredInvoices, invoice)
	}

	// Apply pagination
	start := req.Offset
	if start < 0 {
		start = 0
	}
	end := start + limit
	if end > len(filteredInvoices) {
		end = len(filteredInvoices)
	}

	var paginatedInvoices []*Invoice
	if start < len(filteredInvoices) {
		paginatedInvoices = filteredInvoices[start:end]
	}

	return &ListInvoicesResponse{
		Invoices: paginatedInvoices,
		Total:    len(filteredInvoices),
		Limit:    limit,
		Offset:   req.Offset,
	}, nil
}

// MarkInvoiceAsViewed marks an invoice as viewed by the customer.
func (s *InvoiceServiceImpl) MarkInvoiceAsViewed(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("invoice ID cannot be empty")
	}

	invoice, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := invoice.MarkAsViewed(); err != nil {
		return err
	}

	return s.repository.Update(ctx, invoice)
}

// CancelInvoice cancels an invoice.
func (s *InvoiceServiceImpl) CancelInvoice(ctx context.Context, id string, reason string) error {
	if id == "" {
		return errors.New("invoice ID cannot be empty")
	}

	invoice, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := invoice.Cancel(); err != nil {
		return err
	}

	return s.repository.Update(ctx, invoice)
}

// ProcessPayment processes a payment for an invoice.
func (s *InvoiceServiceImpl) ProcessPayment(ctx context.Context, invoiceID string, payment *Payment) error {
	if invoiceID == "" {
		return errors.New("invoice ID cannot be empty")
	}

	if payment == nil {
		return errors.New("payment cannot be nil")
	}

	invoice, err := s.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Validate payment amount
	isValid, validationType, err := invoice.ValidatePaymentAmount(payment.Amount())
	if err != nil {
		return err
	}

	if !isValid {
		return errors.New("payment validation failed: " + validationType)
	}

	// Update invoice status based on payment
	switch validationType {
	case "sufficient":
		if invoice.Status() == StatusPending || invoice.Status() == StatusPartial {
			if err := invoice.TransitionTo(StatusConfirming); err != nil {
				return err
			}
		}
	case "partial":
		if invoice.Status() == StatusPending {
			if err := invoice.TransitionTo(StatusPartial); err != nil {
				return err
			}
		}
	}

	// Save updated invoice
	return s.repository.Update(ctx, invoice)
}

// GetExpiredInvoices retrieves invoices that have expired.
func (s *InvoiceServiceImpl) GetExpiredInvoices(ctx context.Context) ([]*Invoice, error) {
	return s.repository.FindExpired(ctx)
}

// ProcessExpiredInvoices processes expired invoices.
func (s *InvoiceServiceImpl) ProcessExpiredInvoices(ctx context.Context) error {
	expiredInvoices, err := s.GetExpiredInvoices(ctx)
	if err != nil {
		return err
	}

	for _, invoice := range expiredInvoices {
		if err := invoice.Expire(); err != nil {
			// Log error but continue processing other invoices
			continue
		}

		if err := s.repository.Update(ctx, invoice); err != nil {
			// Log error but continue processing other invoices
			continue
		}
	}

	return nil
}

// GetInvoiceStatus returns the current status of an invoice.
func (s *InvoiceServiceImpl) GetInvoiceStatus(ctx context.Context, id string) (InvoiceStatus, error) {
	if id == "" {
		return "", errors.New("invoice ID cannot be empty")
	}

	invoice, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return "", err
	}

	return invoice.Status(), nil
}

// UpdateInvoiceStatus updates the status of an invoice.
func (s *InvoiceServiceImpl) UpdateInvoiceStatus(ctx context.Context, id string, newStatus InvoiceStatus, reason string) error {
	if id == "" {
		return errors.New("invoice ID cannot be empty")
	}

	if !newStatus.IsValid() {
		return errors.New("invalid invoice status")
	}

	invoice, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := invoice.TransitionTo(newStatus); err != nil {
		return err
	}

	return s.repository.Update(ctx, invoice)
}

// Helper methods

func (s *InvoiceServiceImpl) getExchangeRate(ctx context.Context, from shared.Currency, to shared.CryptoCurrency) (*shared.ExchangeRate, error) {
	// This would typically call an exchange rate service
	// For now, we'll return a mock rate
	rate := "1.0" // Mock rate
	source := "mock_provider"
	duration := 30 * time.Minute

	return shared.NewExchangeRate(rate, from, to, source, duration)
}

func (s *InvoiceServiceImpl) generatePaymentAddress(ctx context.Context, currency shared.CryptoCurrency) (*shared.PaymentAddress, error) {
	// This would typically call a payment address service
	// For now, we'll return a mock address
	var network shared.BlockchainNetwork
	switch currency {
	case shared.CryptoCurrencyUSDT:
		network = shared.NetworkTron
	case shared.CryptoCurrencyBTC:
		network = shared.NetworkBitcoin
	case shared.CryptoCurrencyETH:
		network = shared.NetworkEthereum
	default:
		return nil, errors.New("unsupported cryptocurrency for address generation")
	}

	// Mock address
	address := "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN"
	return shared.NewPaymentAddress(address, network)
}

func (s *InvoiceServiceImpl) generateInvoiceID() string {
	// This would typically use a proper ID generator
	// For now, we'll use a simple timestamp-based ID
	return "inv_" + time.Now().Format("20060102150405")
}

func (s *InvoiceServiceImpl) matchesSearch(invoice *Invoice, searchTerm string) bool {
	// Simple search implementation
	searchTerm = strings.ToLower(searchTerm)

	if strings.Contains(strings.ToLower(invoice.Title()), searchTerm) {
		return true
	}

	if strings.Contains(strings.ToLower(invoice.Description()), searchTerm) {
		return true
	}

	// Search in metadata
	for key, value := range invoice.Metadata() {
		if strings.Contains(strings.ToLower(key), searchTerm) {
			return true
		}
		if str, ok := value.(string); ok {
			if strings.Contains(strings.ToLower(str), searchTerm) {
				return true
			}
		}
	}

	return false
}
