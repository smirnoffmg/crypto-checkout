package invoice

import (
	"context"
	"crypto-checkout/internal/domain/shared"
	"errors"
	"time"
)

// InvoiceService defines the interface for invoice business operations.
type InvoiceService interface {
	// CreateInvoice creates a new invoice with the given parameters.
	CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*Invoice, error)

	// GetInvoice retrieves an invoice by ID.
	GetInvoice(ctx context.Context, id string) (*Invoice, error)

	// GetInvoiceByPaymentAddress retrieves an invoice by payment address.
	GetInvoiceByPaymentAddress(ctx context.Context, address *shared.PaymentAddress) (*Invoice, error)

	// ListInvoices retrieves invoices with the given filters.
	ListInvoices(ctx context.Context, req *ListInvoicesRequest) (*ListInvoicesResponse, error)

	// MarkInvoiceAsViewed marks an invoice as viewed by the customer.
	MarkInvoiceAsViewed(ctx context.Context, id string) error

	// CancelInvoice cancels an invoice.
	CancelInvoice(ctx context.Context, id string, reason string) error

	// ProcessPayment processes a payment for an invoice.
	ProcessPayment(ctx context.Context, invoiceID string, payment *Payment) error

	// GetExpiredInvoices retrieves invoices that have expired.
	GetExpiredInvoices(ctx context.Context) ([]*Invoice, error)

	// ProcessExpiredInvoices processes expired invoices.
	ProcessExpiredInvoices(ctx context.Context) error

	// GetInvoiceStatus returns the current status of an invoice.
	GetInvoiceStatus(ctx context.Context, id string) (InvoiceStatus, error)

	// UpdateInvoiceStatus updates the status of an invoice.
	UpdateInvoiceStatus(ctx context.Context, id string, newStatus InvoiceStatus, reason string) error
}

// CreateInvoiceRequest represents the request to create a new invoice.
type CreateInvoiceRequest struct {
	MerchantID         string
	CustomerID         *string
	Title              string
	Description        string
	Items              []*CreateInvoiceItemRequest
	Tax                *shared.Money
	Currency           shared.Currency
	CryptoCurrency     shared.CryptoCurrency
	PaymentTolerance   *PaymentTolerance
	ExpirationDuration time.Duration
	Metadata           map[string]interface{}
	WebhookURL         *string
	ReturnURL          *string
	CancelURL          *string
}

// CreateInvoiceItemRequest represents a request to create an invoice item.
type CreateInvoiceItemRequest struct {
	Name        string
	Description string
	Quantity    string
	UnitPrice   *shared.Money
}

// ListInvoicesRequest represents the request to list invoices.
type ListInvoicesRequest struct {
	MerchantID    string
	Status        *InvoiceStatus
	CustomerID    *string
	Limit         int
	Offset        int
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Search        *string
}

// ListInvoicesResponse represents the response to list invoices.
type ListInvoicesResponse struct {
	Invoices []*Invoice
	Total    int
	Limit    int
	Offset   int
}

// Payment represents a payment made to an invoice.
type Payment struct {
	id                    string
	invoiceID             string
	transactionHash       *shared.TransactionHash
	amount                *shared.Money
	fromAddress           string
	toAddress             string
	status                PaymentStatus
	confirmations         *shared.ConfirmationCount
	requiredConfirmations *shared.ConfirmationCount
	blockNumber           *int
	blockHash             *string
	networkFee            *shared.Money
	detectedAt            time.Time
	confirmedAt           *time.Time
}

// PaymentStatus represents the status of a payment.
type PaymentStatus string

const (
	// PaymentStatusDetected - Payment detected but not yet confirmed
	PaymentStatusDetected PaymentStatus = "detected"
	// PaymentStatusConfirming - Payment is gaining confirmations
	PaymentStatusConfirming PaymentStatus = "confirming"
	// PaymentStatusConfirmed - Payment is fully confirmed
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	// PaymentStatusFailed - Payment failed or was reverted
	PaymentStatusFailed PaymentStatus = "failed"
	// PaymentStatusOrphaned - Payment was orphaned due to blockchain reorganization
	PaymentStatusOrphaned PaymentStatus = "orphaned"
)

// String returns the string representation of the payment status.
func (ps PaymentStatus) String() string {
	return string(ps)
}

// IsValid returns true if the payment status is valid.
func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusDetected,
		PaymentStatusConfirming,
		PaymentStatusConfirmed,
		PaymentStatusFailed,
		PaymentStatusOrphaned:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the payment status is terminal.
func (ps PaymentStatus) IsTerminal() bool {
	switch ps {
	case PaymentStatusConfirmed, PaymentStatusFailed:
		return true
	default:
		return false
	}
}

// CanTransitionTo returns true if the payment status can transition to the target status.
func (ps PaymentStatus) CanTransitionTo(target PaymentStatus) bool {
	if !target.IsValid() {
		return false
	}

	// Terminal states cannot transition
	if ps.IsTerminal() {
		return false
	}

	// Define valid transitions
	validTransitions := map[PaymentStatus][]PaymentStatus{
		PaymentStatusDetected:   {PaymentStatusConfirming, PaymentStatusFailed},
		PaymentStatusConfirming: {PaymentStatusConfirmed, PaymentStatusOrphaned},
		PaymentStatusOrphaned:   {PaymentStatusDetected, PaymentStatusFailed},
	}

	if transitions, exists := validTransitions[ps]; exists {
		for _, validTarget := range transitions {
			if validTarget == target {
				return true
			}
		}
	}

	return false
}

