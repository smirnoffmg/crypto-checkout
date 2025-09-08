package merchant

import (
	"time"
)

// DomainEvent represents a domain event interface.
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
	AggregateID() string
}

// MerchantCreatedEvent represents the event when a merchant is created.
type MerchantCreatedEvent struct {
	MerchantID   string    `json:"merchant_id"`
	BusinessName string    `json:"business_name"`
	ContactEmail string    `json:"contact_email"`
	Timestamp    time.Time `json:"occurred_at"`
}

// EventType returns the event type.
func (e MerchantCreatedEvent) EventType() string {
	return "merchant.created"
}

// OccurredAt returns when the event occurred.
func (e MerchantCreatedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the aggregate ID.
func (e MerchantCreatedEvent) AggregateID() string {
	return e.MerchantID
}

// MerchantStatusChangedEvent represents the event when a merchant status changes.
type MerchantStatusChangedEvent struct {
	MerchantID string         `json:"merchant_id"`
	FromStatus MerchantStatus `json:"from_status"`
	ToStatus   MerchantStatus `json:"to_status"`
	Reason     string         `json:"reason"`
	Timestamp  time.Time      `json:"occurred_at"`
}

// EventType returns the event type.
func (e MerchantStatusChangedEvent) EventType() string {
	return "merchant.status_changed"
}

// OccurredAt returns when the event occurred.
func (e MerchantStatusChangedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the aggregate ID.
func (e MerchantStatusChangedEvent) AggregateID() string {
	return e.MerchantID
}

// APIKeyGeneratedEvent represents the event when an API key is generated.
type APIKeyGeneratedEvent struct {
	APIKeyID    string    `json:"api_key_id"`
	MerchantID  string    `json:"merchant_id"`
	KeyType     KeyType   `json:"key_type"`
	Permissions []string  `json:"permissions"`
	Timestamp   time.Time `json:"occurred_at"`
}

// EventType returns the event type.
func (e APIKeyGeneratedEvent) EventType() string {
	return "api_key.generated"
}

// OccurredAt returns when the event occurred.
func (e APIKeyGeneratedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the aggregate ID.
func (e APIKeyGeneratedEvent) AggregateID() string {
	return e.MerchantID
}

// APIKeyRevokedEvent represents the event when an API key is revoked.
type APIKeyRevokedEvent struct {
	APIKeyID   string    `json:"api_key_id"`
	MerchantID string    `json:"merchant_id"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"occurred_at"`
}

// EventType returns the event type.
func (e APIKeyRevokedEvent) EventType() string {
	return "api_key.revoked"
}

// OccurredAt returns when the event occurred.
func (e APIKeyRevokedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the aggregate ID.
func (e APIKeyRevokedEvent) AggregateID() string {
	return e.MerchantID
}

// WebhookEndpointCreatedEvent represents the event when a webhook endpoint is created.
type WebhookEndpointCreatedEvent struct {
	EndpointID string    `json:"endpoint_id"`
	MerchantID string    `json:"merchant_id"`
	URL        string    `json:"url"`
	Events     []string  `json:"events"`
	Timestamp  time.Time `json:"occurred_at"`
}

// EventType returns the event type.
func (e WebhookEndpointCreatedEvent) EventType() string {
	return "webhook_endpoint.created"
}

// OccurredAt returns when the event occurred.
func (e WebhookEndpointCreatedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the aggregate ID.
func (e WebhookEndpointCreatedEvent) AggregateID() string {
	return e.MerchantID
}

// WebhookEndpointUpdatedEvent represents the event when a webhook endpoint is updated.
type WebhookEndpointUpdatedEvent struct {
	EndpointID string    `json:"endpoint_id"`
	MerchantID string    `json:"merchant_id"`
	Changes    []string  `json:"changes"`
	Timestamp  time.Time `json:"occurred_at"`
}

// EventType returns the event type.
func (e WebhookEndpointUpdatedEvent) EventType() string {
	return "webhook_endpoint.updated"
}

// OccurredAt returns when the event occurred.
func (e WebhookEndpointUpdatedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the aggregate ID.
func (e WebhookEndpointUpdatedEvent) AggregateID() string {
	return e.MerchantID
}

// WebhookEndpointDeletedEvent represents the event when a webhook endpoint is deleted.
type WebhookEndpointDeletedEvent struct {
	EndpointID string    `json:"endpoint_id"`
	MerchantID string    `json:"merchant_id"`
	Timestamp  time.Time `json:"occurred_at"`
}

// EventType returns the event type.
func (e WebhookEndpointDeletedEvent) EventType() string {
	return "webhook_endpoint.deleted"
}

// OccurredAt returns when the event occurred.
func (e WebhookEndpointDeletedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the aggregate ID.
func (e WebhookEndpointDeletedEvent) AggregateID() string {
	return e.MerchantID
}

// WebhookDeliveryAttemptedEvent represents the event when a webhook delivery is attempted.
type WebhookDeliveryAttemptedEvent struct {
	DeliveryID   string    `json:"delivery_id"`
	EndpointID   string    `json:"endpoint_id"`
	MerchantID   string    `json:"merchant_id"`
	EventName    string    `json:"event_type"`
	Status       string    `json:"status"`
	Attempt      int       `json:"attempt"`
	ResponseCode int       `json:"response_code"`
	Timestamp    time.Time `json:"occurred_at"`
}

// EventType returns the event type.
func (e WebhookDeliveryAttemptedEvent) EventType() string {
	return "webhook_delivery.attempted"
}

// OccurredAt returns when the event occurred.
func (e WebhookDeliveryAttemptedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the aggregate ID.
func (e WebhookDeliveryAttemptedEvent) AggregateID() string {
	return e.MerchantID
}
