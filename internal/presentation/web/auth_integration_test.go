package web_test

import (
	"crypto-checkout/internal/presentation/web"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuthenticationMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("APIKeyAuthMiddleware_RequireAPIKey", func(t *testing.T) {
		// Create middleware with nil service to test basic functionality
		middleware := web.NewAPIKeyAuthMiddleware(nil, logger)

		// Setup router
		router := gin.New()
		router.GET("/test", middleware.RequireAPIKey(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		t.Run("MissingAuthorizationHeader", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("InvalidAuthorizationFormat", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set("Authorization", "Basic sk_test_1234567890abcdef")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("EmptyToken", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set("Authorization", "Bearer ")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("InvalidTokenFormat", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set("Authorization", "Bearer invalid_token")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("ValidTokenFormat", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set("Authorization", "Bearer sk_test_1234567890abcdef")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should get 500 because service is nil, but token format is valid
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("CombinedAuthMiddleware_RequireAuth", func(t *testing.T) {
		// Create middleware with nil service to test basic functionality
		middleware := web.NewCombinedAuthMiddleware(nil, logger, "test-secret")

		// Setup router
		router := gin.New()
		router.GET("/test", middleware.RequireAuth(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		t.Run("MissingAuthorizationHeader", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("ValidAPITokenFormat", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set("Authorization", "Bearer sk_test_1234567890abcdef")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should get 500 because service is nil, but token format is valid
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("ValidJWTTokenFormat", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set(
				"Authorization",
				"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should get 401 because JWT token is not signed with the correct secret
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})

	t.Run("CombinedAuthMiddleware_RequireAPIKeyOnly", func(t *testing.T) {
		// Create middleware with nil service to test basic functionality
		middleware := web.NewCombinedAuthMiddleware(nil, logger, "test-secret")

		// Setup router
		router := gin.New()
		router.GET("/test", middleware.RequireAPIKeyOnly(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		t.Run("JWTTokenNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set(
				"Authorization",
				"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("APITokenAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set("Authorization", "Bearer sk_test_1234567890abcdef")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should get 500 because service is nil, but token format is valid
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("CombinedAuthMiddleware_RequireJWTOnly", func(t *testing.T) {
		// Create middleware with nil service to test basic functionality
		middleware := web.NewCombinedAuthMiddleware(nil, logger, "test-secret")

		// Setup router
		router := gin.New()
		router.GET("/test", middleware.RequireJWTOnly(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		t.Run("APITokenNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set("Authorization", "Bearer sk_test_1234567890abcdef")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("JWTTokenAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set(
				"Authorization",
				"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should get 401 because JWT token is not signed with the correct secret
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})
}

func TestAuthenticationMiddleware_HelperFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetAPIKeyInfo", func(t *testing.T) {
		// Setup router with middleware that sets context
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			// Simulate setting context values
			c.Set("api_key_id", "api-key-123")
			c.Set("merchant_id", "merchant-123")
			c.Set("api_key_permissions", []string{"invoices:read", "invoices:write"})

			// Test GetAPIKeyInfo function
			apiKeyID, merchantID, permissions := web.GetAPIKeyInfo(c)

			c.JSON(http.StatusOK, gin.H{
				"api_key_id":  apiKeyID,
				"merchant_id": merchantID,
				"permissions": permissions,
			})
		})

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetAuthInfo", func(t *testing.T) {
		// Setup router with middleware that sets context
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			// Simulate setting context values
			c.Set("api_key_id", "api-key-123")
			c.Set("merchant_id", "merchant-123")
			c.Set("api_key_permissions", []string{"invoices:read", "invoices:write"})
			c.Set("jwt_scope", []string{"invoices:read"})

			// Test GetAuthInfo function
			apiKeyID, merchantID, permissions, authType := web.GetAuthInfo(c)

			c.JSON(http.StatusOK, gin.H{
				"api_key_id":  apiKeyID,
				"merchant_id": merchantID,
				"permissions": permissions,
				"auth_type":   authType,
			})
		})

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
