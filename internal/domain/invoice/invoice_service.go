package invoice

import (
	"context"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/domain/shared"
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// InvoiceServiceImpl implements the InvoiceService interface.
type InvoiceServiceImpl struct {
	repository Repository
	eventBus   shared.EventBus
	logger     *zap.Logger
}

// NewInvoiceService creates a new InvoiceService implementation.
func NewInvoiceService(repository Repository, eventBus shared.EventBus, logger *zap.Logger) InvoiceService {
	logger.Info("Creating InvoiceService",
		zap.Bool("eventBus_provided", eventBus != nil),
		zap.Bool("repository_provided", repository != nil))

	return &InvoiceServiceImpl{
		repository: repository,
		eventBus:   eventBus,
		logger:     logger,
	}
}

// CreateInvoice creates a new invoice with the given parameters.
func (s *InvoiceServiceImpl) CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*Invoice, error) {
	if err := s.validateCreateInvoiceRequest(req); err != nil {
		return nil, err
	}

	items, pricing, err := s.buildInvoiceItemsAndPricing(req)
	if err != nil {
		return nil, err
	}

	exchangeRate, err := s.getExchangeRate(ctx, req.Currency, req.CryptoCurrency)
	if err != nil {
		return nil, err
	}

	paymentAddress, err := s.generatePaymentAddress(ctx, req.CryptoCurrency)
	if err != nil {
		return nil, err
	}

	paymentTolerance := s.getPaymentTolerance(req)
	expiration := s.getExpiration(req)
	invoiceID := s.generateInvoiceID()

	if err := s.validateInvoiceComponents(invoiceID, req, items, pricing, paymentAddress, exchangeRate, paymentTolerance, expiration); err != nil {
		return nil, err
	}

	invoice, err := s.buildInvoice(
		invoiceID, req, items, pricing, paymentAddress, exchangeRate, paymentTolerance, expiration,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repository.Save(ctx, invoice); err != nil {
		return nil, err
	}

	// Publish invoice created event
	if s.eventBus != nil {
		eventData := createInvoiceEventData(invoice)
		eventData["timestamp"] = time.Now().UTC()
		event := shared.CreateDomainEvent(shared.EventTypeInvoiceCreated, invoice.ID(), "Invoice", eventData, nil)
		if err := s.eventBus.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
			if s.logger != nil {
				s.logger.Error("Failed to publish domain event",
					zap.String("event_type", shared.EventTypeInvoiceCreated),
					zap.String("aggregate_id", invoice.ID()),
					zap.Error(err),
				)
			}
		}
	}

	return invoice, nil
}

// validateCreateInvoiceRequest validates the basic request parameters.
func (s *InvoiceServiceImpl) validateCreateInvoiceRequest(req *CreateInvoiceRequest) error {
	if req == nil {
		return errors.New("create invoice request cannot be nil")
	}
	if req.MerchantID == "" {
		return errors.New("merchant ID is required")
	}
	if req.Title == "" {
		return errors.New("title is required")
	}
	if len(req.Items) == 0 {
		return errors.New("at least one item is required")
	}
	if !req.CryptoCurrency.IsValid() {
		return errors.New("invalid cryptocurrency")
	}
	return nil
}

// buildInvoiceItemsAndPricing creates invoice items and calculates pricing.
func (s *InvoiceServiceImpl) buildInvoiceItemsAndPricing(
	req *CreateInvoiceRequest,
) ([]*InvoiceItem, *InvoicePricing, error) {
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
			return nil, nil, err
		}
		items = append(items, item)
		subtotal = subtotal.Add(item.TotalPrice().Amount())
	}

	subtotalMoney, err := shared.NewMoney(subtotal.String(), req.Currency)
	if err != nil {
		return nil, nil, err
	}

	var taxMoney *shared.Money
	if req.Tax != nil {
		taxMoney = req.Tax
	} else {
		taxMoney, err = shared.NewMoney("0.00", req.Currency)
		if err != nil {
			return nil, nil, err
		}
	}

	totalMoney, err := subtotalMoney.Add(taxMoney)
	if err != nil {
		return nil, nil, err
	}

	pricing, err := NewInvoicePricing(subtotalMoney, taxMoney, totalMoney)
	if err != nil {
		return nil, nil, err
	}

	return items, pricing, nil
}

// getPaymentTolerance returns the payment tolerance, using default if not provided.
func (s *InvoiceServiceImpl) getPaymentTolerance(req *CreateInvoiceRequest) *PaymentTolerance {
	if req.PaymentTolerance != nil {
		return req.PaymentTolerance
	}
	return DefaultPaymentTolerance()
}

