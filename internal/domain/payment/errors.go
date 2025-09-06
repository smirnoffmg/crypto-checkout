package payment

import (
	"crypto-checkout/internal/domain/shared"
)

// PaymentError is an alias for shared.DomainError to maintain consistency.
type PaymentError = shared.DomainError

// NewPaymentError creates a new payment error using shared implementation.
func NewPaymentError(code, message string, err error) *PaymentError {
	return shared.NewDomainError(code, message, err)
}

// Re-export common errors from shared package
var (
	ErrInvalidPayment           = shared.ErrInvalidInput
	ErrPaymentNotFound          = shared.ErrNotFound
	ErrInvalidTransactionHash   = shared.ErrInvalidTransactionHash
	ErrInvalidAddress           = shared.ErrInvalidPaymentAddress
	ErrInvalidConfirmationCount = shared.ErrInvalidConfirmationCount
	ErrInvalidAmount            = shared.ErrInvalidAmount
	ErrInvalidStatus            = shared.ErrInvalidStatus
	ErrInvalidTransition        = shared.ErrInvalidTransition
	ErrTerminalState            = shared.ErrTerminalState
	ErrRepositoryError          = shared.ErrRepositoryError
	ErrServiceError             = shared.ErrServiceError
)

// Payment-specific error codes
const (
	ErrCodeInvalidPaymentStatus      = "INVALID_PAYMENT_STATUS"
	ErrCodeInvalidPaymentTransition  = "INVALID_PAYMENT_TRANSITION"
	ErrCodePaymentNotFound           = "PAYMENT_NOT_FOUND"
	ErrCodeInvalidPaymentAmount      = "INVALID_PAYMENT_AMOUNT"
	ErrCodeInvalidBlockInfo          = "INVALID_BLOCK_INFO"
	ErrCodeInvalidNetworkFee         = "INVALID_NETWORK_FEE"
	ErrCodeInsufficientConfirmations = "INSUFFICIENT_CONFIRMATIONS"
	ErrCodePaymentAlreadyExists      = "PAYMENT_ALREADY_EXISTS"
)

// Payment-specific error constructors

// NewInvalidPaymentStatusError creates an error for invalid payment status.
func NewInvalidPaymentStatusError(status string) *PaymentError {
	return NewPaymentError(ErrCodeInvalidPaymentStatus, "invalid payment status", nil).
		WithDetail("status", status)
}

// NewInvalidPaymentTransitionError creates an error for invalid payment transition.
func NewInvalidPaymentTransitionError(from, to string) *PaymentError {
	return NewPaymentError(ErrCodeInvalidPaymentTransition, "invalid payment transition", nil).
		WithDetail("from", from).
		WithDetail("to", to)
}

// NewPaymentNotFoundError creates an error for payment not found.
func NewPaymentNotFoundError(id string) *PaymentError {
	return NewPaymentError(ErrCodePaymentNotFound, "payment not found", nil).
		WithDetail("payment_id", id)
}

// NewInvalidPaymentAmountError creates an error for invalid payment amount.
func NewInvalidPaymentAmountError(amount string) *PaymentError {
	return NewPaymentError(ErrCodeInvalidPaymentAmount, "invalid payment amount", nil).
		WithDetail("amount", amount)
}

// NewInvalidBlockInfoError creates an error for invalid block information.
func NewInvalidBlockInfoError(reason string) *PaymentError {
	return NewPaymentError(ErrCodeInvalidBlockInfo, "invalid block information", nil).
		WithDetail("reason", reason)
}

// NewInvalidNetworkFeeError creates an error for invalid network fee.
func NewInvalidNetworkFeeError(fee string) *PaymentError {
	return NewPaymentError(ErrCodeInvalidNetworkFee, "invalid network fee", nil).
		WithDetail("fee", fee)
}

// NewInsufficientConfirmationsError creates an error for insufficient confirmations.
func NewInsufficientConfirmationsError(current, required int) *PaymentError {
	return NewPaymentError(ErrCodeInsufficientConfirmations, "insufficient confirmations", nil).
		WithDetail("current", current).
		WithDetail("required", required)
}

// NewPaymentAlreadyExistsError creates an error for duplicate payment.
func NewPaymentAlreadyExistsError(txHash string) *PaymentError {
	return NewPaymentError(ErrCodePaymentAlreadyExists, "payment already exists", nil).
		WithDetail("transaction_hash", txHash)
}

// NewInvalidTransactionHashError creates an error for invalid transaction hash.
func NewInvalidTransactionHashError(hash string) *PaymentError {
	return NewPaymentError(shared.ErrCodeValidationFailed, "invalid transaction hash", nil).
		WithDetail("hash", hash)
}

// NewInvalidAddressError creates an error for invalid payment address.
func NewInvalidAddressError(address string) *PaymentError {
	return NewPaymentError(shared.ErrCodeValidationFailed, "invalid payment address", nil).
		WithDetail("address", address)
}

// NewInvalidConfirmationCountError creates an error for invalid confirmation count.
func NewInvalidConfirmationCountError(count string) *PaymentError {
	return NewPaymentError(shared.ErrCodeValidationFailed, "invalid confirmation count", nil).
		WithDetail("count", count)
}
