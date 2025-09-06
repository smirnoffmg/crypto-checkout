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

// TestMerchantAPICreateInvoice tests the merchant API invoice creation endpoint.
func TestMerchantAPICreateInvoice(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Test valid invoice creation
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
	// Create request with authentication header
	req, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr != nil {
		t.Fatalf("Failed to parse response: %v", unmarshalErr)
	}

	// Verify response structure matches API.md specification
	assert.Contains(t, response, "id")
	assert.Contains(t, response, "items")
	assert.Contains(t, response, "subtotal")
	assert.Contains(t, response, "tax_amount")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "tax_rate")
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "created_at")

	// API.md required fields - these should be implemented
	assert.Contains(t, response, "usdt_amount", "API.md requires usdt_amount field")
	assert.Contains(t, response, "address", "API.md requires address field")
	assert.Contains(t, response, "customer_url", "API.md requires customer_url field")
	assert.Contains(t, response, "expires_at", "API.md requires expires_at field")

	// Verify invoice status is pending
	assert.Equal(t, "created", response["status"])

	// Verify items structure
	items, ok := response["items"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, items, 2)

	// Verify first item
	item1 := items[0].(map[string]interface{})
	assert.Equal(t, "VPN Premium Plan", item1["description"])
	assert.Equal(t, "9.99", item1["unit_price"])
	assert.Equal(t, "1", item1["quantity"])
	assert.Contains(t, item1, "total")
}

// TestMerchantAPICreateInvoiceValidation tests validation errors.
func TestMerchantAPICreateInvoiceValidation(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing items",
			request:        map[string]interface{}{"tax_rate": "0.10"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty items array",
			request: map[string]interface{}{
				"items":    []interface{}{},
				"tax_rate": "0.10",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing tax_rate",
			request: map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"description": "Test Item",
						"unit_price":  "10.00",
						"quantity":    "1",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "negative tax rate",
			request: map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"description": "Test Item",
						"unit_price":  "10.00",
						"quantity":    "1",
					},
				},
				"tax_rate": "-0.10",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid unit_price format",
			request: map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"description": "Test Item",
						"unit_price":  "invalid",
						"quantity":    "1",
					},
				},
				"tax_rate": "0.10",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.request)

			// Create request with authentication
			req, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer sk_test_abc123def456")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestMerchantAPIGetInvoice tests the merchant API get invoice endpoint.
func TestMerchantAPIGetInvoice(t *testing.T) {
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

	// Now get the invoice
	// Get invoice with authentication
	req, err := http.NewRequest("GET", baseURL+"/api/v1/invoices/"+invoiceID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get invoice: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr != nil {
		t.Fatalf("Failed to parse response: %v", unmarshalErr)
	}

	// Verify response structure matches API.md merchant response
	assert.Equal(t, invoiceID, response["id"])
	assert.Contains(t, response, "items")
	assert.Contains(t, response, "subtotal")
	assert.Contains(t, response, "tax_amount")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "tax_rate")
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "created_at")

	// API.md required fields - these should be implemented
	assert.Contains(t, response, "usdt_amount", "API.md requires usdt_amount field")
	assert.Contains(t, response, "address", "API.md requires address field")
	assert.Contains(t, response, "customer_url", "API.md requires customer_url field")
	assert.Contains(t, response, "expires_at", "API.md requires expires_at field")

	// Verify status is pending
	assert.Equal(t, "created", response["status"])
}

// TestMerchantAPIGetInvoiceNotFound tests 404 for non-existent invoice.
func TestMerchantAPIGetInvoiceNotFound(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Create request with authentication
	req, err := http.NewRequest("GET", baseURL+"/api/v1/invoices/non-existent-id", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestMerchantAPIAuthentication tests that merchant endpoints require authentication.
func TestMerchantAPIAuthentication(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Test creating invoice without authentication - should fail
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

	req, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header - should fail

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// API.md requires authentication for merchant endpoints
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Merchant endpoints should require authentication")

	// Verify error response structure matches API.md
	var errorResponse map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		// API.md error structure should have error.type, error.code, error.message
		if errorObj, ok := errorResponse["error"].(map[string]interface{}); ok {
			assert.Contains(t, errorObj, "type", "API.md error should have error.type")
			assert.Contains(t, errorObj, "code", "API.md error should have error.code")
			assert.Contains(t, errorObj, "message", "API.md error should have error.message")
		}
		assert.Contains(t, errorResponse, "request_id", "API.md error should have request_id")
	}
}

// TestMerchantAPICancelInvoice tests the cancel invoice endpoint.
func TestMerchantAPICancelInvoice(t *testing.T) {
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

	// Now cancel the invoice with authentication
	cancelHTTPReq, err := http.NewRequest("POST", baseURL+"/api/v1/invoices/"+invoiceID+"/cancel", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		t.Fatalf("Failed to create cancel request: %v", err)
	}
	cancelHTTPReq.Header.Set("Content-Type", "application/json")
	cancelHTTPReq.Header.Set("Authorization", "Bearer sk_test_abc123def456")

	cancelResp, err := http.DefaultClient.Do(cancelHTTPReq)
	if err != nil {
		t.Fatalf("Failed to cancel invoice: %v", err)
	}
	defer cancelResp.Body.Close()

	// Should return 501 Not Implemented as per current implementation
	assert.Equal(t, http.StatusNotImplemented, cancelResp.StatusCode)

	// Verify error response
	var errorResponse map[string]interface{}
	body, _ := io.ReadAll(cancelResp.Body)
	if unmarshalErr := json.Unmarshal(body, &errorResponse); unmarshalErr == nil {
		assert.Contains(t, errorResponse, "error")
		assert.Contains(t, errorResponse, "message")
	}
}

// TestMerchantAPIListInvoices tests the list invoices endpoint.
func TestMerchantAPIListInvoices(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Create request with authentication
	req, err := http.NewRequest("GET", baseURL+"/api/v1/invoices", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to list invoices: %v", err)
	}
	defer resp.Body.Close()

	// Should return 501 Not Implemented as per current implementation
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)

	// Verify error response
	var errorResponse map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &errorResponse); unmarshalErr == nil {
		assert.Contains(t, errorResponse, "error")
		assert.Contains(t, errorResponse, "message")
	}
}

// TestMerchantAPIGetAnalytics tests the analytics endpoint.
func TestMerchantAPIGetAnalytics(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Create request with authentication
	req, err := http.NewRequest("GET", baseURL+"/api/v1/analytics", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get analytics: %v", err)
	}
	defer resp.Body.Close()

	// Should return 501 Not Implemented as per current implementation
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)

	// Verify error response
	var errorResponse map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &errorResponse); unmarshalErr == nil {
		assert.Contains(t, errorResponse, "error")
		assert.Contains(t, errorResponse, "message")
	}
}
