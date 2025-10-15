package services

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/events"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// PaymentService handles business logic for payments.
type PaymentService struct {
	store        store.PaymentStore
	eventService *EventService
}

// NewPaymentService creates a new PaymentService instance.
func NewPaymentService(store store.PaymentStore, eventService ...*EventService) *PaymentService {
	var evtService *EventService
	if len(eventService) > 0 {
		evtService = eventService[0]
	}
	return &PaymentService{
		store:        store,
		eventService: evtService,
	}
}

// GetPayment retrieves a payment by ID with business logic validation.
func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("payment ID is required")
	}

	payment, err := s.store.GetPayment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// CreatePayment creates a new payment with business logic validation.
func (s *PaymentService) CreatePayment(ctx context.Context, payment *models.Payment) error {
	// Validate required fields
	if payment.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if payment.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if payment.PaymentMethod == "" {
		return fmt.Errorf("payment method is required")
	}

	// Set defaults
	if payment.Currency == "" {
		payment.Currency = models.CurrencyUSD
	}
	if payment.Status == "" {
		payment.Status = models.PaymentStatusPending
	}
	if payment.PaymentNumber == "" {
		payment.PaymentNumber = s.generatePaymentNumber()
	}

	if err := s.store.CreatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	// Publish payment initiated event
	if s.eventService != nil {
		event := events.NewPaymentInitiatedEvent(payment.ID, payment.UserID, *payment.PolicyID, payment.Amount, payment.Currency, payment.PaymentMethod, time.Now())
		if err := s.eventService.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the payment creation
		}
	}

	return nil
}

// ProcessPayment processes a payment through the payment gateway.
func (s *PaymentService) ProcessPayment(ctx context.Context, paymentID uuid.UUID) (*models.Payment, error) {
	// Fetch payment details
	payment, err := s.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	if payment == nil {
		return nil, fmt.Errorf("payment not found: %s", paymentID.String())
	}

	// Check if payment is already processed
	if payment.Status == models.PaymentStatusCompleted {
		return payment, nil
	}

	// Process payment through payment gateway
	result, err := s.processPaymentGateway(ctx, payment)
	if err != nil {
		// Update payment status to failed
		payment.Status = models.PaymentStatusFailed
		now := time.Now()
		payment.FailedAt = &now
		failureReason := err.Error()
		payment.FailureReason = &failureReason
		_ = s.store.UpdatePayment(ctx, payment)

		// Publish payment failed event
		if s.eventService != nil {
			event := events.NewPaymentFailedEvent(payment.ID, payment.UserID, *payment.PolicyID, payment.Amount, payment.Currency, "PAYMENT_FAILED", failureReason, time.Now())
			if publishErr := s.eventService.PublishEvent(ctx, event); publishErr != nil {
				// Log error but don't fail the payment processing
			}
		}

		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	// Update payment with gateway response
	payment.Status = models.PaymentStatusCompleted
	payment.TransactionID = result.TransactionID
	now := time.Now()
	payment.ProcessedAt = &now

	if err := s.store.UpdatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish payment completed event
	if s.eventService != nil {
		event := events.NewPaymentCompletedEvent(payment.ID, payment.UserID, *payment.PolicyID, payment.Amount, payment.Currency, result.TransactionID, time.Now())
		if err := s.eventService.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the payment processing
		}
	}

	return payment, nil
}

// UpdatePayment updates an existing payment with business logic validation.
func (s *PaymentService) UpdatePayment(ctx context.Context, payment *models.Payment) error {
	if payment.ID == uuid.Nil {
		return fmt.Errorf("payment ID is required")
	}

	return s.store.UpdatePayment(ctx, payment)
}

// DeletePayment soft deletes a payment.
func (s *PaymentService) DeletePayment(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("payment ID is required")
	}

	return s.store.DeletePayment(ctx, id)
}

// ListPayments retrieves a list of payments with optional filtering.
func (s *PaymentService) ListPayments(ctx context.Context, opts *models.PaymentListOptions) (*models.ListResponse[*models.Payment], error) {
	if opts == nil {
		opts = models.NewPaymentListOptions()
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid list options: %w", err)
	}

	payments, err := s.store.ListPayments(ctx, opts.UserID, opts.PolicyID, opts.SubscriptionID, opts.Status, opts.GetLimit(), opts.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	total, err := s.store.CountPayments(ctx, opts.UserID, opts.PolicyID, opts.SubscriptionID, opts.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to count payments: %w", err)
	}

	return models.NewListResponse(payments, total, opts.ListOptions), nil
}

// CountPayments returns the total number of payments with optional filtering.
func (s *PaymentService) CountPayments(ctx context.Context, userID *uuid.UUID, status string) (int64, error) {
	return s.store.CountPayments(ctx, userID, nil, nil, status)
}

// PaymentGatewayResult represents the result of a payment gateway transaction
type PaymentGatewayResult struct {
	TransactionID string
	Response      string
	Success       bool
}

// processPaymentGateway simulates payment processing through a gateway
func (s *PaymentService) processPaymentGateway(ctx context.Context, payment *models.Payment) (*PaymentGatewayResult, error) {
	// In a real implementation, you would integrate with:
	// - Stripe
	// - PayPal
	// - Square
	// - Adyen
	// - etc.

	// Simulate network delay
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}

	// Simulate payment processing logic
	// In production, this would make actual API calls to payment gateways

	// Simulate occasional failures for testing
	if payment.Amount > 10000 {
		return nil, fmt.Errorf("payment gateway timeout")
	}

	// Generate mock transaction ID
	transactionID := fmt.Sprintf("txn_%d_%s", time.Now().Unix(), payment.ID.String()[:8])

	return &PaymentGatewayResult{
		TransactionID: transactionID,
		Response:      `{"status":"succeeded","id":"` + transactionID + `"}`,
		Success:       true,
	}, nil
}

// generatePaymentNumber generates a unique payment number.
func (s *PaymentService) generatePaymentNumber() string {
	// Generate a payment number with timestamp and random component
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("PAY-%s", timestamp)
}
