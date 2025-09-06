package shared_test

import (
	"testing"

	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirmationCount(t *testing.T) {
	t.Run("NewConfirmationCount - valid count", func(t *testing.T) {
		count, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		assert.Equal(t, 5, count.Count())
		assert.Equal(t, 5, count.Int())
		assert.False(t, count.IsZero())
	})

	t.Run("NewConfirmationCount - zero count", func(t *testing.T) {
		count, err := shared.NewConfirmationCount(0)
		require.NoError(t, err)
		assert.Equal(t, 0, count.Count())
		assert.True(t, count.IsZero())
	})

	t.Run("NewConfirmationCount - negative count", func(t *testing.T) {
		_, err := shared.NewConfirmationCount(-1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "confirmation count cannot be negative")
	})

	t.Run("Increment", func(t *testing.T) {
		count, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)

		incremented := count.Increment()
		assert.Equal(t, 6, incremented.Count())
		assert.Equal(t, 5, count.Count()) // Original should be unchanged
	})

	t.Run("IsGreaterThan - greater", func(t *testing.T) {
		count1, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		count2, err := shared.NewConfirmationCount(3)
		require.NoError(t, err)

		assert.True(t, count1.IsGreaterThan(count2))
		assert.False(t, count2.IsGreaterThan(count1))
	})

	t.Run("IsGreaterThan - equal", func(t *testing.T) {
		count1, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		count2, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)

		assert.False(t, count1.IsGreaterThan(count2))
	})

	t.Run("IsGreaterThan - nil", func(t *testing.T) {
		count, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)

		assert.False(t, count.IsGreaterThan(nil))
	})

	t.Run("IsGreaterThanOrEqual - greater", func(t *testing.T) {
		count1, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		count2, err := shared.NewConfirmationCount(3)
		require.NoError(t, err)

		assert.True(t, count1.IsGreaterThanOrEqual(count2))
		assert.False(t, count2.IsGreaterThanOrEqual(count1))
	})

	t.Run("IsGreaterThanOrEqual - equal", func(t *testing.T) {
		count1, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		count2, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)

		assert.True(t, count1.IsGreaterThanOrEqual(count2))
		assert.True(t, count2.IsGreaterThanOrEqual(count1))
	})

	t.Run("IsGreaterThanOrEqual - nil", func(t *testing.T) {
		count, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)

		assert.False(t, count.IsGreaterThanOrEqual(nil))
	})

	t.Run("String", func(t *testing.T) {
		count, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		assert.Equal(t, "5", count.String())
	})

	t.Run("Equals - same count", func(t *testing.T) {
		count1, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		count2, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		assert.True(t, count1.Equals(count2))
	})

	t.Run("Equals - different count", func(t *testing.T) {
		count1, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		count2, err := shared.NewConfirmationCount(3)
		require.NoError(t, err)
		assert.False(t, count1.Equals(count2))
	})

	t.Run("Equals - nil count", func(t *testing.T) {
		count, err := shared.NewConfirmationCount(5)
		require.NoError(t, err)
		assert.False(t, count.Equals(nil))
	})
}
