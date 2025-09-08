// Package database provides database connection and migration functionality for the crypto-checkout application.
package database

import (
	"crypto-checkout/pkg/config"
	"fmt"
	"strings"
	"time"

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
func NewConnection(cfg config.DatabaseConfig) (*Connection, error) {
	var db *gorm.DB
	var err error

	// Check if URL is provided (for SQLite or other databases)
	if cfg.URL != "" {
		switch {
		case strings.HasPrefix(cfg.URL, "sqlite://"):
			// SQLite connection
			dbPath := strings.TrimPrefix(cfg.URL, "sqlite://")
			db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent), // Silent for tests
				NowFunc: func() time.Time {
					return time.Now().UTC()
				},
			})
		case strings.HasPrefix(cfg.URL, "file::memory:"):
			// In-memory SQLite connection
			db, err = gorm.Open(sqlite.Open(cfg.URL), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent), // Silent for tests
				NowFunc: func() time.Time {
					return time.Now().UTC()
				},
			})
		default:
			// Other database URLs (PostgreSQL, etc.)
			db, err = gorm.Open(postgres.Open(cfg.URL), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Info),
				NowFunc: func() time.Time {
					return time.Now().UTC()
				},
			})
		}
	} else {
		// PostgreSQL connection using individual config fields
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

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
	if cfg.URL != "" &&
		(strings.HasPrefix(cfg.URL, "sqlite://") || strings.HasPrefix(cfg.URL, "file::memory:")) {
		// SQLite connection pool settings - optimized for concurrency
		// SQLite can handle multiple readers but limited concurrent writers
		sqlDB.SetMaxIdleConns(5)    // Allow more idle connections for better concurrency
		sqlDB.SetMaxOpenConns(10)   // Allow more open connections for concurrent reads
		sqlDB.SetConnMaxLifetime(0) // Keep connections alive

		// Enable WAL mode for better concurrency (if not in-memory)
		if !strings.HasPrefix(cfg.URL, "file::memory:") {
			// Enable WAL mode for better concurrency
			db.Exec("PRAGMA journal_mode=WAL")
			db.Exec("PRAGMA synchronous=NORMAL")
			db.Exec("PRAGMA cache_size=1000")
			db.Exec("PRAGMA temp_store=memory")
		}
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
