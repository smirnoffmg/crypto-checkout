package invoice_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/domain/invoice"
)

func TestInvoice_NewInvoice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		items       []*invoice.InvoiceItem
		taxRate     decimal.Decimal
		expectedErr bool
	}{
		{
			name: "valid invoice with single item",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
			},
			taxRate:     invoice.MustNewDecimal("0.1"),
			expectedErr: false,
		},
		{
			name: "valid invoice with multiple items",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(2)),
				invoice.MustNewInvoiceItem("Product 2", invoice.MustNewUSDTAmount("50.00"), decimal.NewFromInt(1)),
			},
			taxRate:     invoice.MustNewDecimal("0.1"),
			expectedErr: false,
		},
		{
			name:        "invalid invoice with no items",
			items:       []*invoice.InvoiceItem{},
			taxRate:     invoice.MustNewDecimal("0.1"),
			expectedErr: true,
		},
		{
			name: "invalid invoice with negative tax rate",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
			},
			taxRate:     invoice.MustNewDecimal("-0.1"),
			expectedErr: true,
		},
		{
			name: "invalid invoice with tax rate over 100%",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
			},
			taxRate:     invoice.MustNewDecimal("1.5"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inv, err := invoice.NewInvoice(tt.items, tt.taxRate)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Nil(t, inv)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, inv)
				assert.Equal(t, tt.items, inv.Items())
				assert.Equal(t, tt.taxRate, inv.TaxRate())
				assert.Equal(t, invoice.StatusCreated, inv.Status())
				assert.NotEmpty(t, inv.ID())
				assert.WithinDuration(t, time.Now(), inv.CreatedAt(), time.Second)
				assert.Nil(t, inv.PaymentAddress())
				assert.Nil(t, inv.PaidAt())
			}
		})
	}
}

func TestInvoice_CalculateSubtotal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		items          []*invoice.InvoiceItem
		expectedAmount *invoice.USDTAmount
	}{
		{
			name: "single item",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
			},
			expectedAmount: invoice.MustNewUSDTAmount("100.00"),
		},
		{
			name: "multiple items",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(2)),
				invoice.MustNewInvoiceItem("Product 2", invoice.MustNewUSDTAmount("50.00"), decimal.NewFromInt(1)),
			},
			expectedAmount: invoice.MustNewUSDTAmount("250.00"),
		},
		{
			name: "items with decimal quantities",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem(
					"Product 1",
					invoice.MustNewUSDTAmount("10.50"),
					invoice.MustNewDecimal("2.5"),
				),
			},
			expectedAmount: invoice.MustNewUSDTAmount("26.25"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inv, err := invoice.NewInvoice(tt.items, invoice.MustNewDecimal("0.1"))
			require.NoError(t, err)

			subtotal := inv.CalculateSubtotal()
			assert.Equal(t, tt.expectedAmount.String(), subtotal.String())
		})
	}
}

func TestInvoice_CalculateTax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		items          []*invoice.InvoiceItem
		taxRate        decimal.Decimal
		expectedAmount *invoice.USDTAmount
	}{
		{
			name: "10% tax on 100 USDT",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
			},
			taxRate:        invoice.MustNewDecimal("0.1"),
			expectedAmount: invoice.MustNewUSDTAmount("10.00"),
		},
		{
			name: "15% tax on 250 USDT",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(2)),
				invoice.MustNewInvoiceItem("Product 2", invoice.MustNewUSDTAmount("50.00"), decimal.NewFromInt(1)),
			},
			taxRate:        invoice.MustNewDecimal("0.15"),
			expectedAmount: invoice.MustNewUSDTAmount("37.50"),
		},
		{
			name: "no tax",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
			},
			taxRate:        invoice.MustNewDecimal("0.00"),
			expectedAmount: invoice.MustNewUSDTAmount("0.00"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inv, err := invoice.NewInvoice(tt.items, tt.taxRate)
			require.NoError(t, err)

			tax := inv.CalculateTax()
			assert.Equal(t, tt.expectedAmount.String(), tax.String())
		})
	}
}

