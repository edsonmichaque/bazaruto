package models

import (
	"time"

	"github.com/google/uuid"
)

// WebhookConfig represents a webhook configuration for external integrations
type WebhookConfig struct {
	Base
	URL        string            `json:"url" gorm:"not null"`
	Method     string            `json:"method" gorm:"default:POST"`
	Headers    map[string]string `json:"headers" gorm:"type:jsonb"`
	EventTypes []string          `json:"event_types" gorm:"type:jsonb"` // Events this webhook should listen to
	Secret     string            `json:"-" gorm:"column:secret"`        // Webhook secret for signature verification
	IsActive   bool              `json:"is_active" gorm:"default:true"`
	RetryCount int               `json:"retry_count" gorm:"default:3"`
	Timeout    int               `json:"timeout" gorm:"default:30"` // Timeout in seconds

	// Partner/Integration info
	PartnerID *uuid.UUID `json:"partner_id" gorm:"type:uuid"`
	Partner   *Partner   `json:"partner,omitempty" gorm:"foreignKey:PartnerID"`

	// Metadata
	Description string `json:"description"`
	Tags        string `json:"tags"` // Comma-separated tags for categorization
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	Base
	WebhookConfigID uuid.UUID              `json:"webhook_config_id" gorm:"type:uuid;not null"`
	WebhookConfig   *WebhookConfig         `json:"webhook_config,omitempty" gorm:"foreignKey:WebhookConfigID"`
	EventID         uuid.UUID              `json:"event_id" gorm:"type:uuid;not null"`
	EventType       string                 `json:"event_type" gorm:"not null"`
	URL             string                 `json:"url" gorm:"not null"`
	Method          string                 `json:"method" gorm:"not null"`
	Headers         map[string]string      `json:"headers" gorm:"type:jsonb"`
	Payload         map[string]interface{} `json:"payload" gorm:"type:jsonb"`

	// Delivery status
	Status        string     `json:"status" gorm:"not null"` // pending, delivered, failed
	AttemptCount  int        `json:"attempt_count" gorm:"default:0"`
	LastAttemptAt *time.Time `json:"last_attempt_at"`
	DeliveredAt   *time.Time `json:"delivered_at"`
	FailedAt      *time.Time `json:"failed_at"`

	// Response details
	ResponseStatus int    `json:"response_status"`
	ResponseBody   string `json:"response_body"`
	ErrorMessage   string `json:"error_message"`
}

// Webhook delivery status constants
const (
	WebhookStatusPending   = "pending"
	WebhookStatusDelivered = "delivered"
	WebhookStatusFailed    = "failed"
)
