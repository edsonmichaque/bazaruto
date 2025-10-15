package models

import (
	"time"

	"github.com/google/uuid"
)

// Beneficiary represents a beneficiary of an insurance policy.
type Beneficiary struct {
	Base
	PolicyID     uuid.UUID `json:"policy_id" gorm:"not null"`
	FirstName    string    `json:"first_name" gorm:"not null"`
	LastName     string    `json:"last_name" gorm:"not null"`
	Email        string    `json:"email"`
	PhoneNumber  string    `json:"phone_number"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	Relationship string    `json:"relationship" gorm:"not null"` // spouse, child, parent, etc.
	Percentage   float64   `json:"percentage" gorm:"not null"`   // percentage of benefit
	Address      Address   `json:"address" gorm:"embedded;embedded_prefix:address_"`
	Status       string    `json:"status" gorm:"default:active"`

	// Relationships
	Policy Policy `json:"policy,omitempty" gorm:"foreignKey:PolicyID"`
}

// TableName returns the table name for the Beneficiary model.
func (Beneficiary) TableName() string {
	return "beneficiaries"
}
