package events

import (
	"time"

	"github.com/google/uuid"
)

// FraudAnalysisCompletedEvent is published when fraud analysis is completed.
type FraudAnalysisCompletedEvent struct {
	*BaseBusinessEvent
	ClaimID        uuid.UUID `json:"claim_id"`
	UserID         uuid.UUID `json:"user_id"`
	Score          float64   `json:"score"`
	RiskLevel      string    `json:"risk_level"`
	RequiresReview bool      `json:"requires_review"`
	Confidence     float64   `json:"confidence"`
	RiskFactors    []string  `json:"risk_factors"`
	AnalyzedAt     time.Time `json:"analyzed_at"`
}

// NewFraudAnalysisCompletedEvent creates a new fraud analysis completed event.
func NewFraudAnalysisCompletedEvent(claimID, userID uuid.UUID, score float64, riskLevel string, requiresReview bool, confidence float64, riskFactors []string, analyzedAt time.Time) *FraudAnalysisCompletedEvent {
	event := &FraudAnalysisCompletedEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "fraud.analysis_completed",
			EntityID:      claimID,
			EntityType:    "claim",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		ClaimID:        claimID,
		UserID:         userID,
		Score:          score,
		RiskLevel:      riskLevel,
		RequiresReview: requiresReview,
		Confidence:     confidence,
		RiskFactors:    riskFactors,
		AnalyzedAt:     analyzedAt,
	}
	return event
}
