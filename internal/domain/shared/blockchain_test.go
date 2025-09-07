package shared_test

import (
	"testing"

	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/require"
)

func TestBlockchainNetwork(t *testing.T) {
	t.Run("String - valid networks", func(t *testing.T) {
		require.Equal(t, "tron", shared.NetworkTron.String())
		require.Equal(t, "ethereum", shared.NetworkEthereum.String())
		require.Equal(t, "bitcoin", shared.NetworkBitcoin.String())
	})

	t.Run("IsValid - valid networks", func(t *testing.T) {
		require.True(t, shared.NetworkTron.IsValid())
		require.True(t, shared.NetworkEthereum.IsValid())
		require.True(t, shared.NetworkBitcoin.IsValid())
	})

	t.Run("IsValid - invalid network", func(t *testing.T) {
		invalidNetwork := shared.BlockchainNetwork("invalid")
		require.False(t, invalidNetwork.IsValid())
	})
}
