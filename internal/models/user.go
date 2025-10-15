package models

import (
	"time"
)

// User represents an end user who can purchase insurance policies.
type User struct {
	Base
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	FullName     string     `json:"full_name" gorm:"not null"`
	PhoneNumber  string     `json:"phone_number"`
	DateOfBirth  time.Time  `json:"date_of_birth"`
	Address      Address    `json:"address" gorm:"embedded;embedded_prefix:address_"`
	PasswordHash string     `json:"-" gorm:"not null"`
	Status       string     `json:"status" gorm:"default:active"`
	LastLoginAt  *time.Time `json:"last_login_at"`

	// Authentication fields
	EmailVerified bool   `json:"email_verified" gorm:"default:false"`
	Role          string `json:"role" gorm:"default:customer"`
	Permissions   string `json:"permissions" gorm:"type:jsonb"` // JSON array of permissions

	// MFA fields
	MFAEnabled  bool   `json:"mfa_enabled" gorm:"default:false"`
	MFASecret   string `json:"-" gorm:"column:mfa_secret"`
	BackupCodes string `json:"-" gorm:"column:backup_codes;type:jsonb"` // JSON array of backup codes

	// Password reset fields
	ResetToken   string     `json:"-" gorm:"column:reset_token"`
	ResetExpires *time.Time `json:"-" gorm:"column:reset_expires"`

	// Email verification fields
	VerifyToken   string     `json:"-" gorm:"column:verify_token"`
	VerifyExpires *time.Time `json:"-" gorm:"column:verify_expires"`

	// Relationships
	Policies      []Policy       `json:"policies,omitempty" gorm:"foreignKey:UserID"`
	Claims        []Claim        `json:"claims,omitempty" gorm:"foreignKey:UserID"`
	Payments      []Payment      `json:"payments,omitempty" gorm:"foreignKey:UserID"`
	Invoices      []Invoice      `json:"invoices,omitempty" gorm:"foreignKey:UserID"`
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:UserID"`
}

// Address represents a user's address information.
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// TableName returns the table name for the User model.
func (User) TableName() string {
	return "users"
}
