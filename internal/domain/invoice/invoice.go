package invoice

import (
	"errors"
	"time"

	"crypto-checkout/internal/domain/shared"
)

// Invoice represents the main invoice aggregate root.
type Invoice struct {
	id               string
	merchantID       string
	customerID       *string
	title            string
	description      string
	items            []*InvoiceItem
	pricing          *InvoicePricing
	cryptoCurrency   shared.CryptoCurrency
	paymentAddress   *shared.PaymentAddress
	status           InvoiceStatus
	exchangeRate     *shared.ExchangeRate
	paymentTolerance *PaymentTolerance
	expiration       *InvoiceExpiration
	createdAt        time.Time
	updatedAt        time.Time
	paidAt           *time.Time
	viewedAt         *time.Time
	metadata         map[string]interface{}
}

// NewInvoice creates a new Invoice.
func NewInvoice(
	id, merchantID, title, description string,
	items []*InvoiceItem,
	pricing *InvoicePricing,
	cryptoCurrency shared.CryptoCurrency,
	paymentAddress *shared.PaymentAddress,
	exchangeRate *shared.ExchangeRate,
	paymentTolerance *PaymentTolerance,
	expiration *InvoiceExpiration,
	metadata map[string]interface{},
) (*Invoice, error) {
	if id == "" {
		return nil, errors.New("invoice ID cannot be empty")
	}

	if merchantID == "" {
		return nil, errors.New("merchant ID cannot be empty")
	}

	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	if len(title) > 255 {
		return nil, errors.New("title cannot exceed 255 characters")
	}

	if len(description) > 1000 {
		return nil, errors.New("description cannot exceed 1000 characters")
	}

	if len(items) == 0 {
		return nil, errors.New("invoice must have at least one item")
	}

	if pricing == nil {
		return nil, errors.New("pricing cannot be nil")
	}

	if !cryptoCurrency.IsValid() {
		return nil, errors.New("invalid cryptocurrency")
	}

	if paymentAddress == nil {
		return nil, errors.New("payment address cannot be nil")
	}

	if exchangeRate == nil {
		return nil, errors.New("exchange rate cannot be nil")
	}

	if paymentTolerance == nil {
		return nil, errors.New("payment tolerance cannot be nil")
	}

	if expiration == nil {
		return nil, errors.New("expiration cannot be nil")
	}

	// Validate that exchange rate is not expired
	if exchangeRate.IsExpired() {
		return nil, errors.New("exchange rate has expired")
	}

	// Validate that payment address is not expired
	if paymentAddress.IsExpired() {
		return nil, errors.New("payment address has expired")
	}

	now := time.Now().UTC()
	return &Invoice{
		id:               id,
		merchantID:       merchantID,
		title:            title,
		description:      description,
		items:            items,
		pricing:          pricing,
		cryptoCurrency:   cryptoCurrency,
		paymentAddress:   paymentAddress,
		status:           StatusCreated,
		exchangeRate:     exchangeRate,
		paymentTolerance: paymentTolerance,
		expiration:       expiration,
		createdAt:        now,
		updatedAt:        now,
		metadata:         metadata,
	}, nil
}

// ID returns the invoice ID.
func (i *Invoice) ID() string {
	return i.id
}

// MerchantID returns the merchant ID.
func (i *Invoice) MerchantID() string {
	return i.merchantID
}

// CustomerID returns the customer ID if set.
func (i *Invoice) CustomerID() *string {
	return i.customerID
}

// SetCustomerID sets the customer ID.
func (i *Invoice) SetCustomerID(customerID string) {
	i.customerID = &customerID
	i.updatedAt = time.Now().UTC()
}

// Title returns the invoice title.
func (i *Invoice) Title() string {
	return i.title
}

// Description returns the invoice description.
func (i *Invoice) Description() string {
	return i.description
}

// Items returns the invoice items.
func (i *Invoice) Items() []*InvoiceItem {
	return i.items
}

// Pricing returns the invoice pricing.
func (i *Invoice) Pricing() *InvoicePricing {
	return i.pricing
}

// CryptoCurrency returns the cryptocurrency.
func (i *Invoice) CryptoCurrency() shared.CryptoCurrency {
	return i.cryptoCurrency
}

// PaymentAddress returns the payment address.
func (i *Invoice) PaymentAddress() *shared.PaymentAddress {
	return i.paymentAddress
}

// Status returns the current status.
func (i *Invoice) Status() InvoiceStatus {
	return i.status
}

// ExchangeRate returns the exchange rate.
func (i *Invoice) ExchangeRate() *shared.ExchangeRate {
	return i.exchangeRate
}

// PaymentTolerance returns the payment tolerance.
func (i *Invoice) PaymentTolerance() *PaymentTolerance {
	return i.paymentTolerance
}

// Expiration returns the expiration settings.
func (i *Invoice) Expiration() *InvoiceExpiration {
	return i.expiration
}

// CreatedAt returns the creation time.
func (i *Invoice) CreatedAt() time.Time {
	return i.createdAt
}

