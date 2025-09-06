package web

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	Logger         *zap.Logger
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
		Logger:         logger,
		config:         cfg,
		hub:            hub,
	}
}

// RegisterRoutes registers all API routes with the Gin router.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/health", h.healthCheck)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public customer-facing routes (matching API.md spec)
	router.GET("/invoice/:id", h.getPublicInvoice)
	router.GET("/invoice/:id/qr", h.getInvoiceQR)
	router.GET("/invoice/:id/status", h.GetInvoiceStatus)
	router.GET("/invoice/:id/ws", h.serveWS)

	// Public API routes (no authentication required)
	public := router.Group("/api/v1/public")
	{
		public.GET("/invoice/:id", h.GetPublicInvoiceData)
		public.GET("/invoice/:id/status", h.GetPublicInvoiceStatus)
		public.GET("/invoice/:id/events", h.GetPublicInvoiceEvents)
	}

	// API v1 routes (Merchant/Admin API)
	v1 := router.Group("/api/v1")
	{
		// Auth routes (no authentication required for token generation)
		auth := v1.Group("/auth")
		{
			auth.POST("/token", h.generateAuthToken)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(AuthMiddleware(h.Logger))
		{
			// Invoice routes
			invoices := protected.Group("/invoices")
			{
				invoices.POST("", h.CreateInvoice)
				invoices.GET("", h.ListInvoices)
				invoices.GET("/:id", h.GetInvoice)
				invoices.POST("/:id/cancel", h.CancelInvoice)
			}

			// Analytics routes
			analytics := protected.Group("/analytics")
			{
				analytics.GET("", h.GetAnalytics)
			}
		}
	}
}

