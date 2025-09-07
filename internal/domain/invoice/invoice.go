package invoice

import (
	"time"

	"crypto-checkout/internal/domain/shared"

	"github.com/go-playground/validator/v10"
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

// InvoiceValidation represents the validation structure for Invoice creation.
type InvoiceValidation struct {
	ID               string                 `validate:"required,min=1" json:"id"`
	MerchantID       string                 `validate:"required,min=1" json:"merchant_id"`
	Title            string                 `validate:"required,min=1" json:"title"`
	Description      string                 `json:"description"`
	Items            []*InvoiceItem         `validate:"required,min=1,dive" json:"items"`
	Pricing          *InvoicePricing        `validate:"required" json:"pricing"`
	CryptoCurrency   shared.CryptoCurrency  `validate:"required" json:"crypto_currency"`
	PaymentAddress   *shared.PaymentAddress `validate:"required" json:"payment_address"`
	ExchangeRate     *shared.ExchangeRate   `validate:"required" json:"exchange_rate"`
	PaymentTolerance *PaymentTolerance      `validate:"required" json:"payment_tolerance"`
	Expiration       *InvoiceExpiration     `validate:"required" json:"expiration"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// NewInvoice creates a new Invoice with validation using go-playground/validator.
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
	// Create validation struct
	validation := InvoiceValidation{
		ID:               id,
		MerchantID:       merchantID,
		Title:            title,
		Description:      description,
		Items:            items,
		Pricing:          pricing,
		CryptoCurrency:   cryptoCurrency,
		PaymentAddress:   paymentAddress,
		ExchangeRate:     exchangeRate,
		PaymentTolerance: paymentTolerance,
		Expiration:       expiration,
		Metadata:         metadata,
	}

	// Validate using go-playground/validator
	validate := validator.New()
	if err := validate.Struct(validation); err != nil {
		return nil, err
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

// SetViewedAt sets the viewed timestamp.
func (i *Invoice) SetViewedAt(viewedAt *time.Time) {
	i.viewedAt = viewedAt
	i.updatedAt = time.Now().UTC()
}

// SetStatus sets the invoice status.
func (i *Invoice) SetStatus(status InvoiceStatus) {
	i.status = status
	i.updatedAt = time.Now().UTC()
}

// SetPaidAt sets the paid timestamp.
func (i *Invoice) SetPaidAt(paidAt *time.Time) {
	i.paidAt = paidAt
	i.updatedAt = time.Now().UTC()
}

// SetExpiration sets the invoice expiration.
func (i *Invoice) SetExpiration(expiration *InvoiceExpiration) {
	i.expiration = expiration
	i.updatedAt = time.Now().UTC()
}

// SetMetadata sets the invoice metadata.
func (i *Invoice) SetMetadata(metadata map[string]interface{}) {
	i.metadata = metadata
	i.updatedAt = time.Now().UTC()
}

// SetUpdatedAt sets the updated timestamp.
func (i *Invoice) SetUpdatedAt(updatedAt time.Time) {
	i.updatedAt = updatedAt
}

// GetCryptoAmount returns the cryptocurrency amount for this invoice.
func (i *Invoice) GetCryptoAmount() (*shared.Money, error) {
	return i.exchangeRate.Convert(i.pricing.Total())
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
