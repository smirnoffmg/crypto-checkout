package events

import (
	"context"
	"crypto-checkout/internal/domain/shared"

	"go.uber.org/fx"
)

// TestModule provides test-specific event infrastructure dependencies.
//
//nolint:gochecknoglobals // This is a test module that needs to be globally accessible
var TestModule = fx.Module("events-test",
	fx.Provide(
		NewMockEventStore,
		fx.Annotate(
			NewMockEventBus,
			fx.As(new(shared.EventBus)),
		),
	),
)

// MockEventBus is a no-op implementation of EventBus for testing.
type MockEventBus struct {
	eventStore *MockEventStore
}

// NewMockEventBus creates a new mock event bus.
func NewMockEventBus(eventStore *MockEventStore) *MockEventBus {
	return &MockEventBus{
		eventStore: eventStore,
	}
}

// AppendEvents appends events to the event store.
func (m *MockEventBus) AppendEvents(ctx context.Context, aggregateID string, events []*shared.BaseDomainEvent) error {
	return m.eventStore.AppendEvents(ctx, aggregateID, events)
}

func (m *MockEventBus) GetEvents(ctx context.Context, aggregateID string) ([]*shared.BaseDomainEvent, error) {
	return m.eventStore.GetEvents(ctx, aggregateID)
}

func (m *MockEventBus) GetEventsFromVersion(
	ctx context.Context,
	aggregateID string,
	fromVersion int,
) ([]*shared.BaseDomainEvent, error) {
	return m.eventStore.GetEventsFromVersion(ctx, aggregateID, fromVersion)
}

func (m *MockEventBus) GetEventsByType(
	ctx context.Context,
	eventType string,
	limit int,
) ([]*shared.BaseDomainEvent, error) {
	return m.eventStore.GetEventsByType(ctx, eventType, limit)
}

// PublishEvent publishes a single event.
func (m *MockEventBus) PublishEvent(ctx context.Context, event *shared.BaseDomainEvent) error {
	// No-op for testing
	return nil
}

func (m *MockEventBus) PublishEvents(ctx context.Context, events []*shared.BaseDomainEvent) error {
	// No-op for testing
	return nil
}
