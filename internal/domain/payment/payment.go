package payment

import (
	"time"

	"crypto-checkout/internal/domain/shared"

	"github.com/go-playground/validator/v10"
)

// Payment represents a blockchain payment transaction.
type Payment struct {
	id                    shared.PaymentID
	invoiceID             shared.InvoiceID
	amount                *PaymentAmount
	fromAddress           string
	toAddress             *PaymentAddress
	transactionHash       *TransactionHash
	status                PaymentStatus
	confirmations         *ConfirmationCount
	requiredConfirmations int
	blockInfo             *BlockInfo
	networkFee            *NetworkFee
	timestamps            *PaymentTimestamps
}

// PaymentValidation represents the validation structure for Payment creation.
type PaymentValidation struct {
	ID                    string           `validate:"required"`
	InvoiceID             string           `validate:"required"`
	Amount                *PaymentAmount   `validate:"required"`
	FromAddress           string           `validate:"required"`
	ToAddress             *PaymentAddress  `validate:"required"`
	TransactionHash       *TransactionHash `validate:"required"`
	RequiredConfirmations int              `validate:"min=0"`
}

// NewPayment creates a new payment with validation.
func NewPayment(
	id shared.PaymentID,
	invoiceID shared.InvoiceID,
	amount *PaymentAmount,
	fromAddress string,
	toAddress *PaymentAddress,
	transactionHash *TransactionHash,
	requiredConfirmations int,
) (*Payment, error) {
	// Create validation struct
	validation := PaymentValidation{
		ID:                    string(id),
		InvoiceID:             string(invoiceID),
		Amount:                amount,
		FromAddress:           fromAddress,
		ToAddress:             toAddress,
		TransactionHash:       transactionHash,
		RequiredConfirmations: requiredConfirmations,
	}

	// Validate using go-playground/validator
	validate := validator.New()
	if err := validate.Struct(validation); err != nil {
		// Convert validator errors to our custom error format
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				switch fieldErr.Field() {
				case "ID":
					return nil, NewPaymentError(shared.ErrCodeValidationFailed, "payment ID cannot be empty", nil)
				case "InvoiceID":
					return nil, NewPaymentError(shared.ErrCodeValidationFailed, "invoice ID cannot be empty", nil)
				case "Amount":
					return nil, NewPaymentError(shared.ErrCodeValidationFailed, "payment amount cannot be nil", nil)
				case "FromAddress":
					return nil, NewPaymentError(shared.ErrCodeValidationFailed, "from address cannot be empty", nil)
				case "ToAddress":
					return nil, NewPaymentError(shared.ErrCodeValidationFailed, "to address cannot be nil", nil)
				case "TransactionHash":
					return nil, NewPaymentError(shared.ErrCodeValidationFailed, "transaction hash cannot be nil", nil)
				case "RequiredConfirmations":
					return nil, NewPaymentError(shared.ErrCodeValidationFailed, "required confirmations cannot be negative", nil)
				}
			}
		}
		return nil, NewPaymentError(shared.ErrCodeValidationFailed, "validation failed", err)
	}

	confirmations, err := NewConfirmationCount(0)
	if err != nil {
		return nil, err
	}

	timestamps := NewPaymentTimestamps(time.Now().UTC())

	return &Payment{
		id:                    id,
		invoiceID:             invoiceID,
		amount:                amount,
		fromAddress:           fromAddress,
		toAddress:             toAddress,
		transactionHash:       transactionHash,
		status:                StatusDetected,
		confirmations:         confirmations,
		requiredConfirmations: requiredConfirmations,
		timestamps:            timestamps,
	}, nil
}

// ID returns the payment ID.
func (p *Payment) ID() shared.PaymentID {
	return p.id
}

// InvoiceID returns the invoice ID.
func (p *Payment) InvoiceID() shared.InvoiceID {
	return p.invoiceID
}

// Amount returns the payment amount.
func (p *Payment) Amount() *PaymentAmount {
	return p.amount
}

