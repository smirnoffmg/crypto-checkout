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

// TestAPIComplianceHealthCheck tests the health check endpoint.
func TestAPIComplianceHealthCheck(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to get health status from %s/health: %v", baseURL, err)
	}
	defer resp.Body.Close()

	// Read response body for detailed error reporting
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Logf("Warning: Failed to read response body: %v", readErr)
	}

	// Check status code with detailed error
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health check status code mismatch\nExpected: %d %s\nActual: %d %s\nResponse body: %s\nHeaders: %v",
			http.StatusOK, http.StatusText(http.StatusOK),
			resp.StatusCode, http.StatusText(resp.StatusCode),
			string(body), resp.Header)
	}

	// Check content type with detailed error
	actualContentType := resp.Header.Get("Content-Type")
	expectedContentType := "application/json; charset=utf-8"
	if actualContentType != expectedContentType {
		t.Errorf("Health check Content-Type mismatch\nExpected: %s\nActual: %s\nResponse body: %s\nAll headers: %v",
			expectedContentType, actualContentType,
			string(body), resp.Header)
	}

	var response map[string]interface{}
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr != nil {
		t.Fatalf("Failed to parse health check response as JSON: %v\nRaw response: %s", unmarshalErr, string(body))
	}

	// Verify health check response structure with detailed errors
	if status, ok := response["status"]; !ok {
		t.Errorf("Health check response missing 'status' field. Response: %+v", response)
	} else if status != "healthy" {
		t.Errorf("Health check status mismatch\nExpected: 'healthy'\nActual: '%v'\nFull response: %+v", status, response)
	}

	if service, ok := response["service"]; !ok {
		t.Errorf("Health check response missing 'service' field. Response: %+v", response)
	} else if service != "crypto-checkout" {
		t.Errorf("Health check service name mismatch\nExpected: 'crypto-checkout'\nActual: '%v'\nFull response: %+v", service, response)
	}

	// Log response details for debugging
	t.Logf("Health check response: Status=%d, Content-Type=%s, Body=%s", resp.StatusCode, actualContentType, string(body))
}

// TestAPIComplianceErrorHandling tests error response format compliance.
func TestAPIComplianceErrorHandling(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedError  bool
		needsAuth      bool
	}{
		{
			name:           "GET non-existent invoice",
			method:         "GET",
			path:           "/api/v1/invoices/non-existent",
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
			needsAuth:      true,
		},
		{
			name:           "GET non-existent public invoice",
			method:         "GET",
			path:           "/invoice/non-existent",
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name:           "POST invalid JSON",
			method:         "POST",
			path:           "/api/v1/invoices",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			needsAuth:      true,
		},
		{
			name:           "GET invalid QR code",
			method:         "GET",
			path:           "/invoice/non-existent/qr",
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			// Create request with auth header if needed
			var req *http.Request
			switch tt.method {
			case "GET":
				req, err = http.NewRequest("GET", baseURL+tt.path, nil)
			case "POST":
				// Send invalid JSON for POST test
				req, err = http.NewRequest("POST", baseURL+tt.path, bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
			default:
				t.Fatalf("Unsupported method: %s", tt.method)
			}

			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add auth header if needed
			if tt.needsAuth {
				req.Header.Set("Authorization", "Bearer sk_test_abc123def456")
			}

			client := &http.Client{}
			resp, err = client.Do(req)

			if err != nil {
				t.Fatalf("Failed to make request to %s %s: %v", tt.method, tt.path, err)
			}
			defer resp.Body.Close()

			// Read response body for detailed error reporting
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				t.Logf("Warning: Failed to read response body: %v", readErr)
			}

			// Check status code with detailed error
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Status code mismatch for %s %s\nExpected: %d %s\nActual: %d %s\nResponse body: %s\nHeaders: %v",
					tt.method, tt.path,
					tt.expectedStatus, http.StatusText(tt.expectedStatus),
					resp.StatusCode, http.StatusText(resp.StatusCode),
					string(body), resp.Header)
			}

			if tt.expectedError {
				// Verify error response structure
				var errorResponse map[string]interface{}
				if unmarshalErr := json.Unmarshal(body, &errorResponse); unmarshalErr == nil {
					// Current implementation error response structure
					if !assert.Contains(t, errorResponse, "error", "Error response should contain 'error' field. Response: %s", string(body)) {
						t.Logf("Full error response: %+v", errorResponse)
					}

					// TODO: When API.md error structure is implemented, check for:
					// - error.type
					// - error.code
					// - error.message
					// - error.field (for validation errors)
					// - error.documentation_url
					// - request_id

					// For now, verify basic error structure exists
					if errorField, exists := errorResponse["error"]; exists {
						assert.NotEmpty(t, errorField, "Error field should not be empty")
					}
				} else {
					t.Logf("Failed to parse error response as JSON: %v. Raw response: %s", unmarshalErr, string(body))
				}
			}

			// Log response details for debugging
			t.Logf("Response for %s %s: Status=%d, Body=%s", tt.method, tt.path, resp.StatusCode, string(body))
		})
	}
}

