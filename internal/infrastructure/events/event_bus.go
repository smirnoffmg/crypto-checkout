package events

import (
	"context"
	"crypto-checkout/internal/domain/shared"
	"fmt"

	"go.uber.org/zap"
)

// EventBus implements both EventStore and EventPublisher interfaces.
type EventBus struct {
	store     shared.EventStore
	publisher shared.EventPublisher
	logger    *zap.Logger
}

// NewEventBus creates a new event bus that combines event store and publisher.
func NewEventBus(store shared.EventStore, publisher shared.EventPublisher, logger *zap.Logger) *EventBus {
	return &EventBus{
		store:     store,
		publisher: publisher,
		logger:    logger,
	}
}

// AppendEvents stores events and publishes them.
func (b *EventBus) AppendEvents(ctx context.Context, aggregateID string, events []*shared.BaseDomainEvent) error {
	// First, store events in the event store
	if err := b.store.AppendEvents(ctx, aggregateID, events); err != nil {
		return fmt.Errorf("failed to store events: %w", err)
	}

	// Then, publish events to Kafka
	if err := b.publisher.PublishEvents(ctx, events); err != nil {
		// Log error but don't fail the transaction
		// Events are already stored, so we can retry publishing later
		b.logger.Error("Failed to publish events to Kafka",
			zap.String("aggregate_id", aggregateID),
			zap.Int("event_count", len(events)),
			zap.Error(err))
	}

	b.logger.Debug("Successfully processed events",
		zap.String("aggregate_id", aggregateID),
		zap.Int("event_count", len(events)))

	return nil
}

// GetEvents retrieves events from the event store.
func (b *EventBus) GetEvents(ctx context.Context, aggregateID string) ([]*shared.BaseDomainEvent, error) {
	return b.store.GetEvents(ctx, aggregateID)
}

// GetEventsFromVersion retrieves events from a specific version.
func (b *EventBus) GetEventsFromVersion(
	ctx context.Context,
	aggregateID string,
	fromVersion int,
) ([]*shared.BaseDomainEvent, error) {
	return b.store.GetEventsFromVersion(ctx, aggregateID, fromVersion)
}

// GetEventsByType retrieves events by type.
func (b *EventBus) GetEventsByType(
	ctx context.Context,
	eventType string,
	limit int,
) ([]*shared.BaseDomainEvent, error) {
	return b.store.GetEventsByType(ctx, eventType, limit)
}

// PublishEvent publishes a single event.
func (b *EventBus) PublishEvent(ctx context.Context, event *shared.BaseDomainEvent) error {
	return b.publisher.PublishEvent(ctx, event)
}

// PublishEvents publishes multiple events.
func (b *EventBus) PublishEvents(ctx context.Context, events []*shared.BaseDomainEvent) error {
	return b.publisher.PublishEvents(ctx, events)
}
