package database_test

import (
	"testing"

	"crypto-checkout/internal/infrastructure/database"
	"crypto-checkout/pkg/config"
)

func TestConnection(t *testing.T) {
	t.Run("NewConnection", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			URL: "file::memory:",
		}

		conn, err := database.NewConnection(cfg)
		if err != nil {
			t.Fatalf("Failed to create connection: %v", err)
		}

		if conn == nil {
			t.Fatal("Connection should not be nil")
		}

		if conn.DB == nil {
			t.Fatal("Connection.DB should not be nil")
		}
	})

	t.Run("Migrate", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			URL: "file::memory:",
		}

		conn, err := database.NewConnection(cfg)
		if err != nil {
			t.Fatalf("Failed to create connection: %v", err)
		}

		// Run migrations
		err = conn.Migrate()
		if err != nil {
			t.Fatalf("Failed to run migrations: %v", err)
		}

		// Verify tables were created
		if !conn.DB.Migrator().HasTable(&database.InvoiceModel{}) {
			t.Error("InvoiceModel table should exist")
		}

		if !conn.DB.Migrator().HasTable(&database.PaymentModel{}) {
			t.Error("PaymentModel table should exist")
		}
	})

	t.Run("Migrate_Error_Handling", func(t *testing.T) {
		// Create connection with nil database to test error handling
		conn := &database.Connection{DB: nil}

		// This will panic, so we need to recover
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when migrating with nil database")
			}
		}()

		conn.Migrate()
	})
}