// getExpiration returns the expiration, using default if not provided.
func (s *InvoiceServiceImpl) getExpiration(req *CreateInvoiceRequest) *InvoiceExpiration {
	expirationDuration := req.ExpirationDuration
	if expirationDuration == 0 {
		expirationDuration = 30 * time.Minute
	}
	return NewInvoiceExpiration(expirationDuration)
}

// validateInvoiceComponents validates all invoice components.
func (s *InvoiceServiceImpl) validateInvoiceComponents(
	invoiceID string,
	req *CreateInvoiceRequest,
	items []*InvoiceItem,
	pricing *InvoicePricing,
	paymentAddress *shared.PaymentAddress,
	exchangeRate *shared.ExchangeRate,
	paymentTolerance *PaymentTolerance,
	expiration *InvoiceExpiration,
) error {
	if invoiceID == "" {
		return errors.New("invoice ID cannot be empty")
	}
	if len(req.Title) > 255 {
		return errors.New("title cannot exceed 255 characters")
	}
	if len(req.Description) > 1000 {
		return errors.New("description cannot exceed 1000 characters")
	}
	if len(items) == 0 {
		return errors.New("invoice must have at least one item")
	}
	if pricing == nil {
		return errors.New("pricing cannot be nil")
	}
	if paymentAddress == nil {
		return errors.New("payment address cannot be nil")
	}
	if exchangeRate == nil {
		return errors.New("exchange rate cannot be nil")
	}
	if paymentTolerance == nil {
		return errors.New("payment tolerance cannot be nil")
	}
	if expiration == nil {
		return errors.New("expiration cannot be nil")
	}
	if exchangeRate.IsExpired() {
		return errors.New("exchange rate has expired")
	}
	if paymentAddress.IsExpired() {
		return errors.New("payment address has expired")
	}
	return nil
}

