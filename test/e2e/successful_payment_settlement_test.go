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

// TestSuccessfulPaymentWithSettlement tests Scenario 4: Customer pays $9.99 for VPN subscription, settlement processed immediately
func TestSuccessfulPaymentWithSettlement(t *testing.T) {
	baseURL := StartTestApp(t)

	t.Run("Complete Payment and Settlement Flow", func(t *testing.T) {
		// Step 1: Create a merchant
		merchantData := map[string]interface{}{
			"name":          "VPN Service Provider",
			"email":         "vpn@service.com",
			"website":       "https://vpnservice.com",
			"description":   "Premium VPN service",
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
		assert.Equal(t, "USDT", invoiceResponse["crypto_currency"])

		// Step 3: Customer visits payment page and scans QR code
		// Get payment details for the invoice
		paymentDetailsResp, err := http.Get(
			fmt.Sprintf("%s/invoice/%s", baseURL, invoiceID),
		)
		require.NoError(t, err)
		defer paymentDetailsResp.Body.Close()

		assert.Equal(t, http.StatusOK, paymentDetailsResp.StatusCode)

		var paymentDetailsResponse map[string]interface{}
		err = json.NewDecoder(paymentDetailsResp.Body).Decode(&paymentDetailsResponse)
		require.NoError(t, err)

		// Verify payment details are provided
		assert.Contains(t, paymentDetailsResponse, "payment_address")
		assert.Contains(t, paymentDetailsResponse, "crypto_amount")
		assert.Contains(t, paymentDetailsResponse, "qr_code")
		assert.Contains(t, paymentDetailsResponse, "expires_at")

		paymentAddress := paymentDetailsResponse["payment_address"].(string)
		cryptoAmount := paymentDetailsResponse["crypto_amount"].(string)

		require.NotEmpty(t, paymentAddress)
		require.NotEmpty(t, cryptoAmount)

		// Step 4: Customer sends $9.99 USDT to the payment address
		// Simulate payment detection
		paymentData := map[string]interface{}{
			"tx_hash":      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"amount":       cryptoAmount,
			"from_address": "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			"to_address":   paymentAddress,
			"block_number": 12345678,
			"block_hash":   "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"network_fee":  "0.001",
			"detected_at":  time.Now().Format(time.RFC3339),
		}

		paymentReqBody, err := json.Marshal(paymentData)
		require.NoError(t, err)

		// Simulate payment detection via webhook or blockchain monitoring
		paymentResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/payments", baseURL),
			"application/json",
			bytes.NewBuffer(paymentReqBody),
		)
		require.NoError(t, err)
		defer paymentResp.Body.Close()

		assert.Equal(t, http.StatusCreated, paymentResp.StatusCode)

		var paymentResponse map[string]interface{}
		err = json.NewDecoder(paymentResp.Body).Decode(&paymentResponse)
		require.NoError(t, err)

		paymentID := paymentResponse["id"].(string)
		require.NotEmpty(t, paymentID)

		// Step 5: Payment confirmed after sufficient confirmations
		// Simulate confirmation process
		time.Sleep(100 * time.Millisecond) // Allow for async processing

		// Check payment status
		paymentStatusResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/payments/%s", baseURL, paymentID),
		)
		require.NoError(t, err)
		defer paymentStatusResp.Body.Close()

		assert.Equal(t, http.StatusOK, paymentStatusResp.StatusCode)

		var paymentStatusResponse map[string]interface{}
		err = json.NewDecoder(paymentStatusResp.Body).Decode(&paymentStatusResponse)
		require.NoError(t, err)

		// Step 6: Settlement created automatically
		// Check that settlement was created
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
		assert.Len(t, settlements, 1, "Should have one settlement")

		settlement := settlements[0].(map[string]interface{})

		// Step 7: Platform deducts 1% fee ($0.10)
		// Step 8: Merchant receives $9.89 within 30 seconds

		// Verify settlement calculation
		expectedGrossAmount := 9.99
		expectedPlatformFee := 0.10 // 1% of $9.99
		expectedNetAmount := 9.89   // $9.99 - $0.10

		assert.Equal(t, expectedGrossAmount, settlement["gross_amount"])
		assert.Equal(t, expectedPlatformFee, settlement["platform_fee"])
		assert.Equal(t, expectedNetAmount, settlement["net_amount"])
		assert.Equal(t, "completed", settlement["status"])
		assert.Equal(t, invoiceID, settlement["invoice_id"])

		// Step 9: Customer redirected to success page
		// Check invoice status is updated to paid
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
		assert.Contains(t, invoiceStatusResponse, "paid_at")

		// Verify events were emitted
		// In a real system, we would check the event store for:
		// - PaymentConfirmed
		// - InvoicePaid
		// - SettlementCreated
		// - PlatformFeeCollected
		// - SettlementCompleted
	})

	t.Run("Payment with Custom Fee Rate", func(t *testing.T) {
		// Create a merchant with custom fee rate
		merchantData := map[string]interface{}{
			"name":          "Custom Fee Merchant",
			"email":         "customfee@merchant.com",
			"website":       "https://customfee.com",
			"description":   "Merchant with custom fee rate",
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

		// Update merchant to have 0.7% fee rate
		feeUpdateData := map[string]interface{}{
			"platform_fee_percentage": 0.7,
		}

		updateReqBody, err := json.Marshal(feeUpdateData)
		require.NoError(t, err)

		updateReq, err := http.NewRequest(
			"PUT",
			fmt.Sprintf("%s/api/v1/merchants/%s/settings", baseURL, merchantID),
			bytes.NewBuffer(updateReqBody),
		)
		require.NoError(t, err)
		updateReq.Header.Set("Content-Type", "application/json")

		updateResp, err := http.DefaultClient.Do(updateReq)
		require.NoError(t, err)
		defer updateResp.Body.Close()

		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		// Create an invoice
		invoiceData := map[string]interface{}{
			"title":           "Custom Fee Test",
			"description":     "Test with custom fee rate",
			"subtotal":        100.00,
			"tax":             0.00,
			"total":           100.00,
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

		_ = invoiceResponse["id"].(string) // invoiceID not used in this test

		// Simulate payment
		paymentData := map[string]interface{}{
			"tx_hash":      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"amount":       "100.0",
			"from_address": "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			"to_address":   "TTestAddress123456789012345678901234567890",
			"block_number": 12345679,
			"block_hash":   "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"network_fee":  "0.001",
			"detected_at":  time.Now().Format(time.RFC3339),
		}

		paymentReqBody, err := json.Marshal(paymentData)
		require.NoError(t, err)

		paymentResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/payments", baseURL),
			"application/json",
			bytes.NewBuffer(paymentReqBody),
		)
		require.NoError(t, err)
		defer paymentResp.Body.Close()

		assert.Equal(t, http.StatusCreated, paymentResp.StatusCode)

		// Check settlement with custom fee rate
		time.Sleep(100 * time.Millisecond) // Allow for async processing

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
		assert.Len(t, settlements, 1)

		settlement := settlements[0].(map[string]interface{})

		// Verify custom fee calculation
		expectedGrossAmount := 100.00
		expectedPlatformFee := 0.70 // 0.7% of $100.00
		expectedNetAmount := 99.30  // $100.00 - $0.70

		assert.Equal(t, expectedGrossAmount, settlement["gross_amount"])
		assert.Equal(t, expectedPlatformFee, settlement["platform_fee"])
		assert.Equal(t, expectedNetAmount, settlement["net_amount"])
	})

	t.Run("Payment Settlement Timing", func(t *testing.T) {
		// Test that settlement is processed within 30 seconds
		startTime := time.Now()

		// Create merchant and invoice
		merchantData := map[string]interface{}{
			"name":     "Timing Test Merchant",
			"email":    "timing@merchant.com",
			"website":  "https://timing.com",
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

		// Create invoice
		invoiceData := map[string]interface{}{
			"title":           "Timing Test",
			"description":     "Test settlement timing",
			"subtotal":        50.00,
			"tax":             0.00,
			"total":           50.00,
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

		_ = invoiceResponse["id"].(string) // invoiceID not used in this test

		// Simulate payment
		paymentData := map[string]interface{}{
			"tx_hash":      "0xtiming1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"amount":       "50.0",
			"from_address": "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			"to_address":   "TTimingAddress123456789012345678901234567890",
			"block_number": 12345680,
			"block_hash":   "0xtimingabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"network_fee":  "0.001",
			"detected_at":  time.Now().Format(time.RFC3339),
		}

		paymentReqBody, err := json.Marshal(paymentData)
		require.NoError(t, err)

		paymentResp, err := http.Post(
			fmt.Sprintf("%s/api/v1/payments", baseURL),
			"application/json",
			bytes.NewBuffer(paymentReqBody),
		)
		require.NoError(t, err)
		defer paymentResp.Body.Close()

		assert.Equal(t, http.StatusCreated, paymentResp.StatusCode)

		// Wait for settlement to be processed
		var settlementCompleted bool
		maxWaitTime := 30 * time.Second
		checkInterval := 100 * time.Millisecond

		for time.Since(startTime) < maxWaitTime {
			settlementsResp, err := http.Get(
				fmt.Sprintf("%s/api/v1/merchants/%s/settlements", baseURL, merchantID),
			)
			require.NoError(t, err)

			var settlementsResponse map[string]interface{}
			err = json.NewDecoder(settlementsResp.Body).Decode(&settlementsResponse)
			require.NoError(t, err)
			settlementsResp.Body.Close()

			settlements := settlementsResponse["settlements"].([]interface{})
			if len(settlements) > 0 {
				settlement := settlements[0].(map[string]interface{})
				if settlement["status"] == "completed" {
					settlementCompleted = true
					break
				}
			}

			time.Sleep(checkInterval)
		}

		// Verify settlement was completed within 30 seconds
		assert.True(t, settlementCompleted, "Settlement should be completed within 30 seconds")
		assert.True(t, time.Since(startTime) < 30*time.Second, "Settlement should complete within 30 seconds")
	})
}
