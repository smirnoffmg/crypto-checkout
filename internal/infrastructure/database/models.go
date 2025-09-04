package database

import (
	"time"

	"gorm.io/gorm"
)

// InvoiceModel represents the database model for invoices.
type InvoiceModel struct {
	ID             string    `gorm:"primaryKey;type:varchar(32)"`
	Status         string    `gorm:"type:varchar(20);not null"`
	TaxRate        string    `gorm:"type:decimal(5,4);not null"`
	PaymentAddress *string   `gorm:"type:varchar(42)"`
	CreatedAt      time.Time `gorm:"not null"`
	PaidAt         *time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// Relationships
	Items []InvoiceItemModel `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for the InvoiceModel.
func (InvoiceModel) TableName() string {
	return "invoices"
}

// InvoiceItemModel represents the database model for invoice items.
type InvoiceItemModel struct {
	ID          uint   `gorm:"primaryKey"`
	InvoiceID   string `gorm:"type:varchar(32);not null;index"`
	Description string `gorm:"type:text;not null"`
	UnitPrice   string `gorm:"type:decimal(20,2);not null"`
	Quantity    string `gorm:"type:decimal(20,8);not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the InvoiceItemModel.
func (InvoiceItemModel) TableName() string {
	return "invoice_items"
}

// PaymentModel represents the database model for payments.
type PaymentModel struct {
	ID              string         `gorm:"primaryKey;type:varchar(64)"`
	Amount          string         `gorm:"type:decimal(20,2);not null"`
	Address         string         `gorm:"type:varchar(42);not null"`
	TransactionHash string         `gorm:"type:varchar(64);not null;uniqueIndex"`
	Confirmations   int            `gorm:"not null;default:0"`
	Status          string         `gorm:"type:varchar(20);not null"`
	CreatedAt       time.Time      `gorm:"not null"`
	UpdatedAt       time.Time      `gorm:"not null"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the PaymentModel.
func (PaymentModel) TableName() string {
	return "payments"
}
