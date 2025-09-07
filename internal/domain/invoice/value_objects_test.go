package invoice_test

import (
	"testing"
	"time"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/shared"

	"github.com/stretchr/testify/require"
)

func TestPaymentTolerance(t *testing.T) {
	t.Run("NewPaymentTolerance - valid tolerance", func(t *testing.T) {
		tolerance, err := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		require.NoError(t, err)
		require.Equal(t, "0.01", tolerance.UnderpaymentThreshold().String())
		require.Equal(t, "1", tolerance.OverpaymentThreshold().String())
		require.Equal(t, invoice.OverpaymentActionCredit, tolerance.OverpaymentAction())
	})

	t.Run("NewPaymentTolerance - empty underpayment threshold", func(t *testing.T) {
		_, err := invoice.NewPaymentTolerance("", "1.00", invoice.OverpaymentActionCredit)
		require.Error(t, err)
		require.Contains(t, err.Error(), "underpayment threshold cannot be empty")
	})

	t.Run("NewPaymentTolerance - empty overpayment threshold", func(t *testing.T) {
		_, err := invoice.NewPaymentTolerance("0.01", "", invoice.OverpaymentActionCredit)
		require.Error(t, err)
		require.Contains(t, err.Error(), "overpayment threshold cannot be empty")
	})

	t.Run("NewPaymentTolerance - invalid overpayment action", func(t *testing.T) {
		_, err := invoice.NewPaymentTolerance("0.01", "1.00", "invalid")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid overpayment action")
	})

	t.Run("NewPaymentTolerance - negative underpayment threshold", func(t *testing.T) {
		_, err := invoice.NewPaymentTolerance("-0.01", "1.00", invoice.OverpaymentActionCredit)
		require.Error(t, err)
		require.Contains(t, err.Error(), "underpayment threshold cannot be negative")
	})

	t.Run("NewPaymentTolerance - negative overpayment threshold", func(t *testing.T) {
		_, err := invoice.NewPaymentTolerance("0.01", "-1.00", invoice.OverpaymentActionCredit)
		require.Error(t, err)
		require.Contains(t, err.Error(), "overpayment threshold cannot be negative")
	})

	t.Run("NewPaymentTolerance - underpayment threshold too high", func(t *testing.T) {
		_, err := invoice.NewPaymentTolerance("1.5", "1.00", invoice.OverpaymentActionCredit)
		require.Error(t, err)
		require.Contains(t, err.Error(), "underpayment threshold cannot be greater than 1.0")
	})

	t.Run("DefaultPaymentTolerance", func(t *testing.T) {
		tolerance := invoice.DefaultPaymentTolerance()
		require.Equal(t, "0.01", tolerance.UnderpaymentThreshold().String())
		require.Equal(t, "1", tolerance.OverpaymentThreshold().String())
		require.Equal(t, invoice.OverpaymentActionCredit, tolerance.OverpaymentAction())
	})

	t.Run("IsUnderpayment - underpayment", func(t *testing.T) {
		tolerance, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		required, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		received, _ := shared.NewMoney("98.00", shared.CurrencyUSD) // 2% underpayment

		require.True(t, tolerance.IsUnderpayment(required, received))
	})

	t.Run("IsUnderpayment - sufficient payment", func(t *testing.T) {
		tolerance, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		required, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		received, _ := shared.NewMoney("100.00", shared.CurrencyUSD)

		require.False(t, tolerance.IsUnderpayment(required, received))
	})

	t.Run("IsOverpayment - overpayment", func(t *testing.T) {
		tolerance, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		required, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		received, _ := shared.NewMoney("102.00", shared.CurrencyUSD) // $2 overpayment

		require.True(t, tolerance.IsOverpayment(required, received))
	})

	t.Run("IsOverpayment - within tolerance", func(t *testing.T) {
		tolerance, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		required, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		received, _ := shared.NewMoney("100.50", shared.CurrencyUSD) // $0.50 overpayment

		require.False(t, tolerance.IsOverpayment(required, received))
	})

	t.Run("String", func(t *testing.T) {
		tolerance, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		require.Equal(t, "0.01:1:credit_account", tolerance.String())
	})

	t.Run("Equals - same tolerance", func(t *testing.T) {
		tolerance1, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		tolerance2, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		require.True(t, tolerance1.Equals(tolerance2))
	})

	t.Run("Equals - different tolerance", func(t *testing.T) {
		tolerance1, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		tolerance2, _ := invoice.NewPaymentTolerance("0.02", "1.00", invoice.OverpaymentActionCredit)
		require.False(t, tolerance1.Equals(tolerance2))
	})

	t.Run("Equals - nil tolerance", func(t *testing.T) {
		tolerance, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
		require.False(t, tolerance.Equals(nil))
	})
}

