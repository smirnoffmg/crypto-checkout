package database

import (
	"context"
	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/merchant"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/pkg/config"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Module provides database-related dependencies for Fx.
var Module = fx.Module("database",
	fx.Provide(
		NewDatabaseConnection,
		NewGormDBProvider,
		NewInvoiceRepositoryProvider,
		NewPaymentRepositoryProvider,
		NewMerchantRepositoryProvider,
		NewAPIKeyRepositoryProvider,
		NewWebhookEndpointRepositoryProvider,
	),
	fx.Invoke(InitializeDatabase),
)

// NewGormDBProvider provides a *gorm.DB from the database connection.
func NewGormDBProvider(conn *Connection) *gorm.DB {
	return conn.DB
}

// NewDatabaseConnection creates a new database connection.
func NewDatabaseConnection(cfg *config.Config, logger *zap.Logger) (*Connection, error) {
	logger.Info("Connecting to database",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.DBName),
	)

	dbConfig := config.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
		URL:      cfg.Database.URL, // Include the URL field
	}

	conn, err := NewConnection(dbConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	logger.Info("Successfully connected to database")
	return conn, nil
}

// NewInvoiceRepositoryProvider creates a new invoice repository.
func NewInvoiceRepositoryProvider(conn *Connection) invoice.Repository {
	return NewInvoiceRepository(conn.DB)
}

// NewPaymentRepositoryProvider creates a new payment repository.
func NewPaymentRepositoryProvider(conn *Connection) payment.Repository {
	return NewPaymentRepository(conn.DB)
}

// NewMerchantRepositoryProvider creates a new merchant repository.
func NewMerchantRepositoryProvider(conn *Connection, logger *zap.Logger) merchant.MerchantRepository {
	return NewMerchantRepository(conn.DB, logger)
}

// NewAPIKeyRepositoryProvider creates a new API key repository.
func NewAPIKeyRepositoryProvider(conn *Connection, logger *zap.Logger) merchant.APIKeyRepository {
	return NewAPIKeyRepository(conn.DB, logger)
}

// NewWebhookEndpointRepositoryProvider creates a new webhook endpoint repository.
func NewWebhookEndpointRepositoryProvider(conn *Connection, logger *zap.Logger) merchant.WebhookEndpointRepository {
	return NewWebhookEndpointRepository(conn.DB, logger)
}

// InitializeDatabase initializes the database with migrations.
func InitializeDatabase(conn *Connection, logger *zap.Logger, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("Running database migrations")
			if err := conn.Migrate(); err != nil {
				logger.Error("Failed to run database migrations", zap.Error(err))
				return fmt.Errorf("failed to run database migrations: %w", err)
			}
			logger.Info("Database migrations completed successfully")
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info("Closing database connection")
			if err := conn.Close(); err != nil {
				logger.Error("Failed to close database connection", zap.Error(err))
				return fmt.Errorf("failed to close database connection: %w", err)
			}
			logger.Info("Database connection closed")
			return nil
		},
	})
}
