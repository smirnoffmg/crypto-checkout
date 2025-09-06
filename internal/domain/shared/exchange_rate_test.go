package shared_test

import (
	"testing"
	"time"

	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExchangeRate(t *testing.T) {
	t.Run("NewExchangeRate - valid rate", func(t *testing.T) {
		rate, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, "1.5", rate.Rate().String())
		assert.Equal(t, shared.CurrencyUSD, rate.FromCurrency())
		assert.Equal(t, shared.CryptoCurrencyUSDT, rate.ToCurrency())
		assert.Equal(t, "test_provider", rate.Source())
		assert.False(t, rate.IsExpired())
	})

	t.Run("NewExchangeRate - empty rate", func(t *testing.T) {
		_, err := shared.NewExchangeRate("", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exchange rate cannot be empty")
	})

	t.Run("NewExchangeRate - invalid from currency", func(t *testing.T) {
		_, err := shared.NewExchangeRate("1.5", "INVALID", shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid from currency")
	})

	t.Run("NewExchangeRate - invalid to currency", func(t *testing.T) {
		_, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, "INVALID", "test_provider", 30*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid to currency")
	})

	t.Run("NewExchangeRate - empty source", func(t *testing.T) {
		_, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "", 30*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate source cannot be empty")
	})

	t.Run("NewExchangeRate - invalid rate format", func(t *testing.T) {
		_, err := shared.NewExchangeRate("invalid", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid exchange rate format")
	})

	t.Run("NewExchangeRate - zero rate", func(t *testing.T) {
		_, err := shared.NewExchangeRate("0", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exchange rate must be positive")
	})

	t.Run("NewExchangeRate - negative rate", func(t *testing.T) {
		_, err := shared.NewExchangeRate("-1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exchange rate must be positive")
	})

	t.Run("Convert - valid conversion", func(t *testing.T) {
		rate, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)

		amount, err := shared.NewMoney("100.00", shared.CurrencyUSD)
		require.NoError(t, err)

		converted, err := rate.Convert(amount)
		require.NoError(t, err)
		assert.Equal(t, "150.00", converted.String())
		assert.Equal(t, string(shared.CryptoCurrencyUSDT), converted.Currency())
	})

	t.Run("Convert - currency mismatch", func(t *testing.T) {
		rate, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)

		amount, err := shared.NewMoney("100.00", shared.CurrencyEUR)
		require.NoError(t, err)

		_, err = rate.Convert(amount)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "currency mismatch for conversion")
	})

	t.Run("Convert - nil amount", func(t *testing.T) {
		rate, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)

		_, err = rate.Convert(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount cannot be nil")
	})

	t.Run("IsExpired - expired rate", func(t *testing.T) {
		// Create a rate that expires immediately
		rate, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", -1*time.Minute)
		require.NoError(t, err)
		assert.True(t, rate.IsExpired())
	})

	t.Run("String", func(t *testing.T) {
		rate, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, "1.5", rate.String())
	})

	t.Run("Equals - same rate", func(t *testing.T) {
		rate1, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)
		rate2, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)
		assert.True(t, rate1.Equals(rate2))
	})

	t.Run("Equals - different rate", func(t *testing.T) {
		rate1, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)
		rate2, err := shared.NewExchangeRate("2.0", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)
		assert.False(t, rate1.Equals(rate2))
	})

	t.Run("Equals - nil rate", func(t *testing.T) {
		rate, err := shared.NewExchangeRate("1.5", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "test_provider", 30*time.Minute)
		require.NoError(t, err)
		assert.False(t, rate.Equals(nil))
	})
}
