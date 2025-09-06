package web_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/presentation/web"
)

func TestCancelInvoiceEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register the cancel invoice route with auth middleware
	router.POST("/api/v1/invoices/:id/cancel", web.AuthMiddleware(handler.Logger), handler.CancelInvoice)

	t.Run("CancelInvoice_ServiceError", func(t *testing.T) {
		// Given
		invoiceID := "inv_test123"
		request := web.CancelInvoiceRequest{
			Reason: "Customer requested cancellation",
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/"+invoiceID+"/cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		// Since we don't have real services, we expect a service error
		// This tests that the HTTP layer properly handles service errors
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response web.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Error)
		assert.Contains(t, response.Message, "Failed to cancel invoice")
	})

	t.Run("CancelInvoice_NotFound", func(t *testing.T) {
		// Given
		invoiceID := "inv_nonexistent"
		request := web.CancelInvoiceRequest{
			Reason: "Customer requested cancellation",
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/"+invoiceID+"/cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		// Since we don't have real services, we expect a service error
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response web.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Error)
		assert.Contains(t, response.Message, "Failed to cancel invoice")
	})

	t.Run("CancelInvoice_InvalidID", func(t *testing.T) {
		// Given
		invoiceID := ""
		request := web.CancelInvoiceRequest{
			Reason: "Customer requested cancellation",
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/"+invoiceID+"/cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response web.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", response.Error)
		assert.Contains(t, response.Message, "invoice ID")
	})

	t.Run("CancelInvoice_InvalidRequest", func(t *testing.T) {
		// Given
		invoiceID := "inv_test123"
		requestBody := []byte(`{"invalid": json}`)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/"+invoiceID+"/cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", response.Error)
		assert.Contains(t, response.Message, "Invalid JSON")
	})

	t.Run("CancelInvoice_Unauthorized", func(t *testing.T) {
		// Given
		invoiceID := "inv_test123"
		request := web.CancelInvoiceRequest{
			Reason: "Customer requested cancellation",
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/"+invoiceID+"/cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response web.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "authentication_error", response.Error)
		assert.Contains(t, response.Message, "Authorization header")
	})
}
