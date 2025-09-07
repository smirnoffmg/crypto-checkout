package shared_test

import (
	"errors"
	"testing"

	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/require"
)

func TestDomainError(t *testing.T) {
	t.Run("NewDomainError - with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		domainErr := shared.NewDomainError("TEST_CODE", "test message", underlyingErr)

		require.Equal(t, "TEST_CODE", domainErr.Code)
		require.Equal(t, "test message", domainErr.Message)
		require.Equal(t, underlyingErr, domainErr.Err)
		require.NotNil(t, domainErr.Details)
	})

	t.Run("NewDomainError - without underlying error", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)

		require.Equal(t, "TEST_CODE", domainErr.Code)
		require.Equal(t, "test message", domainErr.Message)
		require.Nil(t, domainErr.Err)
		require.NotNil(t, domainErr.Details)
	})

	t.Run("Error - with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		domainErr := shared.NewDomainError("TEST_CODE", "test message", underlyingErr)

		expected := "test message: underlying error"
		require.Equal(t, expected, domainErr.Error())
	})

	t.Run("Error - without underlying error", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)

		require.Equal(t, "test message", domainErr.Error())
	})

	t.Run("Unwrap - with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		domainErr := shared.NewDomainError("TEST_CODE", "test message", underlyingErr)

		require.Equal(t, underlyingErr, domainErr.Unwrap())
	})

	t.Run("Unwrap - without underlying error", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)

		require.Nil(t, domainErr.Unwrap())
	})

	t.Run("WithDetail", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)
		domainErr = domainErr.WithDetail("key1", "value1")

		require.Equal(t, "value1", domainErr.Details["key1"])
	})

	t.Run("WithDetails", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)
		details := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		domainErr = domainErr.WithDetails(details)

		require.Equal(t, "value1", domainErr.Details["key1"])
		require.Equal(t, "value2", domainErr.Details["key2"])
	})

	t.Run("WithDetail - chaining", func(t *testing.T) {
		domainErr := shared.NewDomainError("TEST_CODE", "test message", nil)
		domainErr = domainErr.WithDetail("key1", "value1").WithDetail("key2", "value2")

		require.Equal(t, "value1", domainErr.Details["key1"])
		require.Equal(t, "value2", domainErr.Details["key2"])
	})
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		err := shared.NewValidationError("field", "reason")

		require.Equal(t, shared.ErrCodeValidationFailed, err.Code)
		require.Equal(t, "validation failed", err.Message)
		require.Equal(t, "field", err.Details["field"])
		require.Equal(t, "reason", err.Details["reason"])
	})

	t.Run("NewNotFoundError", func(t *testing.T) {
		err := shared.NewNotFoundError("resource", "id123")

		require.Equal(t, shared.ErrCodeNotFound, err.Code)
		require.Equal(t, "resource not found", err.Message)
		require.Equal(t, "resource", err.Details["resource"])
		require.Equal(t, "id123", err.Details["id"])
	})

	t.Run("NewInvalidTransitionError", func(t *testing.T) {
		err := shared.NewInvalidTransitionError("from_state", "to_state")

		require.Equal(t, shared.ErrCodeInvalidTransition, err.Code)
		require.Equal(t, "invalid transition", err.Message)
		require.Equal(t, "from_state", err.Details["from"])
		require.Equal(t, "to_state", err.Details["to"])
	})

	t.Run("NewBusinessRuleViolationError", func(t *testing.T) {
		err := shared.NewBusinessRuleViolationError("rule", "reason")

		require.Equal(t, shared.ErrCodeBusinessRuleViolation, err.Code)
		require.Equal(t, "business rule violation", err.Message)
		require.Equal(t, "rule", err.Details["rule"])
		require.Equal(t, "reason", err.Details["reason"])
	})

	t.Run("NewCurrencyMismatchError", func(t *testing.T) {
		err := shared.NewCurrencyMismatchError("USD", "EUR")

		require.Equal(t, shared.ErrCodeCurrencyMismatch, err.Code)
		require.Equal(t, "currency mismatch", err.Message)
		require.Equal(t, "USD", err.Details["expected"])
		require.Equal(t, "EUR", err.Details["actual"])
	})

	t.Run("NewInsufficientAmountError", func(t *testing.T) {
		err := shared.NewInsufficientAmountError("100.00", "50.00")

		require.Equal(t, shared.ErrCodeInsufficientAmount, err.Code)
		require.Equal(t, "insufficient amount", err.Message)
		require.Equal(t, "100.00", err.Details["required"])
		require.Equal(t, "50.00", err.Details["provided"])
	})

	t.Run("NewExcessiveAmountError", func(t *testing.T) {
		err := shared.NewExcessiveAmountError("100.00", "150.00")

		require.Equal(t, shared.ErrCodeExcessiveAmount, err.Code)
		require.Equal(t, "excessive amount", err.Message)
		require.Equal(t, "100.00", err.Details["limit"])
		require.Equal(t, "150.00", err.Details["provided"])
	})

	t.Run("NewTerminalStateError", func(t *testing.T) {
		err := shared.NewTerminalStateError("paid", "cancel")

		require.Equal(t, shared.ErrCodeTerminalState, err.Code)
		require.Equal(t, "cannot perform action in terminal state", err.Message)
		require.Equal(t, "paid", err.Details["state"])
		require.Equal(t, "cancel", err.Details["action"])
	})
}

