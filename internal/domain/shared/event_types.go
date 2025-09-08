package shared

// Event type constants for domain events
const (
	// Invoice events
	EventTypeInvoiceCreated       = "invoice.created"
	EventTypeInvoiceStatusChanged = "invoice.status_changed"
	EventTypeInvoicePaid          = "invoice.paid"
	EventTypeInvoiceExpired       = "invoice.expired"
	EventTypeInvoiceCancelled     = "invoice.cancelled"

	// Payment events
	EventTypePaymentDetected      = "payment.detected"
	EventTypePaymentStatusChanged = "payment.status_changed"
	EventTypePaymentConfirmed     = "payment.confirmed"
	EventTypePaymentFailed        = "payment.failed"

	// Integration events
	EventTypeWebhookDelivery = "webhook.delivery"
	EventTypeWebhookRetry    = "webhook.retry"
	EventTypeWebhookFailed   = "webhook.failed"

	// Notification events
	EventTypeNotificationSent   = "notification.sent"
	EventTypeNotificationFailed = "notification.failed"

	// Analytics events
	EventTypeAnalyticsUpdated = "analytics.updated"
	EventTypeAnalyticsReport  = "analytics.report"

	// System events
	EventTypeSystemError   = "system.error"
	EventTypeSystemWarning = "system.warning"
	EventTypeSystemInfo    = "system.info"
)

// Event type categories for routing
const (
	EventCategoryDomain       = "domain"
	EventCategoryIntegration  = "integration"
	EventCategoryNotification = "notification"
	EventCategoryAnalytics    = "analytics"
	EventCategorySystem       = "system"
)

// GetEventCategory returns the category for a given event type
func GetEventCategory(eventType string) string {
	switch eventType {
	case EventTypeInvoiceCreated, EventTypeInvoiceStatusChanged, EventTypeInvoicePaid,
		EventTypeInvoiceExpired, EventTypeInvoiceCancelled,
		EventTypePaymentDetected, EventTypePaymentStatusChanged, EventTypePaymentConfirmed,
		EventTypePaymentFailed:
		return EventCategoryDomain
	case EventTypeWebhookDelivery, EventTypeWebhookRetry, EventTypeWebhookFailed:
		return EventCategoryIntegration
	case EventTypeNotificationSent, EventTypeNotificationFailed:
		return EventCategoryNotification
	case EventTypeAnalyticsUpdated, EventTypeAnalyticsReport:
		return EventCategoryAnalytics
	case EventTypeSystemError, EventTypeSystemWarning, EventTypeSystemInfo:
		return EventCategorySystem
	default:
		return EventCategoryDomain // Default to domain events
	}
}
