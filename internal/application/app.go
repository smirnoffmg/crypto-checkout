package application

import (
	"context"
	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/merchant"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/infrastructure/database"
	"crypto-checkout/internal/infrastructure/events"
	"crypto-checkout/internal/presentation/web"
	"crypto-checkout/pkg/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func GetApp() *fx.App {
	return fx.New(
		fx.Provide(config.NewConfigProvider),
		fx.Provide(NewLogger),
		database.Module,
		events.Module,
		invoice.Module,
		merchant.Module,
		payment.Module,
		web.Module,
		fx.Invoke(StartApplication),
		fx.Invoke(func(log *zap.Logger, graph fx.DotGraph) {
			log.Info("Application modules loaded",
				zap.String("database_module", "database"),
				zap.String("events_module", "events"),
				zap.String("invoice_module", "invoice-service"),
				zap.String("merchant_module", "merchant-service"),
				zap.String("payment_module", "payment-service"),
				zap.String("web_module", "api"))

			// Print dependency graph
			log.Info("Dependency graph", zap.String("graph", string(graph)))
		}),
	)
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
