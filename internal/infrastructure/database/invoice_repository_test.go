package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/shared"
	"crypto-checkout/internal/infrastructure/database"
	"crypto-checkout/pkg/config"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	cfg := config.DatabaseConfig{
		URL: "file::memory:",
	}

	conn, err := database.NewConnection(cfg)
	require.NoError(t, err)

	// Run migrations
	err = conn.Migrate()
	require.NoError(t, err)

	return conn.DB
}

func createTestInvoice(t *testing.T) *invoice.Invoice {
	return createTestInvoiceWithID(t, "test-invoice-id")
}

func createTestInvoiceWithID(t *testing.T, id string) *invoice.Invoice {
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

inv, err := invoice.NewInvoice(
		id,
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
	return inv
}

func TestInvoiceRepository(t *testing.T) {
	t.Run("Save", func(t *testing.T) {
		t.Run("Valid_Invoice", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			inv := createTestInvoice(t)

			err := repo.Save(ctx, inv)
			require.NoError(t, err)

			// Verify invoice was saved
			var model database.InvoiceModel
			err = db.First(&model, "id = ?", inv.ID()).Error
			require.NoError(t, err)
			require.Equal(t, inv.ID(), model.ID)
			require.Equal(t, inv.Title(), model.Title)
		})

		t.Run("Nil_Invoice", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			err := repo.Save(ctx, nil)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid input")
		})

		t.Run("Update_Existing_Invoice", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			// Save initial invoice
			inv := createTestInvoiceWithID(t, "update-test-invoice")
			err := repo.Save(ctx, inv)
			require.NoError(t, err)

			// Update invoice
			inv.SetCustomerID("updated-customer-id")
			inv.SetStatus(invoice.StatusPending)

			err = repo.Save(ctx, inv)
			require.NoError(t, err)

			// Verify update
			var model database.InvoiceModel
			err = db.First(&model, "id = ?", inv.ID()).Error
			require.NoError(t, err)
			require.Equal(t, "updated-customer-id", *model.CustomerID)
			require.Equal(t, "pending", model.Status)
		})
	})

	t.Run("FindByID", func(t *testing.T) {
		t.Run("Existing_Invoice", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			// Save test invoice
			inv := createTestInvoice(t)
			err := repo.Save(ctx, inv)
			require.NoError(t, err)

			// Find by ID
			found, err := repo.FindByID(ctx, inv.ID())
			require.NoError(t, err)
			require.NotNil(t, found)

			require.Equal(t, inv.ID(), found.ID())
			require.Equal(t, inv.Title(), found.Title())
			require.Equal(t, inv.MerchantID(), found.MerchantID())
		})

		t.Run("Non_Existent_Invoice", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			found, err := repo.FindByID(ctx, "non-existent-id")
			require.Error(t, err)
			require.Nil(t, found)
			require.Contains(t, err.Error(), "not found")
		})

		t.Run("Empty_ID", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			found, err := repo.FindByID(ctx, "")
			require.Error(t, err)
			require.Nil(t, found)
			require.Contains(t, err.Error(), "invalid input")
		})
	})

	t.Run("FindByMerchantID", func(t *testing.T) {
		t.Run("Existing_Invoices", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			merchantID := "test-merchant-id"

			// Create and save multiple invoices for the same merchant
			for i := 0; i < 3; i++ {
				inv := createTestInvoiceWithID(t, fmt.Sprintf("invoice-%d", i))
				inv.SetStatus(invoice.StatusCreated) // Reset status
				err := repo.Save(ctx, inv)
				require.NoError(t, err)
			}

			// Find active invoices (since FindByMerchantID doesn't exist)
			invoices, err := repo.FindActive(ctx)
			require.NoError(t, err)
			require.Len(t, invoices, 3)

			for _, inv := range invoices {
				require.Equal(t, merchantID, inv.MerchantID())
			}
		})

		t.Run("No_Active_Invoices", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			invoices, err := repo.FindActive(ctx)
			require.NoError(t, err)
			require.Len(t, invoices, 0)
		})
	})

	t.Run("FindByPaymentAddress", func(t *testing.T) {
		t.Run("Existing_Invoice", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			// Save test invoice
			inv := createTestInvoice(t)
			err := repo.Save(ctx, inv)
			require.NoError(t, err)

			// Find by payment address
			paymentAddress := inv.PaymentAddress()
			require.NotNil(t, paymentAddress)

			found, err := repo.FindByPaymentAddress(ctx, paymentAddress)
			require.NoError(t, err)
			require.NotNil(t, found)

			require.Equal(t, inv.ID(), found.ID())
			require.Equal(t, paymentAddress.String(), found.PaymentAddress().String())
		})

		t.Run("Non_Existent_Address", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			paymentAddress, _ := shared.NewPaymentAddress("TNonExistentAddress123456789012345678901234567890", shared.NetworkTron)

			found, err := repo.FindByPaymentAddress(ctx, paymentAddress)
			require.Error(t, err)
			require.Nil(t, found)
			require.Contains(t, err.Error(), "not found")
		})

		t.Run("Nil_Payment_Address", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			found, err := repo.FindByPaymentAddress(ctx, nil)
			require.Error(t, err)
			require.Nil(t, found)
			require.Contains(t, err.Error(), "invalid input")
		})
	})

	t.Run("FindByStatus", func(t *testing.T) {
		t.Run("Existing_Invoices", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			// Create and save invoices with different statuses
			inv1 := createTestInvoiceWithID(t, "invoice-status-1")
			inv1.SetStatus(invoice.StatusCreated)
			err := repo.Save(ctx, inv1)
			require.NoError(t, err)

			inv2 := createTestInvoiceWithID(t, "invoice-status-2")
			inv2.SetStatus(invoice.StatusPending)
			err = repo.Save(ctx, inv2)
			require.NoError(t, err)

			// Find by status
			invoices, err := repo.FindByStatus(ctx, invoice.StatusCreated)
			require.NoError(t, err)
			require.Len(t, invoices, 1)
			require.Equal(t, invoice.StatusCreated, invoices[0].Status())

			invoices, err = repo.FindByStatus(ctx, invoice.StatusPending)
			require.NoError(t, err)
			require.Len(t, invoices, 1)
			require.Equal(t, invoice.StatusPending, invoices[0].Status())
		})

		t.Run("No_Invoices_With_Status", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			invoices, err := repo.FindByStatus(ctx, invoice.StatusPaid)
			require.NoError(t, err)
			require.Len(t, invoices, 0)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Existing_Invoice", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			// Save test invoice
			inv := createTestInvoice(t)
			err := repo.Save(ctx, inv)
			require.NoError(t, err)

			// Delete invoice
			err = repo.Delete(ctx, inv.ID())
			require.NoError(t, err)

			// Verify deletion (should be soft deleted)
			var model database.InvoiceModel
			err = db.Unscoped().First(&model, "id = ?", inv.ID()).Error
			require.NoError(t, err)
			require.NotNil(t, model.DeletedAt)

			// Should not be found in normal queries
			_, err = repo.FindByID(ctx, inv.ID())
			require.Error(t, err)
		})

		t.Run("Non_Existent_Invoice", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			err := repo.Delete(ctx, "non-existent-id")
			require.Error(t, err)
			require.Contains(t, err.Error(), "not found")
		})

		t.Run("Empty_ID", func(t *testing.T) {
			db := setupTestDB(t)
			repo := database.NewInvoiceRepository(db)
			ctx := context.Background()

			err := repo.Delete(ctx, "")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid input")
		})
	})
}
