package models

import (
	"time"

	"github.com/google/uuid"
)

// Claim represents a claim submitted against an insurance policy.
type Claim struct {
	Base
	ClaimNumber  string     `json:"claim_number" gorm:"uniqueIndex;not null"`
	PolicyID     uuid.UUID  `json:"policy_id" gorm:"not null"`
	UserID       uuid.UUID  `json:"user_id" gorm:"not null"`
	Title        string     `json:"title" gorm:"not null"`
	Description  string     `json:"description" gorm:"not null"`
	ClaimAmount  float64    `json:"claim_amount" gorm:"not null"`
	Currency     string     `json:"currency" gorm:"default:USD"`
	Status       string     `json:"status" gorm:"default:submitted"`
	IncidentDate time.Time  `json:"incident_date" gorm:"not null"`
	ReportedDate time.Time  `json:"reported_date" gorm:"not null"`
	ResolvedDate *time.Time `json:"resolved_date"`
	PaidAmount   float64    `json:"paid_amount" gorm:"default:0"`
	DenialReason *string    `json:"denial_reason"`
	Documents    []Document `json:"documents" gorm:"embedded;embedded_prefix:doc_"`

	// Relationships
	Policy Policy `json:"policy,omitempty" gorm:"foreignKey:PolicyID"`
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Document represents a document attached to a claim.
type Document struct {
	FileName    string    `json:"file_name"`
	FileType    string    `json:"file_type"`
	FileSize    int64     `json:"file_size"`
	UploadDate  time.Time `json:"upload_date"`
	Description string    `json:"description"`
}

// TableName returns the table name for the Claim model.
func (Claim) TableName() string {
	return "claims"
}
