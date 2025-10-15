package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base contains common fields for all models.
type Base struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// BeforeCreate is called before creating a record.
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// Status constants for various model states.
const (
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusPending   = "pending"
	StatusExpired   = "expired"
	StatusCancelled = "cancelled"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// Currency constants.
const (
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
	CurrencyGBP = "GBP"
	CurrencyJPY = "JPY"
	CurrencyCAD = "CAD"
	CurrencyAUD = "AUD"
)

// Payment status constants.
const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
)

// Claim status constants.
const (
	ClaimStatusSubmitted   = "submitted"
	ClaimStatusUnderReview = "under_review"
	ClaimStatusApproved    = "approved"
	ClaimStatusDenied      = "denied"
	ClaimStatusPaid        = "paid"
)

// Policy status constants.
const (
	PolicyStatusActive    = "active"
	PolicyStatusInactive  = "inactive"
	PolicyStatusExpired   = "expired"
	PolicyStatusCancelled = "cancelled"
	PolicyStatusSuspended = "suspended"
)

// Quote status constants.
const (
	QuoteStatusPending = "pending"
	QuoteStatusActive  = "active"
	QuoteStatusExpired = "expired"
	QuoteStatusUsed    = "used"
)
