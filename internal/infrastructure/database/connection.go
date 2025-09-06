// Package database provides database connection and migration functionality for the crypto-checkout application.
package database

import (
	"fmt"
	"strings"
	"time"

	"crypto-checkout/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connection represents a database connection.
type Connection struct {
	DB *gorm.DB
}

// NewConnection creates a new database connection.
func NewConnection(config config.DatabaseConfig) (*Connection, error) {
	var db *gorm.DB
	var err error

	// Check if URL is provided (for SQLite or other databases)
	if config.URL != "" {
		switch {
		case strings.HasPrefix(config.URL, "sqlite://"):
			// SQLite connection
			dbPath := strings.TrimPrefix(config.URL, "sqlite://")
			db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent), // Silent for tests
				NowFunc: func() time.Time {
					return time.Now().UTC()
				},
			})
		case strings.HasPrefix(config.URL, "file::memory:"):
			// In-memory SQLite connection
			db, err = gorm.Open(sqlite.Open(config.URL), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent), // Silent for tests
				NowFunc: func() time.Time {
					return time.Now().UTC()
				},
			})
		default:
			// Other database URLs (PostgreSQL, etc.)
			db, err = gorm.Open(postgres.Open(config.URL), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Info),
				NowFunc: func() time.Time {
					return time.Now().UTC()
				},
			})
		}
	} else {
		// PostgreSQL connection using individual config fields
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Different connection pool settings for SQLite vs PostgreSQL
	if config.URL != "" && strings.HasPrefix(config.URL, "sqlite://") {
		// SQLite connection pool settings
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetConnMaxLifetime(0)
	} else {
		// PostgreSQL connection pool settings
		const (
			maxIdleConns = 10
			maxOpenConns = 100
		)
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return &Connection{DB: db}, nil
}

// Migrate runs database migrations.
func (c *Connection) Migrate() error {
	if err := c.DB.AutoMigrate(
		&InvoiceModel{},
		&PaymentModel{},
	); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}
	return nil
}

// Close closes the database connection.
func (c *Connection) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	if closeErr := sqlDB.Close(); closeErr != nil {
		return fmt.Errorf("failed to close database connection: %w", closeErr)
	}
	return nil
}
