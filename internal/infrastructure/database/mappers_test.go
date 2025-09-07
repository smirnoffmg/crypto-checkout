package database_test

import (
	"testing"
	"time"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/shared"
	"crypto-checkout/internal/infrastructure/database"

	"github.com/stretchr/testify/require"
)

func TestInvoiceMapper(t *testing.T) {
	mapper := database.NewInvoiceMapper()

	t.Run("ToDomain", func(t *testing.T) {
		t.Run("Valid_Model", func(t *testing.T) {
			// Create test data
			itemsJSON := `[{"name": "Test Item", "description": "Test Description", "quantity": "2", "unit_price": "10.00"}]`
			model := &database.InvoiceModel{
				ID:               "test-invoice-id",
				MerchantID:       "test-merchant-id",
				CustomerID:       stringPtr("test-customer-id"),
				Title:            "Test Invoice",
				Description:      "Test Description",
				Items:            itemsJSON,
				Subtotal:         "20.00",
				Tax:              "2.00",
				Total:            "22.00",
				Currency:         "USD",
				CryptoCurrency:   "USDT",
				CryptoAmount:     "22.00",
				PaymentAddress:   stringPtr("TTestAddress123456789012345678901234567890"),
				Status:           "created",
				ExchangeRate:     `{"rate": "1.0", "from": "USD", "to": "USDT", "source": "default", "expires_at": "2025-01-01T00:00:00Z"}`,
				PaymentTolerance: `{"underpayment_threshold": "0.01", "overpayment_threshold": "1.0", "overpayment_action": "credit"}`,
				ExpiresAt:        timePtr(time.Now().Add(30 * time.Minute)),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				PaidAt:           nil,
			}

			// Convert to domain
			domain, err := mapper.ToDomain(model)
			require.NoError(t, err)
			require.NotNil(t, domain)

			// Verify basic fields
			require.Equal(t, "test-invoice-id", domain.ID())
			require.Equal(t, "test-merchant-id", domain.MerchantID())
			require.Equal(t, "test-customer-id", *domain.CustomerID())
			require.Equal(t, "Test Invoice", domain.Title())
			require.Equal(t, "Test Description", domain.Description())
			require.Equal(t, "created", domain.Status().String())

			// Verify items
			items := domain.Items()
			require.Len(t, items, 1)
			require.Equal(t, "Test Item", items[0].Name())
			require.Equal(t, "Test Description", items[0].Description())
			require.Equal(t, "2", items[0].Quantity().String())
			require.Equal(t, "10", items[0].UnitPrice().Amount().String())

			// Verify pricing
			pricing := domain.Pricing()
			require.Equal(t, "20", pricing.Subtotal().Amount().String())
			require.Equal(t, "2", pricing.Tax().Amount().String())
			require.Equal(t, "22", pricing.Total().Amount().String())

			// Verify payment address
			paymentAddress := domain.PaymentAddress()
			require.NotNil(t, paymentAddress)
			require.Equal(t, "TTestAddress123456789012345678901234567890", paymentAddress.String())

			// Verify expiration
			expiration := domain.Expiration()
			require.NotNil(t, expiration)
			require.False(t, expiration.IsExpired())
		})

		t.Run("Nil_Model", func(t *testing.T) {
			domain, err := mapper.ToDomain(nil)
			require.Error(t, err)
			require.Nil(t, domain)
			require.Contains(t, err.Error(), "invoice model cannot be nil")
		})

		t.Run("Empty_Items", func(t *testing.T) {
			model := &database.InvoiceModel{
				ID:             "test-id",
				MerchantID:     "test-merchant",
				Title:          "Test",
				Description:    "Test",
				Items:          "",
				Subtotal:       "10",
				Tax:            "1",
				Total:          "11",
				Currency:       "USD",
				CryptoCurrency: "USDT",
				Status:         "created",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			domain, err := mapper.ToDomain(model)
			require.Error(t, err)
			require.Nil(t, domain)
			require.Contains(t, err.Error(), "invoice must have at least one item")
		})

		t.Run("Invalid_Items_JSON", func(t *testing.T) {
			model := &database.InvoiceModel{
				ID:             "test-id",
				MerchantID:     "test-merchant",
				Title:          "Test",
				Description:    "Test",
				Items:          "invalid json",
				Subtotal:       "10",
				Tax:            "1",
				Total:          "11",
				Currency:       "USD",
				CryptoCurrency: "USDT",
				Status:         "created",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			domain, err := mapper.ToDomain(model)
			require.Error(t, err)
			require.Nil(t, domain)
			require.Contains(t, err.Error(), "failed to parse items JSON")
		})
	})

	t.Run("ToModel", func(t *testing.T) {
		t.Run("Valid_Domain", func(t *testing.T) {
			// Create test domain entity
			items := []*invoice.InvoiceItem{
				createTestInvoiceItem(t, "Test Item", "Test Description", "2", "10.00"),
			}

			subtotal, _ := shared.NewMoney("20.00", shared.CurrencyUSD)
			tax, _ := shared.NewMoney("2.00", shared.CurrencyUSD)
			total, _ := shared.NewMoney("22.00", shared.CurrencyUSD)
			pricing, _ := invoice.NewInvoicePricing(subtotal, tax, total)

			paymentAddress, _ := shared.NewPaymentAddress("TTestAddress123456789012345678901234567890", shared.NetworkTron)
			exchangeRate, _ := shared.NewExchangeRate("1.0", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "default", 30*time.Minute)
			paymentTolerance, _ := invoice.NewPaymentTolerance("0.01", "1.0", invoice.OverpaymentActionCredit)
			expiration := invoice.NewInvoiceExpiration(30 * time.Minute)

			domain, err := invoice.NewInvoice(
				"test-invoice-id",
				"test-merchant-id",
				"Test Invoice",
				"Test Description",
				items,
				pricing,
				shared.CryptoCurrencyUSDT,
				paymentAddress,
				exchangeRate,
				paymentTolerance,
				expiration,
				nil,
			)
			require.NoError(t, err)
			domain.SetCustomerID("test-customer-id")

			// Convert to model
			model := mapper.ToModel(domain)
			require.NotNil(t, model)

			// Verify basic fields
			require.Equal(t, "test-invoice-id", model.ID)
			require.Equal(t, "test-merchant-id", model.MerchantID)
			require.Equal(t, "test-customer-id", *model.CustomerID)
			require.Equal(t, "Test Invoice", model.Title)
			require.Equal(t, "Test Description", model.Description)
			require.Equal(t, "USD", model.Currency)
			require.Equal(t, "USDT", model.CryptoCurrency)
			require.Equal(t, "created", model.Status)

			// Verify pricing
			require.Equal(t, "20", model.Subtotal)
			require.Equal(t, "2", model.Tax)
			require.Equal(t, "22", model.Total)

			// Verify items JSON
			require.Contains(t, model.Items, "Test Item")
			require.Contains(t, model.Items, "Test Description")
			require.Contains(t, model.Items, "2")
			require.Contains(t, model.Items, "10")

			// Verify payment address
			require.NotNil(t, model.PaymentAddress)
			require.Equal(t, "TTestAddress123456789012345678901234567890", *model.PaymentAddress)

			// Verify expiration
			require.NotNil(t, model.ExpiresAt)
			require.True(t, model.ExpiresAt.After(time.Now()))
		})

		t.Run("Nil_Domain", func(t *testing.T) {
			model := mapper.ToModel(nil)
			require.Nil(t, model)
		})

		t.Run("Empty_Items", func(t *testing.T) {
			// Create domain with empty items - this should fail
			subtotal, _ := shared.NewMoney("10", shared.CurrencyUSD)
			tax, _ := shared.NewMoney("1", shared.CurrencyUSD)
			total, _ := shared.NewMoney("11", shared.CurrencyUSD)
			pricing, _ := invoice.NewInvoicePricing(subtotal, tax, total)

			paymentAddress, _ := shared.NewPaymentAddress("TTestAddress123456789012345678901234567890", shared.NetworkTron)
			exchangeRate, _ := shared.NewExchangeRate("1.0", shared.CurrencyUSD, shared.CryptoCurrencyUSDT, "default", 30*time.Minute)
			paymentTolerance, _ := invoice.NewPaymentTolerance("0.01", "1.0", invoice.OverpaymentActionCredit)
			expiration := invoice.NewInvoiceExpiration(30 * time.Minute)

			domain, err := invoice.NewInvoice(
				"test-id",
				"test-merchant",
				"Test",
				"Test",
				[]*invoice.InvoiceItem{}, // Empty items
				pricing,
				shared.CryptoCurrencyUSDT,
				paymentAddress,
				exchangeRate,
				paymentTolerance,
				expiration,
				nil,
			)
			require.Error(t, err)
			require.Nil(t, domain)
			require.Contains(t, err.Error(), "invoice must have at least one item")
		})
	})

	t.Run("ToDomainSlice", func(t *testing.T) {
		t.Run("Valid_Models", func(t *testing.T) {
			paymentAddr1 := "TTestAddress1111111111111111111111111111111111111111"
			paymentAddr2 := "TTestAddress2222222222222222222222222222222222222222"
			models := []database.InvoiceModel{
				{
					ID:             "invoice-1",
					MerchantID:     "merchant-1",
					Title:          "Invoice 1",
					Description:    "Description 1",
					Items:          `[{"name": "Item 1", "description": "Desc 1", "quantity": "1", "unit_price": "10"}]`,
					Subtotal:       "10",
					Tax:            "1",
					Total:          "11",
					Currency:       "USD",
					CryptoCurrency: "USDT",
					PaymentAddress: &paymentAddr1,
					Status:         "created",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				},
				{
					ID:             "invoice-2",
					MerchantID:     "merchant-2",
					Title:          "Invoice 2",
					Description:    "Description 2",
					Items:          `[{"name": "Item 2", "description": "Desc 2", "quantity": "2", "unit_price": "20"}]`,
					Subtotal:       "40",
					Tax:            "4",
					Total:          "44",
					Currency:       "USD",
					CryptoCurrency: "USDT",
					PaymentAddress: &paymentAddr2,
					Status:         "created",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				},
			}

			domains, err := mapper.ToDomainSlice(models)
			require.NoError(t, err)
			require.Len(t, domains, 2)

			require.Equal(t, "invoice-1", domains[0].ID())
			require.Equal(t, "Invoice 1", domains[0].Title())
			require.Equal(t, "invoice-2", domains[1].ID())
			require.Equal(t, "Invoice 2", domains[1].Title())
		})

		t.Run("Empty_Slice", func(t *testing.T) {
			domains, err := mapper.ToDomainSlice([]database.InvoiceModel{})
			require.NoError(t, err)
			require.Len(t, domains, 0)
		})

		t.Run("Invalid_Model", func(t *testing.T) {
			models := []database.InvoiceModel{
				{
					ID:             "invoice-1",
					MerchantID:     "merchant-1",
					Title:          "Invoice 1",
					Description:    "Description 1",
					Items:          "invalid json",
					Subtotal:       "10.00",
					Tax:            "1.00",
					Total:          "11.00",
					Currency:       "USD",
					CryptoCurrency: "USDT",
					Status:         "created",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				},
			}

			domains, err := mapper.ToDomainSlice(models)
			require.Error(t, err)
			require.Nil(t, domains)
			require.Contains(t, err.Error(), "failed to convert model 0")
		})
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func createTestInvoiceItem(t *testing.T, name, description, quantity, unitPrice string) *invoice.InvoiceItem {
	unitPriceMoney, err := shared.NewMoney(unitPrice, shared.CurrencyUSD)
	require.NoError(t, err)

	item, err := invoice.NewInvoiceItem(name, description, quantity, unitPriceMoney)
	require.NoError(t, err)

	return item
}
