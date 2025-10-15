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

// PolicyRenewedEvent is published when a policy is renewed.
type PolicyRenewedEvent struct {
	*BaseBusinessEvent
	OldPolicyID    uuid.UUID `json:"old_policy_id"`
	NewPolicyID    uuid.UUID `json:"new_policy_id"`
	UserID         uuid.UUID `json:"user_id"`
	ProductID      uuid.UUID `json:"product_id"`
	Premium        float64   `json:"premium"`
	Currency       string    `json:"currency"`
	EffectiveDate  time.Time `json:"effective_date"`
	ExpirationDate time.Time `json:"expiration_date"`
	RenewedAt      time.Time `json:"renewed_at"`
}

// NewPolicyRenewedEvent creates a new policy renewed event.
func NewPolicyRenewedEvent(oldPolicyID, newPolicyID, userID, productID uuid.UUID, premium float64, currency string, effectiveDate, expirationDate, renewedAt time.Time) *PolicyRenewedEvent {
	event := &PolicyRenewedEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "policy.renewed",
			EntityID:      newPolicyID,
			EntityType:    "policy",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		OldPolicyID:    oldPolicyID,
		NewPolicyID:    newPolicyID,
		UserID:         userID,
		ProductID:      productID,
		Premium:        premium,
		Currency:       currency,
		EffectiveDate:  effectiveDate,
		ExpirationDate: expirationDate,
		RenewedAt:      renewedAt,
	}
	return event
}

// PolicyCancelledEvent is published when a policy is cancelled.
type PolicyCancelledEvent struct {
	*BaseBusinessEvent
	PolicyID         uuid.UUID `json:"policy_id"`
	UserID           uuid.UUID `json:"user_id"`
	ProductID        uuid.UUID `json:"product_id"`
	RefundAmount     float64   `json:"refund_amount"`
	Currency         string    `json:"currency"`
	CancellationDate time.Time `json:"cancellation_date"`
	EffectiveDate    time.Time `json:"effective_date"`
	Reason           string    `json:"reason"`
}

// NewPolicyCancelledEvent creates a new policy cancelled event.
func NewPolicyCancelledEvent(policyID, userID, productID uuid.UUID, refundAmount float64, currency string, cancellationDate, effectiveDate time.Time, reason string) *PolicyCancelledEvent {
	event := &PolicyCancelledEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "policy.cancelled",
			EntityID:      policyID,
			EntityType:    "policy",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		PolicyID:         policyID,
		UserID:           userID,
		ProductID:        productID,
		RefundAmount:     refundAmount,
		Currency:         currency,
		CancellationDate: cancellationDate,
		EffectiveDate:    effectiveDate,
		Reason:           reason,
	}
	return event
}

// PolicyExpiredEvent is published when a policy expires.
type PolicyExpiredEvent struct {
	*BaseBusinessEvent
	PolicyID       uuid.UUID `json:"policy_id"`
	UserID         uuid.UUID `json:"user_id"`
	ProductID      uuid.UUID `json:"product_id"`
	ExpirationDate time.Time `json:"expiration_date"`
	ExpiredAt      time.Time `json:"expired_at"`
}

// NewPolicyExpiredEvent creates a new policy expired event.
func NewPolicyExpiredEvent(policyID, userID, productID uuid.UUID, expirationDate, expiredAt time.Time) *PolicyExpiredEvent {
	event := &PolicyExpiredEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "policy.expired",
			EntityID:      policyID,
			EntityType:    "policy",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		PolicyID:       policyID,
		UserID:         userID,
		ProductID:      productID,
		ExpirationDate: expirationDate,
		ExpiredAt:      expiredAt,
	}
	return event
}

// GracePeriodExpiredEvent is published when a policy's grace period expires.
type GracePeriodExpiredEvent struct {
	*BaseBusinessEvent
	PolicyID           uuid.UUID `json:"policy_id"`
	UserID             uuid.UUID `json:"user_id"`
	ProductID          uuid.UUID `json:"product_id"`
	GracePeriodEndDate time.Time `json:"grace_period_end_date"`
	ExpiredAt          time.Time `json:"expired_at"`
}

// NewGracePeriodExpiredEvent creates a new grace period expired event.
func NewGracePeriodExpiredEvent(policyID, userID, productID uuid.UUID, gracePeriodEndDate, expiredAt time.Time) *GracePeriodExpiredEvent {
	event := &GracePeriodExpiredEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "policy.grace_period_expired",
			EntityID:      policyID,
			EntityType:    "policy",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		PolicyID:           policyID,
		UserID:             userID,
		ProductID:          productID,
		GracePeriodEndDate: gracePeriodEndDate,
		ExpiredAt:          expiredAt,
	}
	return event
}

// RenewalReminderEvent is published when a renewal reminder is sent.
type RenewalReminderEvent struct {
	*BaseBusinessEvent
	PolicyID        uuid.UUID `json:"policy_id"`
	UserID          uuid.UUID `json:"user_id"`
	ProductID       uuid.UUID `json:"product_id"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	ExpirationDate  time.Time `json:"expiration_date"`
	ReminderSentAt  time.Time `json:"reminder_sent_at"`
}

// NewRenewalReminderEvent creates a new renewal reminder event.
func NewRenewalReminderEvent(policyID, userID, productID uuid.UUID, daysUntilExpiry int, expirationDate, reminderSentAt time.Time) *RenewalReminderEvent {
	event := &RenewalReminderEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "policy.renewal_reminder",
			EntityID:      policyID,
			EntityType:    "policy",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		PolicyID:        policyID,
		UserID:          userID,
		ProductID:       productID,
		DaysUntilExpiry: daysUntilExpiry,
		ExpirationDate:  expirationDate,
		ReminderSentAt:  reminderSentAt,
	}
	return event
}
