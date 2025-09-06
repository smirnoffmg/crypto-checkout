package web

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/pkg/config"
)

const (
	// StackBufferSize is the size of the buffer for stack traces.
	StackBufferSize = 4096
	// MaxErrorChainLength is the maximum length of error chain to prevent infinite loops.
	MaxErrorChainLength = 10
	// Kilo.
	Kilo = 1024
)

// Handler provides HTTP handlers for the crypto-checkout API.
type Handler struct {
	invoiceService invoice.InvoiceService
	paymentService payment.PaymentService
	logger         *zap.Logger
	config         *config.Config
	hub            *Hub
}

// NewHandler creates a new API handler with the required services.
func NewHandler(
	invoiceService invoice.InvoiceService,
	paymentService payment.PaymentService,
	logger *zap.Logger,
	cfg *config.Config,
	hub *Hub,
) *Handler {
	return &Handler{
		invoiceService: invoiceService,
		paymentService: paymentService,
		logger:         logger,
		config:         cfg,
		hub:            hub,
	}
}

// RegisterRoutes registers all API routes with the Gin router.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/health", h.healthCheck)

	// Public customer-facing routes (matching API.md spec)
	router.GET("/invoice/:id", h.getPublicInvoice)
	router.GET("/invoice/:id/qr", h.getInvoiceQR)
	router.GET("/invoice/:id/status", h.getInvoiceStatus)
	router.GET("/invoice/:id/ws", h.serveWS)

	// API v1 routes (Merchant/Admin API) - require authentication
	v1 := router.Group("/api/v1")
	v1.Use(AuthMiddleware(h.logger)) // Apply auth middleware to all v1 routes
	{
		// Invoice routes
		invoices := v1.Group("/invoices")
		{
			invoices.POST("", h.createInvoice)
			invoices.GET("", h.listInvoices)
			invoices.GET("/:id", h.getInvoice)
			invoices.POST("/:id/cancel", h.cancelInvoice)
		}

		// Analytics routes
		analytics := v1.Group("/analytics")
		{
			analytics.GET("", h.getAnalytics)
		}
	}
}

// healthCheck returns the health status of the API.
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "crypto-checkout",
	})
}

// createErrorResponse creates a detailed error response with full debug information.
func (h *Handler) createErrorResponse(errorType, message string, err error) ErrorResponse {
	response := ErrorResponse{
		Error:     errorType,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Add comprehensive debug details if in debug mode
	if h.config.Log.Level == "debug" {
		details := make(map[string]interface{})

		if err != nil {
			details["error_type"] = fmt.Sprintf("%T", err)
			details["error_message"] = err.Error()
			details["error_string"] = err.Error()

			// Add stack trace
			buf := make([]byte, StackBufferSize)
			n := runtime.Stack(buf, false)
			details["stack_trace"] = string(buf[:n])

			// Add error unwrapping chain
			var errorChain []string
			currentErr := err
			for currentErr != nil {
				errorChain = append(errorChain, fmt.Sprintf("%T: %s", currentErr, currentErr.Error()))
				currentErr = fmt.Errorf("%w", currentErr)
				if len(errorChain) > MaxErrorChainLength { // Prevent infinite loops
					break
				}
			}
			details["error_chain"] = errorChain
		}

		// Add system information
		details["go_version"] = runtime.Version()
		details["go_os"] = runtime.GOOS
		details["go_arch"] = runtime.GOARCH
		details["num_goroutines"] = runtime.NumGoroutine()

		// Add memory stats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		details["memory_stats"] = map[string]interface{}{
			"alloc_mb":       m.Alloc / Kilo / Kilo,
			"total_alloc_mb": m.TotalAlloc / Kilo / Kilo,
			"sys_mb":         m.Sys / Kilo / Kilo,
			"num_gc":         m.NumGC,
		}

		response.Details = details
		response.Code = fmt.Sprintf("DEBUG_%d", time.Now().UnixNano())
	}

	return response
}

// getInvoiceStatus returns the payment status for a customer-facing invoice.
func (h *Handler) getInvoiceStatus(c *gin.Context) {
	// TODO: Implement invoice status check for customers
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Not implemented",
		"message": "Invoice status endpoint is not yet implemented",
	})
}

// listInvoices returns a paginated list of invoices for merchants/admins.
func (h *Handler) listInvoices(c *gin.Context) {
	// TODO: Implement invoice listing with filtering and pagination
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Not implemented",
		"message": "List invoices endpoint is not yet implemented",
	})
}

// cancelInvoice cancels an invoice.
func (h *Handler) cancelInvoice(c *gin.Context) {
	// TODO: Implement invoice cancellation
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Not implemented",
		"message": "Cancel invoice endpoint is not yet implemented",
	})
}

// getAnalytics returns analytics data for merchants/admins.
func (h *Handler) getAnalytics(c *gin.Context) {
	// TODO: Implement analytics endpoint
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Not implemented",
		"message": "Analytics endpoint is not yet implemented",
	})
}
