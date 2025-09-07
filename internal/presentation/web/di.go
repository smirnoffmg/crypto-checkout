// Package api provides the API layer setup and HTTP server configuration.
package web

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/pkg/config"
)

const (
	// DebugLogLevel is the debug log level string.
	DebugLogLevel = "debug"
)

//go:embed templates/*
var templatesFS embed.FS

// Module provides the API module for Fx dependency injection.
var Module = fx.Module("api",
	fx.Provide(
		NewGinEngine,
		NewWebSocketHub,
		NewAPIHandler,
		NewHTTPServer,
	),
	fx.Invoke(RegisterRoutes),
)

// setupGinLogging configures Gin to write logs to stdout only.
func setupGinLogging(cfg *config.Config, logger *zap.Logger) {
	// Use stdout for all Gin logging - no file logging
	// This avoids file system issues and race conditions in tests
	// Note: We're setting gin.DefaultWriter which is the intended way to configure Gin's output
	gin.DefaultWriter = os.Stdout //nolint:reassign // This is the intended way to configure Gin's output

	// Enable console color for better readability
	gin.ForceConsoleColor()

	logger.Info("Gin logging configured for stdout",
		zap.String("mode", cfg.Log.Level),
	)
}

// ErrorHandler captures errors and returns a detailed JSON error response.
func ErrorHandler(cfg *config.Config, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Process the request first

		// Debug: Log if middleware is being called
		logger.Debug("ErrorHandler middleware called",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Int("error_count", len(c.Errors)),
			zap.Int("status", c.Writer.Status()),
		)

		// Check if any errors were added to the context
		if len(c.Errors) > 0 {
			// Use the last error
			err := c.Errors.Last().Err
			statusCode := http.StatusInternalServerError
			errorMessage := "An unexpected error occurred"
			errorCode := "INTERNAL_SERVER_ERROR"

			// Map specific errors to HTTP status codes and messages
			switch {
			case errors.Is(err, invoice.ErrInvalidRequest):
				statusCode = http.StatusBadRequest
				errorMessage = err.Error()
				errorCode = "BAD_REQUEST"
			case errors.Is(err, invoice.ErrInvalidUnitPrice):
				statusCode = http.StatusBadRequest
				errorMessage = err.Error()
				errorCode = "INVALID_UNIT_PRICE"
			case errors.Is(err, invoice.ErrInvoiceNotFound), errors.Is(err, invoice.ErrNotFound):
				statusCode = http.StatusNotFound
				errorMessage = err.Error()
				errorCode = "NOT_FOUND"
			case errors.Is(err, invoice.ErrServiceError):
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to process invoice"
				errorCode = "INVOICE_SERVICE_ERROR"
			case errors.Is(err, invoice.ErrInvalidPaymentAddress):
				statusCode = http.StatusBadRequest
				errorMessage = err.Error()
				errorCode = "INVALID_PAYMENT_ADDRESS"
			case errors.Is(err, payment.ErrServiceError):
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to process payment"
				errorCode = "PAYMENT_SERVICE_ERROR"
			case errors.Is(err, payment.ErrPaymentNotFound):
				statusCode = http.StatusNotFound
				errorMessage = err.Error()
				errorCode = "PAYMENT_NOT_FOUND"
			default:
				// Check for common HTTP errors by error message
				errorMsg := err.Error()
				switch {
				case strings.Contains(errorMsg, "invalid character") && strings.Contains(errorMsg, "looking for beginning of value"):
					statusCode = http.StatusBadRequest
					errorMessage = "Invalid JSON format"
					errorCode = "INVALID_JSON"
				case strings.Contains(errorMsg, "EOF"):
					statusCode = http.StatusBadRequest
					errorMessage = "Empty request body"
					errorCode = "EMPTY_BODY"
				case strings.Contains(errorMsg, "not found"):
					statusCode = http.StatusNotFound
					errorMessage = err.Error()
					errorCode = "NOT_FOUND"
				default:
					// Log unexpected errors with stack trace
					logger.Error("Unhandled API error",
						zap.Error(err),
						zap.String("path", c.Request.URL.Path),
						zap.String("method", c.Request.Method),
						zap.Stack("stack_trace"),
					)
				}
			}

			// Create a detailed error response
			errorResponse := (&Handler{Logger: logger, config: cfg}).createErrorResponse(errorCode, errorMessage, err)
			c.AbortWithStatusJSON(statusCode, errorResponse)
			return
		}
	}
}

// NewWebSocketHub creates a new WebSocket hub.
func NewWebSocketHub(logger *zap.Logger) *Hub {
	hub := NewHub(logger)
	go hub.Run()
	return hub
}

// NewAPIHandler creates a new API handler with services.
func NewAPIHandler(
	invoiceService invoice.InvoiceService,
	paymentService payment.PaymentService,
	logger *zap.Logger,
	cfg *config.Config,
	hub *Hub,
) *Handler {
	return NewHandler(invoiceService, paymentService, logger, cfg, hub)
}

const (
	// HTTP timeouts.
	readTimeout     = 15 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 30 * time.Second
)

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(cfg *config.Config) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

// RegisterRoutes registers all API routes with the Gin router.
func RegisterRoutes(
	lc fx.Lifecycle,
	router *gin.Engine,
	handler *Handler,
	server *http.Server,
	logger *zap.Logger,
) {
	// Register API routes
	handler.RegisterRoutes(router)

	// Set the Gin router as the server handler
	server.Handler = router

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("Starting HTTP server", zap.String("addr", server.Addr))
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("HTTP server failed to start", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping HTTP server")
			shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
			defer cancel()
			return server.Shutdown(shutdownCtx)
		},
	})
}
