package web_test

import (
	"bytes"
	"crypto-checkout/internal/presentation/web"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCancelInvoiceEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register routes with auth middleware
	router.POST("/api/v1/invoices", web.AuthMiddleware(handler.Logger), handler.CreateInvoice)
	router.POST("/api/v1/invoices/:id/cancel", web.AuthMiddleware(handler.Logger), handler.CancelInvoice)

	t.Run("CancelInvoice_Success", func(t *testing.T) {
		// Given: First create an invoice to cancel
		createReq := web.CreateInvoiceRequest{
			Title:       "Test Invoice for Cancellation",
			Description: "Test invoice to be cancelled",
			Items: []web.InvoiceItemRequest{
				{
					Name:        "Test Item",
					Description: "Test item for cancellation",
					Quantity:    "1",
					UnitPrice:   "15.00",
				},
			},
			TaxRate: "0.00",
		}

		createBody, err := json.Marshal(createReq)
		require.NoError(t, err)

		createHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/invoices", bytes.NewBuffer(createBody))
		createHTTPReq.Header.Set("Content-Type", "application/json")
		createHTTPReq.Header.Set("Authorization", "Bearer sk_live_test123")

		createW := httptest.NewRecorder()
		router.ServeHTTP(createW, createHTTPReq)

		require.Equal(t, http.StatusCreated, createW.Code)

		var createResponse web.CreateInvoiceResponse
		err = json.Unmarshal(createW.Body.Bytes(), &createResponse)
		require.NoError(t, err)

		invoiceID := createResponse.ID
		require.NotEmpty(t, invoiceID)
		require.Equal(t, "created", createResponse.Status)

		// Now cancel the invoice
		cancelReq := web.CancelInvoiceRequest{
			Reason: "Customer requested cancellation",
		}

		cancelBody, err := json.Marshal(cancelReq)
		require.NoError(t, err)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/invoices/"+invoiceID+"/cancel",
			bytes.NewBuffer(cancelBody),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusOK, w.Code)

		var response web.CancelInvoiceResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, invoiceID, response.ID)
		require.Equal(t, "cancelled", response.Status)
		require.Equal(t, cancelReq.Reason, response.Reason)
		require.NotEmpty(t, response.CancelledAt)
	})

	t.Run("CancelInvoice_NotFound", func(t *testing.T) {
		// Given
		invoiceID := "non-existent-invoice"
		request := web.CancelInvoiceRequest{
			Reason: "Customer requested cancellation",
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/invoices/"+invoiceID+"/cancel",
			bytes.NewBuffer(requestBody),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		// With real services, non-existent invoice should return 404
		require.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "not_found", response.Error)
		require.Contains(t, response.Message, "invoice not found")
	})

	t.Run("CancelInvoice_InvalidID", func(t *testing.T) {
		// Given
		invoiceID := ""
		request := web.CancelInvoiceRequest{
			Reason: "Customer requested cancellation",
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/invoices/"+invoiceID+"/cancel",
			bytes.NewBuffer(requestBody),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusBadRequest, w.Code)

		var response web.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "validation_error", response.Error)
		require.Contains(t, response.Message, "invoice ID")
	})

	t.Run("CancelInvoice_InvalidRequest", func(t *testing.T) {
		// Given
		invoiceID := "inv_test123"
		requestBody := []byte(`{"invalid": json}`)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/invoices/"+invoiceID+"/cancel",
			bytes.NewBuffer(requestBody),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk_live_test123")

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusBadRequest, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "validation_error", response.Error)
		require.Contains(t, response.Message, "Invalid JSON")
	})

	t.Run("CancelInvoice_Unauthorized", func(t *testing.T) {
		// Given
		invoiceID := "inv_test123"
		request := web.CancelInvoiceRequest{
			Reason: "Customer requested cancellation",
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/invoices/"+invoiceID+"/cancel",
			bytes.NewBuffer(requestBody),
		)
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		// When
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusUnauthorized, w.Code)

		var response web.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "authentication_error", response.Error)
		require.Contains(t, response.Message, "Authorization header")
	})
}
