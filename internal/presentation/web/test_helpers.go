package web

import (
	"context"
	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/domain/shared"
	"crypto-checkout/internal/infrastructure/database"
	"crypto-checkout/pkg/config"

	"go.uber.org/zap"
)

// mockEventBus is a no-op implementation of EventBus for testing
type mockEventBus struct{}

// EventStore methods
func (m *mockEventBus) AppendEvents(ctx context.Context, aggregateID string, events []*shared.BaseDomainEvent) error {
	return nil
}

func (m *mockEventBus) GetEvents(ctx context.Context, aggregateID string) ([]*shared.BaseDomainEvent, error) {
	return []*shared.BaseDomainEvent{}, nil
}

func (m *mockEventBus) GetEventsFromVersion(
	ctx context.Context,
	aggregateID string,
	fromVersion int,
) ([]*shared.BaseDomainEvent, error) {
	return []*shared.BaseDomainEvent{}, nil
}

func (m *mockEventBus) GetEventsByType(
	ctx context.Context,
	eventType string,
	limit int,
) ([]*shared.BaseDomainEvent, error) {
	return []*shared.BaseDomainEvent{}, nil
}

// EventPublisher methods
func (m *mockEventBus) PublishEvent(ctx context.Context, event *shared.BaseDomainEvent) error {
	return nil
}

func (m *mockEventBus) PublishEvents(ctx context.Context, events []*shared.BaseDomainEvent) error {
	return nil
}

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

	// Create mock event bus for testing
	mockEventBus := &mockEventBus{}

	// Create real domain services
	invoiceService := invoice.NewInvoiceService(invoiceRepo, mockEventBus, logger)
	paymentService := payment.NewPaymentService(paymentRepo, mockEventBus, logger)

	// Create real handler with real services
	return NewHandler(invoiceService, paymentService, logger, &config.Config{}, nil)
}
