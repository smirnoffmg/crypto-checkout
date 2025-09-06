package database

import (
	"context"
	"fmt"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/pkg/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides database-related dependencies for Fx.
var Module = fx.Module("database",
	fx.Provide(
		NewDatabaseConnection,
		NewInvoiceRepositoryProvider,
		NewPaymentRepositoryProvider,
	),
	fx.Invoke(InitializeDatabase),
)

// NewDatabaseConnection creates a new database connection.
func NewDatabaseConnection(cfg *config.Config, logger *zap.Logger) (*Connection, error) {
	logger.Info("Connecting to database",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.DBName),
	)

	dbConfig := Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
		URL:      cfg.Database.URL, // Include the URL field
	}

	conn, err := NewConnection(dbConfig)
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
