package invoice

import (
	"errors"
	"time"

	"crypto-checkout/internal/domain/shared"

	"github.com/shopspring/decimal"
)

// PaymentTolerance represents payment amount tolerance settings.
type PaymentTolerance struct {
	underpaymentThreshold decimal.Decimal
	overpaymentThreshold  decimal.Decimal
	overpaymentAction     OverpaymentAction
}

// NewPaymentTolerance creates a new PaymentTolerance.
func NewPaymentTolerance(underpaymentThreshold, overpaymentThreshold string, overpaymentAction OverpaymentAction) (*PaymentTolerance, error) {
	if underpaymentThreshold == "" {
		return nil, errors.New("underpayment threshold cannot be empty")
	}

	if overpaymentThreshold == "" {
		return nil, errors.New("overpayment threshold cannot be empty")
	}

	if !overpaymentAction.IsValid() {
		return nil, errors.New("invalid overpayment action")
	}

	underpayment, err := decimal.NewFromString(underpaymentThreshold)
	if err != nil {
		return nil, errors.New("invalid underpayment threshold format")
	}

	overpayment, err := decimal.NewFromString(overpaymentThreshold)
	if err != nil {
		return nil, errors.New("invalid overpayment threshold format")
	}

	if underpayment.IsNegative() {
		return nil, errors.New("underpayment threshold cannot be negative")
	}

	if overpayment.IsNegative() {
		return nil, errors.New("overpayment threshold cannot be negative")
	}

	if underpayment.GreaterThan(decimal.NewFromFloat(1.0)) {
		return nil, errors.New("underpayment threshold cannot be greater than 1.0")
	}

	return &PaymentTolerance{
		underpaymentThreshold: underpayment,
		overpaymentThreshold:  overpayment,
		overpaymentAction:     overpaymentAction,
	}, nil
}

// DefaultPaymentTolerance returns the default payment tolerance settings.
func DefaultPaymentTolerance() *PaymentTolerance {
	return &PaymentTolerance{
		underpaymentThreshold: decimal.NewFromFloat(0.01), // 1%
		overpaymentThreshold:  decimal.NewFromFloat(1.00), // $1.00
		overpaymentAction:     OverpaymentActionCredit,
	}
}

// UnderpaymentThreshold returns the underpayment threshold.
func (pt *PaymentTolerance) UnderpaymentThreshold() decimal.Decimal {
	return pt.underpaymentThreshold
}

// OverpaymentThreshold returns the overpayment threshold.
func (pt *PaymentTolerance) OverpaymentThreshold() decimal.Decimal {
	return pt.overpaymentThreshold
}

// OverpaymentAction returns the overpayment action.
func (pt *PaymentTolerance) OverpaymentAction() OverpaymentAction {
	return pt.overpaymentAction
}

// IsUnderpayment returns true if the received amount is below the required amount by more than the threshold.
func (pt *PaymentTolerance) IsUnderpayment(required, received *shared.Money) bool {
	if required == nil || received == nil {
		return false
	}

	if required.Currency() != received.Currency() {
		return false
	}

	requiredAmount := required.Amount()
	receivedAmount := received.Amount()

	if receivedAmount.GreaterThanOrEqual(requiredAmount) {
		return false
	}

	shortfall := requiredAmount.Sub(receivedAmount)
	threshold := requiredAmount.Mul(pt.underpaymentThreshold)

	return shortfall.GreaterThan(threshold)
}

// IsOverpayment returns true if the received amount exceeds the required amount by more than the threshold.
func (pt *PaymentTolerance) IsOverpayment(required, received *shared.Money) bool {
	if required == nil || received == nil {
		return false
	}

	if required.Currency() != received.Currency() {
		return false
	}

	requiredAmount := required.Amount()
	receivedAmount := received.Amount()

	if receivedAmount.LessThanOrEqual(requiredAmount) {
		return false
	}

	excess := receivedAmount.Sub(requiredAmount)
	return excess.GreaterThan(pt.overpaymentThreshold)
}

// String returns the string representation of the payment tolerance.
func (pt *PaymentTolerance) String() string {
	return pt.underpaymentThreshold.String() + ":" + pt.overpaymentThreshold.String() + ":" + pt.overpaymentAction.String()
}

// Equals returns true if this payment tolerance equals the other.
func (pt *PaymentTolerance) Equals(other *PaymentTolerance) bool {
	if other == nil {
		return false
	}
	return pt.underpaymentThreshold.Equal(other.underpaymentThreshold) &&
		pt.overpaymentThreshold.Equal(other.overpaymentThreshold) &&
		pt.overpaymentAction == other.overpaymentAction
}

// InvoicePricing represents the pricing breakdown of an invoice.
type InvoicePricing struct {
	subtotal *shared.Money
	tax      *shared.Money
	total    *shared.Money
}

// NewInvoicePricing creates a new InvoicePricing.
func NewInvoicePricing(subtotal, tax, total *shared.Money) (*InvoicePricing, error) {
	if subtotal == nil {
		return nil, errors.New("subtotal cannot be nil")
	}

	if tax == nil {
		return nil, errors.New("tax cannot be nil")
	}

	if total == nil {
		return nil, errors.New("total cannot be nil")
	}

	// Validate currency consistency
	if subtotal.Currency() != tax.Currency() || subtotal.Currency() != total.Currency() {
		return nil, errors.New("all amounts must have the same currency")
	}

	// Validate that total = subtotal + tax
	calculatedTotal, err := subtotal.Add(tax)
	if err != nil {
		return nil, errors.New("failed to calculate total")
	}

	if !calculatedTotal.Equals(total) {
		return nil, errors.New("total must equal subtotal plus tax")
	}

	return &InvoicePricing{
		subtotal: subtotal,
		tax:      tax,
		total:    total,
	}, nil
}

