package models

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents a customer in the insurance domain.
// This is separate from User which is only for authentication/authorization.
type Customer struct {
	Base
	UserID           uuid.UUID              `json:"user_id" gorm:"uniqueIndex;not null"` // Reference to auth user
	CustomerNumber   string                 `json:"customer_number" gorm:"uniqueIndex;not null"`
	FirstName        string                 `json:"first_name" gorm:"not null"`
	LastName         string                 `json:"last_name" gorm:"not null"`
	Email            string                 `json:"email" gorm:"not null"`
	Phone            string                 `json:"phone"`
	DateOfBirth      *time.Time             `json:"date_of_birth"`
	Gender           string                 `json:"gender"`
	MaritalStatus    string                 `json:"marital_status"`
	Nationality      string                 `json:"nationality"`
	Occupation       string                 `json:"occupation"`
	Employer         string                 `json:"employer"`
	AnnualIncome     float64                `json:"annual_income"`
	CreditScore      int                    `json:"credit_score"`
	RiskProfile      string                 `json:"risk_profile"` // low, medium, high, very_high
	CustomerTier     string                 `json:"customer_tier"` // bronze, silver, gold, platinum
	Status           string                 `json:"status" gorm:"default:active"`
	PreferredContact string                 `json:"preferred_contact"` // email, phone, sms
	Language         string                 `json:"language" gorm:"default:en"`
	Timezone         string                 `json:"timezone"`
	MarketingConsent bool                   `json:"marketing_consent" gorm:"default:false"`
	DataConsent      bool                   `json:"data_consent" gorm:"default:false"`
	KYCStatus        string                 `json:"kyc_status" gorm:"default:pending"` // pending, verified, rejected, expired
	AMLStatus        string                 `json:"aml_status" gorm:"default:pending"` // pending, cleared, flagged, under_review
	LastKYCUpdate    *time.Time             `json:"last_kyc_update"`
	LastAMLUpdate    *time.Time             `json:"last_aml_update"`
	Addresses        []CustomerAddress      `json:"addresses" gorm:"foreignKey:CustomerID"`
	Documents        []CustomerDocument     `json:"documents" gorm:"foreignKey:CustomerID"`
	RiskFactors      []CustomerRiskFactor   `json:"risk_factors" gorm:"foreignKey:CustomerID"`
	Preferences      map[string]interface{} `json:"preferences" gorm:"type:jsonb"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Policies  []Policy   `json:"policies,omitempty" gorm:"foreignKey:CustomerID"`
	Claims    []Claim    `json:"claims,omitempty" gorm:"foreignKey:CustomerID"`
	Payments  []Payment  `json:"payments,omitempty" gorm:"foreignKey:CustomerID"`
	Quotes    []Quote    `json:"quotes,omitempty" gorm:"foreignKey:CustomerID"`
}

// CustomerAddress represents a customer's address.
type CustomerAddress struct {
	Base
	CustomerID  uuid.UUID `json:"customer_id" gorm:"not null"`
	Type        string    `json:"type"` // home, work, mailing, billing
	Line1       string    `json:"line1" gorm:"not null"`
	Line2       string    `json:"line2"`
	City        string    `json:"city" gorm:"not null"`
	State       string    `json:"state" gorm:"not null"`
	PostalCode  string    `json:"postal_code" gorm:"not null"`
	Country     string    `json:"country" gorm:"not null"`
	IsPrimary   bool      `json:"is_primary" gorm:"default:false"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
}

// CustomerDocument represents a customer's document.
type CustomerDocument struct {
	Base
	CustomerID  uuid.UUID `json:"customer_id" gorm:"not null"`
	Type        string    `json:"type"` // passport, driver_license, national_id, utility_bill, bank_statement
	DocumentID  string    `json:"document_id" gorm:"not null"`
	IssuingAuthority string `json:"issuing_authority"`
	IssueDate   *time.Time `json:"issue_date"`
	ExpiryDate  *time.Time `json:"expiry_date"`
	Status      string    `json:"status" gorm:"default:pending"` // pending, verified, rejected, expired
	FileURL     string    `json:"file_url"`
	FileHash    string    `json:"file_hash"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// CustomerRiskFactor represents a risk factor associated with a customer.
type CustomerRiskFactor struct {
	Base
	CustomerID  uuid.UUID `json:"customer_id" gorm:"not null"`
	Factor      string    `json:"factor" gorm:"not null"` // age, occupation, location, credit_score, etc.
	Value       string    `json:"value" gorm:"not null"`
	Impact      float64   `json:"impact"` // Risk impact multiplier
	Severity    string    `json:"severity"` // low, medium, high, critical
	Source      string    `json:"source"` // internal, external, manual
	ValidFrom   time.Time `json:"valid_from" gorm:"not null"`
	ValidTo     *time.Time `json:"valid_to"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for the Customer model.
func (Customer) TableName() string {
	return "customers"
}

// TableName returns the table name for the CustomerAddress model.
func (CustomerAddress) TableName() string {
	return "customer_addresses"
}

// TableName returns the table name for the CustomerDocument model.
func (CustomerDocument) TableName() string {
	return "customer_documents"
}

// TableName returns the table name for the CustomerRiskFactor model.
func (CustomerRiskFactor) TableName() string {
	return "customer_risk_factors"
}

// GetFullName returns the customer's full name.
func (c *Customer) GetFullName() string {
	return c.FirstName + " " + c.LastName
}

// GetPrimaryAddress returns the customer's primary address.
func (c *Customer) GetPrimaryAddress() *CustomerAddress {
	for i := range c.Addresses {
		if c.Addresses[i].IsPrimary && c.Addresses[i].IsActive {
			return &c.Addresses[i]
		}
	}
	return nil
}

// GetActiveRiskFactors returns active risk factors for the customer.
func (c *Customer) GetActiveRiskFactors() []CustomerRiskFactor {
	var activeFactors []CustomerRiskFactor
	now := time.Now()
	
	for _, factor := range c.RiskFactors {
		if factor.IsActive && 
		   factor.ValidFrom.Before(now) && 
		   (factor.ValidTo == nil || factor.ValidTo.After(now)) {
			activeFactors = append(activeFactors, factor)
		}
	}
	
	return activeFactors
}

// IsKYCVerified returns true if the customer's KYC is verified.
func (c *Customer) IsKYCVerified() bool {
	return c.KYCStatus == "verified"
}

// IsAMLCleared returns true if the customer's AML check is cleared.
func (c *Customer) IsAMLCleared() bool {
	return c.AMLStatus == "cleared"
}

// GetAge returns the customer's age in years.
func (c *Customer) GetAge() int {
	if c.DateOfBirth == nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - c.DateOfBirth.Year()
	if now.YearDay() < c.DateOfBirth.YearDay() {
		age--
	}
	return age
}

// IsHighRisk returns true if the customer has a high risk profile.
func (c *Customer) IsHighRisk() bool {
	return c.RiskProfile == "high" || c.RiskProfile == "very_high"
}

// GetCustomerTierLevel returns the numeric level of the customer tier.
func (c *Customer) GetCustomerTierLevel() int {
	switch c.CustomerTier {
	case "bronze":
		return 1
	case "silver":
		return 2
	case "gold":
		return 3
	case "platinum":
		return 4
	default:
		return 0
	}
}