// TestAPIComplianceContentTypes tests that endpoints return correct content types.
func TestAPIComplianceContentTypes(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Create an invoice first
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
	createResp, postErr := createClient.Do(createHTTPReq)
	if postErr != nil {
		t.Fatalf("Failed to create invoice: %v", postErr)
	}
	defer createResp.Body.Close()

	var createResponse map[string]interface{}
	createBody, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(createBody, &createResponse)
	invoiceID := createResponse["id"].(string)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedType   string
		expectedStatus int
		needsAuth      bool
	}{
		{
			name:           "Health check",
			method:         "GET",
			path:           "/health",
			expectedType:   "application/json; charset=utf-8",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Create invoice",
			method:         "POST",
			path:           "/api/v1/invoices",
			expectedType:   "application/json; charset=utf-8",
			expectedStatus: http.StatusCreated,
			needsAuth:      true,
		},
		{
			name:           "Get invoice (merchant)",
			method:         "GET",
			path:           "/api/v1/invoices/" + invoiceID,
			expectedType:   "application/json; charset=utf-8",
			expectedStatus: http.StatusOK,
			needsAuth:      true,
		},
		{
			name:           "View invoice (customer)",
			method:         "GET",
			path:           "/invoice/" + invoiceID,
			expectedType:   "text/html; charset=utf-8",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var requestErr error

			// Create request with auth header if needed
			var req *http.Request
			switch tt.method {
			case "GET":
				req, requestErr = http.NewRequest("GET", baseURL+tt.path, nil)
			case "POST":
				req, requestErr = http.NewRequest("POST", baseURL+tt.path, bytes.NewBuffer(reqBody))
				req.Header.Set("Content-Type", "application/json")
			default:
				t.Fatalf("Unsupported method: %s", tt.method)
			}

			if requestErr != nil {
				t.Fatalf("Failed to create request: %v", requestErr)
			}

			// Add auth header if needed
			if tt.needsAuth {
				req.Header.Set("Authorization", "Bearer sk_test_abc123def456")
			}

			client := &http.Client{}
			resp, requestErr = client.Do(req)

			if requestErr != nil {
				t.Fatalf("Failed to make request to %s %s: %v", tt.method, tt.path, requestErr)
			}
			defer resp.Body.Close()

			// Read response body for detailed error reporting
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				t.Logf("Warning: Failed to read response body: %v", readErr)
			}

			// Check status code with detailed error
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Status code mismatch for %s %s\nExpected: %d %s\nActual: %d %s\nResponse body: %s\nHeaders: %v",
					tt.method, tt.path,
					tt.expectedStatus, http.StatusText(tt.expectedStatus),
					resp.StatusCode, http.StatusText(resp.StatusCode),
					string(body), resp.Header)
			}

			// Check content type with detailed error
			actualContentType := resp.Header.Get("Content-Type")
			if actualContentType != tt.expectedType {
				t.Errorf("Content-Type mismatch for %s %s\nExpected: %s\nActual: %s\nResponse body: %s\nAll headers: %v",
					tt.method, tt.path,
					tt.expectedType, actualContentType,
					string(body), resp.Header)
			}

			// Log response details for debugging
			t.Logf("Response for %s %s: Status=%d, Content-Type=%s, Body=%s",
				tt.method, tt.path, resp.StatusCode, actualContentType, string(body))
		})
	}
}

