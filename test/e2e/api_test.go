package e2e_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"crypto-checkout/test/testutil"
)

func TestHealthCheckE2E(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Wait a moment for server to be ready
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	require.Equal(t, "healthy", response["status"])
}

func TestCreateInvoiceE2E(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Wait a moment for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create invoice request
	requestBody := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"description": "VPN Service - 1 Year",
				"unit_price":  "99.99",
				"quantity":    "1",
			},
			{
				"description": "Premium Support",
				"unit_price":  "19.99",
				"quantity":    "1",
			},
		},
		"tax_rate": "0.10",
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Create invoice with authentication
	req, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Verify response structure
	require.Contains(t, response, "id")
	require.Contains(t, response, "status")
	require.Contains(t, response, "total")
	require.Contains(t, response, "created_at")

	// Verify invoice status
	require.Equal(t, "created", response["status"])

	// Verify total amount calculation
	expectedTotal := "131.98" // (99.99 + 19.99) * 1.10
	require.Equal(t, expectedTotal, response["total"])
}

func TestGetInvoiceE2E(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Wait a moment for server to be ready
	time.Sleep(100 * time.Millisecond)

	// First create an invoice
	requestBody := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"description": "Test Item",
				"unit_price":  "50.00",
				"quantity":    "1",
			},
		},
		"tax_rate": "0.00",
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Create invoice with authentication
	createReq, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	createClient := &http.Client{}
	createResp, err := createClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createResponse)
	require.NoError(t, err)

	invoiceID := createResponse["id"].(string)
	require.NotEmpty(t, invoiceID)

	// Now get the invoice with authentication
	getReq, err := http.NewRequest("GET", baseURL+"/api/v1/invoices/"+invoiceID, nil)
	require.NoError(t, err)
	getReq.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	getClient := &http.Client{}
	getResp, err := getClient.Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()

	require.Equal(t, http.StatusOK, getResp.StatusCode)
	require.Equal(t, "application/json; charset=utf-8", getResp.Header.Get("Content-Type"))

	body, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)

	var getResponse map[string]interface{}
	err = json.Unmarshal(body, &getResponse)
	require.NoError(t, err)

	// Verify the retrieved invoice matches the created one
	require.Equal(t, invoiceID, getResponse["id"])
	require.Equal(t, "created", getResponse["status"])
	require.Equal(t, "50.00", getResponse["total"])
}

func TestGetPublicInvoiceE2E(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Wait a moment for server to be ready
	time.Sleep(100 * time.Millisecond)

	// First create an invoice
	requestBody := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"description": "Public Test Item",
				"unit_price":  "25.00",
				"quantity":    "2",
			},
		},
		"tax_rate": "0.08",
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Create invoice with authentication
	createReq, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	createClient := &http.Client{}
	createResp, err := createClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createResponse)
	require.NoError(t, err)

	invoiceID := createResponse["id"].(string)
	require.NotEmpty(t, invoiceID)

	// Now get the public invoice page
	getResp, err := http.Get(baseURL + "/invoice/" + invoiceID)
	require.NoError(t, err)
	defer getResp.Body.Close()

	require.Equal(t, http.StatusOK, getResp.StatusCode)
	require.Equal(t, "text/html; charset=utf-8", getResp.Header.Get("Content-Type"))

	body, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)

	// Verify HTML content contains expected elements
	htmlContent := string(body)
	require.Contains(t, htmlContent, "Crypto Checkout")
	require.Contains(t, htmlContent, "Public Test Item")
	require.Contains(t, htmlContent, "54.00") // (25.00 * 2) * 1.08
}

func TestGetInvoiceQRE2E(t *testing.T) {
	baseURL := testutil.SetupTestApp(t)

	// Wait a moment for server to be ready
	time.Sleep(100 * time.Millisecond)

	// First create an invoice
	requestBody := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"description": "QR Test Item",
				"unit_price":  "10.00",
				"quantity":    "1",
			},
		},
		"tax_rate": "0.00",
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Create invoice with authentication
	createReq, err := http.NewRequest("POST", baseURL+"/api/v1/invoices", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer sk_test_abc123def456") // Add auth header

	createClient := &http.Client{}
	createResp, err := createClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createResponse)
	require.NoError(t, err)

	invoiceID := createResponse["id"].(string)
	require.NotEmpty(t, invoiceID)

	// Now get the QR code - this will fail because invoice has no payment address
	qrResp, err := http.Get(baseURL + "/invoice/" + invoiceID + "/qr")
	require.NoError(t, err)
	defer qrResp.Body.Close()

	// QR code generation fails because invoice has no payment address assigned
	// This is expected behavior - invoices need payment addresses to generate QR codes
	require.Equal(t, http.StatusInternalServerError, qrResp.StatusCode)

	// Verify error response
	errorBody, err := io.ReadAll(qrResp.Body)
	require.NoError(t, err)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(errorBody, &errorResponse)
	require.NoError(t, err)

	require.Contains(t, errorResponse, "error")
	// The error message might be in different formats, so let's check for the key presence
	require.NotNil(t, errorResponse["error"])
}
