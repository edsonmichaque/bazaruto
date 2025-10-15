package events

import (
	"time"

	"github.com/google/uuid"
)

// ClaimSubmittedEvent is published when a new claim is submitted.
type ClaimSubmittedEvent struct {
	*BaseBusinessEvent
	ClaimID      uuid.UUID `json:"claim_id"`
	UserID       uuid.UUID `json:"user_id"`
	PolicyID     uuid.UUID `json:"policy_id"`
	ClaimAmount  float64   `json:"claim_amount"`
	Currency     string    `json:"currency"`
	ClaimType    string    `json:"claim_type"`
	Description  string    `json:"description"`
	IncidentDate time.Time `json:"incident_date"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

// NewClaimSubmittedEvent creates a new claim submitted event.
func NewClaimSubmittedEvent(claimID, userID, policyID uuid.UUID, claimAmount float64, currency, claimType, description string, incidentDate, submittedAt time.Time) *ClaimSubmittedEvent {
	event := &ClaimSubmittedEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "claim.submitted",
			EntityID:      claimID,
			EntityType:    "claim",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		ClaimID:      claimID,
		UserID:       userID,
		PolicyID:     policyID,
		ClaimAmount:  claimAmount,
		Currency:     currency,
		ClaimType:    claimType,
		Description:  description,
		IncidentDate: incidentDate,
		SubmittedAt:  submittedAt,
	}
	return event
}
