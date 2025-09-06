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

func TestPublicInvoiceEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register the public invoice route
	router.GET("/api/v1/public/invoice/:id", handler.GetPublicInvoiceData)

	t.Run("GetPublicInvoice_ServiceError", func(t *testing.T) {
		// Given
		invoiceID := "inv_test123"

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		// Since we don't have real services, we expect a service error
		// This tests that the HTTP layer properly handles service errors
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Error)
		assert.Contains(t, response.Message, "Failed to retrieve invoice")
	})

	t.Run("GetPublicInvoice_NotFound", func(t *testing.T) {
		// Given
		invoiceID := "inv_nonexistent"

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "not_found", response.Error)
		assert.Contains(t, response.Message, "invoice not found")
	})

	t.Run("GetPublicInvoice_InvalidID", func(t *testing.T) {
		// Given
		invoiceID := ""

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID, nil)
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

func TestPublicInvoiceStatusEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register the public invoice status route
	router.GET("/api/v1/public/invoice/:id/status", handler.GetPublicInvoiceStatus)

	t.Run("GetPublicInvoiceStatus_Success", func(t *testing.T) {
		// Given
		invoiceID := "inv_test123"

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID+"/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)

		var response web.PublicInvoiceStatusResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, invoiceID, response.ID)
		assert.NotEmpty(t, response.Status)
		assert.NotEmpty(t, response.Timestamp)
	})

	t.Run("GetPublicInvoiceStatus_NotFound", func(t *testing.T) {
		// Given
		invoiceID := "inv_nonexistent"

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID+"/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "not_found", response.Error)
		assert.Contains(t, response.Message, "invoice not found")
	})
}

func TestPublicInvoiceEventsEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := web.CreateTestHandler()

	// Register the public invoice events route
	router.GET("/api/v1/public/invoice/:id/events", handler.GetPublicInvoiceEvents)

	t.Run("GetPublicInvoiceEvents_Success", func(t *testing.T) {
		// Given
		invoiceID := "inv_test123"

		// When
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+invoiceID+"/events", nil)
		req.Header.Set("Accept", "text/event-stream")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
		assert.Equal(t, "keep-alive", w.Header().Get("Connection"))

		// Verify SSE format
		body := w.Body.String()
		assert.Contains(t, body, "data:")
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
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response web.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "not_found", response.Error)
		assert.Contains(t, response.Message, "invoice not found")
	})
}
