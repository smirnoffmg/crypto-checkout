package e2e_test

import (
	"bytes"
	"crypto-checkout/test/testutil"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMerchantAPICreateInvoice tests the merchant API invoice creation endpoint.
func TestMerchantAPICreateInvoice(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Test valid invoice creation
	createReq := map[string]interface{}{
		"title":       "Test Invoice",
		"description": "Test Invoice Description",
		"items": []map[string]interface{}{
			{
				"name":        "VPN Premium Plan",
				"description": "VPN Premium Plan",
				"unit_price":  "9.99",
				"quantity":    "1",
			},
			{
				"name":        "Additional Static IP",
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

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr != nil {
		t.Fatalf("Failed to parse response: %v", unmarshalErr)
	}

	// Verify response structure matches API.md specification
	require.Contains(t, response, "id")
	require.Contains(t, response, "items")
	require.Contains(t, response, "subtotal")
	require.Contains(t, response, "tax_amount")
	require.Contains(t, response, "total")
	require.Contains(t, response, "tax_rate")
	require.Contains(t, response, "status")
	require.Contains(t, response, "created_at")

	// API.md required fields - these should be implemented
	require.Contains(t, response, "usdt_amount", "API.md requires usdt_amount field")
	require.Contains(t, response, "address", "API.md requires address field")
	require.Contains(t, response, "customer_url", "API.md requires customer_url field")
	require.Contains(t, response, "expires_at", "API.md requires expires_at field")

	// Verify invoice status is pending
	require.Equal(t, "created", response["status"])

	// Verify items structure
	items, ok := response["items"].([]interface{})
	require.True(t, ok)
	require.Len(t, items, 2)

	// Verify first item
	item1 := items[0].(map[string]interface{})
	require.Equal(t, "VPN Premium Plan", item1["description"])
	require.Equal(t, "9.99", item1["unit_price"])
	require.Equal(t, "1", item1["quantity"])
	require.Contains(t, item1, "total")
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
				"title":       "Test Invoice",
				"description": "Test Invoice Description",
				"items": []map[string]interface{}{
					{
						"name":        "Test Item",
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
				"title":       "Test Invoice",
				"description": "Test Invoice Description",
				"items": []map[string]interface{}{
					{
						"name":        "Test Item",
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
				"title":       "Test Invoice",
				"description": "Test Invoice Description",
				"items": []map[string]interface{}{
					{
						"name":        "Test Item",
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

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestMerchantAPIGetInvoice tests the merchant API get invoice endpoint.
func TestMerchantAPIGetInvoice(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// First create an invoice
	createReq := map[string]interface{}{
		"title":       "Test Invoice",
		"description": "Test Invoice Description",
		"items": []map[string]interface{}{
			{
				"name":        "Test Item",
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
	req, err := http.NewRequest("GET", baseURL+"/api/v1/invoices/"+invoiceID, http.NoBody)
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

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr != nil {
		t.Fatalf("Failed to parse response: %v", unmarshalErr)
	}

	// Verify response structure matches API.md merchant response
	require.Equal(t, invoiceID, response["id"])
	require.Contains(t, response, "items")
	require.Contains(t, response, "subtotal")
	require.Contains(t, response, "tax_amount")
	require.Contains(t, response, "total")
	require.Contains(t, response, "tax_rate")
	require.Contains(t, response, "status")
	require.Contains(t, response, "created_at")

	// API.md required fields - these should be implemented
	require.Contains(t, response, "usdt_amount", "API.md requires usdt_amount field")
	require.Contains(t, response, "address", "API.md requires address field")
	require.Contains(t, response, "customer_url", "API.md requires customer_url field")
	require.Contains(t, response, "expires_at", "API.md requires expires_at field")

	// Verify status is pending
	require.Equal(t, "created", response["status"])
}

// TestMerchantAPIGetInvoiceNotFound tests 404 for non-existent invoice.
func TestMerchantAPIGetInvoiceNotFound(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Create request with authentication
	req, err := http.NewRequest("GET", baseURL+"/api/v1/invoices/non-existent-id", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestMerchantAPIAuthentication tests that merchant endpoints require authentication.
func TestMerchantAPIAuthentication(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Test creating invoice without authentication - should fail
	createReq := map[string]interface{}{
		"title":       "Test Invoice",
		"description": "Test Invoice Description",
		"items": []map[string]interface{}{
			{
				"name":        "Test Item",
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
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Merchant endpoints should require authentication")

	// Verify error response structure matches API.md
	var errorResponse map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		// API.md error structure should have error.type, error.code, error.message
		if errorObj, ok := errorResponse["error"].(map[string]interface{}); ok {
			require.Contains(t, errorObj, "type", "API.md error should have error.type")
			require.Contains(t, errorObj, "code", "API.md error should have error.code")
			require.Contains(t, errorObj, "message", "API.md error should have error.message")
		}
		require.Contains(t, errorResponse, "request_id", "API.md error should have request_id")
	}
}

// TestMerchantAPICancelInvoice tests the cancel invoice endpoint.
func TestMerchantAPICancelInvoice(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// First create an invoice
	createReq := map[string]interface{}{
		"title":       "Test Invoice",
		"description": "Test Invoice Description",
		"items": []map[string]interface{}{
			{
				"name":        "Test Item",
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
	cancelReq := map[string]interface{}{
		"reason": "Test cancellation",
	}
	cancelReqBody, _ := json.Marshal(cancelReq)
	cancelHTTPReq, err := http.NewRequest(
		"POST",
		baseURL+"/api/v1/invoices/"+invoiceID+"/cancel",
		bytes.NewBuffer(cancelReqBody),
	)
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

	// Should return 200 OK as the cancel invoice endpoint is implemented
	require.Equal(t, http.StatusOK, cancelResp.StatusCode)

	// Verify response structure
	var response map[string]interface{}
	body, _ := io.ReadAll(cancelResp.Body)
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr == nil {
		require.Contains(t, response, "id")
		require.Contains(t, response, "status")
		require.Contains(t, response, "reason")
		require.Contains(t, response, "cancelled_at")
	}
}

// TestMerchantAPIListInvoices tests the list invoices endpoint.
func TestMerchantAPIListInvoices(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Create request with authentication
	req, err := http.NewRequest("GET", baseURL+"/api/v1/invoices", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to list invoices: %v", err)
	}
	defer resp.Body.Close()

	// Should return 200 OK as the list invoices endpoint is implemented
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify response structure
	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr == nil {
		require.Contains(t, response, "invoices")
		require.Contains(t, response, "total")
		require.Contains(t, response, "page")
		require.Contains(t, response, "limit")
	}
}

// TestMerchantAPIGetAnalytics tests the analytics endpoint.
func TestMerchantAPIGetAnalytics(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Create request with authentication
	req, err := http.NewRequest("GET", baseURL+"/api/v1/analytics", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get analytics: %v", err)
	}
	defer resp.Body.Close()

	// Should return 200 OK as the analytics endpoint is implemented
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify response structure
	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr == nil {
		require.Contains(t, response, "summary")
		require.Contains(t, response, "invoices")
		require.Contains(t, response, "payments")

		// Check nested structure
		if invoices, ok := response["invoices"].(map[string]interface{}); ok {
			require.Contains(t, invoices, "by_status")
		}
		if payments, ok := response["payments"].(map[string]interface{}); ok {
			require.Contains(t, payments, "by_status")
		}
	}
}