// Subtotal returns the subtotal amount.
func (ip *InvoicePricing) Subtotal() *shared.Money {
	return ip.subtotal
}

// Tax returns the tax amount.
func (ip *InvoicePricing) Tax() *shared.Money {
	return ip.tax
}

// Total returns the total amount.
func (ip *InvoicePricing) Total() *shared.Money {
	return ip.total
}

// String returns the string representation of the invoice pricing.
func (ip *InvoicePricing) String() string {
	return "Subtotal: " + ip.subtotal.String() + ", Tax: " + ip.tax.String() + ", Total: " + ip.total.String()
}

// Equals returns true if this invoice pricing equals the other.
func (ip *InvoicePricing) Equals(other *InvoicePricing) bool {
	if other == nil {
		return false
	}
	return ip.subtotal.Equals(other.subtotal) &&
		ip.tax.Equals(other.tax) &&
		ip.total.Equals(other.total)
}

// InvoiceItem represents a line item in an invoice.
type InvoiceItem struct {
	name        string
	description string
	quantity    decimal.Decimal
	unitPrice   *shared.Money
	totalPrice  *shared.Money
}

// NewInvoiceItem creates a new InvoiceItem.
func NewInvoiceItem(name, description string, quantity string, unitPrice *shared.Money) (*InvoiceItem, error) {
	if name == "" {
		return nil, errors.New("item name cannot be empty")
	}

	if len(name) > 255 {
		return nil, errors.New("item name cannot exceed 255 characters")
	}

	if len(description) > 1000 {
		return nil, errors.New("item description cannot exceed 1000 characters")
	}

	if quantity == "" {
		return nil, errors.New("quantity cannot be empty")
	}

	if unitPrice == nil {
		return nil, errors.New("unit price cannot be nil")
	}

	qty, err := decimal.NewFromString(quantity)
	if err != nil {
		return nil, errors.New("invalid quantity format")
	}

	if qty.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("quantity must be positive")
	}

	// Calculate total price
	totalPrice, err := unitPrice.Multiply(qty)
	if err != nil {
		return nil, errors.New("failed to calculate total price")
	}

	return &InvoiceItem{
		name:        name,
		description: description,
		quantity:    qty,
		unitPrice:   unitPrice,
		totalPrice:  totalPrice,
	}, nil
}

// Name returns the item name.
func (ii *InvoiceItem) Name() string {
	return ii.name
}

// Description returns the item description.
func (ii *InvoiceItem) Description() string {
	return ii.description
}

// Quantity returns the item quantity.
func (ii *InvoiceItem) Quantity() decimal.Decimal {
	return ii.quantity
}

// UnitPrice returns the unit price.
func (ii *InvoiceItem) UnitPrice() *shared.Money {
	return ii.unitPrice
}

// TotalPrice returns the total price.
func (ii *InvoiceItem) TotalPrice() *shared.Money {
	return ii.totalPrice
}

// String returns the string representation of the invoice item.
func (ii *InvoiceItem) String() string {
	return ii.name + " x" + ii.quantity.String() + " @ " + ii.unitPrice.String() + " = " + ii.totalPrice.String()
}

// Equals returns true if this invoice item equals the other.
func (ii *InvoiceItem) Equals(other *InvoiceItem) bool {
	if other == nil {
		return false
	}
	return ii.name == other.name &&
		ii.description == other.description &&
		ii.quantity.Equal(other.quantity) &&
		ii.unitPrice.Equals(other.unitPrice) &&
		ii.totalPrice.Equals(other.totalPrice)
}

// InvoiceExpiration represents invoice expiration settings.
type InvoiceExpiration struct {
	expiresAt time.Time
	duration  time.Duration
}

// NewInvoiceExpiration creates a new InvoiceExpiration.
func NewInvoiceExpiration(duration time.Duration) *InvoiceExpiration {
	expiresAt := time.Now().UTC().Add(duration)
	return &InvoiceExpiration{
		expiresAt: expiresAt,
		duration:  duration,
	}
}

// NewInvoiceExpirationWithTime creates a new InvoiceExpiration with a specific expiration time.
func NewInvoiceExpirationWithTime(expiresAt time.Time) (*InvoiceExpiration, error) {
	if expiresAt.Before(time.Now().UTC()) {
		return nil, errors.New("expiration time must be in the future")
	}

	duration := time.Until(expiresAt)
	return &InvoiceExpiration{
		expiresAt: expiresAt,
		duration:  duration,
	}, nil
}

// ExpiresAt returns the expiration time.
func (ie *InvoiceExpiration) ExpiresAt() time.Time {
	return ie.expiresAt
}

// Duration returns the expiration duration.
func (ie *InvoiceExpiration) Duration() time.Duration {
	return ie.duration
}

// IsExpired returns true if the invoice has expired.
func (ie *InvoiceExpiration) IsExpired() bool {
	return time.Now().UTC().After(ie.expiresAt)
}

// TimeRemaining returns the time remaining until expiration.
func (ie *InvoiceExpiration) TimeRemaining() time.Duration {
	remaining := time.Until(ie.expiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// String returns the string representation of the invoice expiration.
func (ie *InvoiceExpiration) String() string {
	return "Expires at: " + ie.expiresAt.Format(time.RFC3339) + " (in " + ie.duration.String() + ")"
}

// Equals returns true if this invoice expiration equals the other.
func (ie *InvoiceExpiration) Equals(other *InvoiceExpiration) bool {
	if other == nil {
		return false
	}
	return ie.expiresAt.Equal(other.expiresAt) && ie.duration == other.duration
}
