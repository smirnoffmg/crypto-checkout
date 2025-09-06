package invoice

import (
	"errors"

	"crypto-checkout/internal/domain/shared"
)

// Invoice-specific domain errors
var (
	// Invoice creation errors
	ErrInvalidInvoiceID        = errors.New("invalid invoice ID")
	ErrInvalidMerchantID       = errors.New("invalid merchant ID")
	ErrNoItems                 = errors.New("invoice must have at least one item")
	ErrInvalidPricing          = errors.New("invalid pricing")
	ErrInvalidCryptocurrency   = errors.New("invalid cryptocurrency")
	ErrInvalidPaymentTolerance = errors.New("invalid payment tolerance")
	ErrInvalidExpiration       = errors.New("invalid expiration")

	// Invoice status errors
	ErrInvoiceAlreadyViewed = errors.New("invoice already marked as viewed")
	ErrCannotViewInvoice    = errors.New("can only mark created invoices as viewed")
	ErrCannotCancelInvoice  = errors.New("cannot cancel invoice in terminal state")
	ErrCannotExpireInvoice  = errors.New("cannot auto-expire invoices with partial payments")
	ErrCannotMarkAsPaid     = errors.New("can only mark confirming invoices as paid")
	ErrCannotRefundInvoice  = errors.New("can only refund paid invoices")

	// Invoice item errors
	ErrInvalidItemName        = errors.New("invalid item name")
	ErrInvalidItemDescription = errors.New("invalid item description")
	ErrInvalidQuantity        = errors.New("invalid quantity")
	ErrInvalidUnitPrice       = errors.New("invalid unit price")
	ErrInvalidTotalPrice      = errors.New("invalid total price")

	// Payment tolerance errors
	ErrInvalidUnderpaymentThreshold = errors.New("invalid underpayment threshold")
	ErrInvalidOverpaymentThreshold  = errors.New("invalid overpayment threshold")
	ErrInvalidOverpaymentAction     = errors.New("invalid overpayment action")

	// Payment errors
	ErrInvalidPaymentID             = errors.New("invalid payment ID")
	ErrInvalidFromAddress           = errors.New("invalid from address")
	ErrInvalidToAddress             = errors.New("invalid to address")
	ErrInvalidRequiredConfirmations = errors.New("invalid required confirmations")
	ErrInvalidPaymentStatus         = errors.New("invalid payment status")
	ErrInvalidPaymentTransition     = errors.New("invalid payment status transition")
	ErrPaymentValidationFailed      = errors.New("payment validation failed")
	ErrUnderpayment                 = errors.New("payment amount is below the minimum threshold")

	// Service errors
	ErrInvoiceNotFound            = errors.New("invoice not found")
	ErrPaymentNotFound            = errors.New("payment not found")
	ErrInvalidCreateRequest       = errors.New("invalid create invoice request")
	ErrInvalidListRequest         = errors.New("invalid list invoices request")
	ErrExchangeRateServiceError   = errors.New("exchange rate service error")
	ErrPaymentAddressServiceError = errors.New("payment address service error")

	// Repository errors
	ErrInvoiceSaveError   = errors.New("failed to save invoice")
	ErrInvoiceUpdateError = errors.New("failed to update invoice")
	ErrInvoiceDeleteError = errors.New("failed to delete invoice")
	ErrInvoiceFindError   = errors.New("failed to find invoice")
	ErrInvoiceExistsError = errors.New("failed to check invoice existence")
)

// Re-export commonly used shared errors for convenience
var (
	// Generic validation errors
	ErrInvalidID          = shared.ErrInvalidID
	ErrInvalidTitle       = shared.ErrInvalidTitle
	ErrInvalidDescription = shared.ErrInvalidDescription
	ErrInvalidAmount      = shared.ErrInvalidAmount
	ErrInvalidCurrency    = shared.ErrInvalidCurrency
	ErrInvalidStatus      = shared.ErrInvalidStatus
	ErrInvalidTransition  = shared.ErrInvalidTransition
	ErrInvalidInput       = shared.ErrInvalidInput

	// Money and currency errors
	ErrInvalidMoneyAmount  = shared.ErrInvalidMoneyAmount
	ErrCurrencyMismatch    = shared.ErrCurrencyMismatch
	ErrNegativeAmount      = shared.ErrNegativeAmount
	ErrZeroAmount          = shared.ErrZeroAmount
	ErrInvalidAmountFormat = shared.ErrInvalidAmountFormat

	// Payment and blockchain errors
	ErrInvalidPaymentAddress    = shared.ErrInvalidPaymentAddress
	ErrInvalidTransactionHash   = shared.ErrInvalidTransactionHash
	ErrInvalidNetwork           = shared.ErrInvalidNetwork
	ErrExpiredPaymentAddress    = shared.ErrExpiredPaymentAddress
	ErrExpiredExchangeRate      = shared.ErrExpiredExchangeRate
	ErrInvalidExchangeRate      = shared.ErrInvalidExchangeRate
	ErrInvalidConfirmationCount = shared.ErrInvalidConfirmationCount

	// Service and repository errors
	ErrNotFound          = shared.ErrNotFound
	ErrAlreadyExists     = shared.ErrAlreadyExists
	ErrRepositoryError   = shared.ErrRepositoryError
	ErrServiceError      = shared.ErrServiceError
	ErrIDGenerationError = shared.ErrIDGenerationError
	ErrInvalidRequest    = shared.ErrInvalidRequest

	// State and lifecycle errors
	ErrInvalidState     = shared.ErrInvalidState
	ErrCannotTransition = shared.ErrCannotTransition
	ErrAlreadyInState   = shared.ErrAlreadyInState
	ErrExpired          = shared.ErrExpired
	ErrTerminalState    = shared.ErrTerminalState

	// Business logic errors
	ErrInsufficientAmount    = shared.ErrInsufficientAmount
	ErrExcessiveAmount       = shared.ErrExcessiveAmount
	ErrValidationFailed      = shared.ErrValidationFailed
	ErrBusinessRuleViolation = shared.ErrBusinessRuleViolation
)

