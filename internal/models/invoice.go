package models

import (
	"time"

	"github.com/google/uuid"
)

// Invoice represents a billing invoice.
type Invoice struct {
	Base
	InvoiceNumber  string        `json:"invoice_number" gorm:"uniqueIndex;not null"`
	UserID         uuid.UUID     `json:"user_id" gorm:"not null"`
	PolicyID       *uuid.UUID    `json:"policy_id"`
	SubscriptionID *uuid.UUID    `json:"subscription_id"`
	Amount         float64       `json:"amount" gorm:"not null"`
	Currency       string        `json:"currency" gorm:"default:USD"`
	Tax            float64       `json:"tax" gorm:"default:0"`
	Total          float64       `json:"total" gorm:"not null"`
	Status         string        `json:"status" gorm:"default:pending"`
	DueDate        time.Time     `json:"due_date" gorm:"not null"`
	PaidAt         *time.Time    `json:"paid_at"`
	PaymentID      *uuid.UUID    `json:"payment_id"`
	Items          []InvoiceItem `json:"items" gorm:"embedded;embedded_prefix:item_"`

	// Relationships
	User         User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Policy       *Policy       `json:"policy,omitempty" gorm:"foreignKey:PolicyID"`
	Subscription *Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
	Payment      *Payment      `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
}

// InvoiceItem represents an item on an invoice.
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

// TableName returns the table name for the Invoice model.
func (Invoice) TableName() string {
	return "invoices"
}
