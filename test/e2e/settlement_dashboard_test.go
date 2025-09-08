package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestViewSettlementDashboard tests Scenario 3: Merchant checks earnings and fee breakdown
func TestViewSettlementDashboard(t *testing.T) {
	baseURL := StartTestApp(t)

	t.Run("Complete Settlement Dashboard Flow", func(t *testing.T) {
		// Step 1: Create a merchant with some transaction history
		merchantData := map[string]interface{}{
			"name":          "Dashboard Test Merchant",
			"email":         "dashboard@merchant.com",
			"website":       "https://dashboard.merchant.com",
			"description":   "Merchant for testing settlement dashboard",
			"business_type": "technology",
			"country":       "US",
			"currency":      "USD",
		}

		reqBody, err := json.Marshal(merchantData)
		require.NoError(t, err)

		resp, err := http.Post(
			fmt.Sprintf("%s/api/v1/merchants", baseURL),
			"application/json",
			bytes.NewBuffer(reqBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var merchantResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&merchantResponse)
		require.NoError(t, err)

		merchantID := merchantResponse["id"].(string)
		require.NotEmpty(t, merchantID)

		// Step 2: Create some invoices and simulate payments for settlement data
		invoices := []map[string]interface{}{
			{
				"title":           "VPN Subscription - Monthly",
				"description":     "Monthly VPN subscription",
				"subtotal":        9.99,
				"tax":             0.00,
				"total":           9.99,
				"currency":        "USD",
				"crypto_currency": "USDT",
			},
			{
				"title":           "VPN Subscription - Yearly",
				"description":     "Yearly VPN subscription",
				"subtotal":        99.99,
				"tax":             0.00,
				"total":           99.99,
				"currency":        "USD",
				"crypto_currency": "USDT",
			},
			{
				"title":           "Premium Add-on",
				"description":     "Premium features add-on",
				"subtotal":        19.99,
				"tax":             0.00,
				"total":           19.99,
				"currency":        "USD",
				"crypto_currency": "USDT",
			},
		}

		var createdInvoices []string
		for _, invoiceData := range invoices {
			invoiceReqBody, err := json.Marshal(invoiceData)
			require.NoError(t, err)

			invoiceResp, err := http.Post(
				fmt.Sprintf("%s/api/v1/invoices", baseURL),
				"application/json",
				bytes.NewBuffer(invoiceReqBody),
			)
			require.NoError(t, err)
			invoiceResp.Body.Close()

			assert.Equal(t, http.StatusCreated, invoiceResp.StatusCode)
			// In a real test, we would extract the invoice ID from the response
			createdInvoices = append(createdInvoices, "mock-invoice-id")
		}

		// Step 3: Simulate payments and settlements
		// In a real system, this would involve actual payment processing
		// For testing, we'll create mock settlement data

		// Step 4: Merchant logs into dashboard and views settlements page
		settlementsResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/merchants/%s/settlements", baseURL, merchantID),
		)
		require.NoError(t, err)
		defer settlementsResp.Body.Close()

		// Should return 200 OK with settlement data
		assert.Equal(t, http.StatusOK, settlementsResp.StatusCode)

		var settlementsResponse map[string]interface{}
		err = json.NewDecoder(settlementsResp.Body).Decode(&settlementsResponse)
		require.NoError(t, err)

		// Step 5: Verify fee breakdown is displayed
		settlements := settlementsResponse["settlements"].([]interface{})

		// Calculate expected totals
		expectedGrossTotal := 9.99 + 99.99 + 19.99                   // $129.97
		expectedPlatformFee := expectedGrossTotal * 0.01             // 1% = $1.30
		expectedNetTotal := expectedGrossTotal - expectedPlatformFee // $128.67

		// Verify settlement summary
		summary := settlementsResponse["summary"].(map[string]interface{})
		assert.Equal(t, expectedGrossTotal, summary["total_gross_amount"])
		assert.Equal(t, expectedPlatformFee, summary["total_platform_fees"])
		assert.Equal(t, expectedNetTotal, summary["total_net_amount"])
		assert.Equal(t, 3, summary["transaction_count"])
		assert.Equal(t, 1.0, summary["average_fee_rate"])

		// Step 6: Verify individual settlement details
		for _, settlement := range settlements {
			settlementData := settlement.(map[string]interface{})

			// Each settlement should have gross, fee, and net amounts
			assert.Contains(t, settlementData, "gross_amount")
			assert.Contains(t, settlementData, "platform_fee")
			assert.Contains(t, settlementData, "net_amount")
			assert.Contains(t, settlementData, "invoice_id")
			assert.Contains(t, settlementData, "settled_at")

			// Verify fee calculation is correct
			grossAmount := settlementData["gross_amount"].(float64)
			platformFee := settlementData["platform_fee"].(float64)
			netAmount := settlementData["net_amount"].(float64)

			expectedFee := grossAmount * 0.01
			expectedNet := grossAmount - expectedFee

			assert.Equal(t, expectedFee, platformFee, "Platform fee should be 1% of gross amount")
			assert.Equal(t, expectedNet, netAmount, "Net amount should be gross minus fee")
		}

		// Step 7: Test settlement report download
		reportResp, err := http.Get(
			fmt.Sprintf(
				"%s/api/v1/merchants/%s/settlements/report?format=csv&start_date=2024-01-01&end_date=2024-12-31",
				baseURL,
				merchantID,
			),
		)
		require.NoError(t, err)
		defer reportResp.Body.Close()

		// Should return 200 OK with CSV report
		assert.Equal(t, http.StatusOK, reportResp.StatusCode)
		assert.Equal(t, "text/csv", reportResp.Header.Get("Content-Type"))
		assert.Contains(t, reportResp.Header.Get("Content-Disposition"), "attachment")

		// Step 8: Test settlement analytics
		analyticsResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/merchants/%s/settlements/analytics", baseURL, merchantID),
		)
		require.NoError(t, err)
		defer analyticsResp.Body.Close()

		assert.Equal(t, http.StatusOK, analyticsResp.StatusCode)

		var analyticsResponse map[string]interface{}
		err = json.NewDecoder(analyticsResp.Body).Decode(&analyticsResponse)
		require.NoError(t, err)

		// Verify analytics data
		assert.Contains(t, analyticsResponse, "daily_volume")
		assert.Contains(t, analyticsResponse, "fee_breakdown")
		assert.Contains(t, analyticsResponse, "settlement_trends")
		assert.Contains(t, analyticsResponse, "success_rate")
	})

	t.Run("Settlement Dashboard with Date Filters", func(t *testing.T) {
		// Create a merchant for testing
		merchantData := map[string]interface{}{
			"name":     "Date Filter Test Merchant",
			"email":    "datefilter@merchant.com",
			"website":  "https://datefilter.com",
			"country":  "US",
			"currency": "USD",
		}

		reqBody, err := json.Marshal(merchantData)
		require.NoError(t, err)

		resp, err := http.Post(
			fmt.Sprintf("%s/api/v1/merchants", baseURL),
			"application/json",
			bytes.NewBuffer(reqBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var merchantResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&merchantResponse)
		require.NoError(t, err)

		merchantID := merchantResponse["id"].(string)

		// Test settlements with date range filter
		startDate := "2024-01-01"
		endDate := "2024-01-31"

		settlementsResp, err := http.Get(
			fmt.Sprintf(
				"%s/api/v1/merchants/%s/settlements?start_date=%s&end_date=%s",
				baseURL,
				merchantID,
				startDate,
				endDate,
			),
		)
		require.NoError(t, err)
		defer settlementsResp.Body.Close()

		assert.Equal(t, http.StatusOK, settlementsResp.StatusCode)

		var settlementsResponse map[string]interface{}
		err = json.NewDecoder(settlementsResp.Body).Decode(&settlementsResponse)
		require.NoError(t, err)

		// Verify date filtering works
		assert.Contains(t, settlementsResponse, "settlements")
		assert.Contains(t, settlementsResponse, "summary")
		assert.Contains(t, settlementsResponse, "date_range")

		dateRange := settlementsResponse["date_range"].(map[string]interface{})
		assert.Equal(t, startDate, dateRange["start_date"])
		assert.Equal(t, endDate, dateRange["end_date"])
	})

	t.Run("Settlement Dashboard with Status Filters", func(t *testing.T) {
		// Create a merchant for testing
		merchantData := map[string]interface{}{
			"name":     "Status Filter Test Merchant",
			"email":    "statusfilter@merchant.com",
			"website":  "https://statusfilter.com",
			"country":  "US",
			"currency": "USD",
		}

		reqBody, err := json.Marshal(merchantData)
		require.NoError(t, err)

		resp, err := http.Post(
			fmt.Sprintf("%s/api/v1/merchants", baseURL),
			"application/json",
			bytes.NewBuffer(reqBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var merchantResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&merchantResponse)
		require.NoError(t, err)

		merchantID := merchantResponse["id"].(string)

		// Test settlements with status filter
		statusFilter := "completed"

		settlementsResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/merchants/%s/settlements?status=%s", baseURL, merchantID, statusFilter),
		)
		require.NoError(t, err)
		defer settlementsResp.Body.Close()

		assert.Equal(t, http.StatusOK, settlementsResp.StatusCode)

		var settlementsResponse map[string]interface{}
		err = json.NewDecoder(settlementsResp.Body).Decode(&settlementsResponse)
		require.NoError(t, err)

		// Verify status filtering works
		settlements := settlementsResponse["settlements"].([]interface{})
		for _, settlement := range settlements {
			settlementData := settlement.(map[string]interface{})
			assert.Equal(t, statusFilter, settlementData["status"])
		}
	})
}