func TestSharedErrors(t *testing.T) {
	t.Run("Generic validation errors", func(t *testing.T) {
		require.Equal(t, "invalid ID", shared.ErrInvalidID.Error())
		require.Equal(t, "invalid title", shared.ErrInvalidTitle.Error())
		require.Equal(t, "invalid description", shared.ErrInvalidDescription.Error())
		require.Equal(t, "invalid amount", shared.ErrInvalidAmount.Error())
		require.Equal(t, "invalid currency", shared.ErrInvalidCurrency.Error())
		require.Equal(t, "invalid status", shared.ErrInvalidStatus.Error())
		require.Equal(t, "invalid status transition", shared.ErrInvalidTransition.Error())
		require.Equal(t, "invalid input", shared.ErrInvalidInput.Error())
	})

	t.Run("Money and currency errors", func(t *testing.T) {
		require.Equal(t, "invalid money amount", shared.ErrInvalidMoneyAmount.Error())
		require.Equal(t, "currency mismatch", shared.ErrCurrencyMismatch.Error())
		require.Equal(t, "amount cannot be negative", shared.ErrNegativeAmount.Error())
		require.Equal(t, "amount cannot be zero", shared.ErrZeroAmount.Error())
		require.Equal(t, "invalid amount format", shared.ErrInvalidAmountFormat.Error())
	})

	t.Run("Payment and blockchain errors", func(t *testing.T) {
		require.Equal(t, "invalid payment address", shared.ErrInvalidPaymentAddress.Error())
		require.Equal(t, "invalid transaction hash", shared.ErrInvalidTransactionHash.Error())
		require.Equal(t, "invalid blockchain network", shared.ErrInvalidNetwork.Error())
		require.Equal(t, "payment address has expired", shared.ErrExpiredPaymentAddress.Error())
		require.Equal(t, "exchange rate has expired", shared.ErrExpiredExchangeRate.Error())
		require.Equal(t, "invalid exchange rate", shared.ErrInvalidExchangeRate.Error())
		require.Equal(t, "invalid confirmation count", shared.ErrInvalidConfirmationCount.Error())
	})

	t.Run("Service and repository errors", func(t *testing.T) {
		require.Equal(t, "not found", shared.ErrNotFound.Error())
		require.Equal(t, "already exists", shared.ErrAlreadyExists.Error())
		require.Equal(t, "repository error", shared.ErrRepositoryError.Error())
		require.Equal(t, "service error", shared.ErrServiceError.Error())
		require.Equal(t, "ID generation error", shared.ErrIDGenerationError.Error())
		require.Equal(t, "invalid request", shared.ErrInvalidRequest.Error())
	})

	t.Run("State and lifecycle errors", func(t *testing.T) {
		require.Equal(t, "invalid state", shared.ErrInvalidState.Error())
		require.Equal(t, "cannot transition to target state", shared.ErrCannotTransition.Error())
		require.Equal(t, "already in target state", shared.ErrAlreadyInState.Error())
		require.Equal(t, "expired", shared.ErrExpired.Error())
		require.Equal(t, "cannot perform action in terminal state", shared.ErrTerminalState.Error())
	})

	t.Run("Business logic errors", func(t *testing.T) {
		require.Equal(t, "insufficient amount", shared.ErrInsufficientAmount.Error())
		require.Equal(t, "excessive amount", shared.ErrExcessiveAmount.Error())
		require.Equal(t, "validation failed", shared.ErrValidationFailed.Error())
		require.Equal(t, "business rule violation", shared.ErrBusinessRuleViolation.Error())
	})
}

func TestErrorCodes(t *testing.T) {
	t.Run("Error codes are defined", func(t *testing.T) {
		require.Equal(t, "INVALID_INPUT", shared.ErrCodeInvalidInput)
		require.Equal(t, "INVALID_STATUS", shared.ErrCodeInvalidStatus)
		require.Equal(t, "INVALID_TRANSITION", shared.ErrCodeInvalidTransition)
		require.Equal(t, "NOT_FOUND", shared.ErrCodeNotFound)
		require.Equal(t, "ALREADY_EXISTS", shared.ErrCodeAlreadyExists)
		require.Equal(t, "EXPIRED", shared.ErrCodeExpired)
		require.Equal(t, "INSUFFICIENT_AMOUNT", shared.ErrCodeInsufficientAmount)
		require.Equal(t, "EXCESSIVE_AMOUNT", shared.ErrCodeExcessiveAmount)
		require.Equal(t, "SERVICE_ERROR", shared.ErrCodeServiceError)
		require.Equal(t, "REPOSITORY_ERROR", shared.ErrCodeRepositoryError)
		require.Equal(t, "VALIDATION_FAILED", shared.ErrCodeValidationFailed)
		require.Equal(t, "BUSINESS_RULE_VIOLATION", shared.ErrCodeBusinessRuleViolation)
		require.Equal(t, "CURRENCY_MISMATCH", shared.ErrCodeCurrencyMismatch)
		require.Equal(t, "INVALID_STATE", shared.ErrCodeInvalidState)
		require.Equal(t, "TERMINAL_STATE", shared.ErrCodeTerminalState)
	})
}
