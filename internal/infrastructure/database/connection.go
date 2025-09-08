// Package database provides database connection and migration functionality for the crypto-checkout application.
package database

import (
	"crypto-checkout/pkg/config"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connection represents a database connection.
type Connection struct {
	DB     *gorm.DB
	Logger *zap.Logger
}

// NewConnection creates a new database connection.
func NewConnection(cfg config.DatabaseConfig, logger *zap.Logger) (*Connection, error) {
	var db *gorm.DB
	var err error

	// Check if URL is provided (for SQLite or other databases)
	if cfg.URL != "" {
		switch {
		case strings.HasPrefix(cfg.URL, "sqlite://"):
			// SQLite connection
			dbPath := strings.TrimPrefix(cfg.URL, "sqlite://")
			db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
				Logger: gormlogger.Default.LogMode(gormlogger.Silent), // Silent for tests
				NowFunc: func() time.Time {
					return time.Now().UTC()
				},
			})
		case strings.HasPrefix(cfg.URL, "file::memory:"):
			// In-memory SQLite connection
			db, err = gorm.Open(sqlite.Open(cfg.URL), &gorm.Config{
				Logger: gormlogger.Default.LogMode(gormlogger.Silent), // Silent for tests
				NowFunc: func() time.Time {
					return time.Now().UTC()
				},
			})
		default:
			// Other database URLs (PostgreSQL, etc.)
			db, err = gorm.Open(postgres.Open(cfg.URL), &gorm.Config{
				Logger: gormlogger.Default.LogMode(gormlogger.Info),
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
			Logger: gormlogger.Default.LogMode(gormlogger.Info),
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

	return &Connection{DB: db, Logger: logger}, nil
}

// Migrate runs database migrations.
func (c *Connection) Migrate() error {
	c.Logger.Info("Starting database migration")

	// Handle existing data before running AutoMigrate
	if err := c.migrateExistingData(); err != nil {
		c.Logger.Error("Failed to migrate existing data", zap.Error(err))
		return fmt.Errorf("failed to migrate existing data: %w", err)
	}

	// Run GORM AutoMigrate
	c.Logger.Info("Running GORM AutoMigrate")
	if err := c.DB.AutoMigrate(
		&InvoiceModel{},
		&PaymentModel{},
	); err != nil {
		c.Logger.Error("Failed to run GORM AutoMigrate", zap.Error(err))
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	c.Logger.Info("Database migration completed successfully")
	return nil
}

// migrateExistingData handles migration of existing data before schema changes.
func (c *Connection) migrateExistingData() error {
	c.Logger.Info("Checking for existing data migration needs")

	// Check if invoices table exists and has data
	if c.DB.Migrator().HasTable(&InvoiceModel{}) {
		var count int64
		if err := c.DB.Raw("SELECT COUNT(*) FROM invoices").Scan(&count).Error; err != nil {
			c.Logger.Error("Failed to count existing invoices", zap.Error(err))
			return fmt.Errorf("failed to count existing invoices: %w", err)
		}

		if count > 0 {
			c.Logger.Info("Found existing invoices", zap.Int64("count", count))
			c.Logger.Warn("Migration may fail if existing data has NULL values for new NOT NULL columns")
		} else {
			c.Logger.Info("No existing invoices found")
		}
	} else {
		c.Logger.Info("Invoices table does not exist yet")
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
