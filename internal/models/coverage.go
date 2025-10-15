package models

import (
	"github.com/google/uuid"
)

// Coverage represents coverage details for an insurance product.
type Coverage struct {
	Base
	ProductID     uuid.UUID `json:"product_id" gorm:"not null"`
	CoverageType  string    `json:"coverage_type" gorm:"not null"` // medical, dental, vision, etc.
	CoverageName  string    `json:"coverage_name" gorm:"not null"`
	Description   string    `json:"description"`
	CoverageLimit float64   `json:"coverage_limit"`
	Deductible    float64   `json:"deductible"`
	Copay         float64   `json:"copay"`
	Coinsurance   float64   `json:"coinsurance"` // percentage
	IsIncluded    bool      `json:"is_included" gorm:"default:true"`
	SortOrder     int       `json:"sort_order" gorm:"default:0"`

	// Relationships
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

// TableName returns the table name for the Coverage model.
func (Coverage) TableName() string {
	return "coverages"
}
