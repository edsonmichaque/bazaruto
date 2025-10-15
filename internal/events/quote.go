package events

import (
	"time"

	"github.com/google/uuid"
)

// QuoteCreatedEvent is published when a new quote is created.
type QuoteCreatedEvent struct {
	*BaseBusinessEvent
	QuoteID        uuid.UUID `json:"quote_id"`
	UserID         uuid.UUID `json:"user_id"`
	ProductID      uuid.UUID `json:"product_id"`
	CoverageAmount float64   `json:"coverage_amount"`
	Currency       string    `json:"currency"`
	ValidUntil     time.Time `json:"valid_until"`
}

// NewQuoteCreatedEvent creates a new quote created event.
func NewQuoteCreatedEvent(quoteID, userID, productID uuid.UUID, coverageAmount float64, currency string, validUntil time.Time) *QuoteCreatedEvent {
	event := &QuoteCreatedEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "quote.created",
			EntityID:      quoteID,
			EntityType:    "quote",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		QuoteID:        quoteID,
		UserID:         userID,
		ProductID:      productID,
		CoverageAmount: coverageAmount,
		Currency:       currency,
		ValidUntil:     validUntil,
	}
	return event
}

// QuoteCalculatedEvent is published when a quote premium is calculated.
type QuoteCalculatedEvent struct {
	*BaseBusinessEvent
	QuoteID      uuid.UUID `json:"quote_id"`
	UserID       uuid.UUID `json:"user_id"`
	BasePremium  float64   `json:"base_premium"`
	FinalPremium float64   `json:"final_premium"`
	Currency     string    `json:"currency"`
	CalculatedAt time.Time `json:"calculated_at"`
}

// NewQuoteCalculatedEvent creates a new quote calculated event.
func NewQuoteCalculatedEvent(quoteID, userID uuid.UUID, basePremium, finalPremium float64, currency string, calculatedAt time.Time) *QuoteCalculatedEvent {
	event := &QuoteCalculatedEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "quote.calculated",
			EntityID:      quoteID,
			EntityType:    "quote",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		QuoteID:      quoteID,
		UserID:       userID,
		BasePremium:  basePremium,
		FinalPremium: finalPremium,
		Currency:     currency,
		CalculatedAt: calculatedAt,
	}
	return event
}

