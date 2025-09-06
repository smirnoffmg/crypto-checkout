package shared_test

import (
	"testing"

	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionHash(t *testing.T) {
	t.Run("NewTransactionHash - valid hash", func(t *testing.T) {
		hash := "a7b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
		transactionHash, err := shared.NewTransactionHash(hash)
		require.NoError(t, err)
		assert.Equal(t, hash, transactionHash.Hash())
		assert.Equal(t, hash, transactionHash.String())
	})

	t.Run("NewTransactionHash - empty hash", func(t *testing.T) {
		_, err := shared.NewTransactionHash("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction hash cannot be empty")
	})

	t.Run("NewTransactionHash - hash too short", func(t *testing.T) {
		_, err := shared.NewTransactionHash("short")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction hash format is too short")
	})

	t.Run("Equals - same hash", func(t *testing.T) {
		hash := "a7b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
		hash1, err := shared.NewTransactionHash(hash)
		require.NoError(t, err)
		hash2, err := shared.NewTransactionHash(hash)
		require.NoError(t, err)
		assert.True(t, hash1.Equals(hash2))
	})

	t.Run("Equals - different hash", func(t *testing.T) {
		hash1, err := shared.NewTransactionHash("a7b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456")
		require.NoError(t, err)
		hash2, err := shared.NewTransactionHash("b8c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456")
		require.NoError(t, err)
		assert.False(t, hash1.Equals(hash2))
	})

	t.Run("Equals - nil hash", func(t *testing.T) {
		hash, err := shared.NewTransactionHash("a7b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456")
		require.NoError(t, err)
		assert.False(t, hash.Equals(nil))
	})
}
