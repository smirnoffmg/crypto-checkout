package shared

import (
	"context"
	"encoding/json"
	"time"
)

// BaseDomainEvent provides common fields for all domain events.
type BaseDomainEvent struct {
	EventType     string                 `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	EventVersion  int                    `json:"event_version"`
	OccurredAt    time.Time              `json:"occurred_at"`
	EventData     interface{}            `json:"event_data"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ToJSON converts the event to JSON bytes.
func (e BaseDomainEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// CreateDomainEvent creates a new domain event using the factory pattern.
func CreateDomainEvent(
	eventType, aggregateID, aggregateType string,
	eventData interface{},
	metadata map[string]interface{},
) *BaseDomainEvent {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &BaseDomainEvent{
		EventType:     eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventVersion:  1,
		OccurredAt:    time.Now().UTC(),
		EventData:     eventData,
		Metadata:      metadata,
	}
}

// FromJSON creates an event from JSON bytes.
func FromJSON(data []byte) (*BaseDomainEvent, error) {
	var event BaseDomainEvent
	err := json.Unmarshal(data, &event)
	return &event, err
}

// EventStore represents the event store interface for persisting domain events.
type EventStore interface {
	AppendEvents(ctx context.Context, aggregateID string, events []*BaseDomainEvent) error
	GetEvents(ctx context.Context, aggregateID string) ([]*BaseDomainEvent, error)
	GetEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]*BaseDomainEvent, error)
	GetEventsByType(ctx context.Context, eventType string, limit int) ([]*BaseDomainEvent, error)
}

// EventPublisher represents the interface for publishing domain events.
type EventPublisher interface {
	PublishEvent(ctx context.Context, event *BaseDomainEvent) error
	PublishEvents(ctx context.Context, events []*BaseDomainEvent) error
}

// EventBus represents the event bus interface for both storing and publishing events.
type EventBus interface {
	EventStore
	EventPublisher
}

// EventHandler represents a handler for domain events.
type EventHandler interface {
	HandleEvent(ctx context.Context, event *BaseDomainEvent) error
	EventTypes() []string
}

// EventHandlerRegistry manages event handlers.
type EventHandlerRegistry interface {
	RegisterHandler(handler EventHandler)
	GetHandlers(eventType string) []EventHandler
	GetAllHandlers() map[string][]EventHandler
}