func TestInvoicePricing(t *testing.T) {
	t.Run("NewInvoicePricing - valid pricing", func(t *testing.T) {
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		tax, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		total, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

		pricing, err := invoice.NewInvoicePricing(subtotal, tax, total)
		require.NoError(t, err)
		require.Equal(t, subtotal, pricing.Subtotal())
		require.Equal(t, tax, pricing.Tax())
		require.Equal(t, total, pricing.Total())
	})

	t.Run("NewInvoicePricing - nil subtotal", func(t *testing.T) {
		tax, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		total, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

		_, err := invoice.NewInvoicePricing(nil, tax, total)
		require.Error(t, err)
		require.Contains(t, err.Error(), "subtotal cannot be nil")
	})

	t.Run("NewInvoicePricing - nil tax", func(t *testing.T) {
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		total, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

		_, err := invoice.NewInvoicePricing(subtotal, nil, total)
		require.Error(t, err)
		require.Contains(t, err.Error(), "tax cannot be nil")
	})

	t.Run("NewInvoicePricing - nil total", func(t *testing.T) {
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		tax, _ := shared.NewMoney("10.00", shared.CurrencyUSD)

		_, err := invoice.NewInvoicePricing(subtotal, tax, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "total cannot be nil")
	})

	t.Run("NewInvoicePricing - currency mismatch", func(t *testing.T) {
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		tax, _ := shared.NewMoney("10.00", shared.CurrencyEUR)
		total, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

		_, err := invoice.NewInvoicePricing(subtotal, tax, total)
		require.Error(t, err)
		require.Contains(t, err.Error(), "all amounts must have the same currency")
	})

	t.Run("NewInvoicePricing - total mismatch", func(t *testing.T) {
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		tax, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		total, _ := shared.NewMoney("120.00", shared.CurrencyUSD) // Wrong total

		_, err := invoice.NewInvoicePricing(subtotal, tax, total)
		require.Error(t, err)
		require.Contains(t, err.Error(), "total must equal subtotal plus tax")
	})

	t.Run("String", func(t *testing.T) {
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		tax, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		total, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

		pricing, _ := invoice.NewInvoicePricing(subtotal, tax, total)
		expected := "Subtotal: 100.00, Tax: 10.00, Total: 110.00"
		require.Equal(t, expected, pricing.String())
	})

	t.Run("Equals - same pricing", func(t *testing.T) {
		subtotal1, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		tax1, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		total1, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

		subtotal2, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		tax2, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		total2, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

		pricing1, _ := invoice.NewInvoicePricing(subtotal1, tax1, total1)
		pricing2, _ := invoice.NewInvoicePricing(subtotal2, tax2, total2)

		require.True(t, pricing1.Equals(pricing2))
	})

	t.Run("Equals - different pricing", func(t *testing.T) {
		subtotal1, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		tax1, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		total1, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

		subtotal2, _ := shared.NewMoney("200.00", shared.CurrencyUSD)
		tax2, _ := shared.NewMoney("20.00", shared.CurrencyUSD)
		total2, _ := shared.NewMoney("220.00", shared.CurrencyUSD)

		pricing1, _ := invoice.NewInvoicePricing(subtotal1, tax1, total1)
		pricing2, _ := invoice.NewInvoicePricing(subtotal2, tax2, total2)

		require.False(t, pricing1.Equals(pricing2))
	})

	t.Run("Equals - nil pricing", func(t *testing.T) {
		subtotal, _ := shared.NewMoney("100.00", shared.CurrencyUSD)
		tax, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		total, _ := shared.NewMoney("110.00", shared.CurrencyUSD)

		pricing, _ := invoice.NewInvoicePricing(subtotal, tax, total)
		require.False(t, pricing.Equals(nil))
	})
}

