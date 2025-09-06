package testutil

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/infrastructure/database"
	"crypto-checkout/internal/presentation/web"
	"crypto-checkout/pkg/config"
)

const (
	contextTimeout = 15 * time.Second
)

// StartTestApp starts the app on a random free port.
func StartTestApp(t *testing.T, opts ...fx.Option) string {
	t.Helper()

	// Pre-bind to :0 to get a free port
	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to allocate port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close() // Close immediately after getting the port

	var srv *http.Server
	opts = append(opts,
		fx.Decorate(func(s *http.Server) *http.Server {
			s.Addr = addr // use reserved port
			return s
		}),
		fx.Populate(&srv),
	)
	app := fx.New(opts...)

	startCtx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()
	if startErr := app.Start(startCtx); startErr != nil {
		t.Fatalf("failed to start app: %v", startErr)
	}
	t.Cleanup(func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), contextTimeout)
		defer stopCancel()
		_ = app.Stop(stopCtx)
	})

	return "http://" + addr
}

// CreateTestLogger creates a test logger.
func CreateTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// SetupTestApp sets up the test application with common configuration.
func SetupTestApp(t *testing.T) string {
	t.Helper()

	return StartTestApp(t,
		// Supply test configuration
		fx.Supply(&config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 0, // Will be set by testutil
			},
			Database: config.DatabaseConfig{
				URL: "file::memory:?cache=shared", // In-memory SQLite for testing
			},
			Log: config.LogConfig{
				Level: "debug",
				Dir:   "logs",
			},
		}),
		// Supply test logger
		fx.Supply(CreateTestLogger()),
		// Provide all dependencies
		database.Module,
		invoice.Module,
		payment.Module,
		web.Module,
		// Set Gin to test mode
		fx.Invoke(func() {
			gin.SetMode(gin.TestMode)
		}),
	)
}
