package handlers

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/events"
	"github.com/edsonmichaque/bazaruto/internal/jobs"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/pkg/event"
	"github.com/edsonmichaque/bazaruto/pkg/job"
	"go.uber.org/zap"
)

// PaymentInitiatedHandler handles payment initiation events.
type PaymentInitiatedHandler struct {
	paymentService *services.PaymentService
	dispatcher     job.Dispatcher
	logger         *logger.Logger
}

// NewPaymentInitiatedHandler creates a new payment initiated event handler.
func NewPaymentInitiatedHandler(paymentService *services.PaymentService, dispatcher job.Dispatcher, logger *logger.Logger) *PaymentInitiatedHandler {
	return &PaymentInitiatedHandler{
		paymentService: paymentService,
		dispatcher:     dispatcher,
		logger:         logger,
	}
}

// Handle processes a payment initiated event.
func (h *PaymentInitiatedHandler) Handle(ctx context.Context, event event.Event) error {
	paymentEvent, ok := event.(*events.PaymentInitiatedEvent)
	if !ok {
		return fmt.Errorf("expected PaymentInitiatedEvent, got %T", event)
	}

	h.logger.Info("Processing payment initiated event",
		zap.String("payment_id", paymentEvent.PaymentID.String()),
		zap.String("user_id", paymentEvent.UserID.String()),
		zap.Float64("amount", paymentEvent.Amount),
		zap.String("method", paymentEvent.PaymentMethod))

	// Dispatch payment processing job
	processJob := &jobs.ProcessPaymentJob{
		PaymentID:      paymentEvent.PaymentID,
		PaymentService: h.paymentService,
	}

	if err := h.dispatcher.PerformWithContext(ctx, processJob); err != nil {
		h.logger.Error("Failed to dispatch payment processing job",
			zap.Error(err),
			zap.String("payment_id", paymentEvent.PaymentID.String()))
		return fmt.Errorf("failed to dispatch payment processing job: %w", err)
	}

	h.logger.Info("Payment processing job dispatched successfully",
		zap.String("payment_id", paymentEvent.PaymentID.String()))

	return nil
}

// CanHandle returns true if this handler can process the given event type.
func (h *PaymentInitiatedHandler) CanHandle(eventType string) bool {
	return eventType == "payment.initiated"
}

// HandlerName returns a unique name for this handler.
func (h *PaymentInitiatedHandler) HandlerName() string {
	return "payment_initiated_handler"
}

// PaymentCompletedHandler handles payment completion events.
type PaymentCompletedHandler struct {
	dispatcher job.Dispatcher
	logger     *logger.Logger
}

// NewPaymentCompletedHandler creates a new payment completed event handler.
func NewPaymentCompletedHandler(dispatcher job.Dispatcher, logger *logger.Logger) *PaymentCompletedHandler {
	return &PaymentCompletedHandler{
		dispatcher: dispatcher,
		logger:     logger,
	}
}

// Handle processes a payment completed event.
func (h *PaymentCompletedHandler) Handle(ctx context.Context, event event.Event) error {
	paymentEvent, ok := event.(*events.PaymentCompletedEvent)
	if !ok {
		return fmt.Errorf("expected PaymentCompletedEvent, got %T", event)
	}

	h.logger.Info("Processing payment completed event",
		zap.String("payment_id", paymentEvent.PaymentID.String()),
		zap.String("user_id", paymentEvent.UserID.String()),
		zap.Float64("amount", paymentEvent.Amount),
		zap.String("transaction_id", paymentEvent.TransactionID))

	// Dispatch payment confirmation email job
	emailJob := &jobs.SendEmailJob{
		To:      fmt.Sprintf("user-%s@example.com", paymentEvent.UserID.String()), // In real app, get from user service
		Subject: "Payment Confirmation",
		Body:    fmt.Sprintf("Your payment of %s %s has been processed successfully. Transaction ID: %s", paymentEvent.Amount, paymentEvent.Currency, paymentEvent.TransactionID),
		From:    "payments@bazaruto.com",
	}

	if err := h.dispatcher.PerformWithContext(ctx, emailJob); err != nil {
		h.logger.Error("Failed to dispatch payment confirmation email job",
			zap.Error(err),
			zap.String("payment_id", paymentEvent.PaymentID.String()))
		return fmt.Errorf("failed to dispatch payment confirmation email job: %w", err)
	}

	h.logger.Info("Payment confirmation email job dispatched successfully",
		zap.String("payment_id", paymentEvent.PaymentID.String()))

	return nil
}

// CanHandle returns true if this handler can process the given event type.
func (h *PaymentCompletedHandler) CanHandle(eventType string) bool {
	return eventType == "payment.completed"
}

// HandlerName returns a unique name for this handler.
func (h *PaymentCompletedHandler) HandlerName() string {
	return "payment_completed_handler"
}

// PaymentFailedHandler handles payment failure events.
type PaymentFailedHandler struct {
	dispatcher job.Dispatcher
	logger     *logger.Logger
}

// NewPaymentFailedHandler creates a new payment failed event handler.
func NewPaymentFailedHandler(dispatcher job.Dispatcher, logger *logger.Logger) *PaymentFailedHandler {
	return &PaymentFailedHandler{
		dispatcher: dispatcher,
		logger:     logger,
	}
}

// Handle processes a payment failed event.
func (h *PaymentFailedHandler) Handle(ctx context.Context, event event.Event) error {
	paymentEvent, ok := event.(*events.PaymentFailedEvent)
	if !ok {
		return fmt.Errorf("expected PaymentFailedEvent, got %T", event)
	}

	h.logger.Info("Processing payment failed event",
		zap.String("payment_id", paymentEvent.PaymentID.String()),
		zap.String("user_id", paymentEvent.UserID.String()),
		zap.Float64("amount", paymentEvent.Amount),
		zap.String("failure_reason", paymentEvent.ErrorMessage))

	// Dispatch payment failure notification email job
	emailJob := &jobs.SendEmailJob{
		To:      fmt.Sprintf("user-%s@example.com", paymentEvent.UserID.String()), // In real app, get from user service
		Subject: "Payment Failed",
		Body:    fmt.Sprintf("Your payment of %s %s has failed. Reason: %s", paymentEvent.Amount, paymentEvent.Currency, paymentEvent.ErrorMessage),
		From:    "payments@bazaruto.com",
	}

	if err := h.dispatcher.PerformWithContext(ctx, emailJob); err != nil {
		h.logger.Error("Failed to dispatch payment failure notification email job",
			zap.Error(err),
			zap.String("payment_id", paymentEvent.PaymentID.String()))
		return fmt.Errorf("failed to dispatch payment failure notification email job: %w", err)
	}

	h.logger.Info("Payment failure notification email job dispatched successfully",
		zap.String("payment_id", paymentEvent.PaymentID.String()))

	return nil
}

// CanHandle returns true if this handler can process the given event type.
func (h *PaymentFailedHandler) CanHandle(eventType string) bool {
	return eventType == "payment.failed"
}

// HandlerName returns a unique name for this handler.
func (h *PaymentFailedHandler) HandlerName() string {
	return "payment_failed_handler"
}
