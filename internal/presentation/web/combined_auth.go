package web

import (
	"crypto-checkout/internal/domain/merchant"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CombinedAuthMiddleware provides flexible authentication supporting both API keys and JWT tokens.
type CombinedAuthMiddleware struct {
	apiKeyAuth *APIKeyAuthMiddleware
	jwtAuth    *JWTAuthMiddleware
	logger     *zap.Logger
}

// NewCombinedAuthMiddleware creates a new combined authentication middleware.
func NewCombinedAuthMiddleware(
	apiKeyService merchant.APIKeyService,
	logger *zap.Logger,
	jwtSecret string,
) *CombinedAuthMiddleware {
	return &CombinedAuthMiddleware{
		apiKeyAuth: NewAPIKeyAuthMiddleware(apiKeyService, logger),
		jwtAuth:    NewJWTAuthMiddleware(apiKeyService, logger, jwtSecret),
		logger:     logger,
	}
}

// RequireAuth validates either API key or JWT token authentication.
func (m *CombinedAuthMiddleware) RequireAuth() gin.HandlerFunc {
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

		// Determine token type and validate accordingly
		if isValidAPIToken(token) {
			// API key authentication
			m.logger.Debug("Using API key authentication", zap.String("token", maskToken(token)))
			m.apiKeyAuth.RequireAPIKey()(c)
		} else {
			// JWT token authentication
			m.logger.Debug("Using JWT token authentication", zap.String("token", maskToken(token)))
			m.jwtAuth.RequireJWT()(c)
		}
	}
}

// RequirePermission validates authentication and checks for specific permission.
func (m *CombinedAuthMiddleware) RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, validate authentication
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Get permissions from context (could be from API key or JWT)
		permissions, exists := c.Get("api_key_permissions")
		if !exists {
			// Try to get scope from JWT
			scope, scopeExists := c.Get("jwt_scope")
			if !scopeExists {
				m.logger.Error("No permissions or scope found in context")
				c.JSON(
					http.StatusInternalServerError,
					createAuthErrorResponse("authentication_error", "MISSING_PERMISSIONS", "Permissions not available"),
				)
				c.Abort()
				return
			}

			// Use JWT scope as permissions
			permissions = scope
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

		// Check if token has required permission
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
				m.logger.Debug("Token lacks required permission",
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
					"Token does not have required permission: "+requiredPermission,
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

// RequireMerchantOwnership validates authentication and merchant ownership.
func (m *CombinedAuthMiddleware) RequireMerchantOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, validate authentication
		m.RequireAuth()(c)
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

		// Check if token's merchant matches URL merchant
		if contextMerchantIDStr, ok := contextMerchantID.(string); ok && contextMerchantIDStr != urlMerchantID {
			apiKeyID, _ := c.Get("api_key_id")
			if apiKeyIDStr, ok := apiKeyID.(string); ok {
				m.logger.Debug("Token merchant mismatch",
					zap.String("api_key_id", apiKeyIDStr),
					zap.String("token_merchant_id", contextMerchantIDStr),
					zap.String("url_merchant_id", urlMerchantID),
				)
			}
			c.JSON(
				http.StatusForbidden,
				createAuthErrorResponse(
					"authorization_error",
					"MERCHANT_MISMATCH",
					"Token does not belong to the specified merchant",
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

// RequireAPIKeyOnly validates only API key authentication (no JWT).
func (m *CombinedAuthMiddleware) RequireAPIKeyOnly() gin.HandlerFunc {
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

		// Only allow API key authentication
		if !isValidAPIToken(token) {
			m.logger.Debug("JWT token not allowed for this endpoint", zap.String("token", maskToken(token)))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse(
					"authentication_error",
					"INVALID_TOKEN_TYPE",
					"Only API key authentication is allowed",
				),
			)
			c.Abort()
			return
		}

		// Use API key authentication
		m.apiKeyAuth.RequireAPIKey()(c)
	}
}

// RequireJWTOnly validates only JWT token authentication (no API keys).
func (m *CombinedAuthMiddleware) RequireJWTOnly() gin.HandlerFunc {
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

		// Only allow JWT token authentication
		if isValidAPIToken(token) {
			m.logger.Debug("API key not allowed for this endpoint", zap.String("token", maskToken(token)))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse(
					"authentication_error",
					"INVALID_TOKEN_TYPE",
					"Only JWT token authentication is allowed",
				),
			)
			c.Abort()
			return
		}

		// Use JWT authentication
		m.jwtAuth.RequireJWT()(c)
	}
}

// GetAuthInfo extracts authentication information from the request context.
func GetAuthInfo(c *gin.Context) (apiKeyID, merchantID string, permissions []string, authType string) {
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

	// Determine authentication type
	if _, hasScope := c.Get("jwt_scope"); hasScope {
		authType = "jwt"
	} else {
		authType = "api_key"
	}

	return apiKeyID, merchantID, permissions, authType
}
