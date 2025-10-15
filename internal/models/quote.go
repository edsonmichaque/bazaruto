package models

import (
	"time"

	"github.com/google/uuid"
)

// Quote represents a price quote for an insurance product.
type Quote struct {
	Base
	ProductID   uuid.UUID    `json:"product_id" gorm:"not null"`
	UserID      uuid.UUID    `json:"user_id" gorm:"not null"`
	QuoteNumber string       `json:"quote_number" gorm:"uniqueIndex;not null"`
	BasePrice   float64      `json:"base_price" gorm:"not null"`
	FinalPrice  float64      `json:"final_price" gorm:"not null"`
	Currency    string       `json:"currency" gorm:"default:USD"`
	Discount    float64      `json:"discount" gorm:"default:0"`
	Tax         float64      `json:"tax" gorm:"default:0"`
	Status      string       `json:"status" gorm:"default:pending"`
	ValidUntil  time.Time    `json:"valid_until" gorm:"not null"`
	RiskFactors []RiskFactor `json:"risk_factors" gorm:"type:json"`

	// Relationships
	Product  Product  `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	User     User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Policies []Policy `json:"policies,omitempty" gorm:"foreignKey:QuoteID"`
}

// RiskFactor represents risk assessment factors for a quote.
type RiskFactor struct {
	Factor string  `json:"factor"`
	Value  string  `json:"value"`
	Impact float64 `json:"impact"` // multiplier for price adjustment
}

// TableName returns the table name for the Quote model.
func (Quote) TableName() string {
	return "quotes"
}