// buildInvoice creates the invoice entity.
func (s *InvoiceServiceImpl) buildInvoice(
	invoiceID string,
	req *CreateInvoiceRequest,
	items []*InvoiceItem,
	pricing *InvoicePricing,
	paymentAddress *shared.PaymentAddress,
	exchangeRate *shared.ExchangeRate,
	paymentTolerance *PaymentTolerance,
	expiration *InvoiceExpiration,
) (*Invoice, error) {
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

	if req.CustomerID != nil {
		invoice.SetCustomerID(*req.CustomerID)
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
func (s *InvoiceServiceImpl) GetInvoiceByPaymentAddress(
	ctx context.Context,
	address *shared.PaymentAddress,
) (*Invoice, error) {
	if address == nil {
		return nil, errors.New("payment address cannot be nil")
	}

	return s.repository.FindByPaymentAddress(ctx, address)
}

// ListInvoices retrieves invoices with the given filters.
func (s *InvoiceServiceImpl) ListInvoices(
	ctx context.Context,
	req *ListInvoicesRequest,
) (*ListInvoicesResponse, error) {
	if err := s.validateListInvoicesRequest(req); err != nil {
		return nil, err
	}

	limit := s.normalizeLimit(req.Limit)
	invoices, err := s.fetchInvoices(ctx, req)
	if err != nil {
		return nil, err
	}

	filteredInvoices := s.filterInvoices(invoices, req)
	paginatedInvoices := s.paginateInvoices(filteredInvoices, req.Offset, limit)

	return &ListInvoicesResponse{
		Invoices: paginatedInvoices,
		Total:    len(filteredInvoices),
		Limit:    limit,
		Offset:   req.Offset,
	}, nil
}

// validateListInvoicesRequest validates the list invoices request.
func (s *InvoiceServiceImpl) validateListInvoicesRequest(req *ListInvoicesRequest) error {
	if req == nil {
		return errors.New("list invoices request cannot be nil")
	}
	if req.MerchantID == "" {
		return errors.New("merchant ID is required")
	}
	return nil
}

// normalizeLimit normalizes the limit to a valid range.
func (s *InvoiceServiceImpl) normalizeLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

// fetchInvoices retrieves invoices from the repository.
func (s *InvoiceServiceImpl) fetchInvoices(ctx context.Context, req *ListInvoicesRequest) ([]*Invoice, error) {
	if req.Status != nil {
		return s.repository.FindByStatus(ctx, *req.Status)
	}
	return s.repository.FindActive(ctx)
}

// filterInvoices applies filtering logic to the invoices.
func (s *InvoiceServiceImpl) filterInvoices(invoices []*Invoice, req *ListInvoicesRequest) []*Invoice {
	filteredInvoices := make([]*Invoice, 0)
	for _, invoice := range invoices {
		if s.shouldIncludeInvoice(invoice, req) {
			filteredInvoices = append(filteredInvoices, invoice)
		}
	}
	return filteredInvoices
}

// shouldIncludeInvoice determines if an invoice should be included in the results.
func (s *InvoiceServiceImpl) shouldIncludeInvoice(invoice *Invoice, req *ListInvoicesRequest) bool {
	// Filter by merchant ID
	if invoice.MerchantID() != req.MerchantID {
		return false
	}

	// Filter by customer ID if provided
	if req.CustomerID != nil {
		if invoice.CustomerID() == nil || *invoice.CustomerID() != *req.CustomerID {
			return false
		}
	}

	// Filter by date range if provided
	if req.CreatedAfter != nil && invoice.CreatedAt().Before(*req.CreatedAfter) {
		return false
	}
	if req.CreatedBefore != nil && invoice.CreatedAt().After(*req.CreatedBefore) {
		return false
	}

	// Filter by search term if provided
	if req.Search != nil {
		if !s.matchesSearch(invoice, *req.Search) {
			return false
		}
	}

	return true
}

// paginateInvoices applies pagination to the filtered invoices.
func (s *InvoiceServiceImpl) paginateInvoices(invoices []*Invoice, offset, limit int) []*Invoice {
	start := offset
	if start < 0 {
		start = 0
	}
	end := start + limit
	if end > len(invoices) {
		end = len(invoices)
	}

	if start >= len(invoices) {
		return []*Invoice{}
	}

	return invoices[start:end]
}

// MarkInvoiceAsViewed marks an invoice as viewed by the customer using FSM.
func (s *InvoiceServiceImpl) MarkInvoiceAsViewed(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("invoice ID cannot be empty")
	}

	invoice, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to find invoice for viewing",
				zap.String("invoice_id", id),
				zap.Error(err),
			)
		}
		return err
	}

	// Business logic validation
	if invoice.Status() != StatusCreated {
		return errors.New("can only mark created invoices as viewed")
	}
	if invoice.ViewedAt() != nil {
		return errors.New("invoice already marked as viewed")
	}

	// Use FSM to transition from created to pending when viewed
	fsm := NewInvoiceFSM(invoice)
	if err := fsm.Event(ctx, "view"); err != nil {
		return err
	}

	// Mark as viewed (set viewedAt timestamp)
	now := time.Now().UTC()
	invoice.SetViewedAt(&now)

	if err := s.repository.Update(ctx, invoice); err != nil {
		return err
	}

	// Publish invoice status changed event
	if s.eventBus != nil {
		eventData := createInvoiceEventData(invoice)
		eventData["from_status"] = StatusCreated.String()
		eventData["to_status"] = invoice.Status().String()
		eventData["reason"] = "viewed_by_customer"
		eventData["viewed_at"] = now

		eventData["timestamp"] = time.Now().UTC()
		event := shared.CreateDomainEvent(shared.EventTypeInvoiceStatusChanged, invoice.ID(), "Invoice", eventData, nil)
		if err := s.eventBus.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
			if s.logger != nil {
				s.logger.Error("Failed to publish domain event",
					zap.String("event_type", shared.EventTypeInvoiceStatusChanged),
					zap.String("aggregate_id", invoice.ID()),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// CancelInvoice cancels an invoice using FSM.
func (s *InvoiceServiceImpl) CancelInvoice(ctx context.Context, id, reason string) error {
	if id == "" {
		return errors.New("invoice ID cannot be empty")
	}

	invoice, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Business logic validation
	if invoice.Status().IsTerminal() {
		return errors.New("cannot cancel invoice in terminal state")
	}

	// Use FSM to transition to cancelled status
	fsm := NewInvoiceFSM(invoice)
	if err := fsm.Event(ctx, "cancel"); err != nil {
		return err
	}

	if err := s.repository.Update(ctx, invoice); err != nil {
		return err
	}

	// Publish invoice cancelled event
	if s.eventBus != nil {
		eventData := createInvoiceEventData(invoice)
		eventData["reason"] = reason
		eventData["cancelled_at"] = time.Now().UTC()

		eventData["timestamp"] = time.Now().UTC()
		event := shared.CreateDomainEvent(shared.EventTypeInvoiceCancelled, invoice.ID(), "Invoice", eventData, nil)
		if err := s.eventBus.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
			if s.logger != nil {
				s.logger.Error("Failed to publish domain event",
					zap.String("event_type", shared.EventTypeInvoiceCancelled),
					zap.String("aggregate_id", invoice.ID()),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// ProcessPayment processes a payment for an invoice using FSM.
func (s *InvoiceServiceImpl) ProcessPayment(ctx context.Context, invoiceID string, paymentTx *payment.Payment) error {
	if invoiceID == "" {
		return errors.New("invoice ID cannot be empty")
	}

	if paymentTx == nil {
		return errors.New("payment cannot be nil")
	}

	invoice, err := s.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Validate payment amount (business logic moved to service)
	validationType, err := s.validatePaymentAmount(invoice, paymentTx.Amount().Amount())
	if err != nil {
		return err
	}

	// Use FSM to update invoice status based on payment
	if err := s.processPaymentWithFSM(ctx, invoice, validationType); err != nil {
		return err
	}

	// Save updated invoice
	if err := s.repository.Update(ctx, invoice); err != nil {
		return err
	}

	// Publish payment processed event
	if s.eventBus != nil {
		eventData := createInvoiceEventData(invoice)
		eventData["payment_amount"] = paymentTx.Amount().Amount()
		eventData["payment_validation"] = validationType
		eventData["processed_at"] = time.Now().UTC()

		eventData["timestamp"] = time.Now().UTC()
		event := shared.CreateDomainEvent(shared.EventTypeInvoiceStatusChanged, invoice.ID(), "Invoice", eventData, nil)
		if err := s.eventBus.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
			if s.logger != nil {
				s.logger.Error("Failed to publish domain event",
					zap.String("event_type", shared.EventTypeInvoiceStatusChanged),
					zap.String("aggregate_id", invoice.ID()),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// GetExpiredInvoices retrieves invoices that have expired.
func (s *InvoiceServiceImpl) GetExpiredInvoices(ctx context.Context) ([]*Invoice, error) {
	return s.repository.FindExpired(ctx)
}

// ProcessExpiredInvoices processes expired invoices using FSM.
func (s *InvoiceServiceImpl) ProcessExpiredInvoices(ctx context.Context) error {
	expiredInvoices, err := s.GetExpiredInvoices(ctx)
	if err != nil {
		return err
	}

	for _, invoice := range expiredInvoices {
		// Business logic validation
		if invoice.Status().IsTerminal() {
			continue // Skip terminal invoices
		}
		// Special case: partial payments should not auto-expire
		if invoice.Status() == StatusPartial {
			continue // Skip partial payment invoices
		}
		// Check if invoice has actually expired
		if !invoice.Expiration().IsExpired() {
			continue // Skip invoices that haven't expired yet
		}

		// Use FSM to transition to expired status
		fsm := NewInvoiceFSM(invoice)
		if err := fsm.Event(ctx, "expire"); err != nil {
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

// ProcessExpiredInvoice processes a specific expired invoice by ID using FSM.
// This is useful for testing and manual intervention.
func (s *InvoiceServiceImpl) ProcessExpiredInvoice(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("invoice ID cannot be empty")
	}

	invoice, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Business logic validation
	if invoice.Expiration().IsExpired() && !invoice.Status().IsTerminal() {
		// Special case: partial payments should not auto-expire
		if invoice.Status() != StatusPartial {
			// Use FSM to transition to expired status
			fsm := NewInvoiceFSM(invoice)
			if err := fsm.Event(ctx, "expire"); err != nil {
				return err
			}

			if err := s.repository.Update(ctx, invoice); err != nil {
				return err
			}

			// Publish invoice expired event
			if s.eventBus != nil {
				eventData := createInvoiceEventData(invoice)
				eventData["expired_at"] = time.Now().UTC()
				eventData["expires_at"] = invoice.Expiration().ExpiresAt()

				eventData["timestamp"] = time.Now().UTC()
				event := shared.CreateDomainEvent(
					shared.EventTypeInvoiceExpired,
					invoice.ID(),
					"Invoice",
					eventData,
					nil,
				)
				if err := s.eventBus.PublishEvent(ctx, event); err != nil {
					// Log error but don't fail the operation
					if s.logger != nil {
						s.logger.Error("Failed to publish domain event",
							zap.String("event_type", shared.EventTypeInvoiceExpired),
							zap.String("aggregate_id", invoice.ID()),
							zap.Error(err),
						)
					}
				}
			}

			return nil
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

// UpdateInvoiceStatus updates the status of an invoice using FSM.
func (s *InvoiceServiceImpl) UpdateInvoiceStatus(
	ctx context.Context,
	id string,
	newStatus InvoiceStatus,
	reason string,
) error {
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

	// Use FSM to transition to new status
	fsm := NewInvoiceFSM(invoice)
	if err := fsm.TransitionTo(newStatus); err != nil {
		return err
	}

	if err := s.repository.Update(ctx, invoice); err != nil {
		return err
	}

	// Publish invoice status changed event
	if s.eventBus != nil {
		eventData := createInvoiceEventData(invoice)
		eventData["from_status"] = invoice.Status().String()
		eventData["to_status"] = newStatus.String()
		eventData["reason"] = reason
		eventData["updated_at"] = time.Now().UTC()

		eventData["timestamp"] = time.Now().UTC()
		event := shared.CreateDomainEvent(shared.EventTypeInvoiceStatusChanged, invoice.ID(), "Invoice", eventData, nil)
		if err := s.eventBus.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
			if s.logger != nil {
				s.logger.Error("Failed to publish domain event",
					zap.String("event_type", shared.EventTypeInvoiceStatusChanged),
					zap.String("aggregate_id", invoice.ID()),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// Helper methods

func (s *InvoiceServiceImpl) getExchangeRate(
	ctx context.Context,
	from shared.Currency,
	to shared.CryptoCurrency,
) (*shared.ExchangeRate, error) {
	// This would typically call an exchange rate service
	// For now, we'll return a mock rate
	rate := "1.0" // Mock rate
	source := "mock_provider"
	duration := 30 * time.Minute

	exchangeRate, err := shared.NewExchangeRate(rate, from, to, source, duration)
	if err != nil && s.logger != nil {
		s.logger.Error("Failed to get exchange rate",
			zap.String("currency", string(from)),
			zap.String("crypto_currency", string(to)),
			zap.Error(err),
		)
	}
	return exchangeRate, err
}

func (s *InvoiceServiceImpl) generatePaymentAddress(
	ctx context.Context,
	currency shared.CryptoCurrency,
) (*shared.PaymentAddress, error) {
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
		err := errors.New("unsupported cryptocurrency for address generation")
		if s.logger != nil {
			s.logger.Error("Failed to generate payment address",
				zap.String("crypto_currency", string(currency)),
				zap.Error(err),
			)
		}
		return nil, err
	}

	// Mock address
	address := "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN"
	paymentAddress, err := shared.NewPaymentAddress(address, network)
	if err != nil && s.logger != nil {
		s.logger.Error("Failed to generate payment address",
			zap.String("crypto_currency", string(currency)),
			zap.Error(err),
		)
	}
	return paymentAddress, err
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

// validatePaymentAmount validates if a payment amount is acceptable (business logic moved from domain).
func (s *InvoiceServiceImpl) validatePaymentAmount(invoice *Invoice, paymentAmount *shared.Money) (string, error) {
	if paymentAmount == nil {
		return "", errors.New("payment amount cannot be nil")
	}

	requiredAmount, err := invoice.GetCryptoAmount()
	if err != nil {
		return "", err
	}

	// Check currency match
	if paymentAmount.Currency() != requiredAmount.Currency() {
		return "", errors.New("payment currency does not match invoice currency")
	}

	// Check if payment is sufficient
	if paymentAmount.GreaterThanOrEqual(requiredAmount) {
		return "sufficient", nil
	}

	// Check if underpayment is within tolerance
	if invoice.PaymentTolerance().IsUnderpayment(requiredAmount, paymentAmount) {
		return "", errors.New("payment amount is below the minimum threshold")
	}

	return "partial", nil
}

// processPaymentWithFSM processes payment using FSM to reduce cyclomatic complexity.
func (s *InvoiceServiceImpl) processPaymentWithFSM(ctx context.Context, invoice *Invoice, validationType string) error {
	fsm := NewInvoiceFSM(invoice)

	switch validationType {
	case "sufficient":
		if invoice.Status() == StatusPending || invoice.Status() == StatusPartial {
			return fsm.Event(ctx, "full_payment")
		}
	case "partial":
		if invoice.Status() == StatusPending {
			return fsm.Event(ctx, "partial_payment")
		}
	}

	return nil
}
