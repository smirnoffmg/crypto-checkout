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

func TestInvoiceIntegration(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler with real services
	handler := web.CreateTestHandler()

	// Register routes
	router.POST("/api/v1/invoices", web.AuthMiddleware(handler.Logger), handler.CreateInvoice)
	router.POST("/api/v1/invoices/:id/cancel", web.AuthMiddleware(handler.Logger), handler.CancelInvoice)
	router.GET("/api/v1/invoices/:id", web.AuthMiddleware(handler.Logger), handler.GetInvoice)

	t.Run("CreateAndCancelInvoice", func(t *testing.T) {
		// Step 1: Create an invoice
		createReq := web.CreateInvoiceRequest{
			Items: []web.InvoiceItemRequest{
				{
					Description: "Test item",
					Quantity:    "1",
					UnitPrice:   "10.00",
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

		if createW.Code != http.StatusCreated {
			t.Logf("Create invoice response code: %d, body: %s", createW.Code, createW.Body.String())
		}
		require.Equal(t, http.StatusCreated, createW.Code)

		var createResponse web.CreateInvoiceResponse
		err = json.Unmarshal(createW.Body.Bytes(), &createResponse)
		require.NoError(t, err)

		invoiceID := createResponse.ID
		require.NotEmpty(t, invoiceID)
		assert.Equal(t, "created", createResponse.Status)

		// Step 2: Get the invoice to verify it was created
		getReq := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/"+invoiceID, nil)
		getReq.Header.Set("Authorization", "Bearer sk_live_test123")

		getW := httptest.NewRecorder()
		router.ServeHTTP(getW, getReq)

		require.Equal(t, http.StatusOK, getW.Code)

		var getResponse web.CreateInvoiceResponse
		err = json.Unmarshal(getW.Body.Bytes(), &getResponse)
		require.NoError(t, err)

		assert.Equal(t, invoiceID, getResponse.ID)
		assert.Equal(t, "created", getResponse.Status)

		// Step 3: Cancel the invoice
		cancelReq := web.CancelInvoiceRequest{
			Reason: "Customer requested cancellation",
		}

		cancelBody, err := json.Marshal(cancelReq)
		require.NoError(t, err)

		cancelHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/"+invoiceID+"/cancel", bytes.NewBuffer(cancelBody))
		cancelHTTPReq.Header.Set("Content-Type", "application/json")
		cancelHTTPReq.Header.Set("Authorization", "Bearer sk_live_test123")

		cancelW := httptest.NewRecorder()
		router.ServeHTTP(cancelW, cancelHTTPReq)

		require.Equal(t, http.StatusOK, cancelW.Code)

		var cancelResponse web.CancelInvoiceResponse
		err = json.Unmarshal(cancelW.Body.Bytes(), &cancelResponse)
		require.NoError(t, err)

		assert.Equal(t, invoiceID, cancelResponse.ID)
		assert.Equal(t, "cancelled", cancelResponse.Status)
		assert.Equal(t, cancelReq.Reason, cancelResponse.Reason)
		assert.NotEmpty(t, cancelResponse.CancelledAt)

		// Step 4: Verify the invoice is cancelled by getting it again
		getReq2 := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/"+invoiceID, nil)
		getReq2.Header.Set("Authorization", "Bearer sk_live_test123")

		getW2 := httptest.NewRecorder()
		router.ServeHTTP(getW2, getReq2)

		require.Equal(t, http.StatusOK, getW2.Code)

		var getResponse2 web.CreateInvoiceResponse
		err = json.Unmarshal(getW2.Body.Bytes(), &getResponse2)
		require.NoError(t, err)

		assert.Equal(t, invoiceID, getResponse2.ID)
		assert.Equal(t, "cancelled", getResponse2.Status)
	})

	t.Run("CancelNonExistentInvoice", func(t *testing.T) {
		// Given
		invoiceID := "non-existent-invoice"
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
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "not_found", response.Error)
		assert.Contains(t, response.Message, "invoice not found")
	})
}