// TestAPIComplianceStatusCodes tests that endpoints return correct HTTP status codes.
func TestAPIComplianceStatusCodes(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
		needsAuth      bool
	}{
		{
			name:           "Health check success",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			description:    "Health check should return 200 OK",
		},
		{
			name:           "Create invoice success",
			method:         "POST",
			path:           "/api/v1/invoices",
			expectedStatus: http.StatusCreated,
			description:    "Invoice creation should return 201 Created",
			needsAuth:      true,
		},
		{
			name:           "Get non-existent invoice",
			method:         "GET",
			path:           "/api/v1/invoices/non-existent",
			expectedStatus: http.StatusNotFound,
			description:    "Non-existent invoice should return 404 Not Found",
			needsAuth:      true,
		},
		{
			name:           "Invalid JSON request",
			method:         "POST",
			path:           "/api/v1/invoices",
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid JSON should return 400 Bad Request",
			needsAuth:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := makeTestRequest(t, baseURL, tt.method, tt.path, tt.expectedStatus, tt.needsAuth)
			defer resp.Body.Close()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, tt.description)
		})
	}
}

// makeTestRequest makes a test request and returns the response.
func makeTestRequest(t *testing.T, baseURL, method, path string, expectedStatus int, needsAuth bool) *http.Response {
	t.Helper()

	var resp *http.Response
	var requestErr error

	// Create request with auth header if needed
	var req *http.Request
	switch method {
	case "GET":
		req, requestErr = http.NewRequest("GET", baseURL+path, nil)
	case "POST":
		if path == "/api/v1/invoices" {
			if expectedStatus == http.StatusCreated {
				req, requestErr = makeValidInvoiceRequestWithAuth(t, baseURL, path, needsAuth)
			} else {
				req, requestErr = makeInvalidInvoiceRequestWithAuth(t, baseURL, path, needsAuth)
			}
		}
	}

	if requestErr != nil {
		t.Fatalf("Failed to create request: %v", requestErr)
	}

	// Add auth header if needed
	if needsAuth {
		req.Header.Set("Authorization", "Bearer sk_test_abc123def456")
	}

	client := &http.Client{}
	resp, requestErr = client.Do(req)

	if requestErr != nil {
		t.Fatalf("Failed to make %s request to %s%s: %v", method, baseURL, path, requestErr)
	}

	// Log request details for debugging
	t.Logf("Made %s request to %s%s, got status: %d", method, baseURL, path, resp.StatusCode)

	return resp
}

// makeValidInvoiceRequest makes a valid invoice creation request.
func makeValidInvoiceRequest(t *testing.T, baseURL, path string) (*http.Response, error) {
	t.Helper()

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
	reqBody, marshalErr := json.Marshal(createReq)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal request: %v", marshalErr)
	}
	return http.Post(baseURL+path, "application/json", bytes.NewBuffer(reqBody))
}

// makeInvalidInvoiceRequest makes an invalid invoice creation request.
func makeInvalidInvoiceRequest(t *testing.T, baseURL, path string) (*http.Response, error) {
	t.Helper()
	return http.Post(baseURL+path, "application/json", bytes.NewBufferString("invalid json"))
}