func TestInvoice_CalculateTotal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		items          []*invoice.InvoiceItem
		taxRate        decimal.Decimal
		expectedAmount *invoice.USDTAmount
	}{
		{
			name: "100 USDT + 10% tax = 110 USDT",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
			},
			taxRate:        invoice.MustNewDecimal("0.1"),
			expectedAmount: invoice.MustNewUSDTAmount("110.00"),
		},
		{
			name: "250 USDT + 15% tax = 287.50 USDT",
			items: []*invoice.InvoiceItem{
				invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(2)),
				invoice.MustNewInvoiceItem("Product 2", invoice.MustNewUSDTAmount("50.00"), decimal.NewFromInt(1)),
			},
			taxRate:        invoice.MustNewDecimal("0.15"),
			expectedAmount: invoice.MustNewUSDTAmount("287.50"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inv, err := invoice.NewInvoice(tt.items, tt.taxRate)
			require.NoError(t, err)

			total := inv.CalculateTotal()
			assert.Equal(t, tt.expectedAmount.String(), total.String())
		})
	}
}

func TestInvoice_AssignPaymentAddress(t *testing.T) {
	t.Parallel()

	inv, err := invoice.NewInvoice(
		[]*invoice.InvoiceItem{
			invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
		},
		invoice.MustNewDecimal("0.1"),
	)
	require.NoError(t, err)

	address := invoice.MustNewPaymentAddress("TLyqzVGLV1srkB7dToTAEqgDSfPtXRJZYH")

	err = inv.AssignPaymentAddress(address)
	require.NoError(t, err)
	assert.Equal(t, address, inv.PaymentAddress())
}

func TestInvoice_MarkAsPaid(t *testing.T) {
	t.Parallel()

	inv, err := invoice.NewInvoice(
		[]*invoice.InvoiceItem{
			invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
		},
		invoice.MustNewDecimal("0.1"),
	)
	require.NoError(t, err)

	// Initially created
	assert.Equal(t, invoice.StatusCreated, inv.Status())
	assert.Nil(t, inv.PaidAt())

	// Mark as viewed first
	err = inv.MarkAsViewed()
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPending, inv.Status())

	// Now mark as paid (which completes and confirms)
	err = inv.MarkAsPaid()
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPaid, inv.Status())
	assert.NotNil(t, inv.PaidAt())
	assert.WithinDuration(t, time.Now(), *inv.PaidAt(), time.Second)
}

func TestInvoice_Expire(t *testing.T) {
	t.Parallel()

	inv, err := invoice.NewInvoice(
		[]*invoice.InvoiceItem{
			invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
		},
		invoice.MustNewDecimal("0.1"),
	)
	require.NoError(t, err)

	// Can expire from created state
	err = inv.Expire()
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusExpired, inv.Status())
}

func TestInvoice_Cancel(t *testing.T) {
	t.Parallel()

	inv, err := invoice.NewInvoice(
		[]*invoice.InvoiceItem{
			invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
		},
		invoice.MustNewDecimal("0.1"),
	)
	require.NoError(t, err)

	// Can cancel from created state
	err = inv.Cancel()
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusCancelled, inv.Status())
}

func TestInvoice_IsPaid(t *testing.T) {
	t.Parallel()

	inv, err := invoice.NewInvoice(
		[]*invoice.InvoiceItem{
			invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
		},
		invoice.MustNewDecimal("0.1"),
	)
	require.NoError(t, err)

	assert.False(t, inv.IsPaid())

	// First mark as viewed, then mark as paid
	err = inv.MarkAsViewed()
	require.NoError(t, err)

	err = inv.MarkAsPaid()
	require.NoError(t, err)

	assert.True(t, inv.IsPaid())
}

func TestInvoice_IsPending(t *testing.T) {
	t.Parallel()

	inv, err := invoice.NewInvoice(
		[]*invoice.InvoiceItem{
			invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
		},
		invoice.MustNewDecimal("0.1"),
	)
	require.NoError(t, err)

	// Initially created, not pending
	assert.False(t, inv.IsPending())

	// Mark as viewed to become pending
	err = inv.MarkAsViewed()
	require.NoError(t, err)
	assert.True(t, inv.IsPending())

	// Mark as paid to no longer be pending
	err = inv.MarkAsPaid()
	require.NoError(t, err)

	assert.False(t, inv.IsPending())
}

func TestInvoice_FSMTransitions(t *testing.T) {
	t.Parallel()

	inv := createTestInvoice(t)

	// Test Created -> Pending
	assert.Equal(t, invoice.StatusCreated, inv.Status())
	err := inv.MarkAsViewed()
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPending, inv.Status())

	// Test Pending -> Partial
	err = inv.MarkAsPartial()
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPartial, inv.Status())

	// Test Partial -> Confirming
	err = inv.MarkAsCompleted()
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusConfirming, inv.Status())

	// Test Confirming -> Paid
	err = inv.MarkAsConfirmed()
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusPaid, inv.Status())
	assert.NotNil(t, inv.PaidAt())

	// Test Paid -> Refunded
	err = inv.Refund()
	require.NoError(t, err)
	assert.Equal(t, invoice.StatusRefunded, inv.Status())
}

