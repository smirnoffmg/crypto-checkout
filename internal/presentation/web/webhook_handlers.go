package web

import (
	"crypto-checkout/internal/domain/merchant"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WebhookHandlers handles webhook endpoint-related HTTP requests.
type WebhookHandlers struct {
	webhookService merchant.WebhookEndpointService
	logger         *zap.Logger
}

// NewWebhookHandlers creates a new webhook handlers instance.
func NewWebhookHandlers(webhookService merchant.WebhookEndpointService, logger *zap.Logger) *WebhookHandlers {
	return &WebhookHandlers{
		webhookService: webhookService,
		logger:         logger,
	}
}

// checkService checks if the service is initialized and returns an error response if not.
func (h *WebhookHandlers) checkService(c *gin.Context) bool {
	if h.webhookService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service not initialized"})
		return false
	}
	return true
}

// CreateWebhookEndpoint handles POST /merchants/:merchant_id/webhook-endpoints
func (h *WebhookHandlers) CreateWebhookEndpoint(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	merchantID := c.Param("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant ID is required"})
		return
	}

	var req merchant.CreateWebhookEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create webhook endpoint request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.MerchantID = merchantID

	ctx := c.Request.Context()
	resp, err := h.webhookService.CreateWebhookEndpoint(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create webhook endpoint", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook endpoint"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetWebhookEndpoint handles GET /webhook-endpoints/:id
func (h *WebhookHandlers) GetWebhookEndpoint(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	endpointID := c.Param("id")
	if endpointID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook endpoint ID is required"})
		return
	}

	req := &merchant.GetWebhookEndpointRequest{
		EndpointID: endpointID,
	}

	ctx := c.Request.Context()
	resp, err := h.webhookService.GetWebhookEndpoint(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get webhook endpoint", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook endpoint not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListWebhookEndpoints handles GET /merchants/:merchant_id/webhook-endpoints
func (h *WebhookHandlers) ListWebhookEndpoints(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	merchantID := c.Param("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant ID is required"})
		return
	}

	req := &merchant.ListWebhookEndpointsRequest{
		MerchantID: merchantID,
		Limit:      20, // Default limit
		Offset:     0,  // Default offset
	}

	// Parse query parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := merchant.EndpointStatus(statusStr)
		if status.IsValid() {
			req.Status = &status
		}
	}

	ctx := c.Request.Context()
	resp, err := h.webhookService.ListWebhookEndpoints(ctx, req)
	if err != nil {
		h.logger.Error("Failed to list webhook endpoints", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list webhook endpoints"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateWebhookEndpoint handles PUT /webhook-endpoints/:id
func (h *WebhookHandlers) UpdateWebhookEndpoint(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	endpointID := c.Param("id")
	if endpointID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook endpoint ID is required"})
		return
	}

	var req merchant.UpdateWebhookEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind update webhook endpoint request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.EndpointID = endpointID

	ctx := c.Request.Context()
	resp, err := h.webhookService.UpdateWebhookEndpoint(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to update webhook endpoint", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook endpoint"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteWebhookEndpoint handles DELETE /webhook-endpoints/:id
func (h *WebhookHandlers) DeleteWebhookEndpoint(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	endpointID := c.Param("id")
	if endpointID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook endpoint ID is required"})
		return
	}

	req := &merchant.DeleteWebhookEndpointRequest{
		EndpointID: endpointID,
	}

	ctx := c.Request.Context()
	resp, err := h.webhookService.DeleteWebhookEndpoint(ctx, req)
	if err != nil {
		h.logger.Error("Failed to delete webhook endpoint", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook endpoint"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// TestWebhookEndpoint handles POST /webhook-endpoints/:id/test
func (h *WebhookHandlers) TestWebhookEndpoint(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	endpointID := c.Param("id")
	if endpointID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook endpoint ID is required"})
		return
	}

	req := &merchant.TestWebhookEndpointRequest{
		EndpointID: endpointID,
	}

	ctx := c.Request.Context()
	resp, err := h.webhookService.TestWebhookEndpoint(ctx, req)
	if err != nil {
		h.logger.Error("Failed to test webhook endpoint", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test webhook endpoint"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RegisterWebhookRoutes registers webhook endpoint-related routes.
func (h *WebhookHandlers) RegisterWebhookRoutes(r *gin.RouterGroup) {
	// Webhook endpoint routes
	webhooks := r.Group("/webhook-endpoints")
	webhooks.GET("/:id", h.GetWebhookEndpoint)
	webhooks.PUT("/:id", h.UpdateWebhookEndpoint)
	webhooks.DELETE("/:id", h.DeleteWebhookEndpoint)
	webhooks.POST("/:id/test", h.TestWebhookEndpoint)

	// Merchant-specific webhook endpoint routes - use different path to avoid conflicts
	merchantWebhooks := r.Group("/merchant-webhooks")
	merchantWebhooks.POST("/:merchant_id", h.CreateWebhookEndpoint)
	merchantWebhooks.GET("/:merchant_id", h.ListWebhookEndpoints)
}
