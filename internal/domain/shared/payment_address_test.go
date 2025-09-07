package shared_test

import (
	"testing"
	"time"

	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/require"
)

func TestPaymentAddress(t *testing.T) {
	t.Run("NewPaymentAddress - valid address", func(t *testing.T) {
		address, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron)
		require.NoError(t, err)
		require.Equal(t, "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", address.Address())
		require.Equal(t, shared.NetworkTron, address.Network())
		require.False(t, address.IsExpired())
		require.NotNil(t, address.GeneratedAt())
	})

	t.Run("NewPaymentAddress - empty address", func(t *testing.T) {
		_, err := shared.NewPaymentAddress("", shared.NetworkTron)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address cannot be empty")
	})

	t.Run("NewPaymentAddress - invalid network", func(t *testing.T) {
		_, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", "invalid")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid blockchain network")
	})

	t.Run("NewPaymentAddress - address too short", func(t *testing.T) {
		_, err := shared.NewPaymentAddress("short", shared.NetworkTron)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address format is too short")
	})

	t.Run("NewPaymentAddressWithExpiry - valid address with expiry", func(t *testing.T) {
		expiresAt := time.Now().UTC().Add(30 * time.Minute)
		address, err := shared.NewPaymentAddressWithExpiry("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron, expiresAt)
		require.NoError(t, err)
		require.Equal(t, expiresAt, *address.ExpiresAt())
		require.False(t, address.IsExpired())
	})

	t.Run("NewPaymentAddressWithExpiry - expired address", func(t *testing.T) {
		expiresAt := time.Now().UTC().Add(-1 * time.Hour)
		_, err := shared.NewPaymentAddressWithExpiry("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron, expiresAt)
		require.Error(t, err)
		require.Contains(t, err.Error(), "expiration time must be in the future")
	})

	t.Run("IsExpired - non-expired address", func(t *testing.T) {
		futureExpiry := time.Now().UTC().Add(1 * time.Hour)
		address, err := shared.NewPaymentAddressWithExpiry("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron, futureExpiry)
		require.NoError(t, err)
		require.False(t, address.IsExpired())
	})

	t.Run("String", func(t *testing.T) {
		address, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron)
		require.NoError(t, err)
		require.Equal(t, "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", address.String())
	})

	t.Run("Equals - same address", func(t *testing.T) {
		address1, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron)
		require.NoError(t, err)
		address2, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron)
		require.NoError(t, err)
		require.True(t, address1.Equals(address2))
	})

	t.Run("Equals - different address", func(t *testing.T) {
		address1, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron)
		require.NoError(t, err)
		address2, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN2", shared.NetworkTron)
		require.NoError(t, err)
		require.False(t, address1.Equals(address2))
	})

	t.Run("Equals - different network", func(t *testing.T) {
		address1, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron)
		require.NoError(t, err)
		address2, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkEthereum)
		require.NoError(t, err)
		require.False(t, address1.Equals(address2))
	})

	t.Run("Equals - nil address", func(t *testing.T) {
		address, err := shared.NewPaymentAddress("TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN", shared.NetworkTron)
		require.NoError(t, err)
		require.False(t, address.Equals(nil))
	})
}
