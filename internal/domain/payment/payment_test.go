package payment_test

import (
	"testing"
	"time"

	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/require"
)

func TestPayment(t *testing.T) {
	t.Run("NewPayment - valid payment", func(t *testing.T) {
		testPayment := createTestPayment()
		require.NotNil(t, testPayment)

		require.Equal(t, "test-payment-id", string(testPayment.ID()))
		require.Equal(t, "test-invoice-id", string(testPayment.InvoiceID()))
		require.Equal(t, "test-from-address", testPayment.FromAddress())
		require.Equal(t, payment.StatusDetected, testPayment.Status())
		require.Equal(t, 0, testPayment.Confirmations().Int())
		require.Equal(t, 6, testPayment.RequiredConfirmations())
		require.False(t, testPayment.IsConfirmed())
		require.False(t, testPayment.IsTerminal())
		require.True(t, testPayment.IsActive())
	})

	t.Run("NewPayment - empty ID", func(t *testing.T) {
		_, err := payment.NewPayment(
			"",
			"test-invoice-id",
			createTestPaymentAmount(),
			"test-from-address",
			createTestPaymentAddress(),
			createTestTransactionHash(),
			6,
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "payment ID cannot be empty")
	})

	t.Run("NewPayment - empty invoice ID", func(t *testing.T) {
		_, err := payment.NewPayment(
			"test-payment-id",
			"",
			createTestPaymentAmount(),
			"test-from-address",
			createTestPaymentAddress(),
			createTestTransactionHash(),
			6,
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invoice ID cannot be empty")
	})

	t.Run("NewPayment - nil amount", func(t *testing.T) {
		_, err := payment.NewPayment(
			"test-payment-id",
			"test-invoice-id",
			nil,
			"test-from-address",
			createTestPaymentAddress(),
			createTestTransactionHash(),
			6,
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "payment amount cannot be nil")
	})

	t.Run("NewPayment - empty from address", func(t *testing.T) {
		_, err := payment.NewPayment(
			"test-payment-id",
			"test-invoice-id",
			createTestPaymentAmount(),
			"",
			createTestPaymentAddress(),
			createTestTransactionHash(),
			6,
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "from address cannot be empty")
	})

	t.Run("NewPayment - nil to address", func(t *testing.T) {
		_, err := payment.NewPayment(
			"test-payment-id",
			"test-invoice-id",
			createTestPaymentAmount(),
			"test-from-address",
			nil,
			createTestTransactionHash(),
			6,
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "to address cannot be nil")
	})

	t.Run("NewPayment - nil transaction hash", func(t *testing.T) {
		_, err := payment.NewPayment(
			"test-payment-id",
			"test-invoice-id",
			createTestPaymentAmount(),
			"test-from-address",
			createTestPaymentAddress(),
			nil,
			6,
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "transaction hash cannot be nil")
	})

	t.Run("NewPayment - negative required confirmations", func(t *testing.T) {
		_, err := payment.NewPayment(
			"test-payment-id",
			"test-invoice-id",
			createTestPaymentAmount(),
			"test-from-address",
			createTestPaymentAddress(),
			createTestTransactionHash(),
			-1,
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "required confirmations cannot be negative")
	})

	t.Run("UpdateConfirmations", func(t *testing.T) {
		testPayment := createTestPayment()

		err := testPayment.UpdateConfirmations(nil, 3)
		require.NoError(t, err)
		require.Equal(t, 3, testPayment.Confirmations().Int())
		require.False(t, testPayment.IsConfirmed())

		err = testPayment.UpdateConfirmations(nil, 6)
		require.NoError(t, err)
		require.Equal(t, 6, testPayment.Confirmations().Int())
		require.True(t, testPayment.IsConfirmed())
	})

	t.Run("UpdateConfirmations - negative count", func(t *testing.T) {
		testPayment := createTestPayment()

		err := testPayment.UpdateConfirmations(nil, -1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid confirmation count")
	})

	t.Run("UpdateBlockInfo", func(t *testing.T) {
		testPayment := createTestPayment()

		err := testPayment.UpdateBlockInfo(12345, "blockhash123")
		require.NoError(t, err)

		blockInfo := testPayment.BlockInfo()
		require.NotNil(t, blockInfo)
		require.Equal(t, int64(12345), blockInfo.Number())
		require.Equal(t, "blockhash123", blockInfo.Hash())
	})

	t.Run("UpdateBlockInfo - negative block number", func(t *testing.T) {
		testPayment := createTestPayment()

		err := testPayment.UpdateBlockInfo(-1, "blockhash123")
		require.Error(t, err)
		require.Contains(t, err.Error(), "block number cannot be negative")
	})

	t.Run("UpdateBlockInfo - empty hash", func(t *testing.T) {
		testPayment := createTestPayment()

		err := testPayment.UpdateBlockInfo(12345, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "block hash cannot be empty")
	})

	t.Run("UpdateNetworkFee", func(t *testing.T) {
		testPayment := createTestPayment()

		fee, err := shared.NewMoney("10.00", shared.CurrencyUSD)
		require.NoError(t, err)

		err = testPayment.UpdateNetworkFee(fee, shared.CryptoCurrencyBTC)
		require.NoError(t, err)

		networkFee := testPayment.NetworkFee()
		require.NotNil(t, networkFee)
		require.Equal(t, fee, networkFee.Fee())
		require.Equal(t, shared.CryptoCurrencyBTC, networkFee.Currency())
	})

	t.Run("UpdateNetworkFee - nil fee", func(t *testing.T) {
		testPayment := createTestPayment()

		err := testPayment.UpdateNetworkFee(nil, shared.CryptoCurrencyBTC)
		require.Error(t, err)
		require.Contains(t, err.Error(), "fee cannot be nil")
	})

	t.Run("SetStatus - for testing", func(t *testing.T) {
		testPayment := createTestPayment()

		testPayment.SetStatus(payment.StatusConfirmed)
		require.Equal(t, payment.StatusConfirmed, testPayment.Status())
		require.True(t, testPayment.IsTerminal())
		require.False(t, testPayment.IsActive())
	})

	t.Run("SetConfirmations - for testing", func(t *testing.T) {
		testPayment := createTestPayment()

		err := testPayment.SetConfirmations(5)
		require.NoError(t, err)
		require.Equal(t, 5, testPayment.Confirmations().Int())
	})

	t.Run("SetConfirmedAt - for testing", func(t *testing.T) {
		testPayment := createTestPayment()

		confirmedAt := time.Now().UTC()
		testPayment.SetConfirmedAt(confirmedAt)

		require.Equal(t, confirmedAt, *testPayment.ConfirmedAt())
	})
}

// Helper functions to create test objects
func createTestPayment() *payment.Payment {
	p, _ := payment.NewPayment(
		"test-payment-id",
		"test-invoice-id",
		createTestPaymentAmount(),
		"test-from-address",
		createTestPaymentAddress(),
		createTestTransactionHash(),
		6,
	)
	return p
}

func createTestPaymentAmount() *payment.PaymentAmount {
	amount, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
	paymentAmount, _ := payment.NewPaymentAmount(amount, shared.CryptoCurrencyUSDT)
	return paymentAmount
}

func createTestPaymentAddress() *payment.PaymentAddress {
	addr, _ := payment.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron)
	return addr
}

func createTestTransactionHash() *payment.TransactionHash {
	hash, _ := payment.NewTransactionHash("a7b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456")
	return hash
}
