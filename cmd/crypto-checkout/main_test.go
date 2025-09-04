package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"crypto-checkout/internal/pkg/config"
)

func TestApplication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []fx.Option
		wantErr bool
	}{
		{
			name: "should start application successfully",
			options: []fx.Option{
				fx.Provide(func() *zap.Logger {
					return zaptest.NewLogger(t)
				}),
				fx.Invoke(func(log *zap.Logger) {
					log.Info("Application started successfully")
				}),
			},
			wantErr: false,
		},
		{
			name: "should handle application lifecycle",
			options: []fx.Option{
				fx.Provide(func() *zap.Logger {
					return zaptest.NewLogger(t)
				}),
				fx.Invoke(func(lc fx.Lifecycle, log *zap.Logger) {
					lc.Append(fx.Hook{
						OnStart: func(_ context.Context) error {
							log.Info("Starting application")
							return nil
						},
						OnStop: func(_ context.Context) error {
							log.Info("Stopping application")
							return nil
						},
					})
				}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fx.New(tt.options...)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := app.Start(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("App.Start() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				err = app.Stop(ctx)
				if err != nil {
					t.Errorf("App.Stop() error = %v", err)
				}
			}
		})
	}
}

func TestApplicationWithTestApp(t *testing.T) {
	t.Parallel()

	var logger *zap.Logger

	app := fxtest.New(t,
		fx.Provide(func() *zap.Logger {
			return zaptest.NewLogger(t)
		}),
		fx.Populate(&logger),
		fx.Invoke(func(log *zap.Logger) {
			log.Info("Test application started")
		}),
	)

	app.RequireStart()
	defer app.RequireStop()

	if logger == nil {
		t.Error("Expected logger to be populated")
	}
}

// TestValidateApp tests the application structure validation
// This is a smoke test that ensures the application can be created without errors.
func TestValidateApp(t *testing.T) {
	t.Parallel()

	// Create the application options
	appOptions := createAppOptions()

	// Validate the application structure
	err := fx.ValidateApp(appOptions...)
	require.NoError(t, err, "Application validation should pass")
}

// TestApplicationSmoke tests the complete application lifecycle
// This is a comprehensive smoke test following the patterns from the Kaspersky article.
func TestApplicationSmoke(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		testFunc    func(t *testing.T)
	}{
		{
			name:        "application_creation",
			description: "should create application without errors",
			testFunc:    testApplicationCreation,
		},
		{
			name:        "dependency_injection",
			description: "should resolve all dependencies correctly",
			testFunc:    testDependencyInjection,
		},
		{
			name:        "lifecycle_management",
			description: "should handle start/stop lifecycle correctly",
			testFunc:    testLifecycleManagement,
		},
		{
			name:        "graceful_shutdown",
			description: "should shutdown gracefully within timeout",
			testFunc:    testGracefulShutdown,
		},
		{
			name:        "concurrent_startup",
			description: "should handle concurrent startup scenarios",
			testFunc:    testConcurrentStartup,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("Running smoke test: %s - %s", tt.name, tt.description)
			tt.testFunc(t)
		})
	}
}

// testApplicationCreation verifies that the application can be created.
func testApplicationCreation(t *testing.T) {
	appOptions := createAppOptions()

	app := fx.New(appOptions...)
	require.NotNil(t, app, "Application should be created successfully")

	// Test that the application can be validated
	err := fx.ValidateApp(appOptions...)
	assert.NoError(t, err, "Application should be valid")
}

// testDependencyInjection verifies that all dependencies are resolved correctly.
func testDependencyInjection(t *testing.T) {
	var logger *zap.Logger
	var lifecycle fx.Lifecycle

	app := fxtest.New(t, append(createAppOptions(), fx.Populate(&logger, &lifecycle))...)
	app.RequireStart()
	defer app.RequireStop()

	// Verify dependencies are injected
	require.NotNil(t, logger, "Logger should not be nil")
	require.NotNil(t, lifecycle, "Lifecycle should not be nil")

	// Verify logger is functional
	logger.Info("Dependency injection test completed")
}

