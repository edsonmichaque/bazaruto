package models

// Partner represents an insurance provider/company.
type Partner struct {
	Base
	Name           string  `json:"name" gorm:"not null"`
	Description    string  `json:"description"`
	Website        string  `json:"website"`
	Email          string  `json:"email"`
	PhoneNumber    string  `json:"phone_number"`
	Address        Address `json:"address" gorm:"embedded;embedded_prefix:address_"`
	LicenseNumber  string  `json:"license_number" gorm:"uniqueIndex"`
	Status         string  `json:"status" gorm:"default:active"`
	CommissionRate float64 `json:"commission_rate" gorm:"default:0.1"` // 10% default commission

	// Relationships
	Products []Product `json:"products,omitempty" gorm:"foreignKey:PartnerID"`
}

// TableName returns the table name for the Partner model.
func (Partner) TableName() string {
	return "partners"
}
