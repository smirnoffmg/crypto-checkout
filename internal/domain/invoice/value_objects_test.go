package invoice_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/domain/invoice"
)

func TestUSDTAmount_NewUSDTAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       string
		expectedErr bool
	}{
		{
			name:        "valid positive amount",
			value:       "100.00",
			expectedErr: false,
		},
		{
			name:        "valid amount with many decimals",
			value:       "100.123456789",
			expectedErr: false,
		},
		{
			name:        "valid zero amount",
			value:       "0.00",
			expectedErr: false,
		},
		{
			name:        "valid small amount",
			value:       "0.01",
			expectedErr: false,
		},
		{
			name:        "invalid negative amount",
			value:       "-100.00",
			expectedErr: true,
		},
		{
			name:        "invalid empty string",
			value:       "",
			expectedErr: true,
		},
		{
			name:        "invalid non-numeric string",
			value:       "abc",
			expectedErr: true,
		},
		{
			name:        "invalid amount with letters",
			value:       "100.00abc",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			amount, err := invoice.NewUSDTAmount(tt.value)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Nil(t, amount)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, amount)
				// For amounts with many decimals, expect rounded to 2 decimal places
				expectedString := tt.value
				if tt.value == "100.123456789" {
					expectedString = "100.12"
				}
				assert.Equal(t, expectedString, amount.String())
			}
		})
	}
}

func TestUSDTAmount_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		amount1        string
		amount2        string
		expectedResult string
		expectedErr    bool
	}{
		{
			name:           "add two positive amounts",
			amount1:        "100.00",
			amount2:        "50.00",
			expectedResult: "150.00",
			expectedErr:    false,
		},
		{
			name:           "add amounts with decimals",
			amount1:        "100.50",
			amount2:        "25.25",
			expectedResult: "125.75",
			expectedErr:    false,
		},
		{
			name:           "add zero amount",
			amount1:        "100.00",
			amount2:        "0.00",
			expectedResult: "100.00",
			expectedErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			amount1, err := invoice.NewUSDTAmount(tt.amount1)
			require.NoError(t, err)

			amount2, err := invoice.NewUSDTAmount(tt.amount2)
			require.NoError(t, err)

			result, err := amount1.Add(amount2)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result.String())
			}
		})
	}
}

func TestUSDTAmount_Multiply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		amount         string
		multiplier     decimal.Decimal
		expectedResult string
		expectedErr    bool
	}{
		{
			name:           "multiply by 2",
			amount:         "100.00",
			multiplier:     decimal.NewFromInt(2),
			expectedResult: "200.00",
			expectedErr:    false,
		},
		{
			name:           "multiply by decimal",
			amount:         "100.00",
			multiplier:     invoice.MustNewDecimal("0.1"),
			expectedResult: "10.00",
			expectedErr:    false,
		},
		{
			name:           "multiply by zero",
			amount:         "100.00",
			multiplier:     decimal.Zero,
			expectedResult: "0.00",
			expectedErr:    false,
		},
		{
			name:           "multiply with decimal result",
			amount:         "10.50",
			multiplier:     invoice.MustNewDecimal("2.5"),
			expectedResult: "26.25",
			expectedErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			amount, err := invoice.NewUSDTAmount(tt.amount)
			require.NoError(t, err)

			result, err := amount.Multiply(tt.multiplier)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result.String())
			}
		})
	}
}

func TestPaymentAddress_NewPaymentAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		address     string
		expectedErr bool
	}{
		{
			name:        "valid Tron address",
			address:     "TXYZabc123def456ghi789jkl012mno345pqr678stu901vwx234yz",
			expectedErr: false,
		},
		{
			name:        "valid Tron address with different format",
			address:     "TLyqzVGLV1srkB7dToTAEqgDSfPtXRJZYH",
			expectedErr: false,
		},
		{
			name:        "invalid empty address",
			address:     "",
			expectedErr: true,
		},
		{
			name:        "invalid short address",
			address:     "TXYZ",
			expectedErr: true,
		},
		{
			name:        "invalid address with invalid characters",
			address:     "TXYZabc123def456ghi789jkl012mno345pqr678stu901vwx234yz!",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			addr, err := invoice.NewPaymentAddress(tt.address)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Nil(t, addr)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, addr)
				assert.Equal(t, tt.address, addr.String())
			}
		})
	}
}

func TestPaymentAddress_Equals(t *testing.T) {
	t.Parallel()

	addr1, err := invoice.NewPaymentAddress("TXYZabc123def456ghi789jkl012mno345pqr678stu901vwx234yz")
	require.NoError(t, err)

	addr2, err := invoice.NewPaymentAddress("TXYZabc123def456ghi789jkl012mno345pqr678stu901vwx234yz")
	require.NoError(t, err)

	addr3, err := invoice.NewPaymentAddress("TLyqzVGLV1srkB7dToTAEqgDSfPtXRJZYH")
	require.NoError(t, err)

	assert.True(t, addr1.Equals(addr2))
	assert.False(t, addr1.Equals(addr3))
}