// InvoiceError is an alias for shared.DomainError for invoice-specific errors.
type InvoiceError = shared.DomainError

// NewInvoiceError creates a new invoice error.
func NewInvoiceError(code, message string, err error) *InvoiceError {
	return shared.NewDomainError(code, message, err)
}

// Invoice-specific error codes
const (
	ErrCodeInvalidInvoiceID             = "INVALID_INVOICE_ID"
	ErrCodeInvalidMerchantID            = "INVALID_MERCHANT_ID"
	ErrCodeNoItems                      = "NO_ITEMS"
	ErrCodeInvalidPricing               = "INVALID_PRICING"
	ErrCodeInvalidCryptocurrency        = "INVALID_CRYPTOCURRENCY"
	ErrCodeInvalidPaymentTolerance      = "INVALID_PAYMENT_TOLERANCE"
	ErrCodeInvalidExpiration            = "INVALID_EXPIRATION"
	ErrCodeInvoiceAlreadyViewed         = "INVOICE_ALREADY_VIEWED"
	ErrCodeCannotViewInvoice            = "CANNOT_VIEW_INVOICE"
	ErrCodeCannotCancelInvoice          = "CANNOT_CANCEL_INVOICE"
	ErrCodeCannotExpireInvoice          = "CANNOT_EXPIRE_INVOICE"
	ErrCodeCannotMarkAsPaid             = "CANNOT_MARK_AS_PAID"
	ErrCodeCannotRefundInvoice          = "CANNOT_REFUND_INVOICE"
	ErrCodeInvalidItemName              = "INVALID_ITEM_NAME"
	ErrCodeInvalidItemDescription       = "INVALID_ITEM_DESCRIPTION"
	ErrCodeInvalidQuantity              = "INVALID_QUANTITY"
	ErrCodeInvalidUnitPrice             = "INVALID_UNIT_PRICE"
	ErrCodeInvalidTotalPrice            = "INVALID_TOTAL_PRICE"
	ErrCodeInvalidUnderpaymentThreshold = "INVALID_UNDERPAYMENT_THRESHOLD"
	ErrCodeInvalidOverpaymentThreshold  = "INVALID_OVERPAYMENT_THRESHOLD"
	ErrCodeInvalidOverpaymentAction     = "INVALID_OVERPAYMENT_ACTION"
	ErrCodeInvalidPaymentID             = "INVALID_PAYMENT_ID"
	ErrCodeInvalidFromAddress           = "INVALID_FROM_ADDRESS"
	ErrCodeInvalidToAddress             = "INVALID_TO_ADDRESS"
	ErrCodeInvalidRequiredConfirmations = "INVALID_REQUIRED_CONFIRMATIONS"
	ErrCodeInvalidPaymentStatus         = "INVALID_PAYMENT_STATUS"
	ErrCodeInvalidPaymentTransition     = "INVALID_PAYMENT_TRANSITION"
	ErrCodePaymentValidationFailed      = "PAYMENT_VALIDATION_FAILED"
	ErrCodeUnderpayment                 = "UNDERPAYMENT"
	ErrCodeInvoiceNotFound              = "INVOICE_NOT_FOUND"
	ErrCodePaymentNotFound              = "PAYMENT_NOT_FOUND"
	ErrCodeInvalidCreateRequest         = "INVALID_CREATE_REQUEST"
	ErrCodeInvalidListRequest           = "INVALID_LIST_REQUEST"
	ErrCodeExchangeRateServiceError     = "EXCHANGE_RATE_SERVICE_ERROR"
	ErrCodePaymentAddressServiceError   = "PAYMENT_ADDRESS_SERVICE_ERROR"
	ErrCodeInvoiceSaveError             = "INVOICE_SAVE_ERROR"
	ErrCodeInvoiceUpdateError           = "INVOICE_UPDATE_ERROR"
	ErrCodeInvoiceDeleteError           = "INVOICE_DELETE_ERROR"
	ErrCodeInvoiceFindError             = "INVOICE_FIND_ERROR"
	ErrCodeInvoiceExistsError           = "INVOICE_EXISTS_ERROR"
)

// Re-export shared error codes for convenience
const (
	ErrCodeInvalidInput          = shared.ErrCodeInvalidInput
	ErrCodeInvalidStatus         = shared.ErrCodeInvalidStatus
	ErrCodeInvalidTransition     = shared.ErrCodeInvalidTransition
	ErrCodeNotFound              = shared.ErrCodeNotFound
	ErrCodeAlreadyExists         = shared.ErrCodeAlreadyExists
	ErrCodeExpired               = shared.ErrCodeExpired
	ErrCodeInsufficientAmount    = shared.ErrCodeInsufficientAmount
	ErrCodeExcessiveAmount       = shared.ErrCodeExcessiveAmount
	ErrCodeServiceError          = shared.ErrCodeServiceError
	ErrCodeRepositoryError       = shared.ErrCodeRepositoryError
	ErrCodeValidationFailed      = shared.ErrCodeValidationFailed
	ErrCodeBusinessRuleViolation = shared.ErrCodeBusinessRuleViolation
	ErrCodeCurrencyMismatch      = shared.ErrCodeCurrencyMismatch
	ErrCodeInvalidState          = shared.ErrCodeInvalidState
	ErrCodeTerminalState         = shared.ErrCodeTerminalState
)
