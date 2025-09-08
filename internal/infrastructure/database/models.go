package database

import (
	"time"

	"gorm.io/gorm"
)

// InvoiceModel represents the database model for invoices.
type InvoiceModel struct {
	ID               string  `gorm:"primaryKey;type:uuid"`
	MerchantID       string  `gorm:"type:uuid;not null;index"`
	CustomerID       *string `gorm:"type:uuid;index"` // Made optional to match domain model
	Title            string  `gorm:"type:varchar(255);not null"`
	Description      string  `gorm:"type:text"`
	Items            string  `gorm:"type:jsonb"` // Store items as JSONB as per DB.md
	Subtotal         string  `gorm:"type:decimal(20,2);not null"`
	Tax              string  `gorm:"type:decimal(20,2);not null;default:0"`
	Total            string  `gorm:"type:decimal(20,2);not null"`
	Currency         string  `gorm:"type:varchar(3);not null"`
	CryptoCurrency   string  `gorm:"type:varchar(10);not null"`
	CryptoAmount     string  `gorm:"type:decimal(20,8);not null"`
	PaymentAddress   *string `gorm:"type:varchar(42)"`
	Status           string  `gorm:"type:varchar(20);not null"`
	ExchangeRate     string  `gorm:"type:jsonb"`
	PaymentTolerance string  `gorm:"type:jsonb"`
	ExpiresAt        *time.Time
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
	PaidAt           *time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the InvoiceModel.
func (InvoiceModel) TableName() string {
	return "invoices"
}

// PaymentModel represents the database model for payments.
type PaymentModel struct {
	ID                    string    `gorm:"primaryKey;type:uuid"`
	InvoiceID             string    `gorm:"type:uuid;not null;index"`
	TxHash                string    `gorm:"type:varchar(64);not null;uniqueIndex"` // Changed from TransactionHash to match DB.md
	Amount                string    `gorm:"type:decimal(20,8);not null"`
	FromAddress           string    `gorm:"type:varchar(42);not null"`
	ToAddress             string    `gorm:"type:varchar(42);not null"`
	Status                string    `gorm:"type:varchar(20);not null"`
	Confirmations         int       `gorm:"not null;default:0"`
	RequiredConfirmations int       `gorm:"not null;default:1"`
	BlockNumber           *int64    `gorm:"type:bigint"`
	BlockHash             *string   `gorm:"type:varchar(64)"`
	NetworkFee            *string   `gorm:"type:decimal(20,8)"`
	DetectedAt            time.Time `gorm:"not null"`
	ConfirmedAt           *time.Time
	CreatedAt             time.Time      `gorm:"not null"`
	DeletedAt             gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the PaymentModel.
func (PaymentModel) TableName() string {
	return "payments"
}

// MerchantModel represents the database model for merchants.
type MerchantModel struct {
	ID           string         `gorm:"primaryKey;type:uuid"`
	BusinessName string         `gorm:"type:varchar(255);not null"`
	ContactEmail string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Status       string         `gorm:"type:varchar(20);not null"`
	Settings     string         `gorm:"type:jsonb;not null"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the MerchantModel.
func (MerchantModel) TableName() string {
	return "merchants"
}

// APIKeyModel represents the database model for API keys.
type APIKeyModel struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	MerchantID  string `gorm:"type:uuid;not null;index"`
	KeyHash     string `gorm:"type:varchar(64);not null;uniqueIndex"`
	KeyType     string `gorm:"type:varchar(10);not null"`
	Permissions string `gorm:"type:jsonb;not null"`
	Status      string `gorm:"type:varchar(20);not null"`
	Name        string `gorm:"type:varchar(100)"`
	LastUsedAt  *time.Time
	ExpiresAt   *time.Time
	CreatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the APIKeyModel.
func (APIKeyModel) TableName() string {
	return "api_keys"
}

// WebhookEndpointModel represents the database model for webhook endpoints.
type WebhookEndpointModel struct {
	ID           string         `gorm:"primaryKey;type:uuid"`
	MerchantID   string         `gorm:"type:uuid;not null;index"`
	URL          string         `gorm:"type:varchar(500);not null"`
	Events       string         `gorm:"type:jsonb;not null"`
	Secret       string         `gorm:"type:varchar(255);not null"`
	Status       string         `gorm:"type:varchar(20);not null"`
	MaxRetries   int            `gorm:"not null;default:5"`
	RetryBackoff string         `gorm:"type:varchar(20);not null"`
	Timeout      int            `gorm:"not null;default:30"`
	AllowedIPs   string         `gorm:"type:jsonb"`
	Headers      string         `gorm:"type:jsonb"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the WebhookEndpointModel.
func (WebhookEndpointModel) TableName() string {
	return "webhook_endpoints"
}