func TestInvoice_FSMInvalidTransitions(t *testing.T) {
	t.Parallel()

	inv := createTestInvoice(t)

	// Try invalid transition from Created to Partial
	err := inv.MarkAsPartial()
	require.Error(t, err)
	assert.Equal(t, invoice.StatusCreated, inv.Status())

	// Try invalid transition from Created to Completed
	err = inv.MarkAsCompleted()
	require.Error(t, err)
	assert.Equal(t, invoice.StatusCreated, inv.Status())

	// Try invalid transition from Created to Confirmed
	err = inv.MarkAsConfirmed()
	require.Error(t, err)
	assert.Equal(t, invoice.StatusCreated, inv.Status())
}

func TestInvoice_FSMStateProperties(t *testing.T) {
	t.Parallel()

	inv := createTestInvoice(t)

	// Created state
	assert.True(t, inv.IsActive())
	assert.False(t, inv.IsTerminal())
	assert.False(t, inv.IsPaid())
	assert.False(t, inv.IsPending())

	// Move to Pending
	err := inv.MarkAsViewed()
	require.NoError(t, err)
	assert.True(t, inv.IsActive())
	assert.False(t, inv.IsTerminal())
	assert.False(t, inv.IsPaid())
	assert.True(t, inv.IsPending())

	// Move to Paid
	err = inv.MarkAsPaid()
	require.NoError(t, err)
	assert.False(t, inv.IsActive())
	assert.False(t, inv.IsTerminal())
	assert.True(t, inv.IsPaid())
	assert.False(t, inv.IsPending())

	// Move to Refunded
	err = inv.Refund()
	require.NoError(t, err)
	assert.False(t, inv.IsActive())
	assert.True(t, inv.IsTerminal())
	assert.True(t, inv.IsPaid())
	assert.False(t, inv.IsPending())
}

func TestInvoice_FSMPermittedTriggers(t *testing.T) {
	t.Parallel()

	inv := createTestInvoice(t)

	// Created state should allow viewed, expired, cancelled
	triggers := inv.GetPermittedTriggers()
	assert.Contains(t, triggers, invoice.TriggerViewed)
	assert.Contains(t, triggers, invoice.TriggerExpired)
	assert.Contains(t, triggers, invoice.TriggerCancelled)
	assert.NotContains(t, triggers, invoice.TriggerPartial)
	assert.NotContains(t, triggers, invoice.TriggerCompleted)

	// Move to Pending
	err := inv.MarkAsViewed()
	require.NoError(t, err)

	// Pending state should allow partial, completed, expired, cancelled
	triggers = inv.GetPermittedTriggers()
	assert.Contains(t, triggers, invoice.TriggerPartial)
	assert.Contains(t, triggers, invoice.TriggerCompleted)
	assert.Contains(t, triggers, invoice.TriggerExpired)
	assert.Contains(t, triggers, invoice.TriggerCancelled)
	assert.NotContains(t, triggers, invoice.TriggerViewed)
}

func TestInvoice_FSMCanTransition(t *testing.T) {
	t.Parallel()

	inv := createTestInvoice(t)

	// Created state
	assert.True(t, inv.CanTransition(invoice.TriggerViewed))
	assert.True(t, inv.CanTransition(invoice.TriggerExpired))
	assert.True(t, inv.CanTransition(invoice.TriggerCancelled))
	assert.False(t, inv.CanTransition(invoice.TriggerPartial))
	assert.False(t, inv.CanTransition(invoice.TriggerCompleted))

	// Move to Pending
	err := inv.MarkAsViewed()
	require.NoError(t, err)

	// Pending state
	assert.False(t, inv.CanTransition(invoice.TriggerViewed))
	assert.True(t, inv.CanTransition(invoice.TriggerPartial))
	assert.True(t, inv.CanTransition(invoice.TriggerCompleted))
	assert.True(t, inv.CanTransition(invoice.TriggerExpired))
	assert.True(t, inv.CanTransition(invoice.TriggerCancelled))
}

// createTestInvoice creates a test invoice for testing purposes.
func createTestInvoice(t *testing.T) *invoice.Invoice {
	t.Helper()

	inv, err := invoice.NewInvoice(
		[]*invoice.InvoiceItem{
			invoice.MustNewInvoiceItem("Product 1", invoice.MustNewUSDTAmount("100.00"), decimal.NewFromInt(1)),
		},
		invoice.MustNewDecimal("0.1"),
	)
	require.NoError(t, err)
	return inv
}
