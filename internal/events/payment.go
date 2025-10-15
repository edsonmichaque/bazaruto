package events

import (
	"time"

	"github.com/google/uuid"
)

// PaymentInitiatedEvent is published when a payment is initiated.
type PaymentInitiatedEvent struct {
	*BaseBusinessEvent
	PaymentID     uuid.UUID `json:"payment_id"`
	UserID        uuid.UUID `json:"user_id"`
	PolicyID      uuid.UUID `json:"policy_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	PaymentMethod string    `json:"payment_method"`
	InitiatedAt   time.Time `json:"initiated_at"`
}

// NewPaymentInitiatedEvent creates a new payment initiated event.
func NewPaymentInitiatedEvent(paymentID, userID, policyID uuid.UUID, amount float64, currency, paymentMethod string, initiatedAt time.Time) *PaymentInitiatedEvent {
	event := &PaymentInitiatedEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "payment.initiated",
			EntityID:      paymentID,
			EntityType:    "payment",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		PaymentID:     paymentID,
		UserID:        userID,
		PolicyID:      policyID,
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: paymentMethod,
		InitiatedAt:   initiatedAt,
	}
	return event
}

// PaymentCompletedEvent is published when a payment is completed successfully.
type PaymentCompletedEvent struct {
	*BaseBusinessEvent
	PaymentID     uuid.UUID `json:"payment_id"`
	UserID        uuid.UUID `json:"user_id"`
	PolicyID      uuid.UUID `json:"policy_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	TransactionID string    `json:"transaction_id"`
	CompletedAt   time.Time `json:"completed_at"`
}

// NewPaymentCompletedEvent creates a new payment completed event.
func NewPaymentCompletedEvent(paymentID, userID, policyID uuid.UUID, amount float64, currency, transactionID string, completedAt time.Time) *PaymentCompletedEvent {
	event := &PaymentCompletedEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "payment.completed",
			EntityID:      paymentID,
			EntityType:    "payment",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		PaymentID:     paymentID,
		UserID:        userID,
		PolicyID:      policyID,
		Amount:        amount,
		Currency:      currency,
		TransactionID: transactionID,
		CompletedAt:   completedAt,
	}
	return event
}

// PaymentFailedEvent is published when a payment fails.
type PaymentFailedEvent struct {
	*BaseBusinessEvent
	PaymentID    uuid.UUID `json:"payment_id"`
	UserID       uuid.UUID `json:"user_id"`
	PolicyID     uuid.UUID `json:"policy_id"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	ErrorCode    string    `json:"error_code"`
	ErrorMessage string    `json:"error_message"`
	FailedAt     time.Time `json:"failed_at"`
}

// NewPaymentFailedEvent creates a new payment failed event.
func NewPaymentFailedEvent(paymentID, userID, policyID uuid.UUID, amount float64, currency, errorCode, errorMessage string, failedAt time.Time) *PaymentFailedEvent {
	event := &PaymentFailedEvent{
		BaseBusinessEvent: &BaseBusinessEvent{
			EventID:       uuid.New(),
			EventType:     "payment.failed",
			EntityID:      paymentID,
			EntityType:    "payment",
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
			EventMetadata: make(map[string]interface{}),
		},
		PaymentID:    paymentID,
		UserID:       userID,
		PolicyID:     policyID,
		Amount:       amount,
		Currency:     currency,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		FailedAt:     failedAt,
	}
	return event
}