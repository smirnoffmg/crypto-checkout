package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/presentation/web"
)

func TestListInvoicesEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register the list invoices route with auth middleware
	router.GET("/api/v1/invoices", web.AuthMiddleware(handler.Logger), handler.ListInvoices)

	t.Run("ListInvoices_Success", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices", nil)
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)

		var response web.ListInvoicesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response.Invoices)
		assert.GreaterOrEqual(t, response.Total, 0)
		assert.GreaterOrEqual(t, response.Page, 1)
		assert.GreaterOrEqual(t, response.Limit, 1)
	})

	t.Run("ListInvoices_WithPagination", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices?page=2&limit=10", nil)
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)

		var response web.ListInvoicesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.Page)
		assert.Equal(t, 10, response.Limit)
	})

	t.Run("ListInvoices_WithStatusFilter", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices?status=pending", nil)
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)

		var response web.ListInvoicesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// All returned invoices should have the filtered status
		for _, inv := range response.Invoices {
			assert.Equal(t, "pending", inv.Status)
		}
	})

	t.Run("ListInvoices_Unauthorized", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices", nil)
		// No Authorization header

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "authentication_error", response.Error)
		assert.Contains(t, response.Message, "Authorization header")
	})
}
