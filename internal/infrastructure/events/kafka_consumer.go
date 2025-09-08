package events

import (
	"context"
	"crypto-checkout/internal/domain/shared"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// KafkaConsumer implements event consumption from Kafka.
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	logger   *zap.Logger
	handlers map[string][]shared.EventHandler
	mu       sync.RWMutex
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler.
type ConsumerGroupHandler struct {
	handlers map[string][]shared.EventHandler
	logger   *zap.Logger
}

// NewKafkaConsumer creates a new Kafka consumer.
func NewKafkaConsumer(brokers []string, groupID string, logger *zap.Logger) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Session.Timeout = 20 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &KafkaConsumer{
		consumer: consumer,
		logger:   logger,
		handlers: make(map[string][]shared.EventHandler),
	}, nil
}

// RegisterHandler registers an event handler.
func (c *KafkaConsumer) RegisterHandler(handler shared.EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, eventType := range handler.EventTypes() {
		c.handlers[eventType] = append(c.handlers[eventType], handler)
		c.logger.Info("Registered event handler",
			zap.String("event_type", eventType),
			zap.String("handler_type", fmt.Sprintf("%T", handler)))
	}
}

// Start starts consuming events from the specified topics.
func (c *KafkaConsumer) Start(ctx context.Context, topics []string) error {
	handler := &ConsumerGroupHandler{
		handlers: c.getHandlers(),
		logger:   c.logger,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Stopping Kafka consumer")
				return
			default:
				err := c.consumer.Consume(ctx, topics, handler)
				if err != nil {
					c.logger.Error("Error consuming from Kafka", zap.Error(err))
					time.Sleep(5 * time.Second) // Backoff on error
				}
			}
		}
	}()

	// Handle errors
	go func() {
		for err := range c.consumer.Errors() {
			c.logger.Error("Kafka consumer error", zap.Error(err))
		}
	}()

	c.logger.Info("Started Kafka consumer", zap.Strings("topics", topics))
	return nil
}

// Close closes the Kafka consumer.
func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}

// getHandlers returns a copy of the handlers map.
func (c *KafkaConsumer) getHandlers() map[string][]shared.EventHandler {
	c.mu.RLock()
	defer c.mu.RUnlock()

	handlers := make(map[string][]shared.EventHandler)
	for eventType, handlerList := range c.handlers {
		handlers[eventType] = make([]shared.EventHandler, len(handlerList))
		copy(handlers[eventType], handlerList)
	}
	return handlers
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (h *ConsumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			err := h.handleMessage(session.Context(), message)
			if err != nil {
				h.logger.Error("Failed to handle message",
					zap.String("topic", message.Topic),
					zap.Int32("partition", message.Partition),
					zap.Int64("offset", message.Offset),
					zap.Error(err))
				// Continue processing other messages
			} else {
				session.MarkMessage(message, "")
			}

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleMessage processes a single Kafka message.
func (h *ConsumerGroupHandler) handleMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	// Parse event from message
	var event shared.BaseDomainEvent
	err := json.Unmarshal(message.Value, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Get event type from headers or event data
	eventType := event.EventType
	if eventType == "" {
		// Try to get from headers
		for _, header := range message.Headers {
			if string(header.Key) == "event_type" {
				eventType = string(header.Value)
				break
			}
		}
	}

	if eventType == "" {
		return fmt.Errorf("event type not found in message")
	}

	// Find handlers for this event type
	handlers, exists := h.handlers[eventType]
	if !exists {
		// Try wildcard handler
		handlers, exists = h.handlers["*"]
		if !exists {
			h.logger.Debug("No handlers found for event type",
				zap.String("event_type", eventType))
			return nil
		}
	}

	// Process event with all registered handlers
	for _, handler := range handlers {
		err := handler.HandleEvent(ctx, &event)
		if err != nil {
			h.logger.Error("Handler failed to process event",
				zap.String("event_type", eventType),
				zap.String("aggregate_id", event.AggregateID),
				zap.String("handler_type", fmt.Sprintf("%T", handler)),
				zap.Error(err))
			// Continue with other handlers
		}
	}

	return nil
}
