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

func TestInvoiceStatusEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register routes
	router.POST("/api/v1/invoices", web.AuthMiddleware(handler.Logger), handler.CreateInvoice)
	router.GET("/invoice/:id/status", handler.GetInvoiceStatus)

	t.Run("GetInvoiceStatus_Success", func(t *testing.T) {
		// Given: First create an invoice to check status
		createReq := web.CreateInvoiceRequest{
			Title:       "Status Test Invoice",
			Description: "Test invoice for status checking",
			Items: []web.InvoiceItemRequest{
				{
					Name:        "Test Item",
					Description: "Test item for status check",
					Quantity:    "1",
					UnitPrice:   "20.00",
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

		// Now check the invoice status
		req := httptest.NewRequest(http.MethodGet, "/invoice/"+invoiceID+"/status", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusOK, w.Code)

		var response web.PublicInvoiceStatusResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, invoiceID, response.ID)
		require.Equal(t, "created", response.Status)
		require.NotEmpty(t, response.Timestamp)
	})

	t.Run("GetInvoiceStatus_NotFound", func(t *testing.T) {
		// Given
		invoiceID := "non-existent-invoice"

		// When
		req := httptest.NewRequest(http.MethodGet, "/invoice/"+invoiceID+"/status", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		// With real services, non-existent invoice should return 404
		require.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "not_found", response.Error)
		require.Contains(t, response.Message, "invoice not found")
	})

	t.Run("GetInvoiceStatus_InvalidID", func(t *testing.T) {
		// Given
		invoiceID := ""

		// When
		req := httptest.NewRequest(http.MethodGet, "/invoice/"+invoiceID+"/status", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusBadRequest, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "validation_error", response.Error)
		require.Contains(t, response.Message, "invoice ID")
	})
}