// UpdatedAt returns the last update time.
func (i *Invoice) UpdatedAt() time.Time {
	return i.updatedAt
}

// PaidAt returns the payment completion time if paid.
func (i *Invoice) PaidAt() *time.Time {
	return i.paidAt
}

// ViewedAt returns the first view time if viewed.
func (i *Invoice) ViewedAt() *time.Time {
	return i.viewedAt
}

// Metadata returns the metadata.
func (i *Invoice) Metadata() map[string]interface{} {
	return i.metadata
}

// IsExpired returns true if the invoice has expired.
func (i *Invoice) IsExpired() bool {
	return i.expiration.IsExpired()
}

// IsActive returns true if the invoice is in an active state.
func (i *Invoice) IsActive() bool {
	return i.status.IsActive()
}

// IsTerminal returns true if the invoice is in a terminal state.
func (i *Invoice) IsTerminal() bool {
	return i.status.IsTerminal()
}

// CanTransitionTo returns true if the invoice can transition to the target status.
func (i *Invoice) CanTransitionTo(target InvoiceStatus) bool {
	return i.status.CanTransitionTo(target)
}

// MarkAsViewed marks the invoice as viewed by the customer.
func (i *Invoice) MarkAsViewed() error {
	if i.status != StatusCreated {
		return errors.New("can only mark created invoices as viewed")
	}

	if i.viewedAt != nil {
		return errors.New("invoice already marked as viewed")
	}

	now := time.Now().UTC()
	i.viewedAt = &now
	i.status = StatusPending
	i.updatedAt = now

	return nil
}

// TransitionTo transitions the invoice to a new status.
func (i *Invoice) TransitionTo(newStatus InvoiceStatus) error {
	if !i.CanTransitionTo(newStatus) {
		return errors.New("invalid status transition from " + i.status.String() + " to " + newStatus.String())
	}

	now := time.Now().UTC()
	i.status = newStatus
	i.updatedAt = now

	// Set paidAt when transitioning to paid status
	if newStatus == StatusPaid && i.paidAt == nil {
		i.paidAt = &now
	}

	return nil
}

// Cancel cancels the invoice.
func (i *Invoice) Cancel() error {
	if i.IsTerminal() {
		return errors.New("cannot cancel invoice in terminal state")
	}

	return i.TransitionTo(StatusCancelled)
}

// Expire expires the invoice.
func (i *Invoice) Expire() error {
	if i.IsTerminal() {
		return errors.New("cannot expire invoice in terminal state")
	}

	// Special case: partial payments should not auto-expire
	if i.status == StatusPartial {
		return errors.New("cannot auto-expire invoices with partial payments")
	}

	return i.TransitionTo(StatusExpired)
}

// MarkAsPaid marks the invoice as paid.
func (i *Invoice) MarkAsPaid() error {
	if i.status != StatusConfirming {
		return errors.New("can only mark confirming invoices as paid")
	}

	return i.TransitionTo(StatusPaid)
}

// MarkAsRefunded marks the invoice as refunded.
func (i *Invoice) MarkAsRefunded() error {
	if i.status != StatusPaid {
		return errors.New("can only refund paid invoices")
	}

	return i.TransitionTo(StatusRefunded)
}

// GetCryptoAmount returns the cryptocurrency amount for this invoice.
func (i *Invoice) GetCryptoAmount() (*shared.Money, error) {
	return i.exchangeRate.Convert(i.pricing.Total())
}

// ValidatePaymentAmount validates if a payment amount is acceptable.
func (i *Invoice) ValidatePaymentAmount(paymentAmount *shared.Money) (bool, string, error) {
	if paymentAmount == nil {
		return false, "", errors.New("payment amount cannot be nil")
	}

	requiredAmount, err := i.GetCryptoAmount()
	if err != nil {
		return false, "", err
	}

	// Check currency match
	if paymentAmount.Currency() != requiredAmount.Currency() {
		return false, "currency_mismatch", errors.New("payment currency does not match invoice currency")
	}

	// Check if payment is sufficient
	if paymentAmount.GreaterThanOrEqual(requiredAmount) {
		return true, "sufficient", nil
	}

	// Check if underpayment is within tolerance
	if i.paymentTolerance.IsUnderpayment(requiredAmount, paymentAmount) {
		return false, "underpayment", errors.New("payment amount is below the minimum threshold")
	}

	return true, "partial", nil
}

// String returns the string representation of the invoice.
func (i *Invoice) String() string {
	return "Invoice[" + i.id + "] " + i.title + " - " + i.status.String()
}

// Equals returns true if this invoice equals the other.
func (i *Invoice) Equals(other *Invoice) bool {
	if other == nil {
		return false
	}
	return i.id == other.id
}

// SetStatus sets the invoice status (for testing purposes).
func (i *Invoice) SetStatus(status InvoiceStatus) {
	i.status = status
	i.updatedAt = time.Now().UTC()
}

// SetExpiration sets the invoice expiration (for testing purposes).
func (i *Invoice) SetExpiration(expiration *InvoiceExpiration) {
	i.expiration = expiration
	i.updatedAt = time.Now().UTC()
}
