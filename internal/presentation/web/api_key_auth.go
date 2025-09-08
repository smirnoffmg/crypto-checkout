package web

import (
	"crypto-checkout/internal/domain/merchant"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// APIKeyAuthMiddleware validates API key authentication using the merchant domain.
type APIKeyAuthMiddleware struct {
	apiKeyService merchant.APIKeyService
	logger        *zap.Logger
}

// NewAPIKeyAuthMiddleware creates a new API key authentication middleware.
func NewAPIKeyAuthMiddleware(apiKeyService merchant.APIKeyService, logger *zap.Logger) *APIKeyAuthMiddleware {
	return &APIKeyAuthMiddleware{
		apiKeyService: apiKeyService,
		logger:        logger,
	}
}

// RequireAPIKey validates API key authentication for protected endpoints.
func (m *APIKeyAuthMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Debug("Missing Authorization header")
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse(
					"authentication_error",
					"MISSING_AUTH_HEADER",
					"Authorization header is required",
				),
			)
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.logger.Debug("Invalid Authorization header format", zap.String("header", authHeader))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse(
					"authentication_error",
					"INVALID_AUTH_FORMAT",
					"Authorization header must be 'Bearer <token>'",
				),
			)
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			m.logger.Debug("Empty token in Authorization header")
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "EMPTY_TOKEN", "Token cannot be empty"),
			)
			c.Abort()
			return
		}

		// Validate token format (API.md specifies sk_live_* or sk_test_*)
		if !isValidAPIToken(token) {
			m.logger.Debug("Invalid API token format", zap.String("token", maskToken(token)))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "INVALID_TOKEN", "Invalid API key format"),
			)
			c.Abort()
			return
		}

		// Check if service is available
		if m.apiKeyService == nil {
			m.logger.Error("API key service not initialized")
			c.JSON(
				http.StatusInternalServerError,
				createAuthErrorResponse(
					"internal_error",
					"SERVICE_UNAVAILABLE",
					"Authentication service not available",
				),
			)
			c.Abort()
			return
		}

		// Validate API key against database
		ctx := c.Request.Context()
		req := &merchant.ValidateAPIKeyRequest{
			RawKey: token,
		}

		resp, err := m.apiKeyService.ValidateAPIKey(ctx, req)
		if err != nil {
			m.logger.Debug("API key validation failed", zap.String("token", maskToken(token)), zap.Error(err))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "INVALID_API_KEY", "Invalid or expired API key"),
			)
			c.Abort()
			return
		}

		if !resp.Valid {
			m.logger.Debug("API key is invalid", zap.String("token", maskToken(token)))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "INVALID_API_KEY", "Invalid or expired API key"),
			)
			c.Abort()
			return
		}

		// Store API key information in context for use by handlers
		if resp.APIKey != nil {
			c.Set("api_key_id", resp.APIKey.ID())
			c.Set("merchant_id", resp.APIKey.MerchantID())
			c.Set("api_key_permissions", resp.APIKey.Permissions())

			m.logger.Debug("API key authentication successful",
				zap.String("api_key_id", resp.APIKey.ID()),
				zap.String("merchant_id", resp.APIKey.MerchantID()),
				zap.String("token", maskToken(token)),
			)
		}

		c.Next()
	}
}

// RequirePermission validates API key authentication and checks for specific permission.
func (m *APIKeyAuthMiddleware) RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, validate the API key
		m.RequireAPIKey()(c)
		if c.IsAborted() {
			return
		}

		// Get permissions from context
		permissions, exists := c.Get("api_key_permissions")
		if !exists {
			m.logger.Error("API key permissions not found in context")
			c.JSON(
				http.StatusInternalServerError,
				createAuthErrorResponse(
					"authentication_error",
					"MISSING_PERMISSIONS",
					"API key permissions not available",
				),
			)
			c.Abort()
			return
		}

		permissionList, ok := permissions.([]string)
		if !ok {
			m.logger.Error("Invalid permissions type in context")
			c.JSON(
				http.StatusInternalServerError,
				createAuthErrorResponse("authentication_error", "INVALID_PERMISSIONS", "Invalid permissions format"),
			)
			c.Abort()
			return
		}

		// Check if API key has required permission
		hasPermission := false
		for _, permission := range permissionList {
			if permission == requiredPermission || permission == "*" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			apiKeyID, _ := c.Get("api_key_id")
			if apiKeyIDStr, ok := apiKeyID.(string); ok {
				m.logger.Debug("API key lacks required permission",
					zap.String("api_key_id", apiKeyIDStr),
					zap.String("required_permission", requiredPermission),
					zap.Strings("available_permissions", permissionList),
				)
			}
			c.JSON(
				http.StatusForbidden,
				createAuthErrorResponse(
					"authorization_error",
					"INSUFFICIENT_PERMISSIONS",
					"API key does not have required permission: "+requiredPermission,
				),
			)
			c.Abort()
			return
		}

		m.logger.Debug("Permission check successful",
			zap.String("required_permission", requiredPermission),
		)

		c.Next()
	}
}

// RequireMerchantOwnership validates that the API key belongs to the merchant specified in the URL.
func (m *APIKeyAuthMiddleware) RequireMerchantOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, validate the API key
		m.RequireAPIKey()(c)
		if c.IsAborted() {
			return
		}

		// Get merchant ID from context
		contextMerchantID, exists := c.Get("merchant_id")
		if !exists {
			m.logger.Error("Merchant ID not found in context")
			c.JSON(
				http.StatusInternalServerError,
				createAuthErrorResponse("authentication_error", "MISSING_MERCHANT_ID", "Merchant ID not available"),
			)
			c.Abort()
			return
		}

		// Get merchant ID from URL parameter
		urlMerchantID := c.Param("merchant_id")
		if urlMerchantID == "" {
			// Try alternative parameter names
			urlMerchantID = c.Param("id")
		}

		if urlMerchantID == "" {
			m.logger.Debug("No merchant ID found in URL parameters")
			c.Next() // Allow if no merchant ID in URL
			return
		}

		// Check if API key's merchant matches URL merchant
		if contextMerchantIDStr, ok := contextMerchantID.(string); ok && contextMerchantIDStr != urlMerchantID {
			apiKeyID, _ := c.Get("api_key_id")
			if apiKeyIDStr, ok := apiKeyID.(string); ok {
				m.logger.Debug("API key merchant mismatch",
					zap.String("api_key_id", apiKeyIDStr),
					zap.String("api_key_merchant_id", contextMerchantIDStr),
					zap.String("url_merchant_id", urlMerchantID),
				)
			}
			c.JSON(
				http.StatusForbidden,
				createAuthErrorResponse(
					"authorization_error",
					"MERCHANT_MISMATCH",
					"API key does not belong to the specified merchant",
				),
			)
			c.Abort()
			return
		}

		m.logger.Debug("Merchant ownership check successful",
			zap.String("merchant_id", urlMerchantID),
		)

		c.Next()
	}
}

// GetAPIKeyInfo extracts API key information from the request context.
func GetAPIKeyInfo(c *gin.Context) (apiKeyID, merchantID string, permissions []string) {
	if val, exists := c.Get("api_key_id"); exists {
		if str, ok := val.(string); ok {
			apiKeyID = str
		}
	}
	if val, exists := c.Get("merchant_id"); exists {
		if str, ok := val.(string); ok {
			merchantID = str
		}
	}
	if val, exists := c.Get("api_key_permissions"); exists {
		if perms, ok := val.([]string); ok {
			permissions = perms
		}
	}

	return apiKeyID, merchantID, permissions
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
