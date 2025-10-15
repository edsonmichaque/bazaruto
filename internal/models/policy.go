package models

import (
	"time"

	"github.com/google/uuid"
)

// Policy represents an active insurance policy.
type Policy struct {
	Base
	PolicyNumber     string     `json:"policy_number" gorm:"uniqueIndex;not null"`
	ProductID        uuid.UUID  `json:"product_id" gorm:"not null"`
	UserID           uuid.UUID  `json:"user_id" gorm:"not null"`
	QuoteID          *uuid.UUID `json:"quote_id"`
	Premium          float64    `json:"premium" gorm:"not null"`
	Currency         string     `json:"currency" gorm:"default:USD"`
	CoverageAmount   float64    `json:"coverage_amount" gorm:"not null"`
	Status           string     `json:"status" gorm:"default:active"`
	EffectiveDate    time.Time  `json:"effective_date" gorm:"not null"`
	ExpirationDate   time.Time  `json:"expiration_date" gorm:"not null"`
	RenewalDate      *time.Time `json:"renewal_date"`
	AutoRenew        bool       `json:"auto_renew" gorm:"default:false"`
	PaymentFrequency string     `json:"payment_frequency" gorm:"default:monthly"` // monthly, quarterly, annually

	// Relationships
	Product       Product        `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	User          User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Quote         *Quote         `json:"quote,omitempty" gorm:"foreignKey:QuoteID"`
	Claims        []Claim        `json:"claims,omitempty" gorm:"foreignKey:PolicyID"`
	Payments      []Payment      `json:"payments,omitempty" gorm:"foreignKey:PolicyID"`
	Invoices      []Invoice      `json:"invoices,omitempty" gorm:"foreignKey:PolicyID"`
	Beneficiaries []Beneficiary  `json:"beneficiaries,omitempty" gorm:"foreignKey:PolicyID"`
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:PolicyID"`
}

// TableName returns the table name for the Policy model.
func (Policy) TableName() string {
	return "policies"
}
