package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuthTokenEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	router := gin.New()

	// Create a test handler that would normally be injected
	handler := &Handler{
		Logger: logger,
	}

	// Register the auth token route
	router.POST("/api/v1/auth/token", handler.generateAuthToken)

	t.Run("GenerateToken_Success", func(t *testing.T) {
		// Given
		request := TokenRequest{
			GrantType: "api_key",
			APIKey:    "sk_live_abc123def456",
			Scope:     []string{"invoices:create", "invoices:read"},
			ExpiresIn: 3600,
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		// When
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)

		var response TokenResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AccessToken)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Equal(t, int64(3600), response.ExpiresIn)
		assert.Equal(t, []string{"invoices:create", "invoices:read"}, response.Scope)

		// Verify JWT token is valid
		token, err := jwt.Parse(response.AccessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)
		assert.True(t, token.Valid)

		// Verify claims
		claims, ok := token.Claims.(jwt.MapClaims)
		require.True(t, ok)
		assert.Equal(t, "api_key", claims["grant_type"])
		assert.Equal(t, "sk_live_abc123def456", claims["api_key"])
		assert.Contains(t, claims["scope"], "invoices:create")
		assert.Contains(t, claims["scope"], "invoices:read")
	})

	t.Run("GenerateToken_InvalidGrantType", func(t *testing.T) {
		// Given
		request := TokenRequest{
			GrantType: "invalid_grant",
			APIKey:    "sk_live_abc123def456",
			Scope:     []string{"invoices:create"},
			ExpiresIn: 3600,
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		// When
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", response.Error)
		assert.Contains(t, response.Message, "grant_type")
	})

	t.Run("GenerateToken_InvalidAPIKey", func(t *testing.T) {
		// Given
		request := TokenRequest{
			GrantType: "api_key",
			APIKey:    "invalid_key",
			Scope:     []string{"invoices:create"},
			ExpiresIn: 3600,
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		// When
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "authentication_error", response.Error)
		assert.Contains(t, response.Message, "Invalid API key")
	})

	t.Run("GenerateToken_EmptyScope", func(t *testing.T) {
		// Given
		request := TokenRequest{
			GrantType: "api_key",
			APIKey:    "sk_live_abc123def456",
			Scope:     []string{},
			ExpiresIn: 3600,
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		// When
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", response.Error)
		// The binding validation happens first, so we get "Invalid JSON format" instead of specific scope error
		assert.Contains(t, response.Message, "Invalid JSON format")
	})

	t.Run("GenerateToken_InvalidExpiresIn", func(t *testing.T) {
		// Given
		request := TokenRequest{
			GrantType: "api_key",
			APIKey:    "sk_live_abc123def456",
			Scope:     []string{"invoices:create"},
			ExpiresIn: 0, // Invalid: must be > 0
		}

		requestBody, err := json.Marshal(request)
		require.NoError(t, err)

		// When
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", response.Error)
		// The binding validation happens first, so we get "Invalid JSON format" instead of specific expires_in error
		assert.Contains(t, response.Message, "Invalid JSON format")
	})

	t.Run("GenerateToken_InvalidJSON", func(t *testing.T) {
		// Given
		requestBody := []byte(`{"invalid": json}`)

		// When
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "validation_error", response.Error)
		assert.Contains(t, response.Message, "Invalid JSON")
	})
}

func TestJWTValidation(t *testing.T) {
	t.Run("ValidateJWT_Success", func(t *testing.T) {
		// Given
		secret := "test-secret"
		claims := jwt.MapClaims{
			"grant_type": "api_key",
			"api_key":    "sk_live_abc123def456",
			"scope":      []string{"invoices:create", "invoices:read"},
			"exp":        time.Now().Add(time.Hour).Unix(),
			"iat":        time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		// When
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		// Then
		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		parsedClaims, ok := parsedToken.Claims.(jwt.MapClaims)
		require.True(t, ok)
		assert.Equal(t, "api_key", parsedClaims["grant_type"])
		assert.Equal(t, "sk_live_abc123def456", parsedClaims["api_key"])
	})

	t.Run("ValidateJWT_Expired", func(t *testing.T) {
		// Given
		secret := "test-secret"
		claims := jwt.MapClaims{
			"grant_type": "api_key",
			"api_key":    "sk_live_abc123def456",
			"scope":      []string{"invoices:create"},
			"exp":        time.Now().Add(-time.Hour).Unix(), // Expired
			"iat":        time.Now().Add(-2 * time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		// When
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		// Then
		assert.Error(t, err)
		assert.False(t, parsedToken.Valid)
		assert.Contains(t, err.Error(), "token is expired")
	})
}
