// Package main is the entry point for the crypto-checkout application.
package main

import (
	"context"
	"flag"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"crypto-checkout/internal/pkg/config"
)

func main() {
	// Parse command line flags
	healthCheck := flag.Bool("health-check", false, "Run health check and exit")
	flag.Parse()

	if *healthCheck {
		// Simple health check - just exit with 0
		os.Exit(0)
	}

	fx.New(
		fx.Provide(config.NewConfigProvider),
		fx.Provide(NewLogger),
		fx.Invoke(StartApplication),
	).Run()
}

// NewLogger creates a new logger based on configuration.
func NewLogger(cfg *config.Config) *zap.Logger {
	var logger *zap.Logger
	var err error

	switch cfg.Log.Level {
	case "debug":
		logger, err = zap.NewDevelopment()
	default:
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic(err)
	}
	return logger
}

// StartApplication starts the application with lifecycle management.
func StartApplication(lc fx.Lifecycle, log *zap.Logger, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info("Starting crypto-checkout application",
				zap.String("host", cfg.Server.Host),
				zap.Int("port", cfg.Server.Port),
				zap.String("log_level", cfg.Log.Level))
			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Info("Stopping crypto-checkout application")
			return nil
		},
	})
}
