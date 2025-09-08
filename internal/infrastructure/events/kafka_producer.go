package events

import (
	"context"
	"crypto-checkout/internal/domain/shared"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// KafkaProducer implements EventPublisher using Kafka.
type KafkaProducer struct {
	producer sarama.SyncProducer
	logger   *zap.Logger
	topics   map[string]string // event type -> topic mapping
}

// NewKafkaProducer creates a new Kafka producer.
func NewKafkaProducer(config *KafkaConfig, logger *zap.Logger) (*KafkaProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 3
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Compression = sarama.CompressionSnappy
	saramaConfig.Producer.Flush.Frequency = 100 * time.Millisecond

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer: producer,
		logger:   logger,
		topics:   config.Topics,
	}, nil
}

// PublishEvent publishes a single domain event to Kafka.
func (p *KafkaProducer) PublishEvent(ctx context.Context, event *shared.BaseDomainEvent) error {
	return p.PublishEvents(ctx, []*shared.BaseDomainEvent{event})
}

// PublishEvents publishes multiple domain events to Kafka.
func (p *KafkaProducer) PublishEvents(ctx context.Context, events []*shared.BaseDomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	messages := make([]*sarama.ProducerMessage, 0, len(events))

	for _, event := range events {
		topic, exists := p.topics[event.EventType]
		if !exists {
			// Default to domain events topic
			topic = p.topics["*"]
		}

		eventData, err := json.Marshal(event)
		if err != nil {
			p.logger.Error("Failed to marshal event",
				zap.String("event_type", event.EventType),
				zap.String("aggregate_id", event.AggregateID),
				zap.Error(err))
			continue
		}

		message := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(event.AggregateID),
			Value: sarama.ByteEncoder(eventData),
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("event_type"),
					Value: []byte(event.EventType),
				},
				{
					Key:   []byte("aggregate_id"),
					Value: []byte(event.AggregateID),
				},
				{
					Key:   []byte("aggregate_type"),
					Value: []byte(event.AggregateType),
				},
				{
					Key:   []byte("event_version"),
					Value: []byte(fmt.Sprintf("%d", event.EventVersion)),
				},
				{
					Key:   []byte("occurred_at"),
					Value: []byte(event.OccurredAt.Format(time.RFC3339)),
				},
			},
		}

		messages = append(messages, message)
	}

	// Send messages in batch
	err := p.producer.SendMessages(messages)
	if err != nil {
		p.logger.Error("Failed to send events to Kafka",
			zap.Int("event_count", len(events)),
			zap.Error(err))
		return fmt.Errorf("failed to send events to Kafka: %w", err)
	}

	p.logger.Debug("Successfully published events to Kafka",
		zap.Int("event_count", len(events)))

	return nil
}

// Close closes the Kafka producer.
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}
