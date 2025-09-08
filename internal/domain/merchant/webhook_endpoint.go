package merchant

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// WebhookEndpoint represents a webhook endpoint entity within the Merchant aggregate.
type WebhookEndpoint struct {
	id           string
	merchantID   string
	url          string
	events       []string
	secret       string
	status       EndpointStatus
	maxRetries   int
	retryBackoff BackoffStrategy
	timeout      int
	allowedIPs   []string
	headers      map[string]string
	createdAt    time.Time
	updatedAt    time.Time
}

// WebhookEndpointValidation represents the validation structure for WebhookEndpoint creation.
type WebhookEndpointValidation struct {
	ID           string          `validate:"required,min=1"  json:"id"`
	MerchantID   string          `validate:"required,min=1"  json:"merchant_id"`
	URL          string          `validate:"required,url"    json:"url"`
	Events       []string        `validate:"required,min=1"  json:"events"`
	Secret       string          `validate:"required,min=32" json:"secret"`
	Status       EndpointStatus  `validate:"required"        json:"status"`
	MaxRetries   int             `validate:"min=0,max=10"    json:"max_retries"`
	RetryBackoff BackoffStrategy `validate:"required"        json:"retry_backoff"`
	Timeout      int             `validate:"min=5,max=60"    json:"timeout"`
}

// NewWebhookEndpoint creates a new WebhookEndpoint with validation.
func NewWebhookEndpoint(
	id, merchantID, url string,
	events []string,
	secret string,
	maxRetries int,
	retryBackoff BackoffStrategy,
	timeout int,
	allowedIPs []string,
	headers map[string]string,
) (*WebhookEndpoint, error) {
	if id == "" {
		return nil, errors.New("webhook endpoint ID is required")
	}
	if merchantID == "" {
		return nil, errors.New("merchant ID is required")
	}
	if url == "" {
		return nil, errors.New("URL is required")
	}
	if len(events) == 0 {
		return nil, errors.New("at least one event type is required")
	}
	if len(secret) < 32 {
		return nil, errors.New("secret must be at least 32 characters")
	}
	if maxRetries < 0 || maxRetries > 10 {
		return nil, errors.New("max retries must be between 0 and 10")
	}
	if !retryBackoff.IsValid() {
		return nil, fmt.Errorf("invalid retry backoff strategy: %s", retryBackoff)
	}
	if timeout < 5 || timeout > 60 {
		return nil, errors.New("timeout must be between 5 and 60 seconds")
	}

	now := time.Now()
	endpoint := &WebhookEndpoint{
		id:           id,
		merchantID:   merchantID,
		url:          url,
		events:       events,
		secret:       secret,
		status:       EndpointStatusActive,
		maxRetries:   maxRetries,
		retryBackoff: retryBackoff,
		timeout:      timeout,
		allowedIPs:   allowedIPs,
		headers:      headers,
		createdAt:    now,
		updatedAt:    now,
	}

	// Validate using go-playground/validator
	validation := WebhookEndpointValidation{
		ID:           endpoint.id,
		MerchantID:   endpoint.merchantID,
		URL:          endpoint.url,
		Events:       endpoint.events,
		Secret:       endpoint.secret,
		Status:       endpoint.status,
		MaxRetries:   endpoint.maxRetries,
		RetryBackoff: endpoint.retryBackoff,
		Timeout:      endpoint.timeout,
	}

	validate := validator.New()
	if err := validate.Struct(validation); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return endpoint, nil
}

// ID returns the webhook endpoint ID.
func (w *WebhookEndpoint) ID() string {
	return w.id
}

// MerchantID returns the merchant ID.
func (w *WebhookEndpoint) MerchantID() string {
	return w.merchantID
}

// URL returns the webhook URL.
func (w *WebhookEndpoint) URL() string {
	return w.url
}

// Events returns the subscribed events.
func (w *WebhookEndpoint) Events() []string {
	return w.events
}

// Secret returns the webhook secret.
func (w *WebhookEndpoint) Secret() string {
	return w.secret
}

// Status returns the endpoint status.
func (w *WebhookEndpoint) Status() EndpointStatus {
	return w.status
}

// MaxRetries returns the maximum retry count.
func (w *WebhookEndpoint) MaxRetries() int {
	return w.maxRetries
}

// RetryBackoff returns the retry backoff strategy.
func (w *WebhookEndpoint) RetryBackoff() BackoffStrategy {
	return w.retryBackoff
}

