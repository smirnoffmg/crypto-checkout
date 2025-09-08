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

// TestAdjustPlatformFeeRate tests Scenario 2: Platform admin negotiates custom fee rate with high-volume merchant
func TestAdjustPlatformFeeRate(t *testing.T) {
	baseURL := StartTestApp(t)

	t.Run("Complete Fee Rate Adjustment Flow", func(t *testing.T) {
		// Step 1: Create a high-volume merchant first
		merchantData := map[string]interface{}{
			"name":          "High Volume VPN Service",
			"email":         "enterprise@vpnservice.com",
			"website":       "https://enterprise.vpnservice.com",
			"description":   "Enterprise VPN service with 100k+ monthly volume",
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

		// Verify initial fee rate is 1% (default)
		settings := merchantResponse["settings"].(map[string]interface{})
		assert.Equal(t, 1.0, settings["platform_fee_percentage"])

		// Step 2: Admin reviews merchant volume (>$100k/month)
		// In a real system, this would be calculated from actual transaction data
		// For testing, we'll simulate this by checking merchant analytics

		// Step 3: Admin negotiates reduced rate from 1% to 0.7%
		feeUpdateData := map[string]interface{}{
			"platform_fee_percentage": 0.7,
			"reason":                  "High volume merchant - negotiated rate",
			"effective_date":          "2024-01-01T00:00:00Z",
		}

		updateReqBody, err := json.Marshal(feeUpdateData)
		require.NoError(t, err)

		// Update merchant settings with new fee rate
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

		// Should return 200 OK for successful update
		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updateResponse map[string]interface{}
		err = json.NewDecoder(updateResp.Body).Decode(&updateResponse)
		require.NoError(t, err)

		// Verify new fee rate is set
		updatedSettings := updateResponse["settings"].(map[string]interface{})
		assert.Equal(t, 0.7, updatedSettings["platform_fee_percentage"])

		// Step 4: Merchant receives notification of new rate
		// This would typically be handled by a notification service
		// For now, we verify the settings were updated

		// Step 5: New rate applies to future settlements
		// Verify the merchant can be retrieved with updated settings
		getResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/merchants/%s", baseURL, merchantID),
		)
		require.NoError(t, err)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var getMerchantResponse map[string]interface{}
		err = json.NewDecoder(getResp.Body).Decode(&getMerchantResponse)
		require.NoError(t, err)

		finalSettings := getMerchantResponse["settings"].(map[string]interface{})
		assert.Equal(t, 0.7, finalSettings["platform_fee_percentage"])

		// Verify fee rate history is maintained
		// In a real system, we would track fee rate changes for audit purposes
		assert.Contains(t, updateResponse, "updated_at")
	})

	t.Run("Fee Rate Validation", func(t *testing.T) {
		// Create a merchant for testing
		merchantData := map[string]interface{}{
			"name":     "Test Merchant",
			"email":    "test@merchant.com",
			"website":  "https://test.com",
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

		// Test fee rate below minimum (0.1%)
		invalidFeeData := map[string]interface{}{
			"platform_fee_percentage": 0.05, // Below 0.1% minimum
		}

		invalidReqBody, err := json.Marshal(invalidFeeData)
		require.NoError(t, err)

		invalidReq, err := http.NewRequest(
			"PUT",
			fmt.Sprintf("%s/api/v1/merchants/%s/settings", baseURL, merchantID),
			bytes.NewBuffer(invalidReqBody),
		)
		require.NoError(t, err)
		invalidReq.Header.Set("Content-Type", "application/json")

		invalidResp, err := http.DefaultClient.Do(invalidReq)
		require.NoError(t, err)
		defer invalidResp.Body.Close()

		// Should return 400 Bad Request for invalid fee rate
		assert.Equal(t, http.StatusBadRequest, invalidResp.StatusCode)

		// Test fee rate above maximum (5.0%)
		excessiveFeeData := map[string]interface{}{
			"platform_fee_percentage": 6.0, // Above 5.0% maximum
		}

		excessiveReqBody, err := json.Marshal(excessiveFeeData)
		require.NoError(t, err)

		excessiveReq, err := http.NewRequest(
			"PUT",
			fmt.Sprintf("%s/api/v1/merchants/%s/settings", baseURL, merchantID),
			bytes.NewBuffer(excessiveReqBody),
		)
		require.NoError(t, err)
		excessiveReq.Header.Set("Content-Type", "application/json")

		excessiveResp, err := http.DefaultClient.Do(excessiveReq)
		require.NoError(t, err)
		defer excessiveResp.Body.Close()

		// Should return 400 Bad Request for excessive fee rate
		assert.Equal(t, http.StatusBadRequest, excessiveResp.StatusCode)
	})

	t.Run("Fee Rate Audit Trail", func(t *testing.T) {
		// Create a merchant for testing
		merchantData := map[string]interface{}{
			"name":     "Audit Test Merchant",
			"email":    "audit@merchant.com",
			"website":  "https://audit.com",
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

		// Make multiple fee rate changes
		feeRates := []float64{1.0, 0.8, 0.6, 0.7}

		for i, feeRate := range feeRates {
			feeUpdateData := map[string]interface{}{
				"platform_fee_percentage": feeRate,
				"reason":                  fmt.Sprintf("Fee adjustment #%d", i+1),
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
			updateResp.Body.Close()

			assert.Equal(t, http.StatusOK, updateResp.StatusCode)
		}

		// Verify fee rate history can be retrieved
		historyResp, err := http.Get(
			fmt.Sprintf("%s/api/v1/merchants/%s/fee-history", baseURL, merchantID),
		)
		require.NoError(t, err)
		defer historyResp.Body.Close()

		// Should return 200 OK with fee rate history
		assert.Equal(t, http.StatusOK, historyResp.StatusCode)

		var historyResponse map[string]interface{}
		err = json.NewDecoder(historyResp.Body).Decode(&historyResponse)
		require.NoError(t, err)

		// Verify history contains all fee rate changes
		history := historyResponse["fee_history"].([]interface{})
		assert.Len(t, history, len(feeRates), "Should have history for all fee rate changes")

		// Verify the latest fee rate is correct
		latestEntry := history[len(history)-1].(map[string]interface{})
		assert.Equal(t, 0.7, latestEntry["fee_percentage"])
	})
}