// testLifecycleManagement verifies start/stop lifecycle.
func testLifecycleManagement(t *testing.T) {
	startCalled := make(chan struct{})
	stopCalled := make(chan struct{})

	app := fxtest.New(t, append(createAppOptions(), fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				close(startCalled)
				return nil
			},
			OnStop: func(_ context.Context) error {
				close(stopCalled)
				return nil
			},
		})
	}))...)
	app.RequireStart()

	// Verify OnStart was called
	select {
	case <-startCalled:
		// Good, OnStart was called
	case <-time.After(1 * time.Second):
		t.Error("OnStart should be called within timeout")
	}

	// Verify OnStop is not called yet
	select {
	case <-stopCalled:
		t.Error("OnStop should not be called yet")
	default:
		// Good, OnStop not called yet
	}

	app.RequireStop()

	// Verify OnStop was called
	select {
	case <-stopCalled:
		// Good, OnStop was called
	case <-time.After(1 * time.Second):
		t.Error("OnStop should be called within timeout")
	}
}

// testGracefulShutdown verifies graceful shutdown within timeout.
func testGracefulShutdown(t *testing.T) {
	shutdownComplete := make(chan struct{})
	shutdownStarted := make(chan struct{})

	app := fxtest.New(t, append(createAppOptions(), fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				return nil
			},
			OnStop: func(ctx context.Context) error {
				close(shutdownStarted)

				// Simulate some cleanup work
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
				}

				close(shutdownComplete)
				return nil
			},
		})
	}))...)
	app.RequireStart()

	// Stop the application to trigger shutdown
	app.RequireStop()

	// Verify shutdown was initiated
	select {
	case <-shutdownStarted:
		// Good, shutdown was initiated
	case <-time.After(1 * time.Second):
		t.Error("Shutdown should be initiated within timeout")
	}

	// Verify shutdown completed
	select {
	case <-shutdownComplete:
		// Good, shutdown completed
	case <-time.After(1 * time.Second):
		t.Error("Shutdown should complete within timeout")
	}
}

// testConcurrentStartup verifies concurrent startup scenarios.
func testConcurrentStartup(t *testing.T) {
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	// Test concurrent application creation and validation
	for i := range numGoroutines {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					results <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			appOptions := createAppOptions()
			err := fx.ValidateApp(appOptions...)
			results <- err
		}(i)
	}

	// Collect results
	for range numGoroutines {
		select {
		case err := <-results:
			require.NoError(t, err, "Concurrent validation should not fail")
		case <-time.After(5 * time.Second):
			t.Error("Concurrent validation timed out")
		}
	}
}

// createAppOptions creates the application options for testing
// This mirrors the structure from the main application.
func createAppOptions() []fx.Option {
	return []fx.Option{
		fx.Provide(config.NewConfigProvider),
		fx.Provide(NewLogger),
		fx.Invoke(StartApplication),
	}
}

// TestApplicationIntegration tests the complete application integration.
func TestApplicationIntegration(t *testing.T) {
	t.Parallel()

	// Test with real application structure
	app := fxtest.New(t, createAppOptions()...)

	// Verify application starts
	app.RequireStart()

	// Verify application stops
	app.RequireStop()
}

// TestApplicationErrorHandling tests error scenarios.
func TestApplicationErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		options     []fx.Option
		expectError bool
	}{
		{
			name:        "valid_application",
			options:     createAppOptions(),
			expectError: false,
		},
		{
			name: "missing_dependency",
			options: []fx.Option{
				// No logger provided
				fx.Invoke(func(_ *zap.Logger) {
					// This should fail because no logger is provided
				}),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := fx.ValidateApp(tt.options...)
			if tt.expectError {
				assert.Error(t, err, "Expected validation to fail")
			} else {
				assert.NoError(t, err, "Expected validation to pass")
			}
		})
	}
}
