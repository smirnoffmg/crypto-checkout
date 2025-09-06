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
