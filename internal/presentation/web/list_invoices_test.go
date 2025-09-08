package web_test

import (
	"crypto-checkout/internal/presentation/web"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
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
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices", http.NoBody)
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusOK, w.Code)

		var response web.ListInvoicesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.NotNil(t, response.Invoices)
		require.GreaterOrEqual(t, response.Total, 0)
		require.GreaterOrEqual(t, response.Page, 1)
		require.GreaterOrEqual(t, response.Limit, 1)
	})

	t.Run("ListInvoices_WithPagination", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices?page=2&limit=10", http.NoBody)
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusOK, w.Code)

		var response web.ListInvoicesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, 2, response.Page)
		require.Equal(t, 10, response.Limit)
	})

	t.Run("ListInvoices_WithStatusFilter", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices?status=pending", http.NoBody)
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusOK, w.Code)

		var response web.ListInvoicesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// All returned invoices should have the filtered status
		for _, inv := range response.Invoices {
			require.Equal(t, "pending", inv.Status)
		}
	})

	t.Run("ListInvoices_Unauthorized", func(t *testing.T) {
		// Given
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices", http.NoBody)
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
