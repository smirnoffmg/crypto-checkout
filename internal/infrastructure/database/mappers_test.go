package database_test

import (
	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/shared"
	"crypto-checkout/internal/infrastructure/database"
	"testing"
	"time"

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
				ExchangeRate:     `{"rate": "1.0", "from": "USD", "to": "USDT", "source": "default", "locked_at": "2024-12-31T23:30:00Z", "expires_at": "2025-01-01T00:00:00Z"}`,
				PaymentTolerance: `{"underpayment_threshold": "0.01", "overpayment_threshold": "1.00", "overpayment_action": "credit_account"}`,
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

			// Test that empty items are properly rejected by domain validation
			domain, err := mapper.ToDomain(model)
			require.Error(t, err)
			require.Nil(t, domain)
			require.Contains(t, err.Error(), "Field validation for 'Items' failed on the 'min' tag")
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

			paymentAddress, _ := shared.NewPaymentAddress(
				"TTestAddress123456789012345678901234567890",
				shared.NetworkTron,
			)
			exchangeRate, _ := shared.NewExchangeRate(
				"1.0",
				shared.CurrencyUSD,
				shared.CryptoCurrencyUSDT,
				"default",
				30*time.Minute,
			)
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

			paymentAddress, _ := shared.NewPaymentAddress(
				"TTestAddress123456789012345678901234567890",
				shared.NetworkTron,
			)
			exchangeRate, _ := shared.NewExchangeRate(
				"1.0",
				shared.CurrencyUSD,
				shared.CryptoCurrencyUSDT,
				"default",
				30*time.Minute,
			)
			paymentTolerance, _ := invoice.NewPaymentTolerance("0.01", "1.0", invoice.OverpaymentActionCredit)
			expiration := invoice.NewInvoiceExpiration(30 * time.Minute)

			// Test that empty items are properly rejected
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
			require.Contains(t, err.Error(), "Field validation for 'Items' failed on the 'min' tag")
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

	t.Run("JSONB_Serialization", func(t *testing.T) {
		t.Run("SerializeExchangeRate", func(t *testing.T) {
			// Create a test exchange rate
			exchangeRate, err := shared.NewExchangeRate(
				"1.5",
				shared.CurrencyUSD,
				shared.CryptoCurrencyUSDT,
				"test-source",
				30*time.Minute,
			)
			require.NoError(t, err)

			// Test serialization
			jsonStr, err := mapper.SerializeExchangeRate(exchangeRate)
			require.NoError(t, err)
			require.NotEmpty(t, jsonStr)

			// Verify JSON structure
			require.Contains(t, jsonStr, `"rate":"1.5"`)
			require.Contains(t, jsonStr, `"from":"USD"`)
			require.Contains(t, jsonStr, `"to":"USDT"`)
			require.Contains(t, jsonStr, `"source":"test-source"`)
			require.Contains(t, jsonStr, `"locked_at"`)
			require.Contains(t, jsonStr, `"expires_at"`)
		})

		t.Run("SerializeExchangeRate_Nil", func(t *testing.T) {
			jsonStr, err := mapper.SerializeExchangeRate(nil)
			require.NoError(t, err)
			require.Empty(t, jsonStr)
		})

		t.Run("SerializePaymentTolerance", func(t *testing.T) {
			// Create a test payment tolerance
			paymentTolerance, err := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
			require.NoError(t, err)

			// Test serialization
			jsonStr, err := mapper.SerializePaymentTolerance(paymentTolerance)
			require.NoError(t, err)
			require.NotEmpty(t, jsonStr)

			// Verify JSON structure
			require.Contains(t, jsonStr, `"underpayment_threshold":"0.01"`)
			require.Contains(t, jsonStr, `"overpayment_threshold":"1.00"`)
			require.Contains(t, jsonStr, `"overpayment_action":"credit_account"`)
		})

		t.Run("SerializePaymentTolerance_Nil", func(t *testing.T) {
			jsonStr, err := mapper.SerializePaymentTolerance(nil)
			require.NoError(t, err)
			require.Empty(t, jsonStr)
		})

		t.Run("ToModel_WithExchangeRateAndPaymentTolerance", func(t *testing.T) {
			// Create test invoice with exchange rate and payment tolerance
			items := []*invoice.InvoiceItem{
				createTestInvoiceItem(t, "Test Item", "Test Description", "2", "10.00"),
			}
			pricing := createTestInvoicePricing(t, "20.00", "2.00", "22.00")
			exchangeRate, _ := shared.NewExchangeRate(
				"1.5",
				shared.CurrencyUSD,
				shared.CryptoCurrencyUSDT,
				"test-source",
				30*time.Minute,
			)
			paymentTolerance, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)
			paymentAddress, _ := shared.NewPaymentAddress(
				"0x1234567890123456789012345678901234567890",
				shared.NetworkEthereum,
			)
			expiration := invoice.NewInvoiceExpiration(30 * time.Minute)

			inv, err := invoice.NewInvoice(
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

			// Convert to model
			model := mapper.ToModel(inv)
			require.NotNil(t, model)

			// Verify exchange rate serialization
			require.NotEmpty(t, model.ExchangeRate)
			require.Contains(t, model.ExchangeRate, `"rate":"1.5"`)
			require.Contains(t, model.ExchangeRate, `"from":"USD"`)
			require.Contains(t, model.ExchangeRate, `"to":"USDT"`)
			require.Contains(t, model.ExchangeRate, `"source":"test-source"`)

			// Verify payment tolerance serialization
			require.NotEmpty(t, model.PaymentTolerance)
			require.Contains(t, model.PaymentTolerance, `"underpayment_threshold":"0.01"`)
			require.Contains(t, model.PaymentTolerance, `"overpayment_threshold":"1.00"`)
			require.Contains(t, model.PaymentTolerance, `"overpayment_action":"credit_account"`)
		})

		t.Run("ToModel_Serialization_EdgeCases", func(t *testing.T) {
			// Test that serialization methods handle nil values correctly
			// This tests the edge case handling in the ToModel method

			// Test nil exchange rate serialization
			exchangeRateJSON, err := mapper.SerializeExchangeRate(nil)
			require.NoError(t, err)
			require.Empty(t, exchangeRateJSON)

			// Test nil payment tolerance serialization
			paymentToleranceJSON, err := mapper.SerializePaymentTolerance(nil)
			require.NoError(t, err)
			require.Empty(t, paymentToleranceJSON)
		})

		t.Run("DeserializeExchangeRate", func(t *testing.T) {
			// Test deserialization of valid exchange rate JSON
			jsonStr := `{"rate":"1.5","from":"USD","to":"USDT","source":"test-source","locked_at":"2024-01-01T00:00:00Z","expires_at":"2024-01-01T00:30:00Z"}`

			exchangeRate, err := mapper.DeserializeExchangeRate(jsonStr)
			require.NoError(t, err)
			require.NotNil(t, exchangeRate)
			require.Equal(t, "1.5", exchangeRate.Rate().String())
			require.Equal(t, string(shared.CurrencyUSD), string(exchangeRate.FromCurrency()))
			require.Equal(t, string(shared.CryptoCurrencyUSDT), string(exchangeRate.ToCurrency()))
			require.Equal(t, "test-source", exchangeRate.Source())
		})

		t.Run("DeserializeExchangeRate_Empty", func(t *testing.T) {
			exchangeRate, err := mapper.DeserializeExchangeRate("")
			require.NoError(t, err)
			require.Nil(t, exchangeRate)
		})

		t.Run("DeserializeExchangeRate_InvalidJSON", func(t *testing.T) {
			exchangeRate, err := mapper.DeserializeExchangeRate("invalid json")
			require.Error(t, err)
			require.Nil(t, exchangeRate)
			require.Contains(t, err.Error(), "failed to unmarshal exchange rate")
		})

		t.Run("DeserializePaymentTolerance", func(t *testing.T) {
			// Test deserialization of valid payment tolerance JSON
			jsonStr := `{"underpayment_threshold":"0.01","overpayment_threshold":"1.00","overpayment_action":"credit_account"}`

			paymentTolerance, err := mapper.DeserializePaymentTolerance(jsonStr)
			require.NoError(t, err)
			require.NotNil(t, paymentTolerance)
			require.Equal(t, "0.01", paymentTolerance.UnderpaymentThreshold().StringFixed(2))
			require.Equal(t, "1.00", paymentTolerance.OverpaymentThreshold().StringFixed(2))
			require.Equal(t, string(invoice.OverpaymentActionCredit), string(paymentTolerance.OverpaymentAction()))
		})

		t.Run("DeserializePaymentTolerance_Empty", func(t *testing.T) {
			paymentTolerance, err := mapper.DeserializePaymentTolerance("")
			require.NoError(t, err)
			require.Nil(t, paymentTolerance)
		})

		t.Run("DeserializePaymentTolerance_InvalidJSON", func(t *testing.T) {
			paymentTolerance, err := mapper.DeserializePaymentTolerance("invalid json")
			require.Error(t, err)
			require.Nil(t, paymentTolerance)
			require.Contains(t, err.Error(), "failed to unmarshal payment tolerance")
		})

		t.Run("RoundTrip_Serialization", func(t *testing.T) {
			// Test that serialization and deserialization are inverse operations

			// Create original objects
			originalExchangeRate, _ := shared.NewExchangeRate(
				"1.5",
				shared.CurrencyUSD,
				shared.CryptoCurrencyUSDT,
				"test-source",
				30*time.Minute,
			)
			originalPaymentTolerance, _ := invoice.NewPaymentTolerance("0.01", "1.00", invoice.OverpaymentActionCredit)

			// Serialize
			exchangeRateJSON, err := mapper.SerializeExchangeRate(originalExchangeRate)
			require.NoError(t, err)

			paymentToleranceJSON, err := mapper.SerializePaymentTolerance(originalPaymentTolerance)
			require.NoError(t, err)

			// Deserialize
			deserializedExchangeRate, err := mapper.DeserializeExchangeRate(exchangeRateJSON)
			require.NoError(t, err)

			deserializedPaymentTolerance, err := mapper.DeserializePaymentTolerance(paymentToleranceJSON)
			require.NoError(t, err)

			// Verify they match
			require.Equal(t, originalExchangeRate.Rate().String(), deserializedExchangeRate.Rate().String())
			require.Equal(t, originalExchangeRate.FromCurrency(), deserializedExchangeRate.FromCurrency())
			require.Equal(t, originalExchangeRate.ToCurrency(), deserializedExchangeRate.ToCurrency())
			require.Equal(t, originalExchangeRate.Source(), deserializedExchangeRate.Source())

			require.Equal(
				t,
				originalPaymentTolerance.UnderpaymentThreshold().StringFixed(2),
				deserializedPaymentTolerance.UnderpaymentThreshold().StringFixed(2),
			)
			require.Equal(
				t,
				originalPaymentTolerance.OverpaymentThreshold().StringFixed(2),
				deserializedPaymentTolerance.OverpaymentThreshold().StringFixed(2),
			)
			require.Equal(
				t,
				originalPaymentTolerance.OverpaymentAction(),
				deserializedPaymentTolerance.OverpaymentAction(),
			)
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

func createTestInvoicePricing(t *testing.T, subtotal, tax, total string) *invoice.InvoicePricing {
	subtotalMoney, err := shared.NewMoney(subtotal, shared.CurrencyUSD)
	require.NoError(t, err)

	taxMoney, err := shared.NewMoney(tax, shared.CurrencyUSD)
	require.NoError(t, err)

	totalMoney, err := shared.NewMoney(total, shared.CurrencyUSD)
	require.NoError(t, err)

	pricing, err := invoice.NewInvoicePricing(subtotalMoney, taxMoney, totalMoney)
	require.NoError(t, err)

	return pricing
}
