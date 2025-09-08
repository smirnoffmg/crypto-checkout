package events

import (
	"crypto-checkout/internal/domain/shared"
	"crypto-checkout/pkg/config"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides the events infrastructure dependencies.
var Module = fx.Module("events",
	fx.Provide(
		NewKafkaConfig,
		fx.Annotate(
			NewKafkaProducer,
			fx.As(new(shared.EventPublisher)),
		),
		NewKafkaConsumer,
		fx.Annotate(
			NewPostgreSQLEventStore,
			fx.As(new(shared.EventStore)),
		),
		fx.Annotate(
			NewEventBus,
			fx.As(new(shared.EventBus)),
		),
	),
	fx.Invoke(
		MigrateEventStore,
	),
)

// KafkaConfig holds Kafka configuration.
type KafkaConfig struct {
	Brokers []string
	Topics  map[string]string
}

// NewKafkaConfig creates Kafka configuration from environment.
func NewKafkaConfig(cfg *config.Config, logger *zap.Logger) *KafkaConfig {
	brokers := strings.Split(cfg.Kafka.Brokers, ",")

	topics := map[string]string{
		"*": cfg.Kafka.TopicDomainEvents,
		// Map specific event types to topics
		shared.EventTypeInvoiceCreated:   cfg.Kafka.TopicDomainEvents,
		shared.EventTypeInvoicePaid:      cfg.Kafka.TopicDomainEvents,
		shared.EventTypeInvoiceExpired:   cfg.Kafka.TopicDomainEvents,
		shared.EventTypeInvoiceCancelled: cfg.Kafka.TopicDomainEvents,
		shared.EventTypePaymentDetected:  cfg.Kafka.TopicDomainEvents,
		shared.EventTypePaymentConfirmed: cfg.Kafka.TopicDomainEvents,
		shared.EventTypePaymentFailed:    cfg.Kafka.TopicDomainEvents,
		shared.EventTypeWebhookDelivery:  cfg.Kafka.TopicIntegrations,
		shared.EventTypeNotificationSent: cfg.Kafka.TopicNotifications,
		shared.EventTypeAnalyticsUpdated: cfg.Kafka.TopicAnalytics,
	}

	return &KafkaConfig{
		Brokers: brokers,
		Topics:  topics,
	}
}

// MigrateEventStore runs database migrations for the event store.
func MigrateEventStore(eventStore shared.EventStore) error {
	// Type assert to get the concrete type for migration
	if pgEventStore, ok := eventStore.(*PostgreSQLEventStore); ok {
		return pgEventStore.Migrate()
	}
	// For other implementations (like mocks), skip migration
	return nil
}
