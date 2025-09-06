package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthMiddleware validates API key authentication for merchant endpoints.
func AuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Debug("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, createAuthErrorResponse("authentication_error", "MISSING_AUTH_HEADER", "Authorization header is required"))
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Debug("Invalid Authorization header format", zap.String("header", authHeader))
			c.JSON(http.StatusUnauthorized, createAuthErrorResponse("authentication_error", "INVALID_AUTH_FORMAT", "Authorization header must be 'Bearer <token>'"))
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			logger.Debug("Empty token in Authorization header")
			c.JSON(http.StatusUnauthorized, createAuthErrorResponse("authentication_error", "EMPTY_TOKEN", "Token cannot be empty"))
			c.Abort()
			return
		}

		// Validate token format (API.md specifies sk_live_* or sk_test_*)
		if !isValidAPIToken(token) {
			logger.Debug("Invalid API token format", zap.String("token", maskToken(token)))
			c.JSON(http.StatusUnauthorized, createAuthErrorResponse("authentication_error", "INVALID_TOKEN", "Invalid API key format"))
			c.Abort()
			return
		}

		// For now, accept any valid format token (in production, validate against database)
		logger.Debug("Authentication successful", zap.String("token", maskToken(token)))
		c.Next()
	}
}

// isValidAPIToken validates the API token format according to API.md.
func isValidAPIToken(token string) bool {
	// API.md specifies: sk_live_* or sk_test_*
	return strings.HasPrefix(token, "sk_live_") || strings.HasPrefix(token, "sk_test_")
}

// maskToken masks the token for logging (shows first 8 chars + ...)
func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:8] + "..."
}

// createAuthErrorResponse creates an authentication error response matching API.md format.
func createAuthErrorResponse(errorType, code, message string) ErrorResponse {
	return ErrorResponse{
		Error:     errorType,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: generateRequestID(),
	}
}

// generateRequestID generates a simple request ID for error responses.
func generateRequestID() string {
	// Simple implementation - in production, use proper UUID or correlation ID
	return "req_auth_" + randomString(8)
}

// randomString generates a random string of specified length.
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// JWTSecret is the secret key for JWT signing (in production, use environment variable)
const JWTSecret = "test-secret"

// generateAuthToken handles POST /api/v1/auth/token requests.
// @Summary Generate JWT access token
// @Description Generate a JWT access token using API key authentication for accessing protected endpoints
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body TokenRequest true "Token generation request"
// @Success 200 {object} TokenResponse "JWT token generated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Invalid API key"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/token [post]
func (h *Handler) generateAuthToken(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Failed to bind token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("Invalid JSON format", err))
		return
	}

	// Validate grant type
	if req.GrantType != "api_key" {
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("grant_type must be 'api_key'", nil))
		return
	}

	// Validate API key format
	if !isValidAPIToken(req.APIKey) {
		h.Logger.Debug("Invalid API key format", zap.String("token", maskToken(req.APIKey)))
		c.JSON(http.StatusUnauthorized, createAuthErrorResponse("authentication_error", "INVALID_API_KEY", "Invalid API key format"))
		return
	}

	// Validate scope
	if len(req.Scope) == 0 {
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("scope is required and cannot be empty", nil))
		return
	}

	// Validate expires_in
	if req.ExpiresIn <= 0 || req.ExpiresIn > 86400 { // Max 24 hours
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("expires_in must be between 1 and 86400 seconds", nil))
		return
	}

	// TODO: In production, validate API key against database and check permissions
	// For now, we'll accept any valid format API key

	// Generate JWT token
	token, err := h.generateJWTToken(req.APIKey, req.Scope, req.ExpiresIn)
	if err != nil {
		h.Logger.Error("Failed to generate JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, createAuthErrorResponse("token_generation_error", "TOKEN_GENERATION_FAILED", "Failed to generate access token"))
		return
	}

	response := TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   req.ExpiresIn,
		Scope:       req.Scope,
	}

	c.JSON(http.StatusOK, response)
}

// generateJWTToken creates a JWT token with the specified claims.
func (h *Handler) generateJWTToken(apiKey string, scope []string, expiresIn int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"grant_type": "api_key",
		"api_key":    apiKey,
		"scope":      scope,
		"iat":        now.Unix(),
		"exp":        now.Add(time.Duration(expiresIn) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

// createValidationErrorResponse creates a validation error response.
func createValidationErrorResponse(message string, err error) ErrorResponse {
	response := ErrorResponse{
		Error:     "validation_error",
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err != nil {
		response.Details = map[string]interface{}{
			"error": err.Error(),
		}
	}

	return response
}
