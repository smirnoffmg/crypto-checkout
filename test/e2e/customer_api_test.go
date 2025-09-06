package e2e_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"crypto-checkout/test/testutil"
)

// TestCustomerAPIViewInvoice tests the customer API view invoice endpoint.
func TestCustomerAPIViewInvoice(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// First create an invoice via merchant API
	createReq := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"description": "VPN Premium Plan",
				"unit_price":  "9.99",
				"quantity":    "1",
			},
			{
				"description": "Additional Static IP",
				"unit_price":  "2.50",
				"quantity":    "2",
			},
		},
		"tax_rate": "0.10",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create invoice with authentication
	createHTTPReq, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	createHTTPReq.Header.Set("Content-Type", "application/json")
	createHTTPReq.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	createClient := &http.Client{}
	createResp, err := createClient.Do(createHTTPReq)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}
	defer createResp.Body.Close()

	var createResponse map[string]interface{}
	createBody, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(createBody, &createResponse)
	invoiceID := createResponse["id"].(string)

	// Now view the invoice via customer API
	resp, err := http.Get(baseURL + "/invoice/" + invoiceID)
	if err != nil {
		t.Fatalf("Failed to view invoice: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

	// Verify HTML response contains invoice data
	body, _ := io.ReadAll(resp.Body)
	htmlContent := string(body)

	// Check that HTML contains invoice information
	assert.Contains(t, htmlContent, invoiceID)
	assert.Contains(t, htmlContent, "VPN Premium Plan")
	assert.Contains(t, htmlContent, "Additional Static IP")
	assert.Contains(t, htmlContent, "9.99")
	assert.Contains(t, htmlContent, "2.50")
}

// TestCustomerAPIViewInvoiceNotFound tests 404 for non-existent invoice.
func TestCustomerAPIViewInvoiceNotFound(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	resp, err := http.Get(baseURL + "/invoice/non-existent-id")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestCustomerAPIGetQRCode tests the QR code endpoint.
func TestCustomerAPIGetQRCode(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// First create an invoice
	createReq := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"description": "Test Item",
				"unit_price":  "10.00",
				"quantity":    "1",
			},
		},
		"tax_rate": "0.10",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create invoice with authentication
	createHTTPReq, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	createHTTPReq.Header.Set("Content-Type", "application/json")
	createHTTPReq.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	createClient := &http.Client{}
	createResp, err := createClient.Do(createHTTPReq)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}
	defer createResp.Body.Close()

	var createResponse map[string]interface{}
	createBody, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(createBody, &createResponse)
	invoiceID := createResponse["id"].(string)

	// Test QR code endpoint
	resp, err := http.Get(baseURL + "/invoice/" + invoiceID + "/qr")
	if err != nil {
		t.Fatalf("Failed to get QR code: %v", err)
	}
	defer resp.Body.Close()

	// QR code should return 500 error if no payment address is assigned
	// This is expected behavior as per the current implementation
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// Verify error response
	var errorResponse map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &errorResponse); unmarshalErr == nil {
		assert.Contains(t, errorResponse, "error")
	}
}

// TestCustomerAPIGetQRCodeNotFound tests QR code for non-existent invoice.
func TestCustomerAPIGetQRCodeNotFound(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	resp, err := http.Get(baseURL + "/invoice/non-existent-id/qr")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestCustomerAPIWebSocketConnection tests WebSocket connection for real-time updates.
func TestCustomerAPIWebSocketConnection(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// First create an invoice
	createReq := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"description": "Test Item",
				"unit_price":  "10.00",
				"quantity":    "1",
			},
		},
		"tax_rate": "0.10",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create invoice with authentication
	createHTTPReq, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	createHTTPReq.Header.Set("Content-Type", "application/json")
	createHTTPReq.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	createClient := &http.Client{}
	createResp, err := createClient.Do(createHTTPReq)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}
	defer createResp.Body.Close()

	var createResponse map[string]interface{}
	createBody, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(createBody, &createResponse)
	invoiceID := createResponse["id"].(string)

	// Test WebSocket endpoint
	// Note: This is a basic test that the endpoint exists and returns appropriate response
	// Full WebSocket testing would require a WebSocket client library
	resp, err := http.Get(baseURL + "/invoice/" + invoiceID + "/ws")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket endpoint: %v", err)
	}
	defer resp.Body.Close()

	// WebSocket endpoint should return 400 Bad Request for HTTP GET
	// (WebSocket requires upgrade)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestCustomerAPIWebSocketNotFound tests WebSocket for non-existent invoice.
func TestCustomerAPIWebSocketNotFound(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	resp, err := http.Get(baseURL + "/invoice/non-existent-id/ws")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestCustomerAPIPaymentStatusEndpoint tests the payment status endpoint.
// Note: This endpoint is documented in API.md but not yet implemented
func TestCustomerAPIPaymentStatusEndpoint(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// First create an invoice
	createReq := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"description": "Test Item",
				"unit_price":  "10.00",
				"quantity":    "1",
			},
		},
		"tax_rate": "0.10",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create invoice with authentication
	createHTTPReq, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	createHTTPReq.Header.Set("Content-Type", "application/json")
	createHTTPReq.Header.Set("Authorization", "Bearer sk_test_abc123def456")

	createResp, err := http.DefaultClient.Do(createHTTPReq)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}
	defer createResp.Body.Close()

	var createResponse map[string]interface{}
	createBody, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(createBody, &createResponse)
	invoiceID := createResponse["id"].(string)

	// Test payment status endpoint (should return 501 as not implemented yet)
	resp, err := http.Get(baseURL + "/invoice/" + invoiceID + "/status")
	if err != nil {
		t.Fatalf("Failed to get payment status: %v", err)
	}
	defer resp.Body.Close()

	// This endpoint is not implemented yet, so it should return 501 Not Implemented
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)

	// Verify error response
	var errorResponse map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &errorResponse); unmarshalErr == nil {
		assert.Contains(t, errorResponse, "error")
		assert.Contains(t, errorResponse, "message")
	}
}