// FromAddress returns the sender address.
func (p *Payment) FromAddress() string {
	return p.fromAddress
}

// ToAddress returns the recipient address.
func (p *Payment) ToAddress() *PaymentAddress {
	return p.toAddress
}

// TransactionHash returns the transaction hash.
func (p *Payment) TransactionHash() *TransactionHash {
	return p.transactionHash
}

// Status returns the payment status.
func (p *Payment) Status() PaymentStatus {
	return p.status
}

// Confirmations returns the confirmation count.
func (p *Payment) Confirmations() *ConfirmationCount {
	return p.confirmations
}

// RequiredConfirmations returns the required confirmation count.
func (p *Payment) RequiredConfirmations() int {
	return p.requiredConfirmations
}

// BlockInfo returns the block information.
func (p *Payment) BlockInfo() *BlockInfo {
	return p.blockInfo
}

// NetworkFee returns the network fee.
func (p *Payment) NetworkFee() *NetworkFee {
	return p.networkFee
}

// DetectedAt returns when the payment was detected.
func (p *Payment) DetectedAt() time.Time {
	return p.timestamps.DetectedAt()
}

// ConfirmedAt returns when the payment was confirmed.
func (p *Payment) ConfirmedAt() *time.Time {
	return p.timestamps.ConfirmedAt()
}

// CreatedAt returns when the payment was created.
func (p *Payment) CreatedAt() time.Time {
	return p.timestamps.CreatedAt()
}

// UpdatedAt returns when the payment was last updated.
func (p *Payment) UpdatedAt() time.Time {
	return p.timestamps.UpdatedAt()
}

// IsConfirmed returns true if the payment has sufficient confirmations.
func (p *Payment) IsConfirmed() bool {
	return p.confirmations.Int() >= p.requiredConfirmations
}

// IsTerminal returns true if the payment is in a terminal state.
func (p *Payment) IsTerminal() bool {
	return p.status.IsTerminal()
}

// IsActive returns true if the payment is in an active state.
func (p *Payment) IsActive() bool {
	return p.status.IsActive()
}

// UpdateConfirmations updates the confirmation count.
func (p *Payment) UpdateConfirmations(ctx interface{}, count int) error {
	if count < 0 {
		return NewInvalidConfirmationCountError("confirmation count cannot be negative")
	}

	newConfirmations, err := NewConfirmationCount(count)
	if err != nil {
		return err
	}

	p.confirmations = newConfirmations
	p.timestamps.SetUpdatedAt(time.Now().UTC())

	return nil
}

// UpdateBlockInfo updates the block information.
func (p *Payment) UpdateBlockInfo(blockNumber int64, blockHash string) error {
	blockInfo, err := NewBlockInfo(blockNumber, blockHash)
	if err != nil {
		return err
	}

	p.blockInfo = blockInfo
	p.timestamps.SetUpdatedAt(time.Now().UTC())

	return nil
}

// UpdateNetworkFee updates the network fee.
func (p *Payment) UpdateNetworkFee(fee *shared.Money, currency shared.CryptoCurrency) error {
	networkFee, err := NewNetworkFee(fee, currency)
	if err != nil {
		return err
	}

	p.networkFee = networkFee
	p.timestamps.SetUpdatedAt(time.Now().UTC())

	return nil
}

// SetStatus sets the payment status (for testing purposes).
func (p *Payment) SetStatus(status PaymentStatus) {
	p.status = status
	p.timestamps.SetUpdatedAt(time.Now().UTC())
}

// SetConfirmations sets the confirmation count (for testing purposes).
func (p *Payment) SetConfirmations(count int) error {
	confirmations, err := NewConfirmationCount(count)
	if err != nil {
		return err
	}
	p.confirmations = confirmations
	p.timestamps.SetUpdatedAt(time.Now().UTC())
	return nil
}

// SetConfirmedAt sets the confirmation timestamp (for testing purposes).
func (p *Payment) SetConfirmedAt(confirmedAt time.Time) {
	p.timestamps.SetConfirmedAt(confirmedAt)
}
