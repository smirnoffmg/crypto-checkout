package web

import (
	"crypto-checkout/internal/domain/merchant"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTAuthMiddleware validates JWT token authentication.
type JWTAuthMiddleware struct {
	apiKeyService merchant.APIKeyService
	logger        *zap.Logger
	secret        string
}

// NewJWTAuthMiddleware creates a new JWT authentication middleware.
func NewJWTAuthMiddleware(apiKeyService merchant.APIKeyService, logger *zap.Logger, secret string) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		apiKeyService: apiKeyService,
		logger:        logger,
		secret:        secret,
	}
}

// RequireJWT validates JWT token authentication for protected endpoints.
func (m *JWTAuthMiddleware) RequireJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract and validate token
		tokenString, ok := m.extractTokenFromHeader(c)
		if !ok {
			return
		}

		// Parse and validate JWT token
		claims, ok := m.parseAndValidateJWT(c, tokenString)
		if !ok {
			return
		}

		// Validate grant type
		grantType, ok := claims["grant_type"].(string)
		if !ok || grantType != "api_key" {
			m.logger.Debug("Invalid grant type in JWT", zap.String("grant_type", grantType))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "INVALID_GRANT_TYPE", "Invalid grant type"),
			)
			c.Abort()
			return
		}

		// Extract API key from claims
		apiKey, ok := claims["api_key"].(string)
		if !ok {
			m.logger.Debug("API key not found in JWT claims")
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "MISSING_API_KEY", "API key not found in token"),
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
			RawKey: apiKey,
		}

		resp, err := m.apiKeyService.ValidateAPIKey(ctx, req)
		if err != nil {
			m.logger.Debug("API key validation failed", zap.String("token", maskToken(apiKey)), zap.Error(err))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "INVALID_API_KEY", "Invalid or expired API key"),
			)
			c.Abort()
			return
		}

		if !resp.Valid {
			m.logger.Debug("API key is invalid", zap.String("token", maskToken(apiKey)))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "INVALID_API_KEY", "Invalid or expired API key"),
			)
			c.Abort()
			return
		}

		// Extract scope from claims
		scope, ok := claims["scope"].([]interface{})
		if !ok {
			m.logger.Debug("Scope not found in JWT claims")
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "MISSING_SCOPE", "Scope not found in token"),
			)
			c.Abort()
			return
		}

		// Convert scope to string slice
		scopeStrings := make([]string, len(scope))
		for i, s := range scope {
			if str, ok := s.(string); ok {
				scopeStrings[i] = str
			} else {
				m.logger.Debug("Invalid scope type in JWT claims")
				c.JSON(
					http.StatusUnauthorized,
					createAuthErrorResponse("authentication_error", "INVALID_SCOPE", "Invalid scope format"),
				)
				c.Abort()
				return
			}
		}

		// Store information in context for use by handlers
		if resp.APIKey != nil {
			c.Set("api_key_id", resp.APIKey.ID())
			c.Set("merchant_id", resp.APIKey.MerchantID())
			c.Set("api_key_permissions", resp.APIKey.Permissions())
			c.Set("jwt_scope", scopeStrings)
			c.Set("jwt_expires_at", claims["exp"])

			m.logger.Debug("JWT authentication successful",
				zap.String("api_key_id", resp.APIKey.ID()),
				zap.String("merchant_id", resp.APIKey.MerchantID()),
				zap.Strings("scope", scopeStrings),
			)
		}

		c.Next()
	}
}

// RequireJWTPermission validates JWT token authentication and checks for specific permission.
func (m *JWTAuthMiddleware) RequireJWTPermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, validate the JWT token
		m.RequireJWT()(c)
		if c.IsAborted() {
			return
		}

		// Get scope from context
		scope, exists := c.Get("jwt_scope")
		if !exists {
			m.logger.Error("JWT scope not found in context")
			c.JSON(
				http.StatusInternalServerError,
				createAuthErrorResponse("authentication_error", "MISSING_SCOPE", "JWT scope not available"),
			)
			c.Abort()
			return
		}

		scopeList, ok := scope.([]string)
		if !ok {
			m.logger.Error("Invalid scope type in context")
			c.JSON(
				http.StatusInternalServerError,
				createAuthErrorResponse("authentication_error", "INVALID_SCOPE", "Invalid scope format"),
			)
			c.Abort()
			return
		}

		// Check if JWT has required permission in scope
		hasPermission := false
		for _, permission := range scopeList {
			if permission == requiredPermission || permission == "*" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			apiKeyID, _ := c.Get("api_key_id")
			if apiKeyIDStr, ok := apiKeyID.(string); ok {
				m.logger.Debug("JWT lacks required permission",
					zap.String("api_key_id", apiKeyIDStr),
					zap.String("required_permission", requiredPermission),
					zap.Strings("available_scope", scopeList),
				)
			}
			c.JSON(
				http.StatusForbidden,
				createAuthErrorResponse(
					"authorization_error",
					"INSUFFICIENT_PERMISSIONS",
					"JWT token does not have required permission: "+requiredPermission,
				),
			)
			c.Abort()
			return
		}

		m.logger.Debug("JWT permission check successful",
			zap.String("required_permission", requiredPermission),
		)

		c.Next()
	}
}

// GenerateJWTToken creates a JWT token with the specified claims.
func (m *JWTAuthMiddleware) GenerateJWTToken(apiKey string, scope []string, expiresIn int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"grant_type": "api_key",
		"api_key":    apiKey,
		"scope":      scope,
		"iat":        now.Unix(),
		"exp":        now.Add(time.Duration(expiresIn) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

// extractTokenFromHeader extracts and validates the JWT token from the Authorization header.
func (m *JWTAuthMiddleware) extractTokenFromHeader(c *gin.Context) (string, bool) {
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
		return "", false
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
		return "", false
	}

	// Extract the token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		m.logger.Debug("Empty token in Authorization header")
		c.JSON(
			http.StatusUnauthorized,
			createAuthErrorResponse("authentication_error", "EMPTY_TOKEN", "Token cannot be empty"),
		)
		c.Abort()
		return "", false
	}

	return tokenString, true
}

// parseAndValidateJWT parses and validates the JWT token.
func (m *JWTAuthMiddleware) parseAndValidateJWT(c *gin.Context, tokenString string) (jwt.MapClaims, bool) {
	// Parse and validate JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.secret), nil
	})
	if err != nil {
		m.logger.Debug("JWT token parsing failed", zap.Error(err))
		c.JSON(
			http.StatusUnauthorized,
			createAuthErrorResponse("authentication_error", "INVALID_JWT", "Invalid or expired JWT token"),
		)
		c.Abort()
		return nil, false
	}

	if !token.Valid {
		m.logger.Debug("JWT token is invalid")
		c.JSON(
			http.StatusUnauthorized,
			createAuthErrorResponse("authentication_error", "INVALID_JWT", "Invalid or expired JWT token"),
		)
		c.Abort()
		return nil, false
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		m.logger.Debug("Invalid JWT claims")
		c.JSON(
			http.StatusUnauthorized,
			createAuthErrorResponse("authentication_error", "INVALID_JWT_CLAIMS", "Invalid JWT token claims"),
		)
		c.Abort()
		return nil, false
	}

	return claims, true
}
