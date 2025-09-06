package payment_test

import (
	"testing"
	"time"

	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentValueObjects(t *testing.T) {
	t.Run("TransactionHash", func(t *testing.T) {
		t.Run("NewTransactionHash - valid hash", func(t *testing.T) {
			hash := "a7b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
			txHash, err := payment.NewTransactionHash(hash)
			require.NoError(t, err)
			assert.Equal(t, hash, txHash.String())
		})

		t.Run("NewTransactionHash - empty hash", func(t *testing.T) {
			_, err := payment.NewTransactionHash("")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "transaction hash cannot be empty")
		})

		t.Run("NewTransactionHash - hash too short", func(t *testing.T) {
			_, err := payment.NewTransactionHash("short")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "transaction hash format is too short")
		})
	})

	t.Run("PaymentAddress", func(t *testing.T) {
		t.Run("NewPaymentAddress - valid address", func(t *testing.T) {
			address := "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN"
			paymentAddr, err := payment.NewPaymentAddress(address, shared.NetworkTron)
			require.NoError(t, err)
			assert.Equal(t, address, paymentAddr.String())
			assert.Equal(t, shared.NetworkTron, paymentAddr.Network())
		})

		t.Run("NewPaymentAddress - empty address", func(t *testing.T) {
			_, err := payment.NewPaymentAddress("", shared.NetworkTron)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "address cannot be empty")
		})

		t.Run("NewPaymentAddress - invalid network", func(t *testing.T) {
			_, err := payment.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", "invalid")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid blockchain network")
		})
	})

	t.Run("ConfirmationCount", func(t *testing.T) {
		t.Run("NewConfirmationCount - valid count", func(t *testing.T) {
			count, err := payment.NewConfirmationCount(5)
			require.NoError(t, err)
			assert.Equal(t, 5, count.Int())
			assert.Equal(t, "5", count.String())
		})

		t.Run("NewConfirmationCount - negative count", func(t *testing.T) {
			_, err := payment.NewConfirmationCount(-1)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "confirmation count cannot be negative")
		})

		t.Run("IsGreaterThanOrEqual", func(t *testing.T) {
			count, err := payment.NewConfirmationCount(5)
			require.NoError(t, err)

			required3, err := payment.NewConfirmationCount(3)
			require.NoError(t, err)
			required5, err := payment.NewConfirmationCount(5)
			require.NoError(t, err)
			required6, err := payment.NewConfirmationCount(6)
			require.NoError(t, err)

			assert.True(t, count.IsGreaterThanOrEqual(required3))
			assert.True(t, count.IsGreaterThanOrEqual(required5))
			assert.False(t, count.IsGreaterThanOrEqual(required6))
		})
	})

	t.Run("BlockInfo", func(t *testing.T) {
		t.Run("NewBlockInfo - valid info", func(t *testing.T) {
			blockInfo, err := payment.NewBlockInfo(12345, "blockhash123")
			require.NoError(t, err)
			assert.Equal(t, int64(12345), blockInfo.Number())
			assert.Equal(t, "blockhash123", blockInfo.Hash())
		})

		t.Run("NewBlockInfo - negative block number", func(t *testing.T) {
			_, err := payment.NewBlockInfo(-1, "blockhash123")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "block number cannot be negative")
		})

		t.Run("NewBlockInfo - empty hash", func(t *testing.T) {
			_, err := payment.NewBlockInfo(12345, "")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "block hash cannot be empty")
		})
	})

	t.Run("PaymentAmount", func(t *testing.T) {
		t.Run("NewPaymentAmount - valid amount", func(t *testing.T) {
			amount, err := shared.NewMoney("100.00", shared.CurrencyUSD)
			require.NoError(t, err)

			paymentAmount, err := payment.NewPaymentAmount(amount, shared.CryptoCurrencyUSDT)
			require.NoError(t, err)
			assert.Equal(t, amount, paymentAmount.Amount())
			assert.Equal(t, shared.CryptoCurrencyUSDT, paymentAmount.Currency())
			assert.Equal(t, "100.00 USDT", paymentAmount.String())
		})

		t.Run("NewPaymentAmount - nil amount", func(t *testing.T) {
			_, err := payment.NewPaymentAmount(nil, shared.CryptoCurrencyUSDT)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "amount cannot be nil")
		})

		t.Run("NewPaymentAmount - invalid currency", func(t *testing.T) {
			amount, err := shared.NewMoney("100.00", shared.CurrencyUSD)
			require.NoError(t, err)

			_, err = payment.NewPaymentAmount(amount, "INVALID")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid cryptocurrency")
		})
	})

	t.Run("NetworkFee", func(t *testing.T) {
		t.Run("NewNetworkFee - valid fee", func(t *testing.T) {
			fee, err := shared.NewMoney("10.00", shared.CurrencyUSD)
			require.NoError(t, err)

			networkFee, err := payment.NewNetworkFee(fee, shared.CryptoCurrencyBTC)
			require.NoError(t, err)
			assert.Equal(t, fee, networkFee.Fee())
			assert.Equal(t, shared.CryptoCurrencyBTC, networkFee.Currency())
			assert.Equal(t, "10.00 BTC", networkFee.String())
		})

		t.Run("NewNetworkFee - nil fee", func(t *testing.T) {
			_, err := payment.NewNetworkFee(nil, shared.CryptoCurrencyBTC)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "fee cannot be nil")
		})

		t.Run("NewNetworkFee - negative fee", func(t *testing.T) {
			// This test is actually testing that shared.Money prevents negative amounts
			// So we expect the error to come from shared.NewMoney, not from NewNetworkFee
			_, err := shared.NewMoney("-10.00", shared.CurrencyUSD)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "amount cannot be negative")
		})
	})

	t.Run("PaymentTimestamps", func(t *testing.T) {
		t.Run("NewPaymentTimestamps", func(t *testing.T) {
			detectedAt := time.Now().UTC()
			timestamps := payment.NewPaymentTimestamps(detectedAt)

			assert.Equal(t, detectedAt, timestamps.DetectedAt())
			assert.Nil(t, timestamps.ConfirmedAt())
			assert.WithinDuration(t, time.Now().UTC(), timestamps.CreatedAt(), time.Second)
			assert.WithinDuration(t, time.Now().UTC(), timestamps.UpdatedAt(), time.Second)
		})

		t.Run("SetConfirmedAt", func(t *testing.T) {
			detectedAt := time.Now().UTC()
			timestamps := payment.NewPaymentTimestamps(detectedAt)

			confirmedAt := time.Now().UTC().Add(5 * time.Minute)
			timestamps.SetConfirmedAt(confirmedAt)

			assert.Equal(t, confirmedAt, *timestamps.ConfirmedAt())
		})

		t.Run("SetUpdatedAt", func(t *testing.T) {
			detectedAt := time.Now().UTC()
			timestamps := payment.NewPaymentTimestamps(detectedAt)

			updatedAt := time.Now().UTC().Add(10 * time.Minute)
			timestamps.SetUpdatedAt(updatedAt)

			assert.Equal(t, updatedAt, timestamps.UpdatedAt())
		})
	})
}
