package web

import (
	"crypto-checkout/internal/domain/merchant"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// APIKeyHandlers handles API key-related HTTP requests.
type APIKeyHandlers struct {
	apiKeyService merchant.APIKeyService
	logger        *zap.Logger
}

// NewAPIKeyHandlers creates a new API key handlers instance.
func NewAPIKeyHandlers(apiKeyService merchant.APIKeyService, logger *zap.Logger) *APIKeyHandlers {
	return &APIKeyHandlers{
		apiKeyService: apiKeyService,
		logger:        logger,
	}
}

// checkService checks if the service is initialized and returns an error response if not.
func (h *APIKeyHandlers) checkService(c *gin.Context) bool {
	if h.apiKeyService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service not initialized"})
		return false
	}
	return true
}

// CreateAPIKey handles POST /merchants/:merchant_id/api-keys
func (h *APIKeyHandlers) CreateAPIKey(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	merchantID := c.Param("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant ID is required"})
		return
	}

	var req merchant.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create API key request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.MerchantID = merchantID

	ctx := c.Request.Context()
	resp, err := h.apiKeyService.CreateAPIKey(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetAPIKey handles GET /api-keys/:id
func (h *APIKeyHandlers) GetAPIKey(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	apiKeyID := c.Param("id")
	if apiKeyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	req := &merchant.GetAPIKeyRequest{
		APIKeyID: apiKeyID,
	}

	ctx := c.Request.Context()
	resp, err := h.apiKeyService.GetAPIKey(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get API key", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListAPIKeys handles GET /merchants/:merchant_id/api-keys
func (h *APIKeyHandlers) ListAPIKeys(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	merchantID := c.Param("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant ID is required"})
		return
	}

	req := &merchant.ListAPIKeysRequest{
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
		status := merchant.KeyStatus(statusStr)
		if status.IsValid() {
			req.Status = &status
		}
	}

	if keyTypeStr := c.Query("key_type"); keyTypeStr != "" {
		keyType := merchant.KeyType(keyTypeStr)
		if keyType.IsValid() {
			req.KeyType = &keyType
		}
	}

	ctx := c.Request.Context()
	resp, err := h.apiKeyService.ListAPIKeys(ctx, req)
	if err != nil {
		h.logger.Error("Failed to list API keys", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list API keys"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateAPIKey handles PUT /api-keys/:id
func (h *APIKeyHandlers) UpdateAPIKey(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	apiKeyID := c.Param("id")
	if apiKeyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	var req merchant.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind update API key request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.APIKeyID = apiKeyID

	ctx := c.Request.Context()
	resp, err := h.apiKeyService.UpdateAPIKey(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to update API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update API key"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RevokeAPIKey handles DELETE /api-keys/:id
func (h *APIKeyHandlers) RevokeAPIKey(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	apiKeyID := c.Param("id")
	if apiKeyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	var req merchant.RevokeAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no JSON body, create empty request
		req = merchant.RevokeAPIKeyRequest{}
	}

	req.APIKeyID = apiKeyID

	ctx := c.Request.Context()
	resp, err := h.apiKeyService.RevokeAPIKey(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to revoke API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke API key"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ValidateAPIKey handles POST /api-keys/validate
func (h *APIKeyHandlers) ValidateAPIKey(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	var req merchant.ValidateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind validate API key request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.apiKeyService.ValidateAPIKey(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to validate API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate API key"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RegisterAPIKeyRoutes registers API key-related routes.
func (h *APIKeyHandlers) RegisterAPIKeyRoutes(r *gin.RouterGroup) {
	// API key routes
	apiKeys := r.Group("/api-keys")
	apiKeys.POST("/validate", h.ValidateAPIKey)
	apiKeys.GET("/:id", h.GetAPIKey)
	apiKeys.PUT("/:id", h.UpdateAPIKey)
	apiKeys.DELETE("/:id", h.RevokeAPIKey)

	// Merchant-specific API key routes - use different path to avoid conflicts
	merchantAPIKeys := r.Group("/merchant-api-keys")
	merchantAPIKeys.POST("/:merchant_id", h.CreateAPIKey)
	merchantAPIKeys.GET("/:merchant_id", h.ListAPIKeys)
}
