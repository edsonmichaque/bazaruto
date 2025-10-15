package models

import (
	"time"

	"github.com/google/uuid"
)

// Subscription represents a recurring subscription for insurance policies.
type Subscription struct {
	Base
	SubscriptionNumber string     `json:"subscription_number" gorm:"uniqueIndex;not null"`
	UserID             uuid.UUID  `json:"user_id" gorm:"not null"`
	PolicyID           uuid.UUID  `json:"policy_id" gorm:"not null"`
	Status             string     `json:"status" gorm:"default:active"`
	StartDate          time.Time  `json:"start_date" gorm:"not null"`
	EndDate            *time.Time `json:"end_date"`
	NextBillingDate    time.Time  `json:"next_billing_date" gorm:"not null"`
	BillingFrequency   string     `json:"billing_frequency" gorm:"default:monthly"` // monthly, quarterly, annually
	Amount             float64    `json:"amount" gorm:"not null"`
	Currency           string     `json:"currency" gorm:"default:USD"`
	AutoRenew          bool       `json:"auto_renew" gorm:"default:true"`
	PaymentMethodID    string     `json:"payment_method_id"`

	// Relationships
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Policy   Policy    `json:"policy,omitempty" gorm:"foreignKey:PolicyID"`
	Payments []Payment `json:"payments,omitempty" gorm:"foreignKey:SubscriptionID"`
}

// TableName returns the table name for the Subscription model.
func (Subscription) TableName() string {
	return "subscriptions"
}
