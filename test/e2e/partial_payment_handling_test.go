package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPartialPaymentHandling tests Scenario 5: Customer sends $5.00 instead of $9.99, then completes payment
func TestPartialPaymentHandling(t *testing.T) {
	baseURL := StartTestApp(t)

	t.Run("Complete Partial Payment Flow", func(t *testing.T) {
		// Step 1: Create a merchant
		merchantData := map[string]interface{}{
			"name":          "Partial Payment Test Merchant",
			"email":         "partial@merchant.com",
			"website":       "https://partial.com",
			"description":   "Merchant for testing partial payments",
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

		// Step 2: Create an invoice for $9.99
		invoiceData := map[string]interface{}{
			"title":           "VPN Subscription - Monthly",
			"description":     "Monthly VPN subscription",
			"subtotal":        9.99,
			"tax":             0.00,
			"total":           9.99,
			"currency":        "USD",
			"crypto_currency": "USDT",
		}

		invoiceReqBody, err := json.Marshal(invoiceData)
		require.NoError(t, err)

		invoiceResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/invoices", baseURL),
			"application/json",
			bytes.NewBuffer(invoiceReqBody),
		)
		require.NoError(t, err)
		defer invoiceResp.Body.Close()

		assert.Equal(t, http.StatusCreated, invoiceResp.StatusCode)

		var invoiceResponse map[string]interface{}
		err = json.NewDecoder(invoiceResp.Body).Decode(&invoiceResponse)
		require.NoError(t, err)

		invoiceID := invoiceResponse["id"].(string)
		require.NotEmpty(t, invoiceID)

		// Verify invoice is in pending status
		assert.Equal(t, "pending", invoiceResponse["status"])
		assert.Equal(t, 9.99, invoiceResponse["total"])

		// Step 3: First payment detected ($5.00 USDT)
		firstPaymentData := map[string]interface{}{
			"tx_hash":      "0xfirst1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"amount":       "5.0",
			"from_address": "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			"to_address":   "TTestAddress123456789012345678901234567890",
			"block_number": 12345678,
			"block_hash":   "0xfirstabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"network_fee":  "0.001",
			"detected_at":  time.Now().Format(time.RFC3339),
		}

		firstPaymentReqBody, err := json.Marshal(firstPaymentData)
		require.NoError(t, err)

		firstPaymentResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/payments", baseURL),
			"application/json",
			bytes.NewBuffer(firstPaymentReqBody),
		)
		require.NoError(t, err)
		defer firstPaymentResp.Body.Close()

		assert.Equal(t, http.StatusCreated, firstPaymentResp.StatusCode)

		var firstPaymentResponse map[string]interface{}
		err = json.NewDecoder(firstPaymentResp.Body).Decode(&firstPaymentResponse)
		require.NoError(t, err)

		firstPaymentID := firstPaymentResponse["id"].(string)
		require.NotEmpty(t, firstPaymentID)

		// Step 4: Invoice status updated to "partial"
		time.Sleep(100 * time.Millisecond) // Allow for async processing

		invoiceStatusResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/invoices/%s", baseURL, invoiceID),
		)
		require.NoError(t, err)
		defer invoiceStatusResp.Body.Close()

		assert.Equal(t, http.StatusOK, invoiceStatusResp.StatusCode)

		var invoiceStatusResponse map[string]interface{}
		err = json.NewDecoder(invoiceStatusResp.Body).Decode(&invoiceStatusResponse)
		require.NoError(t, err)

		assert.Equal(t, "partial", invoiceStatusResponse["status"])

		// Verify payment progress (received vs required)
		assert.Contains(t, invoiceStatusResponse, "paid_amount")
		assert.Contains(t, invoiceStatusResponse, "remaining_amount")

		paidAmount := invoiceStatusResponse["paid_amount"].(float64)
		remainingAmount := invoiceStatusResponse["remaining_amount"].(float64)

		assert.Equal(t, 5.0, paidAmount)
		assert.Equal(t, 4.99, remainingAmount)

		// Step 5: Customer sees progress (50% paid)
		// Check payment progress endpoint
		progressResp, err := http.Get(
			fmt.Sprintf("%s/invoice/%s/status", baseURL, invoiceID),
		)
		require.NoError(t, err)
		defer progressResp.Body.Close()

		assert.Equal(t, http.StatusOK, progressResp.StatusCode)

		var progressResponse map[string]interface{}
		err = json.NewDecoder(progressResp.Body).Decode(&progressResponse)
		require.NoError(t, err)

		assert.Equal(t, "partial", progressResponse["status"])
		assert.Contains(t, progressResponse, "progress_percentage")

		progressPercentage := progressResponse["progress_percentage"].(float64)
		assert.Equal(t, 50.05, progressPercentage) // 5.0 / 9.99 * 100

		// Step 6: Second payment sent ($4.99 USDT)
		secondPaymentData := map[string]interface{}{
			"tx_hash":      "0xsecond1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"amount":       "4.99",
			"from_address": "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			"to_address":   "TTestAddress123456789012345678901234567890",
			"block_number": 12345679,
			"block_hash":   "0xsecondabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"network_fee":  "0.001",
			"detected_at":  time.Now().Format(time.RFC3339),
		}

		secondPaymentReqBody, err := json.Marshal(secondPaymentData)
		require.NoError(t, err)

		secondPaymentResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/payments", baseURL),
			"application/json",
			bytes.NewBuffer(secondPaymentReqBody),
		)
		require.NoError(t, err)
		defer secondPaymentResp.Body.Close()

		assert.Equal(t, http.StatusCreated, secondPaymentResp.StatusCode)

		var secondPaymentResponse map[string]interface{}
		err = json.NewDecoder(secondPaymentResp.Body).Decode(&secondPaymentResponse)
		require.NoError(t, err)

		secondPaymentID := secondPaymentResponse["id"].(string)
		require.NotEmpty(t, secondPaymentID)

		// Step 7: Total payment confirmed ($9.99 total)
		time.Sleep(100 * time.Millisecond) // Allow for async processing

		finalInvoiceStatusResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/invoices/%s", baseURL, invoiceID),
		)
		require.NoError(t, err)
		defer finalInvoiceStatusResp.Body.Close()

		assert.Equal(t, http.StatusOK, finalInvoiceStatusResp.StatusCode)

		var finalInvoiceStatusResponse map[string]interface{}
		err = json.NewDecoder(finalInvoiceStatusResp.Body).Decode(&finalInvoiceStatusResponse)
		require.NoError(t, err)

		assert.Equal(t, "paid", finalInvoiceStatusResponse["status"])
		assert.Contains(t, finalInvoiceStatusResponse, "paid_at")

		// Verify total amount received
		finalPaidAmount := finalInvoiceStatusResponse["paid_amount"].(float64)
		assert.Equal(t, 9.99, finalPaidAmount)

		// Step 8: Settlement processed for full amount
		settlementsResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/merchants/%s/settlements", baseURL, merchantID),
		)
		require.NoError(t, err)
		defer settlementsResp.Body.Close()

		assert.Equal(t, http.StatusOK, settlementsResp.StatusCode)

		var settlementsResponse map[string]interface{}
		err = json.NewDecoder(settlementsResp.Body).Decode(&settlementsResponse)
		require.NoError(t, err)

		settlements := settlementsResponse["settlements"].([]interface{})
		assert.Len(t, settlements, 1, "Should have one settlement for the complete payment")

		settlement := settlements[0].(map[string]interface{})

		// Step 9: Apply platform fee to full $9.99
		expectedGrossAmount := 9.99
		expectedPlatformFee := 0.10 // 1% of $9.99
		expectedNetAmount := 9.89   // $9.99 - $0.10

		assert.Equal(t, expectedGrossAmount, settlement["gross_amount"])
		assert.Equal(t, expectedPlatformFee, settlement["platform_fee"])
		assert.Equal(t, expectedNetAmount, settlement["net_amount"])
		assert.Equal(t, "completed", settlement["status"])
		assert.Equal(t, invoiceID, settlement["invoice_id"])

		// Verify both payments are linked to the settlement
		assert.Contains(t, settlement, "payment_ids")
		paymentIDs := settlement["payment_ids"].([]interface{})
		assert.Len(t, paymentIDs, 2)
		assert.Contains(t, paymentIDs, firstPaymentID)
		assert.Contains(t, paymentIDs, secondPaymentID)
	})

	t.Run("Partial Payment with Overpayment", func(t *testing.T) {
		// Create a merchant
		merchantData := map[string]interface{}{
			"name":     "Overpayment Test Merchant",
			"email":    "overpayment@merchant.com",
			"website":  "https://overpayment.com",
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

		_ = merchantResponse["id"].(string) // merchantID not used in this test

		// Create an invoice for $9.99
		invoiceData := map[string]interface{}{
			"title":           "Overpayment Test",
			"description":     "Test overpayment handling",
			"subtotal":        9.99,
			"tax":             0.00,
			"total":           9.99,
			"currency":        "USD",
			"crypto_currency": "USDT",
		}

		invoiceReqBody, err := json.Marshal(invoiceData)
		require.NoError(t, err)

		invoiceResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/invoices", baseURL),
			"application/json",
			bytes.NewBuffer(invoiceReqBody),
		)
		require.NoError(t, err)
		defer invoiceResp.Body.Close()

		assert.Equal(t, http.StatusCreated, invoiceResp.StatusCode)

		var invoiceResponse map[string]interface{}
		err = json.NewDecoder(invoiceResp.Body).Decode(&invoiceResponse)
		require.NoError(t, err)

		invoiceID := invoiceResponse["id"].(string)

		// First payment: $5.00
		firstPaymentData := map[string]interface{}{
			"tx_hash":      "0xoverfirst1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"amount":       "5.0",
			"from_address": "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			"to_address":   "TTestAddress123456789012345678901234567890",
			"block_number": 12345680,
			"block_hash":   "0xoverfirstabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"network_fee":  "0.001",
			"detected_at":  time.Now().Format(time.RFC3339),
		}

		firstPaymentReqBody, err := json.Marshal(firstPaymentData)
		require.NoError(t, err)

		firstPaymentResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/payments", baseURL),
			"application/json",
			bytes.NewBuffer(firstPaymentReqBody),
		)
		require.NoError(t, err)
		defer firstPaymentResp.Body.Close()

		assert.Equal(t, http.StatusCreated, firstPaymentResp.StatusCode)

		// Second payment: $6.00 (overpayment)
		secondPaymentData := map[string]interface{}{
			"tx_hash":      "0xoversecond1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"amount":       "6.0",
			"from_address": "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			"to_address":   "TTestAddress123456789012345678901234567890",
			"block_number": 12345681,
			"block_hash":   "0xoversecondabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"network_fee":  "0.001",
			"detected_at":  time.Now().Format(time.RFC3339),
		}

		secondPaymentReqBody, err := json.Marshal(secondPaymentData)
		require.NoError(t, err)

		secondPaymentResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/payments", baseURL),
			"application/json",
			bytes.NewBuffer(secondPaymentReqBody),
		)
		require.NoError(t, err)
		defer secondPaymentResp.Body.Close()

		assert.Equal(t, http.StatusCreated, secondPaymentResp.StatusCode)

		// Check invoice status
		time.Sleep(100 * time.Millisecond) // Allow for async processing

		invoiceStatusResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/invoices/%s", baseURL, invoiceID),
		)
		require.NoError(t, err)
		defer invoiceStatusResp.Body.Close()

		assert.Equal(t, http.StatusOK, invoiceStatusResp.StatusCode)

		var invoiceStatusResponse map[string]interface{}
		err = json.NewDecoder(invoiceStatusResp.Body).Decode(&invoiceStatusResponse)
		require.NoError(t, err)

		assert.Equal(t, "paid", invoiceStatusResponse["status"])

		// Verify overpayment handling
		paidAmount := invoiceStatusResponse["paid_amount"].(float64)
		overpaymentAmount := invoiceStatusResponse["overpayment_amount"].(float64)

		assert.Equal(t, 11.0, paidAmount)        // 5.0 + 6.0
		assert.Equal(t, 1.01, overpaymentAmount) // 11.0 - 9.99

		// Check settlement
		settlementsResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/merchants/%s/settlements", baseURL, "mock-merchant-id"),
		)
		require.NoError(t, err)
		defer settlementsResp.Body.Close()

		assert.Equal(t, http.StatusOK, settlementsResp.StatusCode)

		var settlementsResponse map[string]interface{}
		err = json.NewDecoder(settlementsResp.Body).Decode(&settlementsResponse)
		require.NoError(t, err)

		settlements := settlementsResponse["settlements"].([]interface{})
		assert.Len(t, settlements, 1)

		settlement := settlements[0].(map[string]interface{})

		// Settlement should be based on invoice amount, not total received
		expectedGrossAmount := 9.99 // Invoice amount, not total received
		expectedPlatformFee := 0.10 // 1% of $9.99
		expectedNetAmount := 9.89   // $9.99 - $0.10

		assert.Equal(t, expectedGrossAmount, settlement["gross_amount"])
		assert.Equal(t, expectedPlatformFee, settlement["platform_fee"])
		assert.Equal(t, expectedNetAmount, settlement["net_amount"])
		assert.Equal(t, "completed", settlement["status"])

		// Verify overpayment is tracked separately
		assert.Contains(t, settlement, "overpayment_amount")
		assert.Equal(t, 1.01, settlement["overpayment_amount"])
	})

	t.Run("Partial Payment Timeline", func(t *testing.T) {
		// Test the timeline described in the scenario
		startTime := time.Now()

		// Create merchant and invoice
		merchantData := map[string]interface{}{
			"name":     "Timeline Test Merchant",
			"email":    "timeline@merchant.com",
			"website":  "https://timeline.com",
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

		_ = merchantResponse["id"].(string) // merchantID not used in this test

		// Create invoice
		invoiceData := map[string]interface{}{
			"title":           "Timeline Test",
			"description":     "Test partial payment timeline",
			"subtotal":        9.99,
			"tax":             0.00,
			"total":           9.99,
			"currency":        "USD",
			"crypto_currency": "USDT",
		}

		invoiceReqBody, err := json.Marshal(invoiceData)
		require.NoError(t, err)

		invoiceResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/invoices", baseURL),
			"application/json",
			bytes.NewBuffer(invoiceReqBody),
		)
		require.NoError(t, err)
		defer invoiceResp.Body.Close()

		assert.Equal(t, http.StatusCreated, invoiceResp.StatusCode)

		var invoiceResponse map[string]interface{}
		err = json.NewDecoder(invoiceResp.Body).Decode(&invoiceResponse)
		require.NoError(t, err)

		invoiceID := invoiceResponse["id"].(string)

		// First payment at 10:00
		firstPaymentTime := startTime.Add(0 * time.Minute)
		firstPaymentData := map[string]interface{}{
			"tx_hash":      "0xtimelinefirst1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"amount":       "5.0",
			"from_address": "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			"to_address":   "TTimelineAddress123456789012345678901234567890",
			"block_number": 12345682,
			"block_hash":   "0xtimelinefirstabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"network_fee":  "0.001",
			"detected_at":  firstPaymentTime.Format(time.RFC3339),
		}

		firstPaymentReqBody, err := json.Marshal(firstPaymentData)
		require.NoError(t, err)

		firstPaymentResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/payments", baseURL),
			"application/json",
			bytes.NewBuffer(firstPaymentReqBody),
		)
		require.NoError(t, err)
		defer firstPaymentResp.Body.Close()

		assert.Equal(t, http.StatusCreated, firstPaymentResp.StatusCode)

		// Second payment at 10:15 (15 minutes later)
		secondPaymentTime := startTime.Add(15 * time.Minute)
		secondPaymentData := map[string]interface{}{
			"tx_hash":      "0xtimelinesecond1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"amount":       "4.99",
			"from_address": "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			"to_address":   "TTimelineAddress123456789012345678901234567890",
			"block_number": 12345683,
			"block_hash":   "0xtimelinesecondabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"network_fee":  "0.001",
			"detected_at":  secondPaymentTime.Format(time.RFC3339),
		}

		secondPaymentReqBody, err := json.Marshal(secondPaymentData)
		require.NoError(t, err)

		secondPaymentResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/payments", baseURL),
			"application/json",
			bytes.NewBuffer(secondPaymentReqBody),
		)
		require.NoError(t, err)
		defer secondPaymentResp.Body.Close()

		assert.Equal(t, http.StatusCreated, secondPaymentResp.StatusCode)

		// Verify timeline is tracked correctly
		time.Sleep(100 * time.Millisecond) // Allow for async processing

		// Check payment timeline
		timelineResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/invoices/%s/timeline", baseURL, invoiceID),
		)
		require.NoError(t, err)
		defer timelineResp.Body.Close()

		assert.Equal(t, http.StatusOK, timelineResp.StatusCode)

		var timelineResponse map[string]interface{}
		err = json.NewDecoder(timelineResp.Body).Decode(&timelineResponse)
		require.NoError(t, err)

		// Verify timeline events
		events := timelineResponse["events"].([]interface{})
		assert.Len(t, events, 4) // created, first_payment, second_payment, completed

		// Verify event timestamps
		createdEvent := events[0].(map[string]interface{})
		firstPaymentEvent := events[1].(map[string]interface{})
		secondPaymentEvent := events[2].(map[string]interface{})
		completedEvent := events[3].(map[string]interface{})

		assert.Equal(t, "created", createdEvent["event_type"])
		assert.Equal(t, "payment_detected", firstPaymentEvent["event_type"])
		assert.Equal(t, "payment_detected", secondPaymentEvent["event_type"])
		assert.Equal(t, "completed", completedEvent["event_type"])

		// Verify payment amounts in timeline
		assert.Equal(t, 5.0, firstPaymentEvent["amount"])
		assert.Equal(t, 4.99, secondPaymentEvent["amount"])
	})
}
