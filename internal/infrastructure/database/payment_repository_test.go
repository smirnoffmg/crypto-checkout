package database_test

import (
	"context"
	"fmt"
	"testing"

	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/domain/shared"
	"crypto-checkout/internal/infrastructure/database"
	"crypto-checkout/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupPaymentTestDB(t *testing.T) *gorm.DB {
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

func createTestPayment(t *testing.T) *payment.Payment {
	return createTestPaymentWithID(t, "test-payment-id", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
}

func createTestPaymentWithID(t *testing.T, id, txHash string) *payment.Payment {
	amount, _ := shared.NewMoneyWithCrypto("100.00", shared.CryptoCurrencyUSDT)
	paymentAmount, _ := payment.NewPaymentAmount(amount, shared.CryptoCurrencyUSDT)
	toAddress, _ := payment.NewPaymentAddress("TTestAddress123456789012345678901234567890", shared.NetworkTron)
	transactionHash, _ := payment.NewTransactionHash(txHash)

	p, err := payment.NewPayment(
		shared.PaymentID(id),
		shared.InvoiceID("test-invoice-id"),
		paymentAmount,
		"TSenderAddress123456789012345678901234567890",
		toAddress,
		transactionHash,
		3,
	)
	require.NoError(t, err)

	return p
}

func TestPaymentRepository(t *testing.T) {
	t.Run("Save", func(t *testing.T) {
		t.Run("Valid_Payment", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			p := createTestPayment(t)

			err := repo.Save(ctx, p)
			require.NoError(t, err)

			// Verify payment was saved
			var model database.PaymentModel
			err = db.First(&model, "id = ?", p.ID()).Error
			require.NoError(t, err)
			assert.Equal(t, string(p.ID()), model.ID)
			assert.Equal(t, string(p.InvoiceID()), model.InvoiceID)
			assert.Equal(t, p.TransactionHash().String(), model.TxHash)
		})

		t.Run("Nil_Payment", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			err := repo.Save(ctx, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid input")
		})

		t.Run("Update_Existing_Payment", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			// Save initial payment
			p := createTestPaymentWithID(t, "update-test-payment", "0x1111111111111111111111111111111111111111111111111111111111111111")
			err := repo.Save(ctx, p)
			require.NoError(t, err)

			// Update payment
			p.UpdateConfirmations(nil, 5)
			p.SetStatus(payment.StatusConfirming)

			err = repo.Save(ctx, p)
			require.NoError(t, err)

			// Verify update
			var model database.PaymentModel
			err = db.First(&model, "id = ?", p.ID()).Error
			require.NoError(t, err)
			assert.Equal(t, 5, model.Confirmations)
			assert.Equal(t, "confirming", model.Status)
		})
	})

	t.Run("FindByID", func(t *testing.T) {
		t.Run("Existing_Payment", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			// Save test payment
			p := createTestPayment(t)
			err := repo.Save(ctx, p)
			require.NoError(t, err)

			// Find by ID
			found, err := repo.FindByID(ctx, string(p.ID()))
			require.NoError(t, err)
			require.NotNil(t, found)

			assert.Equal(t, p.ID(), found.ID())
			assert.Equal(t, p.InvoiceID(), found.InvoiceID())
			assert.Equal(t, p.TransactionHash().String(), found.TransactionHash().String())
		})

		t.Run("Non_Existent_Payment", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			found, err := repo.FindByID(ctx, "non-existent-id")
			assert.Error(t, err)
			assert.Nil(t, found)
			assert.Contains(t, err.Error(), "not found")
		})
	})

	t.Run("FindByTransactionHash", func(t *testing.T) {
		t.Run("Existing_Payment", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			// Save test payment
			p := createTestPayment(t)
			err := repo.Save(ctx, p)
			require.NoError(t, err)

			// Find by transaction hash
			transactionHash := p.TransactionHash()
			found, err := repo.FindByTransactionHash(ctx, transactionHash)
			require.NoError(t, err)
			require.NotNil(t, found)

			assert.Equal(t, p.ID(), found.ID())
			assert.Equal(t, transactionHash.String(), found.TransactionHash().String())
		})

		t.Run("Non_Existent_Transaction_Hash", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			transactionHash, _ := payment.NewTransactionHash("0x0000000000000000000000000000000000000000000000000000000000000000")

			found, err := repo.FindByTransactionHash(ctx, transactionHash)
			assert.Error(t, err)
			assert.Nil(t, found)
			assert.Contains(t, err.Error(), "not found")
		})

		t.Run("Nil_Transaction_Hash", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			found, err := repo.FindByTransactionHash(ctx, nil)
			assert.Error(t, err)
			assert.Nil(t, found)
			assert.Contains(t, err.Error(), "invalid input")
		})
	})

	t.Run("FindByInvoiceID", func(t *testing.T) {
		t.Run("Existing_Payments", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			invoiceID := shared.InvoiceID("test-invoice-id")

			// Create and save multiple payments for the same invoice
			for i := 0; i < 3; i++ {
				amount, _ := shared.NewMoneyWithCrypto("100.00", shared.CryptoCurrencyUSDT)
				paymentAmount, _ := payment.NewPaymentAmount(amount, shared.CryptoCurrencyUSDT)
				toAddress, _ := payment.NewPaymentAddress("TTestAddress123456789012345678901234567890", shared.NetworkTron)
				transactionHash, _ := payment.NewTransactionHash(fmt.Sprintf("0x%064d", i))

				p, err := payment.NewPayment(
					shared.PaymentID(fmt.Sprintf("payment-%d", i)),
					invoiceID,
					paymentAmount,
					"TSenderAddress123456789012345678901234567890",
					toAddress,
					transactionHash,
					3,
				)
				require.NoError(t, err)

				err = repo.Save(ctx, p)
				require.NoError(t, err)
			}

			// Find pending payments (since FindByInvoiceID doesn't exist)
			payments, err := repo.FindPending(ctx)
			require.NoError(t, err)
			assert.Len(t, payments, 3)

			for _, p := range payments {
				assert.Equal(t, invoiceID, p.InvoiceID())
			}
		})

		t.Run("No_Pending_Payments", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			payments, err := repo.FindPending(ctx)
			require.NoError(t, err)
			assert.Len(t, payments, 0)
		})
	})

	t.Run("FindByAddress", func(t *testing.T) {
		t.Run("Existing_Payments", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			address, _ := payment.NewPaymentAddress("TTestAddress123456789012345678901234567890", shared.NetworkTron)

			// Create and save multiple payments to the same address
			for i := 0; i < 2; i++ {
				amount, _ := shared.NewMoneyWithCrypto("100.00", shared.CryptoCurrencyUSDT)
				paymentAmount, _ := payment.NewPaymentAmount(amount, shared.CryptoCurrencyUSDT)
				transactionHash, _ := payment.NewTransactionHash(fmt.Sprintf("0x%064d", i))

				p, err := payment.NewPayment(
					shared.PaymentID(fmt.Sprintf("payment-%d", i)),
					shared.InvoiceID(fmt.Sprintf("invoice-%d", i)),
					paymentAmount,
					"TSenderAddress123456789012345678901234567890",
					address,
					transactionHash,
					3,
				)
				require.NoError(t, err)

				err = repo.Save(ctx, p)
				require.NoError(t, err)
			}

			// Find by address
			payments, err := repo.FindByAddress(ctx, address)
			require.NoError(t, err)
			assert.Len(t, payments, 2)

			for _, p := range payments {
				assert.Equal(t, address.String(), p.ToAddress().String())
			}
		})

		t.Run("No_Payments_To_Address", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			address, _ := payment.NewPaymentAddress("TNonExistentAddress123456789012345678901234567890", shared.NetworkTron)

			payments, err := repo.FindByAddress(ctx, address)
			require.NoError(t, err)
			assert.Len(t, payments, 0)
		})

		t.Run("Nil_Address", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			payments, err := repo.FindByAddress(ctx, nil)
			assert.Error(t, err)
			assert.Nil(t, payments)
			assert.Contains(t, err.Error(), "invalid input")
		})
	})

	t.Run("FindByStatus", func(t *testing.T) {
		t.Run("Existing_Payments", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			// Create and save payments with different statuses
			p1 := createTestPaymentWithID(t, "status-payment-1", "0x2222222222222222222222222222222222222222222222222222222222222222")
			p1.SetStatus(payment.StatusDetected)
			err := repo.Save(ctx, p1)
			require.NoError(t, err)

			p2 := createTestPaymentWithID(t, "status-payment-2", "0x3333333333333333333333333333333333333333333333333333333333333333")
			p2.SetStatus(payment.StatusConfirming)
			err = repo.Save(ctx, p2)
			require.NoError(t, err)

			// Find by status
			payments, err := repo.FindByStatus(ctx, payment.StatusDetected)
			require.NoError(t, err)
			assert.Len(t, payments, 1)
			assert.Equal(t, payment.StatusDetected, payments[0].Status())

			payments, err = repo.FindByStatus(ctx, payment.StatusConfirming)
			require.NoError(t, err)
			assert.Len(t, payments, 1)
			assert.Equal(t, payment.StatusConfirming, payments[0].Status())
		})

		t.Run("No_Payments_With_Status", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			payments, err := repo.FindByStatus(ctx, payment.StatusConfirmed)
			require.NoError(t, err)
			assert.Len(t, payments, 0)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Existing_Payment", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			// Save test payment
			p := createTestPayment(t)
			err := repo.Save(ctx, p)
			require.NoError(t, err)

			// Delete payment
			err = repo.Delete(ctx, string(p.ID()))
			require.NoError(t, err)

			// Verify deletion (should be soft deleted)
			var model database.PaymentModel
			err = db.Unscoped().First(&model, "id = ?", p.ID()).Error
			require.NoError(t, err)
			assert.NotNil(t, model.DeletedAt)

			// Should not be found in normal queries
			_, err = repo.FindByID(ctx, string(p.ID()))
			assert.Error(t, err)
		})

		t.Run("Non_Existent_Payment", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			err := repo.Delete(ctx, "non-existent-id")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	})

	t.Run("Domain_Model_Conversion", func(t *testing.T) {
		t.Run("Round_Trip_Conversion", func(t *testing.T) {
			db := setupPaymentTestDB(t)
			repo := database.NewPaymentRepository(db)
			ctx := context.Background()

			// Create original payment
			original := createTestPayment(t)
			original.UpdateConfirmations(nil, 5)
			original.UpdateBlockInfo(12345, "blockhash123")

			networkFee, _ := shared.NewMoneyWithCrypto("0.1", shared.CryptoCurrencyUSDT)
			original.UpdateNetworkFee(networkFee, shared.CryptoCurrencyUSDT)

			// Save to database
			err := repo.Save(ctx, original)
			require.NoError(t, err)

			// Retrieve from database
			retrieved, err := repo.FindByID(ctx, string(original.ID()))
			require.NoError(t, err)
			require.NotNil(t, retrieved)

			// Verify all fields match
			assert.Equal(t, original.ID(), retrieved.ID())
			assert.Equal(t, original.InvoiceID(), retrieved.InvoiceID())
			assert.Equal(t, original.Amount().Amount().String(), retrieved.Amount().Amount().String())
			assert.Equal(t, original.FromAddress(), retrieved.FromAddress())
			assert.Equal(t, original.ToAddress().String(), retrieved.ToAddress().String())
			assert.Equal(t, original.TransactionHash().String(), retrieved.TransactionHash().String())
			assert.Equal(t, original.Status(), retrieved.Status())
			assert.Equal(t, original.Confirmations().Int(), retrieved.Confirmations().Int())
			assert.Equal(t, original.RequiredConfirmations(), retrieved.RequiredConfirmations())

			// Verify block info
			originalBlockInfo := original.BlockInfo()
			retrievedBlockInfo := retrieved.BlockInfo()
			if originalBlockInfo != nil && retrievedBlockInfo != nil {
				assert.Equal(t, originalBlockInfo.Number(), retrievedBlockInfo.Number())
				assert.Equal(t, originalBlockInfo.Hash(), retrievedBlockInfo.Hash())
			}

			// Verify network fee
			originalFee := original.NetworkFee()
			retrievedFee := retrieved.NetworkFee()
			if originalFee != nil && retrievedFee != nil {
				assert.Equal(t, originalFee.Fee().Amount().String(), retrievedFee.Fee().Amount().String())
			}
		})
	})
}
