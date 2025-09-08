package events

import (
	"context"
	"crypto-checkout/internal/domain/shared"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EventStoreModel represents the database model for storing events.
type EventStoreModel struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AggregateID    string    `gorm:"not null;index:idx_events_aggregate_lookup"`
	AggregateType  string    `gorm:"not null;index:idx_events_type_timeline"`
	EventType      string    `gorm:"not null;index:idx_events_type_timeline"`
	EventVersion   int       `gorm:"not null;index:idx_events_aggregate_lookup"`
	EventData      string    `gorm:"type:jsonb;not null"`
	Metadata       string    `gorm:"type:jsonb"`
	CreatedAt      time.Time `gorm:"not null;index:idx_events_created_at"`
	SequenceNumber int64     `gorm:"autoIncrement;index:idx_events_sequence"`
}

// TableName returns the table name for the EventStoreModel.
func (EventStoreModel) TableName() string {
	return "events"
}

// PostgreSQLEventStore implements EventStore using PostgreSQL.
type PostgreSQLEventStore struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPostgreSQLEventStore creates a new PostgreSQL event store.
func NewPostgreSQLEventStore(db *gorm.DB, logger *zap.Logger) *PostgreSQLEventStore {
	return &PostgreSQLEventStore{
		db:     db,
		logger: logger,
	}
}

// AppendEvents appends events to the event store.
func (s *PostgreSQLEventStore) AppendEvents(
	ctx context.Context,
	aggregateID string,
	events []*shared.BaseDomainEvent,
) error {
	if len(events) == 0 {
		return nil
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get current version for this aggregate
	var currentVersion int
	err := tx.Model(&EventStoreModel{}).
		Where("aggregate_id = ?", aggregateID).
		Select("COALESCE(MAX(event_version), 0)").
		Scan(&currentVersion).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Validate event versions are sequential
	expectedVersion := currentVersion + 1
	for i, event := range events {
		if event.EventVersion != expectedVersion+i {
			tx.Rollback()
			return fmt.Errorf("invalid event version: expected %d, got %d", expectedVersion+i, event.EventVersion)
		}
	}

	// Convert events to database models
	eventModels := make([]EventStoreModel, len(events))
	for i, event := range events {
		eventData, err := json.Marshal(event.EventData)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		metadata, err := json.Marshal(event.Metadata)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to marshal event metadata: %w", err)
		}

		eventModels[i] = EventStoreModel{
			AggregateID:   event.AggregateID,
			AggregateType: event.AggregateType,
			EventType:     event.EventType,
			EventVersion:  event.EventVersion,
			EventData:     string(eventData),
			Metadata:      string(metadata),
			CreatedAt:     event.OccurredAt,
		}
	}

	// Insert events
	if err := tx.Create(&eventModels).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert events: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Successfully appended events to store",
		zap.String("aggregate_id", aggregateID),
		zap.Int("event_count", len(events)))

	return nil
}

// GetEvents retrieves all events for an aggregate.
func (s *PostgreSQLEventStore) GetEvents(ctx context.Context, aggregateID string) ([]*shared.BaseDomainEvent, error) {
	return s.GetEventsFromVersion(ctx, aggregateID, 1)
}

// GetEventsFromVersion retrieves events for an aggregate from a specific version.
func (s *PostgreSQLEventStore) GetEventsFromVersion(
	ctx context.Context,
	aggregateID string,
	fromVersion int,
) ([]*shared.BaseDomainEvent, error) {
	var eventModels []EventStoreModel

	err := s.db.WithContext(ctx).
		Where("aggregate_id = ? AND event_version >= ?", aggregateID, fromVersion).
		Order("event_version ASC").
		Find(&eventModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	events := make([]*shared.BaseDomainEvent, len(eventModels))
	for i := range eventModels {
		event, err := s.modelToEvent(&eventModels[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to event: %w", err)
		}
		events[i] = event
	}

	return events, nil
}

// GetEventsByType retrieves events by type with a limit.
func (s *PostgreSQLEventStore) GetEventsByType(
	ctx context.Context,
	eventType string,
	limit int,
) ([]*shared.BaseDomainEvent, error) {
	var eventModels []EventStoreModel

	query := s.db.WithContext(ctx).
		Where("event_type = ?", eventType).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&eventModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query events by type: %w", err)
	}

	events := make([]*shared.BaseDomainEvent, len(eventModels))
	for i := range eventModels {
		event, err := s.modelToEvent(&eventModels[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to event: %w", err)
		}
		events[i] = event
	}

	return events, nil
}

// modelToEvent converts a database model to a domain event.
func (s *PostgreSQLEventStore) modelToEvent(model *EventStoreModel) (*shared.BaseDomainEvent, error) {
	var eventData interface{}
	if err := json.Unmarshal([]byte(model.EventData), &eventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	var metadata map[string]interface{}
	if model.Metadata != "" {
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event metadata: %w", err)
		}
	}

	return &shared.BaseDomainEvent{
		EventType:     model.EventType,
		AggregateID:   model.AggregateID,
		AggregateType: model.AggregateType,
		EventVersion:  model.EventVersion,
		OccurredAt:    model.CreatedAt,
		EventData:     eventData,
		Metadata:      metadata,
	}, nil
}

// Migrate creates the events table if it doesn't exist.
func (s *PostgreSQLEventStore) Migrate() error {
	return s.db.AutoMigrate(&EventStoreModel{})
}
