package payment_test

import (
	"testing"

	"crypto-checkout/internal/domain/payment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUSDTAmount_NewUSDTAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid amount",
			value:       "100.50",
			expectError: false,
			errorMsg:    "",
		},
		{
			name:        "valid amount with zero decimals",
			value:       "100",
			expectError: false,
			errorMsg:    "",
		},
		{
			name:        "empty amount",
			value:       "",
			expectError: true,
			errorMsg:    "amount cannot be empty",
		},
		{
			name:        "negative amount",
			value:       "-10.50",
			expectError: true,
			errorMsg:    "amount cannot be negative",
		},
		{
			name:        "invalid format",
			value:       "abc",
			expectError: true,
			errorMsg:    "invalid amount format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			amount, err := payment.NewUSDTAmount(tt.value)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, amount)
			} else {
				require.NoError(t, err)
				require.NotNil(t, amount)
				// For non-error cases, check that the amount was created successfully
				// The string representation will be formatted with 2 decimal places
				assert.NotEmpty(t, amount.String())
			}
		})
	}
}

func TestUSDTAmount_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "amount with decimals",
			value:    "100.50",
			expected: "100.50",
		},
		{
			name:     "amount without decimals",
			value:    "100",
			expected: "100.00",
		},
		{
			name:     "amount with many decimals",
			value:    "100.123456",
			expected: "100.12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			amount := payment.MustNewUSDTAmount(tt.value)
			assert.Equal(t, tt.expected, amount.String())
		})
	}
}

func TestUSDTAmount_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		amount1  string
		amount2  string
		expected string
	}{
		{
			name:     "add two amounts",
			amount1:  "100.50",
			amount2:  "50.25",
			expected: "150.75",
		},
		{
			name:     "add with zero",
			amount1:  "100.50",
			amount2:  "0.00",
			expected: "100.50",
		},
		{
			name:     "add with precision",
			amount1:  "100.123",
			amount2:  "50.456",
			expected: "150.58",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			amount1 := payment.MustNewUSDTAmount(tt.amount1)
			amount2 := payment.MustNewUSDTAmount(tt.amount2)

			result, err := amount1.Add(amount2)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestUSDTAmount_Multiply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		amount     string
		multiplier string
		expected   string
	}{
		{
			name:       "multiply by decimal",
			amount:     "100.50",
			multiplier: "1.5",
			expected:   "150.75",
		},
		{
			name:       "multiply by integer",
			amount:     "100.50",
			multiplier: "2",
			expected:   "201.00",
		},
		{
			name:       "multiply with precision",
			amount:     "100.123",
			multiplier: "1.1",
			expected:   "110.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			amount := payment.MustNewUSDTAmount(tt.amount)
			multiplier := payment.MustNewDecimal(tt.multiplier)

			result, err := amount.Multiply(multiplier)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestUSDTAmount_Comparisons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		amount1      string
		amount2      string
		lessThan     bool
		greaterEqual bool
	}{
		{
			name:         "amount1 less than amount2",
			amount1:      "100.50",
			amount2:      "200.00",
			lessThan:     true,
			greaterEqual: false,
		},
		{
			name:         "amount1 greater than amount2",
			amount1:      "200.00",
			amount2:      "100.50",
			lessThan:     false,
			greaterEqual: true,
		},
		{
			name:         "amounts equal",
			amount1:      "100.50",
			amount2:      "100.50",
			lessThan:     false,
			greaterEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			amount1 := payment.MustNewUSDTAmount(tt.amount1)
			amount2 := payment.MustNewUSDTAmount(tt.amount2)

			assert.Equal(t, tt.lessThan, amount1.LessThan(amount2))
			assert.Equal(t, tt.greaterEqual, amount1.GreaterThanOrEqual(amount2))
		})
	}
}

