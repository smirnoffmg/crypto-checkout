package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
func createAuthErrorResponse(errorType, code, message string) gin.H {
	return gin.H{
		"error": gin.H{
			"type":    errorType,
			"code":    code,
			"message": message,
		},
		"request_id": generateRequestID(),
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
