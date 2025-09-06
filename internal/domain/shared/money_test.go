package shared_test

import (
	"testing"

	"crypto-checkout/internal/domain/shared"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoney(t *testing.T) {
	t.Run("NewMoney - valid amount and currency", func(t *testing.T) {
		money, err := shared.NewMoney("100.50", shared.CurrencyUSD)
		require.NoError(t, err)
		assert.Equal(t, "100.50", money.String())
		assert.Equal(t, string(shared.CurrencyUSD), money.Currency())
		assert.Equal(t, "100.50", money.Amount().StringFixed(2))
	})

	t.Run("NewMoney - invalid amount", func(t *testing.T) {
		_, err := shared.NewMoney("invalid", shared.CurrencyUSD)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount format")
	})

	t.Run("NewMoney - negative amount", func(t *testing.T) {
		_, err := shared.NewMoney("-10.00", shared.CurrencyUSD)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount cannot be negative")
	})

	t.Run("NewMoney - empty amount", func(t *testing.T) {
		_, err := shared.NewMoney("", shared.CurrencyUSD)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount cannot be empty")
	})

	t.Run("NewMoney - invalid currency", func(t *testing.T) {
		_, err := shared.NewMoney("100.00", "INVALID")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid currency")
	})

	t.Run("NewMoneyWithCrypto - valid amount and cryptocurrency", func(t *testing.T) {
		money, err := shared.NewMoneyWithCrypto("0.001", shared.CryptoCurrencyBTC)
		require.NoError(t, err)
		assert.Equal(t, "0.00", money.String()) // Rounded to 2 decimal places
		assert.Equal(t, string(shared.CryptoCurrencyBTC), money.Currency())
	})

	t.Run("Add - same currency", func(t *testing.T) {
		money1, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		money2, _ := shared.NewMoney("50.00", shared.CurrencyUSD)

		result, err := money1.Add(money2)
		require.NoError(t, err)
		assert.Equal(t, "150.00", result.String())
		assert.Equal(t, string(shared.CurrencyUSD), result.Currency())
	})

	t.Run("Add - different currencies", func(t *testing.T) {
		money1, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		money2, _ := shared.NewMoney("50.00", shared.CurrencyEUR)

		_, err := money1.Add(money2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "currency mismatch")
	})

	t.Run("Multiply", func(t *testing.T) {
		money, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		multiplier := decimal.NewFromFloat(1.5)

		result, err := money.Multiply(multiplier)
		require.NoError(t, err)
		assert.Equal(t, "150.00", result.String())
		assert.Equal(t, string(shared.CurrencyUSD), result.Currency())
	})

	t.Run("LessThan", func(t *testing.T) {
		money1, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		money2, _ := shared.NewMoney("150.00", shared.CurrencyUSD)

		assert.True(t, money1.LessThan(money2))
		assert.False(t, money2.LessThan(money1))
	})

	t.Run("LessThan - different currencies", func(t *testing.T) {
		money1, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		money2, _ := shared.NewMoney("50.00", shared.CurrencyEUR)

		assert.False(t, money1.LessThan(money2)) // Cannot compare different currencies
	})

	t.Run("GreaterThanOrEqual", func(t *testing.T) {
		money1, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		money2, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		money3, _ := shared.NewMoney("50.00", shared.CurrencyUSD)

		assert.True(t, money1.GreaterThanOrEqual(money2))  // Equal
		assert.True(t, money1.GreaterThanOrEqual(money3))  // Greater
		assert.False(t, money3.GreaterThanOrEqual(money1)) // Less
	})

	t.Run("Equals", func(t *testing.T) {
		money1, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		money2, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		money3, _ := shared.NewMoney("100.00", shared.CurrencyEUR)

		assert.True(t, money1.Equals(money2))  // Same amount and currency
		assert.False(t, money1.Equals(money3)) // Different currency
	})
}
