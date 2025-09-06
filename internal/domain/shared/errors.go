package shared

import "errors"

// Common domain errors that can be shared across different domains

var (
	// Generic validation errors
	ErrInvalidID          = errors.New("invalid ID")
	ErrInvalidTitle       = errors.New("invalid title")
	ErrInvalidDescription = errors.New("invalid description")
	ErrInvalidAmount      = errors.New("invalid amount")
	ErrInvalidCurrency    = errors.New("invalid currency")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrInvalidInput       = errors.New("invalid input")

	// Money and currency errors
	ErrInvalidMoneyAmount  = errors.New("invalid money amount")
	ErrCurrencyMismatch    = errors.New("currency mismatch")
	ErrNegativeAmount      = errors.New("amount cannot be negative")
	ErrZeroAmount          = errors.New("amount cannot be zero")
	ErrInvalidAmountFormat = errors.New("invalid amount format")

	// Payment and blockchain errors
	ErrInvalidPaymentAddress    = errors.New("invalid payment address")
	ErrInvalidTransactionHash   = errors.New("invalid transaction hash")
	ErrInvalidNetwork           = errors.New("invalid blockchain network")
	ErrExpiredPaymentAddress    = errors.New("payment address has expired")
	ErrExpiredExchangeRate      = errors.New("exchange rate has expired")
	ErrInvalidExchangeRate      = errors.New("invalid exchange rate")
	ErrInvalidConfirmationCount = errors.New("invalid confirmation count")

	// Service and repository errors
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrRepositoryError   = errors.New("repository error")
	ErrServiceError      = errors.New("service error")
	ErrIDGenerationError = errors.New("ID generation error")
	ErrInvalidRequest    = errors.New("invalid request")

	// State and lifecycle errors
	ErrInvalidState     = errors.New("invalid state")
	ErrCannotTransition = errors.New("cannot transition to target state")
	ErrAlreadyInState   = errors.New("already in target state")
	ErrExpired          = errors.New("expired")
	ErrTerminalState    = errors.New("cannot perform action in terminal state")

	// Business logic errors
	ErrInsufficientAmount    = errors.New("insufficient amount")
	ErrExcessiveAmount       = errors.New("excessive amount")
	ErrValidationFailed      = errors.New("validation failed")
	ErrBusinessRuleViolation = errors.New("business rule violation")
)

// DomainError represents a domain-specific error with additional context.
type DomainError struct {
	Code    string
	Message string
	Details map[string]interface{}
	Err     error
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new domain error.
func NewDomainError(code, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail to the error.
func (e *DomainError) WithDetail(key string, value interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails adds multiple details to the error.
func (e *DomainError) WithDetails(details map[string]interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// Common error codes
const (
	ErrCodeInvalidInput          = "INVALID_INPUT"
	ErrCodeInvalidStatus         = "INVALID_STATUS"
	ErrCodeInvalidTransition     = "INVALID_TRANSITION"
	ErrCodeNotFound              = "NOT_FOUND"
	ErrCodeAlreadyExists         = "ALREADY_EXISTS"
	ErrCodeExpired               = "EXPIRED"
	ErrCodeInsufficientAmount    = "INSUFFICIENT_AMOUNT"
	ErrCodeExcessiveAmount       = "EXCESSIVE_AMOUNT"
	ErrCodeServiceError          = "SERVICE_ERROR"
	ErrCodeRepositoryError       = "REPOSITORY_ERROR"
	ErrCodeValidationFailed      = "VALIDATION_FAILED"
	ErrCodeBusinessRuleViolation = "BUSINESS_RULE_VIOLATION"
	ErrCodeCurrencyMismatch      = "CURRENCY_MISMATCH"
	ErrCodeInvalidState          = "INVALID_STATE"
	ErrCodeTerminalState         = "TERMINAL_STATE"
)

// Error constructors for common patterns

// NewValidationError creates a validation error.
func NewValidationError(field string, reason string) *DomainError {
	return NewDomainError(ErrCodeValidationFailed, "validation failed", nil).
		WithDetail("field", field).
		WithDetail("reason", reason)
}

// NewNotFoundError creates a not found error.
func NewNotFoundError(resource string, id string) *DomainError {
	return NewDomainError(ErrCodeNotFound, resource+" not found", nil).
		WithDetail("resource", resource).
		WithDetail("id", id)
}

// NewInvalidTransitionError creates an invalid transition error.
func NewInvalidTransitionError(from, to string) *DomainError {
	return NewDomainError(ErrCodeInvalidTransition, "invalid transition", nil).
		WithDetail("from", from).
		WithDetail("to", to)
}

// NewBusinessRuleViolationError creates a business rule violation error.
func NewBusinessRuleViolationError(rule string, reason string) *DomainError {
	return NewDomainError(ErrCodeBusinessRuleViolation, "business rule violation", nil).
		WithDetail("rule", rule).
		WithDetail("reason", reason)
}

// NewCurrencyMismatchError creates a currency mismatch error.
func NewCurrencyMismatchError(expected, actual string) *DomainError {
	return NewDomainError(ErrCodeCurrencyMismatch, "currency mismatch", nil).
		WithDetail("expected", expected).
		WithDetail("actual", actual)
}

// NewInsufficientAmountError creates an insufficient amount error.
func NewInsufficientAmountError(required, provided string) *DomainError {
	return NewDomainError(ErrCodeInsufficientAmount, "insufficient amount", nil).
		WithDetail("required", required).
		WithDetail("provided", provided)
}

// NewExcessiveAmountError creates an excessive amount error.
func NewExcessiveAmountError(limit, provided string) *DomainError {
	return NewDomainError(ErrCodeExcessiveAmount, "excessive amount", nil).
		WithDetail("limit", limit).
		WithDetail("provided", provided)
}

// NewTerminalStateError creates a terminal state error.
func NewTerminalStateError(state string, action string) *DomainError {
	return NewDomainError(ErrCodeTerminalState, "cannot perform action in terminal state", nil).
		WithDetail("state", state).
		WithDetail("action", action)
}
