package invoice_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/domain/invoice"
)

func TestInvoiceItem_NewInvoiceItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		unitPrice   *invoice.USDTAmount
		quantity    decimal.Decimal
		expectedErr bool
	}{
		{
			name:        "valid item",
			description: "Product 1",
			unitPrice:   invoice.MustNewUSDTAmount("100.00"),
			quantity:    decimal.NewFromInt(1),
			expectedErr: false,
		},
		{
			name:        "valid item with decimal quantity",
			description: "Product 1",
			unitPrice:   invoice.MustNewUSDTAmount("10.50"),
			quantity:    invoice.MustNewDecimal("2.5"),
			expectedErr: false,
		},
		{
			name:        "invalid empty description",
			description: "",
			unitPrice:   invoice.MustNewUSDTAmount("100.00"),
			quantity:    decimal.NewFromInt(1),
			expectedErr: true,
		},
		{
			name:        "invalid zero quantity",
			description: "Product 1",
			unitPrice:   invoice.MustNewUSDTAmount("100.00"),
			quantity:    decimal.Zero,
			expectedErr: true,
		},
		{
			name:        "invalid negative quantity",
			description: "Product 1",
			unitPrice:   invoice.MustNewUSDTAmount("100.00"),
			quantity:    decimal.NewFromInt(-1),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			item, err := invoice.NewInvoiceItem(tt.description, tt.unitPrice, tt.quantity)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Nil(t, item)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, item)
				assert.Equal(t, tt.description, item.Description())
				assert.Equal(t, tt.unitPrice, item.UnitPrice())
				assert.Equal(t, tt.quantity, item.Quantity())
			}
		})
	}
}

func TestInvoiceItem_CalculateTotal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		unitPrice      *invoice.USDTAmount
		quantity       decimal.Decimal
		expectedAmount *invoice.USDTAmount
	}{
		{
			name:           "100 USDT * 1 = 100 USDT",
			unitPrice:      invoice.MustNewUSDTAmount("100.00"),
			quantity:       decimal.NewFromInt(1),
			expectedAmount: invoice.MustNewUSDTAmount("100.00"),
		},
		{
			name:           "100 USDT * 2 = 200 USDT",
			unitPrice:      invoice.MustNewUSDTAmount("100.00"),
			quantity:       decimal.NewFromInt(2),
			expectedAmount: invoice.MustNewUSDTAmount("200.00"),
		},
		{
			name:           "10.50 USDT * 2.5 = 26.25 USDT",
			unitPrice:      invoice.MustNewUSDTAmount("10.50"),
			quantity:       invoice.MustNewDecimal("2.5"),
			expectedAmount: invoice.MustNewUSDTAmount("26.25"),
		},
		{
			name:           "0.01 USDT * 1000 = 10.00 USDT",
			unitPrice:      invoice.MustNewUSDTAmount("0.01"),
			quantity:       decimal.NewFromInt(1000),
			expectedAmount: invoice.MustNewUSDTAmount("10.00"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			item, err := invoice.NewInvoiceItem("Test Product", tt.unitPrice, tt.quantity)
			require.NoError(t, err)

			total := item.CalculateTotal()
			assert.Equal(t, tt.expectedAmount.String(), total.String())
		})
	}
}
