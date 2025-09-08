package e2e_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHealthCheckE2E(t *testing.T) {
	baseURL := StartTestApp(t)

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
