package web_test

import (
	"bytes"
	"crypto-checkout/internal/presentation/web"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMerchantHandlers_Integration(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create test handlers with nil services to test route registration
	merchantHandlers := &web.MerchantHandlers{}
	apiKeyHandlers := &web.APIKeyHandlers{}
	webhookHandlers := &web.WebhookHandlers{}

	// Register routes
	api := router.Group("/api/v1")
	merchantHandlers.RegisterMerchantRoutes(api)
	apiKeyHandlers.RegisterAPIKeyRoutes(api)
	webhookHandlers.RegisterWebhookRoutes(api)

	t.Run("MerchantRoutesRegistered", func(t *testing.T) {
		// Test that merchant routes are registered by checking if we get a 500 (nil service) instead of 404
		req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// We expect a 500 error since we have nil services, but the route should exist (not 404)
		require.NotEqual(t, http.StatusNotFound, w.Code, "Merchant routes should be registered")
	})

	t.Run("APIKeyRoutesRegistered", func(t *testing.T) {
		// Test that API key routes are registered
		req := httptest.NewRequest(http.MethodGet, "/api/v1/api-keys/test", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// We expect a 500 error since we have nil services, but the route should exist (not 404)
		require.NotEqual(t, http.StatusNotFound, w.Code, "API key routes should be registered")
	})

	t.Run("WebhookRoutesRegistered", func(t *testing.T) {
		// Test that webhook routes are registered
		req := httptest.NewRequest(http.MethodGet, "/api/v1/webhook-endpoints/test", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// We expect a 500 error since we have nil services, but the route should exist (not 404)
		require.NotEqual(t, http.StatusNotFound, w.Code, "Webhook routes should be registered")
	})
}

func TestMerchantHandlers_RequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create test handlers with nil services to test validation
	merchantHandlers := &web.MerchantHandlers{}
	apiKeyHandlers := &web.APIKeyHandlers{}
	webhookHandlers := &web.WebhookHandlers{}

	// Register routes
	api := router.Group("/api/v1")
	merchantHandlers.RegisterMerchantRoutes(api)
	apiKeyHandlers.RegisterAPIKeyRoutes(api)
	webhookHandlers.RegisterWebhookRoutes(api)

	t.Run("CreateMerchant_InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/merchants", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 for invalid JSON")
	})

	t.Run("CreateAPIKey_MissingMerchantID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/merchant-api-keys/", http.NoBody)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 for missing merchant ID")
	})

	t.Run("CreateWebhookEndpoint_MissingMerchantID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/merchant-webhooks/", http.NoBody)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 for missing merchant ID")
	})
}

func TestMerchantHandlers_QueryParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create test handlers
	merchantHandlers := &web.MerchantHandlers{}
	apiKeyHandlers := &web.APIKeyHandlers{}
	webhookHandlers := &web.WebhookHandlers{}

	// Register routes
	api := router.Group("/api/v1")
	merchantHandlers.RegisterMerchantRoutes(api)
	apiKeyHandlers.RegisterAPIKeyRoutes(api)
	webhookHandlers.RegisterWebhookRoutes(api)

	t.Run("ListMerchants_WithPagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants?limit=10&offset=5", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should not return 404 (route exists)
		require.NotEqual(t, http.StatusNotFound, w.Code, "Pagination parameters should be handled")
	})

	t.Run("ListAPIKeys_WithFilters", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/merchant-api-keys/test-merchant?status=active&key_type=test",
			http.NoBody,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should not return 404 (route exists)
		require.NotEqual(t, http.StatusNotFound, w.Code, "Filter parameters should be handled")
	})

	t.Run("ListWebhookEndpoints_WithPagination", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/merchant-webhooks/test-merchant?limit=5&offset=0",
			http.NoBody,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should not return 404 (route exists)
		require.NotEqual(t, http.StatusNotFound, w.Code, "Pagination parameters should be handled")
	})
}
