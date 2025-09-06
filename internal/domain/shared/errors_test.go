package shared_test

import (
	"errors"
	"testing"

	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/assert"
)

func TestDomainError(t *testing.T) {
	t.Run("NewDomainError - with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		domainErr := shared.NewDomainError("TEST_CODE", "test message", underlyingErr)

		assert.Equal(t, "TEST_CODE", domainErr.Code)
		assert.Equal(t, "test message", domainErr.Message)
		assert.Equal(t, underlyingErr, domainErr.Err)
		assert.NotNil(t, domainErr.Details)
	})

	t.Run("NewDomainError - without underlying error", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)

		assert.Equal(t, "TEST_CODE", domainErr.Code)
		assert.Equal(t, "test message", domainErr.Message)
		assert.Nil(t, domainErr.Err)
		assert.NotNil(t, domainErr.Details)
	})

	t.Run("Error - with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		domainErr := shared.NewDomainError("TEST_CODE", "test message", underlyingErr)

		expected := "test message: underlying error"
		assert.Equal(t, expected, domainErr.Error())
	})

	t.Run("Error - without underlying error", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)

		assert.Equal(t, "test message", domainErr.Error())
	})

	t.Run("Unwrap - with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		domainErr := shared.NewDomainError("TEST_CODE", "test message", underlyingErr)

		assert.Equal(t, underlyingErr, domainErr.Unwrap())
	})

	t.Run("Unwrap - without underlying error", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)

		assert.Nil(t, domainErr.Unwrap())
	})

	t.Run("WithDetail", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)
		domainErr = domainErr.WithDetail("key1", "value1")

		assert.Equal(t, "value1", domainErr.Details["key1"])
	})

	t.Run("WithDetails", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)
		details := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		domainErr = domainErr.WithDetails(details)

		assert.Equal(t, "value1", domainErr.Details["key1"])
		assert.Equal(t, "value2", domainErr.Details["key2"])
	})

	t.Run("WithDetail - chaining", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)
		domainErr = domainErr.WithDetail("key1", "value1").WithDetail("key2", "value2")

		assert.Equal(t, "value1", domainErr.Details["key1"])
		assert.Equal(t, "value2", domainErr.Details["key2"])
	})
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		err := shared.NewValidationError("field", "reason")

		assert.Equal(t, shared.ErrCodeValidationFailed, err.Code)
		assert.Equal(t, "validation failed", err.Message)
		assert.Equal(t, "field", err.Details["field"])
		assert.Equal(t, "reason", err.Details["reason"])
	})

	t.Run("NewNotFoundError", func(t *testing.T) {
		err := shared.NewNotFoundError("resource", "id123")

		assert.Equal(t, shared.ErrCodeNotFound, err.Code)
		assert.Equal(t, "resource not found", err.Message)
		assert.Equal(t, "resource", err.Details["resource"])
		assert.Equal(t, "id123", err.Details["id"])
	})

	t.Run("NewInvalidTransitionError", func(t *testing.T) {
		err := shared.NewInvalidTransitionError("from_state", "to_state")

		assert.Equal(t, shared.ErrCodeInvalidTransition, err.Code)
		assert.Equal(t, "invalid transition", err.Message)
		assert.Equal(t, "from_state", err.Details["from"])
		assert.Equal(t, "to_state", err.Details["to"])
	})

	t.Run("NewBusinessRuleViolationError", func(t *testing.T) {
		err := shared.NewBusinessRuleViolationError("rule", "reason")

		assert.Equal(t, shared.ErrCodeBusinessRuleViolation, err.Code)
		assert.Equal(t, "business rule violation", err.Message)
		assert.Equal(t, "rule", err.Details["rule"])
		assert.Equal(t, "reason", err.Details["reason"])
	})

	t.Run("NewCurrencyMismatchError", func(t *testing.T) {
		err := shared.NewCurrencyMismatchError("USD", "EUR")

		assert.Equal(t, shared.ErrCodeCurrencyMismatch, err.Code)
		assert.Equal(t, "currency mismatch", err.Message)
		assert.Equal(t, "USD", err.Details["expected"])
		assert.Equal(t, "EUR", err.Details["actual"])
	})

	t.Run("NewInsufficientAmountError", func(t *testing.T) {
		err := shared.NewInsufficientAmountError("100.00", "50.00")

		assert.Equal(t, shared.ErrCodeInsufficientAmount, err.Code)
		assert.Equal(t, "insufficient amount", err.Message)
		assert.Equal(t, "100.00", err.Details["required"])
		assert.Equal(t, "50.00", err.Details["provided"])
	})

	t.Run("NewExcessiveAmountError", func(t *testing.T) {
		err := shared.NewExcessiveAmountError("100.00", "150.00")

		assert.Equal(t, shared.ErrCodeExcessiveAmount, err.Code)
		assert.Equal(t, "excessive amount", err.Message)
		assert.Equal(t, "100.00", err.Details["limit"])
		assert.Equal(t, "150.00", err.Details["provided"])
	})

	t.Run("NewTerminalStateError", func(t *testing.T) {
		err := shared.NewTerminalStateError("paid", "cancel")

		assert.Equal(t, shared.ErrCodeTerminalState, err.Code)
		assert.Equal(t, "cannot perform action in terminal state", err.Message)
		assert.Equal(t, "paid", err.Details["state"])
		assert.Equal(t, "cancel", err.Details["action"])
	})
}