// NewPayment creates a new Payment.
func NewPayment(
	id, invoiceID string,
	transactionHash *shared.TransactionHash,
	amount *shared.Money,
	fromAddress, toAddress string,
	requiredConfirmations *shared.ConfirmationCount,
) (*Payment, error) {
	if id == "" {
		return nil, errors.New("payment ID cannot be empty")
	}

	if invoiceID == "" {
		return nil, errors.New("invoice ID cannot be empty")
	}

	if transactionHash == nil {
		return nil, errors.New("transaction hash cannot be nil")
	}

	if amount == nil {
		return nil, errors.New("payment amount cannot be nil")
	}

	if fromAddress == "" {
		return nil, errors.New("from address cannot be empty")
	}

	if toAddress == "" {
		return nil, errors.New("to address cannot be empty")
	}

	if requiredConfirmations == nil {
		return nil, errors.New("required confirmations cannot be nil")
	}

	now := time.Now().UTC()
	zeroConfirmations, _ := shared.NewConfirmationCount(0)
	return &Payment{
		id:                    id,
		invoiceID:             invoiceID,
		transactionHash:       transactionHash,
		amount:                amount,
		fromAddress:           fromAddress,
		toAddress:             toAddress,
		status:                PaymentStatusDetected,
		confirmations:         zeroConfirmations,
		requiredConfirmations: requiredConfirmations,
		detectedAt:            now,
	}, nil
}

// ID returns the payment ID.
func (p *Payment) ID() string {
	return p.id
}

// InvoiceID returns the invoice ID.
func (p *Payment) InvoiceID() string {
	return p.invoiceID
}

// TransactionHash returns the transaction hash.
func (p *Payment) TransactionHash() *shared.TransactionHash {
	return p.transactionHash
}

// Amount returns the payment amount.
func (p *Payment) Amount() *shared.Money {
	return p.amount
}

// FromAddress returns the sender address.
func (p *Payment) FromAddress() string {
	return p.fromAddress
}

// ToAddress returns the recipient address.
func (p *Payment) ToAddress() string {
	return p.toAddress
}

// Status returns the payment status.
func (p *Payment) Status() PaymentStatus {
	return p.status
}

// Confirmations returns the current confirmation count.
func (p *Payment) Confirmations() *shared.ConfirmationCount {
	return p.confirmations
}

// RequiredConfirmations returns the required confirmation count.
func (p *Payment) RequiredConfirmations() *shared.ConfirmationCount {
	return p.requiredConfirmations
}

// BlockNumber returns the block number if confirmed.
func (p *Payment) BlockNumber() *int {
	return p.blockNumber
}

// BlockHash returns the block hash if confirmed.
func (p *Payment) BlockHash() *string {
	return p.blockHash
}

// NetworkFee returns the network fee.
func (p *Payment) NetworkFee() *shared.Money {
	return p.networkFee
}

// DetectedAt returns when the payment was detected.
func (p *Payment) DetectedAt() time.Time {
	return p.detectedAt
}

// ConfirmedAt returns when the payment was confirmed.
func (p *Payment) ConfirmedAt() *time.Time {
	return p.confirmedAt
}

// IsConfirmed returns true if the payment is confirmed.
func (p *Payment) IsConfirmed() bool {
	return p.status == PaymentStatusConfirmed
}

// IsTerminal returns true if the payment is in a terminal state.
func (p *Payment) IsTerminal() bool {
	return p.status.IsTerminal()
}

// CanTransitionTo returns true if the payment can transition to the target status.
func (p *Payment) CanTransitionTo(target PaymentStatus) bool {
	return p.status.CanTransitionTo(target)
}

// TransitionTo transitions the payment to a new status.
func (p *Payment) TransitionTo(newStatus PaymentStatus) error {
	if !p.CanTransitionTo(newStatus) {
		return errors.New("invalid payment status transition from " + p.status.String() + " to " + newStatus.String())
	}

	p.status = newStatus

	// Set confirmedAt when transitioning to confirmed status
	if newStatus == PaymentStatusConfirmed && p.confirmedAt == nil {
		now := time.Now().UTC()
		p.confirmedAt = &now
	}

	return nil
}

// UpdateConfirmations updates the confirmation count.
func (p *Payment) UpdateConfirmations(count *shared.ConfirmationCount) error {
	if count == nil {
		return errors.New("confirmation count cannot be nil")
	}

	p.confirmations = count

	// Auto-transition to confirmed if we have enough confirmations
	if p.confirmations.IsGreaterThanOrEqual(p.requiredConfirmations) {
		return p.TransitionTo(PaymentStatusConfirmed)
	}

	// Auto-transition to confirming if we have at least 1 confirmation
	if p.confirmations.Count() > 0 && p.status == PaymentStatusDetected {
		return p.TransitionTo(PaymentStatusConfirming)
	}

	return nil
}

// String returns the string representation of the payment.
func (p *Payment) String() string {
	return "Payment[" + p.id + "] " + p.amount.String() + " - " + p.status.String()
}

// Equals returns true if this payment equals the other.
func (p *Payment) Equals(other *Payment) bool {
	if other == nil {
		return false
	}
	return p.id == other.id
}