func TestPaymentAddress_NewPaymentAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		address     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid Tron address",
			address:     "TLyqzVGLV1srkB7dToTAEqgDSfPtXRJZYH",
			expectError: false,
			errorMsg:    "",
		},
		{
			name:        "empty address",
			address:     "",
			expectError: true,
			errorMsg:    "address cannot be empty",
		},
		{
			name:        "address too short",
			address:     "T123",
			expectError: true,
			errorMsg:    "address too short",
		},
		{
			name:        "address not starting with T",
			address:     "ALyqzVGLV1srkB7dToTAEqgDSfPtXRJZYH",
			expectError: true,
			errorMsg:    "invalid Tron address format",
		},
		{
			name:        "address with invalid characters",
			address:     "TLyqzVGLV1srkB7dToTAEqgDSfPtXRJZYH!",
			expectError: true,
			errorMsg:    "invalid address format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			address, err := payment.NewPaymentAddress(tt.address)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, address)
			} else {
				require.NoError(t, err)
				require.NotNil(t, address)
				assert.Equal(t, tt.address, address.String())
			}
		})
	}
}

func TestPaymentAddress_Equals(t *testing.T) {
	t.Parallel()

	address1 := payment.MustNewPaymentAddress("TLyqzVGLV1srkB7dToTAEqgDSfPtXRJZYH")
	address2 := payment.MustNewPaymentAddress("TLyqzVGLV1srkB7dToTAEqgDSfPtXRJZYH")
	address3 := payment.MustNewPaymentAddress("TLyqzVGLV1srkB7dToTAEqgDSfPtXRJZYI")

	assert.True(t, address1.Equals(address2))
	assert.False(t, address1.Equals(address3))
}

func TestTransactionHash_NewTransactionHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		hash        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid transaction hash",
			hash:        "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456",
			expectError: false,
			errorMsg:    "",
		},
		{
			name:        "empty hash",
			hash:        "",
			expectError: true,
			errorMsg:    "transaction hash cannot be empty",
		},
		{
			name:        "hash too short",
			hash:        "a1b2c3d4",
			expectError: true,
			errorMsg:    "transaction hash too short",
		},
		{
			name:        "hash with invalid characters",
			hash:        "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456!",
			expectError: true,
			errorMsg:    "invalid transaction hash format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hash, err := payment.NewTransactionHash(tt.hash)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, hash)
			} else {
				require.NoError(t, err)
				require.NotNil(t, hash)
				assert.Equal(t, tt.hash, hash.String())
			}
		})
	}
}

func TestTransactionHash_Equals(t *testing.T) {
	t.Parallel()

	hash1 := payment.MustNewTransactionHash("a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456")
	hash2 := payment.MustNewTransactionHash("a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456")
	hash3 := payment.MustNewTransactionHash("a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123457")

	assert.True(t, hash1.Equals(hash2))
	assert.False(t, hash1.Equals(hash3))
}

func TestConfirmationCount_NewConfirmationCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		count       int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid count",
			count:       5,
			expectError: false,
			errorMsg:    "",
		},
		{
			name:        "zero count",
			count:       0,
			expectError: false,
			errorMsg:    "",
		},
		{
			name:        "negative count",
			count:       -1,
			expectError: true,
			errorMsg:    "confirmation count cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			count, err := payment.NewConfirmationCount(tt.count)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, count)
			} else {
				require.NoError(t, err)
				require.NotNil(t, count)
				assert.Equal(t, tt.count, count.Int())
			}
		})
	}
}

func TestConfirmationCount_Add(t *testing.T) {
	t.Parallel()

	count := payment.MustNewConfirmationCount(5)
	result := count.Add()

	assert.Equal(t, 6, result.Int())
	assert.Equal(t, 5, count.Int()) // Original should be unchanged
}

func TestConfirmationCount_GreaterThanOrEqual(t *testing.T) {
	t.Parallel()

	count1 := payment.MustNewConfirmationCount(5)
	count2 := payment.MustNewConfirmationCount(3)
	count3 := payment.MustNewConfirmationCount(5)

	assert.True(t, count1.GreaterThanOrEqual(count2))
	assert.True(t, count1.GreaterThanOrEqual(count3))
	assert.False(t, count2.GreaterThanOrEqual(count1))
}

func TestConfirmationCount_String(t *testing.T) {
	t.Parallel()

	count := payment.MustNewConfirmationCount(5)
	assert.Equal(t, "5", count.String())
}
