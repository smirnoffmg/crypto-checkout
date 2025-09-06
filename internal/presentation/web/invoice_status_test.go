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

func TestInvoiceStatusEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register the invoice status route
	router.GET("/invoice/:id/status", handler.GetInvoiceStatus)

	t.Run("GetInvoiceStatus_ServiceError", func(t *testing.T) {
		// Given
		invoiceID := "inv_test123"

		// When
		req := httptest.NewRequest(http.MethodGet, "/invoice/"+invoiceID+"/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		// Since we don't have real services, we expect a service error
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Error)
		assert.Contains(t, response.Message, "Failed to retrieve invoice status")
	})

	t.Run("GetInvoiceStatus_NotFound", func(t *testing.T) {
		// Given
		invoiceID := "inv_nonexistent"

		// When
		req := httptest.NewRequest(http.MethodGet, "/invoice/"+invoiceID+"/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// The actual error type depends on how the error is handled in the handler
		// For now, we'll check that it's an error response
		assert.NotEmpty(t, response.Error)
		assert.Contains(t, response.Message, "invoice not found")
	})

	t.Run("GetInvoiceStatus_InvalidID", func(t *testing.T) {
		// Given
		invoiceID := ""

		// When
		req := httptest.NewRequest(http.MethodGet, "/invoice/"+invoiceID+"/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", response.Error)
		assert.Contains(t, response.Message, "invoice ID")
	})
}