func TestSharedErrors(t *testing.T) {
	t.Run("Generic validation errors", func(t *testing.T) {
		assert.Equal(t, "invalid ID", shared.ErrInvalidID.Error())
		assert.Equal(t, "invalid title", shared.ErrInvalidTitle.Error())
		assert.Equal(t, "invalid description", shared.ErrInvalidDescription.Error())
		assert.Equal(t, "invalid amount", shared.ErrInvalidAmount.Error())
		assert.Equal(t, "invalid currency", shared.ErrInvalidCurrency.Error())
		assert.Equal(t, "invalid status", shared.ErrInvalidStatus.Error())
		assert.Equal(t, "invalid status transition", shared.ErrInvalidTransition.Error())
		assert.Equal(t, "invalid input", shared.ErrInvalidInput.Error())
	})

	t.Run("Money and currency errors", func(t *testing.T) {
		assert.Equal(t, "invalid money amount", shared.ErrInvalidMoneyAmount.Error())
		assert.Equal(t, "currency mismatch", shared.ErrCurrencyMismatch.Error())
		assert.Equal(t, "amount cannot be negative", shared.ErrNegativeAmount.Error())
		assert.Equal(t, "amount cannot be zero", shared.ErrZeroAmount.Error())
		assert.Equal(t, "invalid amount format", shared.ErrInvalidAmountFormat.Error())
	})

	t.Run("Payment and blockchain errors", func(t *testing.T) {
		assert.Equal(t, "invalid payment address", shared.ErrInvalidPaymentAddress.Error())
		assert.Equal(t, "invalid transaction hash", shared.ErrInvalidTransactionHash.Error())
		assert.Equal(t, "invalid blockchain network", shared.ErrInvalidNetwork.Error())
		assert.Equal(t, "payment address has expired", shared.ErrExpiredPaymentAddress.Error())
		assert.Equal(t, "exchange rate has expired", shared.ErrExpiredExchangeRate.Error())
		assert.Equal(t, "invalid exchange rate", shared.ErrInvalidExchangeRate.Error())
		assert.Equal(t, "invalid confirmation count", shared.ErrInvalidConfirmationCount.Error())
	})

	t.Run("Service and repository errors", func(t *testing.T) {
		assert.Equal(t, "not found", shared.ErrNotFound.Error())
		assert.Equal(t, "already exists", shared.ErrAlreadyExists.Error())
		assert.Equal(t, "repository error", shared.ErrRepositoryError.Error())
		assert.Equal(t, "service error", shared.ErrServiceError.Error())
		assert.Equal(t, "ID generation error", shared.ErrIDGenerationError.Error())
		assert.Equal(t, "invalid request", shared.ErrInvalidRequest.Error())
	})

	t.Run("State and lifecycle errors", func(t *testing.T) {
		assert.Equal(t, "invalid state", shared.ErrInvalidState.Error())
		assert.Equal(t, "cannot transition to target state", shared.ErrCannotTransition.Error())
		assert.Equal(t, "already in target state", shared.ErrAlreadyInState.Error())
		assert.Equal(t, "expired", shared.ErrExpired.Error())
		assert.Equal(t, "cannot perform action in terminal state", shared.ErrTerminalState.Error())
	})

	t.Run("Business logic errors", func(t *testing.T) {
		assert.Equal(t, "insufficient amount", shared.ErrInsufficientAmount.Error())
		assert.Equal(t, "excessive amount", shared.ErrExcessiveAmount.Error())
		assert.Equal(t, "validation failed", shared.ErrValidationFailed.Error())
		assert.Equal(t, "business rule violation", shared.ErrBusinessRuleViolation.Error())
	})
}

func TestErrorCodes(t *testing.T) {
	t.Run("Error codes are defined", func(t *testing.T) {
		assert.Equal(t, "INVALID_INPUT", shared.ErrCodeInvalidInput)
		assert.Equal(t, "INVALID_STATUS", shared.ErrCodeInvalidStatus)
		assert.Equal(t, "INVALID_TRANSITION", shared.ErrCodeInvalidTransition)
		assert.Equal(t, "NOT_FOUND", shared.ErrCodeNotFound)
		assert.Equal(t, "ALREADY_EXISTS", shared.ErrCodeAlreadyExists)
		assert.Equal(t, "EXPIRED", shared.ErrCodeExpired)
		assert.Equal(t, "INSUFFICIENT_AMOUNT", shared.ErrCodeInsufficientAmount)
		assert.Equal(t, "EXCESSIVE_AMOUNT", shared.ErrCodeExcessiveAmount)
		assert.Equal(t, "SERVICE_ERROR", shared.ErrCodeServiceError)
		assert.Equal(t, "REPOSITORY_ERROR", shared.ErrCodeRepositoryError)
		assert.Equal(t, "VALIDATION_FAILED", shared.ErrCodeValidationFailed)
		assert.Equal(t, "BUSINESS_RULE_VIOLATION", shared.ErrCodeBusinessRuleViolation)
		assert.Equal(t, "CURRENCY_MISMATCH", shared.ErrCodeCurrencyMismatch)
		assert.Equal(t, "INVALID_STATE", shared.ErrCodeInvalidState)
		assert.Equal(t, "TERMINAL_STATE", shared.ErrCodeTerminalState)
	})
}