func TestInvoiceItem(t *testing.T) {
	t.Run("NewInvoiceItem - valid item", func(t *testing.T) {
		unitPrice, _ := shared.NewMoney("10.00", shared.CurrencyUSD)

		item, err := invoice.NewInvoiceItem("Test Item", "Test Description", "2", unitPrice)
		require.NoError(t, err)
		require.Equal(t, "Test Item", item.Name())
		require.Equal(t, "Test Description", item.Description())
		require.Equal(t, "2", item.Quantity().String())
		require.Equal(t, unitPrice, item.UnitPrice())
		require.Equal(t, "20.00", item.TotalPrice().String())
	})

	t.Run("NewInvoiceItem - empty name", func(t *testing.T) {
		unitPrice, _ := shared.NewMoney("10.00", shared.CurrencyUSD)

		_, err := invoice.NewInvoiceItem("", "Test Description", "2", unitPrice)
		require.Error(t, err)
		require.Contains(t, err.Error(), "item name cannot be empty")
	})

	t.Run("NewInvoiceItem - name too long", func(t *testing.T) {
		unitPrice, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		longName := string(make([]byte, 256)) // 256 characters

		_, err := invoice.NewInvoiceItem(longName, "Test Description", "2", unitPrice)
		require.Error(t, err)
		require.Contains(t, err.Error(), "item name cannot exceed 255 characters")
	})

	t.Run("NewInvoiceItem - description too long", func(t *testing.T) {
		unitPrice, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		longDescription := string(make([]byte, 1001)) // 1001 characters

		_, err := invoice.NewInvoiceItem("Test Item", longDescription, "2", unitPrice)
		require.Error(t, err)
		require.Contains(t, err.Error(), "item description cannot exceed 1000 characters")
	})

	t.Run("NewInvoiceItem - empty quantity", func(t *testing.T) {
		unitPrice, _ := shared.NewMoney("10.00", shared.CurrencyUSD)

		_, err := invoice.NewInvoiceItem("Test Item", "Test Description", "", unitPrice)
		require.Error(t, err)
		require.Contains(t, err.Error(), "quantity cannot be empty")
	})

	t.Run("NewInvoiceItem - zero quantity", func(t *testing.T) {
		unitPrice, _ := shared.NewMoney("10.00", shared.CurrencyUSD)

		_, err := invoice.NewInvoiceItem("Test Item", "Test Description", "0", unitPrice)
		require.Error(t, err)
		require.Contains(t, err.Error(), "quantity must be positive")
	})

	t.Run("NewInvoiceItem - negative quantity", func(t *testing.T) {
		unitPrice, _ := shared.NewMoney("10.00", shared.CurrencyUSD)

		_, err := invoice.NewInvoiceItem("Test Item", "Test Description", "-1", unitPrice)
		require.Error(t, err)
		require.Contains(t, err.Error(), "quantity must be positive")
	})

	t.Run("NewInvoiceItem - nil unit price", func(t *testing.T) {
		_, err := invoice.NewInvoiceItem("Test Item", "Test Description", "2", nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unit price cannot be nil")
	})

	t.Run("String", func(t *testing.T) {
		unitPrice, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		item, _ := invoice.NewInvoiceItem("Test Item", "Test Description", "2", unitPrice)

		expected := "Test Item x2 @ 10.00 = 20.00"
		require.Equal(t, expected, item.String())
	})

	t.Run("Equals - same item", func(t *testing.T) {
		unitPrice1, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		unitPrice2, _ := shared.NewMoney("10.00", shared.CurrencyUSD)

		item1, _ := invoice.NewInvoiceItem("Test Item", "Test Description", "2", unitPrice1)
		item2, _ := invoice.NewInvoiceItem("Test Item", "Test Description", "2", unitPrice2)

		require.True(t, item1.Equals(item2))
	})

	t.Run("Equals - different item", func(t *testing.T) {
		unitPrice1, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		unitPrice2, _ := shared.NewMoney("20.00", shared.CurrencyUSD)

		item1, _ := invoice.NewInvoiceItem("Test Item", "Test Description", "2", unitPrice1)
		item2, _ := invoice.NewInvoiceItem("Test Item", "Test Description", "2", unitPrice2)

		require.False(t, item1.Equals(item2))
	})

	t.Run("Equals - nil item", func(t *testing.T) {
		unitPrice, _ := shared.NewMoney("10.00", shared.CurrencyUSD)
		item, _ := invoice.NewInvoiceItem("Test Item", "Test Description", "2", unitPrice)

		require.False(t, item.Equals(nil))
	})
}

func TestInvoiceExpiration(t *testing.T) {
	t.Run("NewInvoiceExpiration - valid duration", func(t *testing.T) {
		duration := 30 * time.Minute
		expiration := invoice.NewInvoiceExpiration(duration)

		require.False(t, expiration.IsExpired())
		require.Equal(t, duration, expiration.Duration())
		require.True(t, expiration.TimeRemaining() > 0)
	})

	t.Run("NewInvoiceExpirationWithTime - valid time", func(t *testing.T) {
		futureTime := time.Now().UTC().Add(30 * time.Minute)
		expiration, err := invoice.NewInvoiceExpirationWithTime(futureTime)
		require.NoError(t, err)

		require.False(t, expiration.IsExpired())
		require.Equal(t, futureTime, expiration.ExpiresAt())
		require.True(t, expiration.TimeRemaining() > 0)
	})

	t.Run("NewInvoiceExpirationWithTime - past time", func(t *testing.T) {
		pastTime := time.Now().UTC().Add(-30 * time.Minute)
		_, err := invoice.NewInvoiceExpirationWithTime(pastTime)
		require.Error(t, err)
		require.Contains(t, err.Error(), "expiration time must be in the future")
	})

	t.Run("IsExpired - expired", func(t *testing.T) {
		pastDuration := -30 * time.Minute
		expiration := invoice.NewInvoiceExpiration(pastDuration)

		require.True(t, expiration.IsExpired())
		require.Equal(t, time.Duration(0), expiration.TimeRemaining())
	})

	t.Run("String", func(t *testing.T) {
		duration := 30 * time.Minute
		expiration := invoice.NewInvoiceExpiration(duration)

		require.Contains(t, expiration.String(), "Expires at:")
		require.Contains(t, expiration.String(), "30m0s")
	})

	t.Run("Equals - same expiration", func(t *testing.T) {
		// Use the same duration to ensure both have the same duration
		duration := 30 * time.Minute
		expiration1 := invoice.NewInvoiceExpiration(duration)
		expiration2 := invoice.NewInvoiceExpiration(duration)

		// They should be equal if created with the same duration
		require.True(t, expiration1.Equals(expiration2))
	})

	t.Run("Equals - different expiration", func(t *testing.T) {
		expiration1 := invoice.NewInvoiceExpiration(30 * time.Minute)
		expiration2 := invoice.NewInvoiceExpiration(60 * time.Minute)

		require.False(t, expiration1.Equals(expiration2))
	})

	t.Run("Equals - nil expiration", func(t *testing.T) {
		expiration := invoice.NewInvoiceExpiration(30 * time.Minute)
		require.False(t, expiration.Equals(nil))
	})
}
