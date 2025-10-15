package models

import (
	"time"

	"github.com/google/uuid"
)

// Payment represents a payment transaction.
type Payment struct {
	Base
	PaymentNumber   string     `json:"payment_number" gorm:"uniqueIndex;not null"`
	UserID          uuid.UUID  `json:"user_id" gorm:"not null"`
	PolicyID        *uuid.UUID `json:"policy_id"`
	SubscriptionID  *uuid.UUID `json:"subscription_id"`
	Amount          float64    `json:"amount" gorm:"not null"`
	Currency        string     `json:"currency" gorm:"default:USD"`
	Status          string     `json:"status" gorm:"default:pending"`
	PaymentMethod   string     `json:"payment_method" gorm:"not null"` // credit_card, bank_transfer, etc.
	PaymentProvider string     `json:"payment_provider"`               // stripe, paypal, etc.
	TransactionID   string     `json:"transaction_id"`
	ProcessedAt     *time.Time `json:"processed_at"`
	FailedAt        *time.Time `json:"failed_at"`
	FailureReason   *string    `json:"failure_reason"`
	RefundAmount    float64    `json:"refund_amount" gorm:"default:0"`
	RefundedAt      *time.Time `json:"refunded_at"`

	// Relationships
	User         User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Policy       *Policy       `json:"policy,omitempty" gorm:"foreignKey:PolicyID"`
	Subscription *Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
}

// TableName returns the table name for the Payment model.
func (Payment) TableName() string {
	return "payments"
}
