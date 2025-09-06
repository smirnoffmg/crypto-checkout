package shared_test

import (
	"testing"

	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/assert"
)

func TestBlockchainNetwork(t *testing.T) {
	t.Run("String - valid networks", func(t *testing.T) {
		assert.Equal(t, "tron", shared.NetworkTron.String())
		assert.Equal(t, "ethereum", shared.NetworkEthereum.String())
		assert.Equal(t, "bitcoin", shared.NetworkBitcoin.String())
	})

	t.Run("IsValid - valid networks", func(t *testing.T) {
		assert.True(t, shared.NetworkTron.IsValid())
		assert.True(t, shared.NetworkEthereum.IsValid())
		assert.True(t, shared.NetworkBitcoin.IsValid())
	})

	t.Run("IsValid - invalid network", func(t *testing.T) {
		invalidNetwork := shared.BlockchainNetwork("invalid")
		assert.False(t, invalidNetwork.IsValid())
	})
}
