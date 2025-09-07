package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/presentation/web"
)

func TestAnalyticsEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register the analytics route with auth middleware
	router.GET("/api/v1/analytics", web.AuthMiddleware(handler.Logger), handler.GetAnalytics)

	t.Run("GetAnalytics_Success", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics", nil)
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusOK, w.Code)

		var response web.AnalyticsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.NotNil(t, response.Summary)
		require.NotNil(t, response.Revenue)
		require.NotNil(t, response.Invoices)
		require.NotNil(t, response.Payments)
	})

	t.Run("GetAnalytics_WithDateRange", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics?start_date=2024-01-01&end_date=2024-01-31", nil)
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusOK, w.Code)

		var response web.AnalyticsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.NotNil(t, response.Summary)
		require.NotNil(t, response.Revenue)
	})

	t.Run("GetAnalytics_Unauthorized", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics", nil)
		// No Authorization header

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusUnauthorized, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "authentication_error", response.Error)
		require.Contains(t, response.Message, "Authorization header")
	})
}
