package web

import (
	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/infrastructure/database"
	"crypto-checkout/pkg/config"

	"go.uber.org/zap"
)

// CreateTestHandler creates a test handler with real services for integration testing
func CreateTestHandler() *Handler {
	// Create real services with in-memory SQLite database
	logger := zap.NewNop()

	// Create in-memory database connection
	db, err := database.NewConnection(config.DatabaseConfig{
		URL: "sqlite://:memory:",
	})
	if err != nil {
		panic("Failed to create test database: " + err.Error())
	}

	// Run migrations
	if err := db.Migrate(); err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}

	// Create real repositories
	invoiceRepo := database.NewInvoiceRepository(db.DB)
	paymentRepo := database.NewPaymentRepository(db.DB)

	// Create real domain services
	invoiceService := invoice.NewInvoiceService(invoiceRepo)
	paymentService := payment.NewPaymentService(paymentRepo)

	// Create real handler with real services
	return NewHandler(invoiceService, paymentService, logger, &config.Config{}, nil)
}
