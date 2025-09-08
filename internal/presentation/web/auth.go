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

// AuthMiddleware validates API key authentication for merchant endpoints.
func AuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Debug("Missing Authorization header")
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
			logger.Debug("Invalid Authorization header format", zap.String("header", authHeader))
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
			logger.Debug("Empty token in Authorization header")
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "EMPTY_TOKEN", "Token cannot be empty"),
			)
			c.Abort()
			return
		}

		// Validate token format (API.md specifies sk_live_* or sk_test_*)
		if !isValidAPIToken(token) {
			logger.Debug("Invalid API token format", zap.String("token", maskToken(token)))
			c.JSON(
				http.StatusUnauthorized,
				createAuthErrorResponse("authentication_error", "INVALID_TOKEN", "Invalid API key format"),
			)
			c.Abort()
			return
		}

		// For now, accept any valid format token (in production, validate against database)
		logger.Debug("Authentication successful", zap.String("token", maskToken(token)))
		c.Next()
	}
}

// createNotFoundErrorResponse creates a not found error response matching API.md format.
func createNotFoundErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Error:     "not_found",
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: generateRequestID(),
	}
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

	// Validate request
	if !h.validateTokenRequest(c, &req) {
		return
	}

	// Validate API key and permissions
	resp, ok := h.validateAPIKeyAndPermissions(c, &req)
	if !ok {
		return
	}

	// Generate JWT token
	token, err := h.generateJWTToken(req.APIKey, req.Scope, req.ExpiresIn)
	if err != nil {
		h.Logger.Error("Failed to generate JWT token", zap.Error(err))
		c.JSON(
			http.StatusInternalServerError,
			createAuthErrorResponse(
				"token_generation_error",
				"TOKEN_GENERATION_FAILED",
				"Failed to generate access token",
			),
		)
		return
	}

	response := TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   req.ExpiresIn,
		Scope:       req.Scope,
	}

	if resp.APIKey != nil {
		h.Logger.Debug("JWT token generated successfully",
			zap.String("api_key_id", resp.APIKey.ID()),
			zap.String("merchant_id", resp.APIKey.MerchantID()),
			zap.Strings("scope", req.Scope),
		)
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

// validateTokenRequest validates the token request parameters.
func (h *Handler) validateTokenRequest(c *gin.Context, req *TokenRequest) bool {
	// Validate grant type
	if req.GrantType != "api_key" {
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("grant_type must be 'api_key'", nil))
		return false
	}

	// Validate API key format
	if !isValidAPIToken(req.APIKey) {
		h.Logger.Debug("Invalid API key format", zap.String("token", maskToken(req.APIKey)))
		c.JSON(
			http.StatusUnauthorized,
			createAuthErrorResponse("authentication_error", "INVALID_API_KEY", "Invalid API key format"),
		)
		return false
	}

	// Validate scope
	if len(req.Scope) == 0 {
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("scope is required and cannot be empty", nil))
		return false
	}

	// Validate expires_in
	if req.ExpiresIn <= 0 || req.ExpiresIn > 86400 { // Max 24 hours
		c.JSON(
			http.StatusBadRequest,
			createValidationErrorResponse("expires_in must be between 1 and 86400 seconds", nil),
		)
		return false
	}

	return true
}

// validateAPIKeyAndPermissions validates the API key and checks permissions.
func (h *Handler) validateAPIKeyAndPermissions(
	c *gin.Context,
	req *TokenRequest,
) (*merchant.ValidateAPIKeyResponse, bool) {
	ctx := c.Request.Context()
	validateReq := &merchant.ValidateAPIKeyRequest{
		RawKey: req.APIKey,
	}

	resp, err := h.APIKeyService.ValidateAPIKey(ctx, validateReq)
	if err != nil {
		h.Logger.Debug("API key validation failed", zap.String("token", maskToken(req.APIKey)), zap.Error(err))
		c.JSON(
			http.StatusUnauthorized,
			createAuthErrorResponse("authentication_error", "INVALID_API_KEY", "Invalid or expired API key"),
		)
		return nil, false
	}

	if !resp.Valid {
		h.Logger.Debug("API key is invalid", zap.String("token", maskToken(req.APIKey)))
		c.JSON(
			http.StatusUnauthorized,
			createAuthErrorResponse("authentication_error", "INVALID_API_KEY", "Invalid or expired API key"),
		)
		return nil, false
	}

	// Check if API key has required permissions for the requested scope
	if resp.APIKey != nil {
		apiKeyPermissions := resp.APIKey.Permissions()
		for _, requestedScope := range req.Scope {
			hasPermission := false
			for _, permission := range apiKeyPermissions {
				if permission == requestedScope || permission == "*" {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				h.Logger.Debug("API key lacks required permission",
					zap.String("api_key_id", resp.APIKey.ID()),
					zap.String("requested_scope", requestedScope),
					zap.Strings("available_permissions", apiKeyPermissions),
				)
				c.JSON(
					http.StatusForbidden,
					createAuthErrorResponse(
						"authorization_error",
						"INSUFFICIENT_PERMISSIONS",
						"API key does not have required permission: "+requestedScope,
					),
				)
				return nil, false
			}
		}
	}

	return resp, true
}
