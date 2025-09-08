package events

import (
	"context"
	"crypto-checkout/internal/domain/shared"
	"sync"
)

// MockEventStore is an in-memory implementation of EventStore for testing.
type MockEventStore struct {
	events map[string][]*shared.BaseDomainEvent
	mu     sync.RWMutex
}

// NewMockEventStore creates a new mock event store.
func NewMockEventStore() *MockEventStore {
	return &MockEventStore{
		events: make(map[string][]*shared.BaseDomainEvent),
	}
}

// AppendEvents appends events to the event store.
func (m *MockEventStore) AppendEvents(ctx context.Context, aggregateID string, events []*shared.BaseDomainEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events[aggregateID] = append(m.events[aggregateID], events...)
	return nil
}

// GetEvents retrieves all events for an aggregate.
func (m *MockEventStore) GetEvents(ctx context.Context, aggregateID string) ([]*shared.BaseDomainEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events, exists := m.events[aggregateID]
	if !exists {
		return []*shared.BaseDomainEvent{}, nil
	}

	// Return a copy to avoid race conditions
	result := make([]*shared.BaseDomainEvent, len(events))
	copy(result, events)
	return result, nil
}

// GetEventsFromVersion retrieves events from a specific version.
func (m *MockEventStore) GetEventsFromVersion(
	ctx context.Context,
	aggregateID string,
	fromVersion int,
) ([]*shared.BaseDomainEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events, exists := m.events[aggregateID]
	if !exists {
		return []*shared.BaseDomainEvent{}, nil
	}

	// Filter events from the specified version
	var result []*shared.BaseDomainEvent
	for _, event := range events {
		if event.EventVersion >= fromVersion {
			result = append(result, event)
		}
	}

	return result, nil
}

// GetEventsByType retrieves events by type.
func (m *MockEventStore) GetEventsByType(
	ctx context.Context,
	eventType string,
	limit int,
) ([]*shared.BaseDomainEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*shared.BaseDomainEvent
	for _, events := range m.events {
		for _, event := range events {
			if event.EventType == eventType {
				result = append(result, event)
				if limit > 0 && len(result) >= limit {
					return result, nil
				}
			}
		}
	}

	return result, nil
}

// Migrate is a no-op for the mock event store.
func (m *MockEventStore) Migrate() error {
	return nil
}