// Timeout returns the request timeout in seconds.
func (w *WebhookEndpoint) Timeout() int {
	return w.timeout
}

// AllowedIPs returns the allowed IP addresses.
func (w *WebhookEndpoint) AllowedIPs() []string {
	return w.allowedIPs
}

// Headers returns the custom headers.
func (w *WebhookEndpoint) Headers() map[string]string {
	return w.headers
}

// CreatedAt returns the creation timestamp.
func (w *WebhookEndpoint) CreatedAt() time.Time {
	return w.createdAt
}

// UpdatedAt returns the last update timestamp.
func (w *WebhookEndpoint) UpdatedAt() time.Time {
	return w.updatedAt
}

// UpdateURL updates the webhook URL.
func (w *WebhookEndpoint) UpdateURL(url string) error {
	if url == "" {
		return errors.New("URL cannot be empty")
	}
	w.url = url
	w.updatedAt = time.Now()
	return nil
}

// UpdateEvents updates the subscribed events.
func (w *WebhookEndpoint) UpdateEvents(events []string) error {
	if len(events) == 0 {
		return errors.New("at least one event type is required")
	}
	w.events = events
	w.updatedAt = time.Now()
	return nil
}

// UpdateSecret updates the webhook secret.
func (w *WebhookEndpoint) UpdateSecret(secret string) error {
	if len(secret) < 32 {
		return errors.New("secret must be at least 32 characters")
	}
	w.secret = secret
	w.updatedAt = time.Now()
	return nil
}

// UpdateMaxRetries updates the maximum retry count.
func (w *WebhookEndpoint) UpdateMaxRetries(maxRetries int) error {
	if maxRetries < 0 || maxRetries > 10 {
		return errors.New("max retries must be between 0 and 10")
	}
	w.maxRetries = maxRetries
	w.updatedAt = time.Now()
	return nil
}

// UpdateTimeout updates the request timeout.
func (w *WebhookEndpoint) UpdateTimeout(timeout int) error {
	if timeout < 5 || timeout > 60 {
		return errors.New("timeout must be between 5 and 60 seconds")
	}
	w.timeout = timeout
	w.updatedAt = time.Now()
	return nil
}

// UpdateAllowedIPs updates the allowed IP addresses.
func (w *WebhookEndpoint) UpdateAllowedIPs(allowedIPs []string) error {
	w.allowedIPs = allowedIPs
	w.updatedAt = time.Now()
	return nil
}

// UpdateHeaders updates the custom headers.
func (w *WebhookEndpoint) UpdateHeaders(headers map[string]string) error {
	w.headers = headers
	w.updatedAt = time.Now()
	return nil
}

// ChangeStatus changes the endpoint status.
func (w *WebhookEndpoint) ChangeStatus(newStatus EndpointStatus) error {
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid status: %s", newStatus)
	}
	w.status = newStatus
	w.updatedAt = time.Now()
	return nil
}

// IsActive checks if the webhook endpoint is active.
func (w *WebhookEndpoint) IsActive() bool {
	return w.status == EndpointStatusActive
}

// IsDisabled checks if the webhook endpoint is disabled.
func (w *WebhookEndpoint) IsDisabled() bool {
	return w.status == EndpointStatusDisabled
}

// IsFailed checks if the webhook endpoint has failed.
func (w *WebhookEndpoint) IsFailed() bool {
	return w.status == EndpointStatusFailed
}

// SubscribeToEvent adds an event to the subscription list.
func (w *WebhookEndpoint) SubscribeToEvent(event string) error {
	if event == "" {
		return errors.New("event cannot be empty")
	}

	// Check if already subscribed
	for _, e := range w.events {
		if e == event {
			return nil // Already subscribed
		}
	}

	w.events = append(w.events, event)
	w.updatedAt = time.Now()
	return nil
}

// UnsubscribeFromEvent removes an event from the subscription list.
func (w *WebhookEndpoint) UnsubscribeFromEvent(event string) error {
	if event == "" {
		return errors.New("event cannot be empty")
	}

	// Find and remove the event
	for i, e := range w.events {
		if e == event {
			w.events = append(w.events[:i], w.events[i+1:]...)
			w.updatedAt = time.Now()
			return nil
		}
	}

	return errors.New("event not found in subscription list")
}

// IsSubscribedToEvent checks if the endpoint is subscribed to a specific event.
func (w *WebhookEndpoint) IsSubscribedToEvent(event string) bool {
	for _, e := range w.events {
		if e == event {
			return true
		}
	}
	return false
}
