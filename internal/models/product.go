package models

import (
	"time"

	"github.com/google/uuid"
)

// Product represents an insurance product in the marketplace.
type Product struct {
	Base
	Name               string     `json:"name" gorm:"not null"`
	Description        string     `json:"description"`
	Category           string     `json:"category" gorm:"not null"`
	PartnerID          uuid.UUID  `json:"partner_id" gorm:"not null"`
	BasePrice          float64    `json:"base_price" gorm:"not null"`
	Currency           string     `json:"currency" gorm:"default:USD"`
	CoverageAmount     float64    `json:"coverage_amount"`
	CoveragePeriod     int        `json:"coverage_period"` // in days
	Deductible         float64    `json:"deductible"`
	Status             string     `json:"status" gorm:"default:active"`
	EffectiveDate      time.Time  `json:"effective_date"`
	ExpirationDate     *time.Time `json:"expiration_date"`
	TermsAndConditions string     `json:"terms_and_conditions"`

	// Relationships
	Partner   Partner    `json:"partner,omitempty" gorm:"foreignKey:PartnerID"`
	Quotes    []Quote    `json:"quotes,omitempty" gorm:"foreignKey:ProductID"`
	Policies  []Policy   `json:"policies,omitempty" gorm:"foreignKey:ProductID"`
	Coverages []Coverage `json:"coverages,omitempty" gorm:"foreignKey:ProductID"`
}

// TableName returns the table name for the Product model.
func (Product) TableName() string {
	return "products"
}
