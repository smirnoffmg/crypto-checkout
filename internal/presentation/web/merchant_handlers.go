package web

import (
	"crypto-checkout/internal/domain/merchant"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MerchantHandlers handles merchant-related HTTP requests.
type MerchantHandlers struct {
	merchantService merchant.MerchantService
	logger          *zap.Logger
}

// NewMerchantHandlers creates a new merchant handlers instance.
func NewMerchantHandlers(merchantService merchant.MerchantService, logger *zap.Logger) *MerchantHandlers {
	return &MerchantHandlers{
		merchantService: merchantService,
		logger:          logger,
	}
}

// checkService checks if the service is initialized and returns an error response if not.
func (h *MerchantHandlers) checkService(c *gin.Context) bool {
	if h.merchantService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service not initialized"})
		return false
	}
	return true
}

// CreateMerchant handles POST /merchants
func (h *MerchantHandlers) CreateMerchant(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	var req merchant.CreateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create merchant request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.merchantService.CreateMerchant(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create merchant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create merchant"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetMerchant handles GET /merchants/:id
func (h *MerchantHandlers) GetMerchant(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	merchantID := c.Param("id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant ID is required"})
		return
	}

	req := &merchant.GetMerchantRequest{
		MerchantID: merchantID,
	}

	ctx := c.Request.Context()
	resp, err := h.merchantService.GetMerchant(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get merchant", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Merchant not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateMerchant handles PUT /merchants/:id
func (h *MerchantHandlers) UpdateMerchant(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	merchantID := c.Param("id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant ID is required"})
		return
	}

	var req merchant.UpdateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind update merchant request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.MerchantID = merchantID

	ctx := c.Request.Context()
	resp, err := h.merchantService.UpdateMerchant(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to update merchant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update merchant"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ChangeMerchantStatus handles PATCH /merchants/:id/status
func (h *MerchantHandlers) ChangeMerchantStatus(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	merchantID := c.Param("id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant ID is required"})
		return
	}

	var req merchant.ChangeMerchantStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind change merchant status request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.MerchantID = merchantID

	ctx := c.Request.Context()
	resp, err := h.merchantService.ChangeMerchantStatus(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to change merchant status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change merchant status"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListMerchants handles GET /merchants
func (h *MerchantHandlers) ListMerchants(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	req := &merchant.ListMerchantsRequest{
		Limit:  20, // Default limit
		Offset: 0,  // Default offset
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
		status := merchant.MerchantStatus(statusStr)
		if status.IsValid() {
			req.Status = &status
		}
	}

	ctx := c.Request.Context()
	resp, err := h.merchantService.ListMerchants(ctx, req)
	if err != nil {
		h.logger.Error("Failed to list merchants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list merchants"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RegisterMerchantRoutes registers merchant-related routes.
func (h *MerchantHandlers) RegisterMerchantRoutes(r *gin.RouterGroup) {
	merchants := r.Group("/merchants")
	merchants.POST("", h.CreateMerchant)
	merchants.GET("", h.ListMerchants)
	merchants.GET("/:id", h.GetMerchant)
	merchants.PUT("/:id", h.UpdateMerchant)
	merchants.PATCH("/:id/status", h.ChangeMerchantStatus)
}
