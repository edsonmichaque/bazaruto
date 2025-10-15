package events

import (
	"time"

	"github.com/google/uuid"
)

// PolicyCreatedEvent is published when a new policy is created.
type PolicyCreatedEvent struct {
	*BaseBusinessEvent
	PolicyID       uuid.UUID `json:"policy_id"`
	UserID         uuid.UUID `json:"user_id"`
	QuoteID        uuid.UUID `json:"quote_id"`
	ProductID      uuid.UUID `json:"product_id"`
	Premium        float64   `json:"premium"`
	Currency       string    `json:"currency"`
	EffectiveDate  time.Time `json:"effective_date"`
	ExpirationDate time.Time `json:"expiration_date"`
	CreatedAt      time.Time `json:"created_at"`
}

// NewPolicyCreatedEvent creates a new policy created event.
func NewPolicyCreatedEvent(policyID, userID, quoteID, productID uuid.UUID, premium float64, currency string, effectiveDate, expirationDate, createdAt time.Time) *PolicyCreatedEvent {
	event := &PolicyCreatedEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "policy.created",
			EntityID:      policyID,
			EntityType:    "policy",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		PolicyID:       policyID,
		UserID:         userID,
		QuoteID:        quoteID,
		ProductID:      productID,
		Premium:        premium,
		Currency:       currency,
		EffectiveDate:  effectiveDate,
		ExpirationDate: expirationDate,
		CreatedAt:      createdAt,
	}
	return event
}