// makeValidInvoiceRequestWithAuth makes a valid invoice creation request with optional auth.
func makeValidInvoiceRequestWithAuth(t *testing.T, baseURL, path string, needsAuth bool) (*http.Request, error) {
	t.Helper()

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
	reqBody, marshalErr := json.Marshal(createReq)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal request: %v", marshalErr)
	}

	req, err := http.NewRequest("POST", baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// makeInvalidInvoiceRequestWithAuth makes an invalid invoice creation request with optional auth.
func makeInvalidInvoiceRequestWithAuth(t *testing.T, baseURL, path string, needsAuth bool) (*http.Request, error) {
	t.Helper()

	req, err := http.NewRequest("POST", baseURL+path, bytes.NewBufferString("invalid json"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TestAPIComplianceInvoiceStructure tests that invoice responses match API.md structure
func TestAPIComplianceInvoiceStructure(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Create invoice with multiple items as per API.md example
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
	req, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	client := &http.Client{}
	resp, postErr := client.Do(req)
	if postErr != nil {
		t.Fatalf("Failed to create invoice: %v", postErr)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr != nil {
		t.Fatalf("Failed to parse response: %v", unmarshalErr)
	}

	// Verify required fields as per API.md specification
	requiredFields := []string{
		"id", "items", "subtotal", "tax_amount", "total", "tax_rate", "status", "created_at",
		"usdt_amount", "address", "customer_url", "expires_at", // API.md required fields
	}

	for _, field := range requiredFields {
		if !assert.Contains(t, response, field, "Response should contain %s field. Full response: %+v", field, response) {
			t.Logf("Missing field '%s' in invoice response. Available fields: %v", field, getMapKeys(response))
		}
	}

	// Verify items structure
	items, ok := response["items"].([]interface{})
	if !assert.True(t, ok, "Items should be an array. Actual type: %T, Value: %+v", response["items"], response["items"]) {
		t.Logf("Full response: %+v", response)
	}
	if !assert.Len(t, items, 2, "Should have 2 items. Actual count: %d", len(items)) {
		t.Logf("Items array: %+v", items)
	}

	// Verify first item structure
	if len(items) > 0 {
		item1, ok := items[0].(map[string]interface{})
		if !assert.True(t, ok, "First item should be a map. Actual type: %T, Value: %+v", items[0], items[0]) {
			t.Logf("Items array: %+v", items)
		} else {
			itemFields := []string{"description", "unit_price", "quantity", "total"}
			for _, field := range itemFields {
				if !assert.Contains(t, item1, field, "Item should contain %s field. Item: %+v", field, item1) {
					t.Logf("Available item fields: %v", getMapKeys(item1))
				}
			}
		}
	}

	// Verify status is created (transitions to pending after being viewed)
	if !assert.Equal(t, "created", response["status"], "New invoice should have created status. Actual status: %v", response["status"]) {
		t.Logf("Full response: %+v", response)
	}

	// Verify numeric fields are strings (as per API.md)
	if !assert.IsType(t, "", response["subtotal"], "Subtotal should be string. Actual type: %T, Value: %v", response["subtotal"], response["subtotal"]) {
		t.Logf("Full response: %+v", response)
	}
	if !assert.IsType(t, "", response["tax_amount"], "Tax amount should be string. Actual type: %T, Value: %v", response["tax_amount"], response["tax_amount"]) {
		t.Logf("Full response: %+v", response)
	}
	if !assert.IsType(t, "", response["total"], "Total should be string. Actual type: %T, Value: %v", response["total"], response["total"]) {
		t.Logf("Full response: %+v", response)
	}
	if !assert.IsType(t, "", response["tax_rate"], "Tax rate should be string. Actual type: %T, Value: %v", response["tax_rate"], response["tax_rate"]) {
		t.Logf("Full response: %+v", response)
	}

	// Log full response for debugging
	t.Logf("Invoice structure test response: %+v", response)
}

// getMapKeys returns the keys of a map[string]interface{} for debugging purposes
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function to create test logger.

// Helper function to create test logger.
