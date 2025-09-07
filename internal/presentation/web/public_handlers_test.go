package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/presentation/web"
)

func TestPublicInvoiceEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register routes
	router.POST("/api/v1/invoices", web.AuthMiddleware(handler.Logger), handler.CreateInvoice)
	router.GET("/api/v1/public/invoice/:id", handler.GetPublicInvoiceData)

	t.Run("GetPublicInvoice_Success", func(t *testing.T) {
		// Given: First create an invoice to retrieve publicly
		createReq := web.CreateInvoiceRequest{
			Title:       "Public Test Invoice",
			Description: "Test invoice for public viewing",
			Items: []web.InvoiceItemRequest{
				{
					Name:        "Test Item",
					Description: "Test item for public view",
					Quantity:    "1",
					UnitPrice:   "25.00",
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

		// Now retrieve the invoice via public endpoint
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusOK, w.Code)

		var response web.PublicInvoiceResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, invoiceID, response.ID)
		require.Equal(t, "created", response.Status)
		require.NotEmpty(t, response.Items)
		require.Equal(t, "25.00", response.Total)
		require.NotEmpty(t, response.CreatedAt)
	})

	t.Run("GetPublicInvoice_NotFound", func(t *testing.T) {
		// Given
		invoiceID := "non-existent-invoice"

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID, nil)
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

	t.Run("GetPublicInvoice_InvalidID", func(t *testing.T) {
		// Given
		invoiceID := "invalid-id-format"

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		// With real services, invalid ID should return 404 (not found)
		require.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "not_found", response.Error)
		require.Contains(t, response.Message, "invoice not found")
	})
}

func TestPublicInvoiceStatusEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register routes
	router.POST("/api/v1/invoices", web.AuthMiddleware(handler.Logger), handler.CreateInvoice)
	router.GET("/api/v1/public/invoice/:id/status", handler.GetPublicInvoiceStatus)

	t.Run("GetPublicInvoiceStatus_Success", func(t *testing.T) {
		// Given: First create an invoice to check status
		createReq := web.CreateInvoiceRequest{
			Title:       "Status Test Invoice",
			Description: "Test invoice for status checking",
			Items: []web.InvoiceItemRequest{
				{
					Name:        "Test Item",
					Description: "Test item for status check",
					Quantity:    "1",
					UnitPrice:   "30.00",
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

		// Now check the invoice status via public endpoint
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID+"/status", nil)
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

	t.Run("GetPublicInvoiceStatus_NotFound", func(t *testing.T) {
		// Given
		invoiceID := "inv_nonexistent"

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID+"/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "not_found", response.Error)
		require.Contains(t, response.Message, "invoice not found")
	})
}

func TestPublicInvoiceEventsEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register routes
	router.POST("/api/v1/invoices", web.AuthMiddleware(handler.Logger), handler.CreateInvoice)
	router.GET("/api/v1/public/invoice/:id/events", handler.GetPublicInvoiceEvents)

	t.Run("GetPublicInvoiceEvents_Success", func(t *testing.T) {
		// Given: First create an invoice to get events for
		createReq := web.CreateInvoiceRequest{
			Title:       "Events Test Invoice",
			Description: "Test invoice for events testing",
			Items: []web.InvoiceItemRequest{
				{
					Name:        "Test Item",
					Description: "Test item for events",
					Quantity:    "1",
					UnitPrice:   "35.00",
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

		// Now get the invoice events via public endpoint
		// Use a context with timeout to prevent hanging on SSE infinite loop
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID+"/events", nil)
		req.Header.Set("Accept", "text/event-stream")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		// Use a goroutine to handle the request since SSE will block
		done := make(chan bool)
		go func() {
			router.ServeHTTP(w, req)
			done <- true
		}()

		// Wait for either completion or timeout
		select {
		case <-done:
			// Request completed
		case <-time.After(200 * time.Millisecond):
			// Timeout - this is expected for SSE endpoints
		}

		// Then
		// For SSE endpoints, we expect the connection to be established
		// The exact status code may vary, but headers should be set correctly
		require.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
		require.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
		require.Equal(t, "keep-alive", w.Header().Get("Connection"))

		// Verify SSE format in the response body
		body := w.Body.String()
		if body != "" {
			require.Contains(t, body, "data:")
		}
	})

	t.Run("GetPublicInvoiceEvents_NotFound", func(t *testing.T) {
		// Given
		invoiceID := "inv_nonexistent"

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID+"/events", nil)
		req.Header.Set("Accept", "text/event-stream")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		require.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, "not_found", response.Error)
		require.Contains(t, response.Message, "invoice not found")
	})
}