// healthCheck returns the health status of the API.
// @Summary Health check
// @Description Check the health status of the API
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "API is healthy"
// @Router /health [get]
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
// @Summary Get invoice status for customers
// @Description Get the current status of an invoice for customer-facing display
// @Tags Customer API
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} PublicInvoiceStatusResponse "Invoice status retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid invoice ID"
// @Failure 404 {object} ErrorResponse "Invoice not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /invoice/{id}/status [get]
func (h *Handler) GetInvoiceStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Debug("Empty invoice ID in status request")
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("invoice ID is required", nil))
		return
	}

	// Get invoice status from service
	status, err := h.invoiceService.GetInvoiceStatus(c.Request.Context(), id)
	if err != nil {
		h.Logger.Error("Failed to get invoice status", zap.Error(err), zap.String("invoice_id", id))
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			c.JSON(http.StatusNotFound, createValidationErrorResponse("invoice not found", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, createValidationErrorResponse("failed to retrieve invoice status", err))
		return
	}

	response := PublicInvoiceStatusResponse{
		ID:        id,
		Status:    status.String(),
		Timestamp: time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// listInvoices returns a paginated list of invoices for merchants/admins.
// @Summary List invoices
// @Description Get a paginated list of invoices with optional filtering
// @Tags Invoices
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param status query string false "Filter by status"
// @Param merchant query string false "Filter by merchant ID"
// @Success 200 {object} ListInvoicesResponse "Invoices retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid API key"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/invoices [get]
func (h *Handler) ListInvoices(c *gin.Context) {
	var req ListInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.Logger.Error("Failed to bind list invoices request", zap.Error(err))
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("Invalid query parameters", err))
		return
	}

	// Get merchant ID from authentication context (for now, use placeholder)
	merchantID := "test-merchant" // TODO: Extract from JWT token

	// Build filter options
	var status *invoice.InvoiceStatus
	if req.Status != "" {
		s := invoice.InvoiceStatus(req.Status)
		status = &s
	}

	filter := &invoice.ListInvoicesRequest{
		MerchantID: merchantID,
		Status:     status,
		Limit:      req.Limit,
		Offset:     (req.Page - 1) * req.Limit,
	}

	// Get invoices from service
	response, err := h.invoiceService.ListInvoices(c.Request.Context(), filter)
	if err != nil {
		h.Logger.Error("Failed to list invoices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, createValidationErrorResponse("Failed to retrieve invoices", err))
		return
	}

	// Convert to response DTOs
	responseInvoices := make([]CreateInvoiceResponse, len(response.Invoices))
	for i, inv := range response.Invoices {
		responseInvoices[i] = ToCreateInvoiceResponse(inv)
	}

	// Calculate total pages
	pages := (response.Total + req.Limit - 1) / req.Limit

	listResponse := ListInvoicesResponse{
		Invoices: responseInvoices,
		Total:    response.Total,
		Page:     req.Page,
		Limit:    req.Limit,
		Pages:    pages,
	}

	c.JSON(http.StatusOK, listResponse)
}

// cancelInvoice cancels an invoice.
// @Summary Cancel an invoice
// @Description Cancel an invoice with a reason
// @Tags Invoices
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Invoice ID"
// @Param request body CancelInvoiceRequest true "Cancellation request"
// @Success 200 {object} CancelInvoiceResponse "Invoice cancelled successfully"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid API key"
// @Failure 404 {object} ErrorResponse "Invoice not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/invoices/{id}/cancel [post]
func (h *Handler) CancelInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Debug("Empty invoice ID in cancel request")
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("invoice ID is required", nil))
		return
	}

	var req CancelInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Failed to bind cancel invoice request", zap.Error(err))
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("Invalid JSON format", err))
		return
	}

	// Cancel the invoice
	err := h.invoiceService.CancelInvoice(c.Request.Context(), id, req.Reason)
	if err != nil {
		h.Logger.Error("Failed to cancel invoice", zap.Error(err), zap.String("invoice_id", id))
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			c.JSON(http.StatusNotFound, createValidationErrorResponse("invoice not found", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, createValidationErrorResponse("Failed to cancel invoice", err))
		return
	}

	// Get the updated invoice to return current status
	inv, err := h.invoiceService.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.Logger.Error("Failed to get updated invoice after cancellation", zap.Error(err), zap.String("invoice_id", id))
		c.JSON(http.StatusInternalServerError, createValidationErrorResponse("Failed to retrieve updated invoice", err))
		return
	}

	response := CancelInvoiceResponse{
		ID:          id,
		Status:      inv.Status().String(),
		Reason:      req.Reason,
		CancelledAt: time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// getAnalytics returns analytics data for merchants/admins.
// @Summary Get analytics data
// @Description Get comprehensive analytics data for merchants and admins
// @Tags Analytics
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param merchant query string false "Filter by merchant ID"
// @Success 200 {object} AnalyticsResponse "Analytics data retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid API key"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/analytics [get]
func (h *Handler) GetAnalytics(c *gin.Context) {
	var req AnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.Logger.Error("Failed to bind analytics request", zap.Error(err))
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("Invalid query parameters", err))
		return
	}

	// TODO: Extract merchant ID from authentication context
	// merchantID := "test-merchant" // TODO: Extract from JWT token

	// TODO: Implement real analytics by aggregating data from invoice and payment services
	// For now, return mock analytics data
	response := AnalyticsResponse{
		Summary: AnalyticsSummary{
			TotalInvoices:     150,
			TotalRevenue:      "12500.00",
			TotalPayments:     120,
			SuccessRate:       80.0,
			AverageAmount:     "83.33",
			PendingInvoices:   25,
			CompletedInvoices: 100,
			CancelledInvoices: 25,
		},
		Revenue: AnalyticsRevenue{
			Total:     "12500.00",
			ThisMonth: "3200.00",
			LastMonth: "2800.00",
			Growth:    "14.29",
		},
		Invoices: AnalyticsInvoices{
			ByStatus: map[string]int{
				"pending":   25,
				"paid":      100,
				"cancelled": 25,
			},
			ByMonth: map[string]int{
				"2024-01": 45,
				"2024-02": 52,
				"2024-03": 53,
			},
		},
		Payments: AnalyticsPayments{
			ByStatus: map[string]int{
				"confirmed": 100,
				"pending":   20,
				"failed":    5,
			},
			ByMonth: map[string]int{
				"2024-01": 40,
				"2024-02": 45,
				"2024-03": 35,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}
