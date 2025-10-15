package handlers

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/edsonmichaque/bazaruto/internal/events"
	"github.com/edsonmichaque/bazaruto/internal/jobs"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/pkg/event"
	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/google/uuid"
)

// WebhookEventHandler handles multiple event types and creates webhook jobs for external integrations
type WebhookEventHandler struct {
	webhookService *services.WebhookService
	dispatcher     job.Dispatcher
	logger         *logger.Logger
}

// NewWebhookEventHandler creates a new webhook event handler.
func NewWebhookEventHandler(webhookService *services.WebhookService, dispatcher job.Dispatcher, logger *logger.Logger) *WebhookEventHandler {
	return &WebhookEventHandler{
		webhookService: webhookService,
		dispatcher:     dispatcher,
		logger:         logger,
	}
}

// Handle processes events and creates webhook jobs for configured webhooks.
func (h *WebhookEventHandler) Handle(ctx context.Context, event event.Event) error {
	eventType := event.Type()

	h.logger.Info("Processing event for webhook delivery",
		zap.String("event_type", eventType),
		zap.String("event_id", event.ID().String()),
		zap.String("aggregate_id", event.AggregateID().String()))

	// Get all webhook configurations that should receive this event type
	webhookConfigs, err := h.webhookService.GetWebhookConfigsForEvent(ctx, eventType)
	if err != nil {
		h.logger.Error("Failed to get webhook configurations for event",
			zap.Error(err),
			zap.String("event_type", eventType))
		return fmt.Errorf("failed to get webhook configurations: %w", err)
	}

	if len(webhookConfigs) == 0 {
		h.logger.Debug("No webhook configurations found for event type",
			zap.String("event_type", eventType))
		return nil
	}

	h.logger.Info("Found webhook configurations for event",
		zap.String("event_type", eventType),
		zap.Int("webhook_count", len(webhookConfigs)))

	// Create webhook jobs for each configuration
	for _, config := range webhookConfigs {
		if err := h.createWebhookJob(ctx, event, config); err != nil {
			h.logger.Error("Failed to create webhook job",
				zap.Error(err),
				zap.String("event_type", eventType),
				zap.String("webhook_url", config.URL))
			// Continue with other webhooks even if one fails
			continue
		}
	}

	return nil
}

// createWebhookJob creates a webhook job and delivery record for a specific configuration
func (h *WebhookEventHandler) createWebhookJob(ctx context.Context, event event.Event, config *models.WebhookConfig) error {
	// Create webhook delivery record
	delivery := &models.WebhookDelivery{
		WebhookConfigID: config.ID,
		EventID:         event.ID(),
		EventType:       event.Type(),
		URL:             config.URL,
		Method:          config.Method,
		Headers:         config.Headers,
		Payload:         h.buildWebhookPayload(event),
		Status:          models.WebhookStatusPending,
		AttemptCount:    0,
	}

	// Save delivery record
	if err := h.webhookService.CreateWebhookDelivery(ctx, delivery); err != nil {
		return fmt.Errorf("failed to create webhook delivery record: %w", err)
	}

	// Create webhook delivery job
	webhookJob := &jobs.WebhookDeliveryJob{
		ID:             uuid.New(),
		DeliveryID:     delivery.ID,
		URL:            config.URL,
		Method:         config.Method,
		Headers:        h.buildWebhookHeaders(config, event),
		Payload:        delivery.Payload,
		EventType:      event.Type(),
		EventID:        event.ID(),
		WebhookService: h.webhookService,
	}

	// Dispatch the webhook job
	if err := h.dispatcher.PerformWithContext(ctx, webhookJob); err != nil {
		// Update delivery status to failed
		delivery.Status = models.WebhookStatusFailed
		delivery.ErrorMessage = err.Error()
		now := time.Now()
		delivery.FailedAt = &now
		_ = h.webhookService.UpdateWebhookDelivery(ctx, delivery)

		return fmt.Errorf("failed to dispatch webhook job: %w", err)
	}

	h.logger.Info("Webhook job created and dispatched successfully",
		zap.String("event_type", event.Type()),
		zap.String("webhook_url", config.URL),
		zap.String("delivery_id", delivery.ID.String()))

	return nil
}

// buildWebhookPayload creates the payload for the webhook request
func (h *WebhookEventHandler) buildWebhookPayload(event event.Event) map[string]interface{} {
	payload := map[string]interface{}{
		"event": map[string]interface{}{
			"id":             event.ID().String(),
			"type":           event.Type(),
			"aggregate_id":   event.AggregateID().String(),
			"aggregate_type": event.AggregateType(),
			"version":        event.Version(),
			"occurred_at":    event.OccurredAt().Format(time.RFC3339),
			"data":           event.Data(),
			"metadata":       event.Metadata(),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Add event-specific data based on event type
	switch event.Type() {
	case "user.registered":
		if userEvent, ok := event.(*events.UserRegisteredEvent); ok {
			payload["user"] = map[string]interface{}{
				"id":        userEvent.UserID.String(),
				"email":     userEvent.Email,
				"full_name": userEvent.FullName,
				"role":      userEvent.Role,
			}
		}
	case "user.logged_in":
		if userEvent, ok := event.(*events.UserLoggedInEvent); ok {
			payload["user"] = map[string]interface{}{
				"id":         userEvent.UserID.String(),
				"email":      userEvent.Email,
				"login_time": userEvent.LoginTime.Format(time.RFC3339),
				"ip_address": userEvent.IPAddress,
			}
		}
	case "quote.created":
		if quoteEvent, ok := event.(*events.QuoteCreatedEvent); ok {
			payload["quote"] = map[string]interface{}{
				"id":              quoteEvent.QuoteID.String(),
				"user_id":         quoteEvent.UserID.String(),
				"product_id":      quoteEvent.ProductID.String(),
				"coverage_amount": quoteEvent.CoverageAmount,
				"currency":        quoteEvent.Currency,
			}
		}
	case "quote.calculated":
		if quoteEvent, ok := event.(*events.QuoteCalculatedEvent); ok {
			payload["quote"] = map[string]interface{}{
				"id":            quoteEvent.QuoteID.String(),
				"user_id":       quoteEvent.UserID.String(),
				"final_premium": quoteEvent.FinalPremium,
				"base_premium":  quoteEvent.BasePremium,
				"currency":      quoteEvent.Currency,
			}
		}
	case "payment.initiated":
		if paymentEvent, ok := event.(*events.PaymentInitiatedEvent); ok {
			payload["payment"] = map[string]interface{}{
				"id":       paymentEvent.PaymentID.String(),
				"user_id":  paymentEvent.UserID.String(),
				"amount":   paymentEvent.Amount,
				"currency": paymentEvent.Currency,
				"method":   paymentEvent.PaymentMethod,
			}
		}
	case "payment.completed":
		if paymentEvent, ok := event.(*events.PaymentCompletedEvent); ok {
			payload["payment"] = map[string]interface{}{
				"id":             paymentEvent.PaymentID.String(),
				"user_id":        paymentEvent.UserID.String(),
				"amount":         paymentEvent.Amount,
				"currency":       paymentEvent.Currency,
				"transaction_id": paymentEvent.TransactionID,
				"completed_at":   paymentEvent.CompletedAt.Format(time.RFC3339),
			}
		}
	case "payment.failed":
		if paymentEvent, ok := event.(*events.PaymentFailedEvent); ok {
			payload["payment"] = map[string]interface{}{
				"id":             paymentEvent.PaymentID.String(),
				"user_id":        paymentEvent.UserID.String(),
				"amount":         paymentEvent.Amount,
				"currency":       paymentEvent.Currency,
				"failure_reason": paymentEvent.ErrorMessage,
				"failed_at":      paymentEvent.FailedAt.Format(time.RFC3339),
			}
		}
	}

	return payload
}

// buildWebhookHeaders creates the headers for the webhook request
func (h *WebhookEventHandler) buildWebhookHeaders(config *models.WebhookConfig, event event.Event) map[string]string {
	headers := make(map[string]string)

	// Copy existing headers from config
	for key, value := range config.Headers {
		headers[key] = value
	}

	// Add webhook signature if secret is configured
	if config.Secret != "" {
		// In a real implementation, you would generate HMAC signature
		// For now, we'll add a placeholder
		headers["X-Webhook-Signature"] = "sha256=placeholder_signature"
	}

	// Add event-specific headers
	headers["X-Event-ID"] = event.ID().String()
	headers["X-Event-Type"] = event.Type()
	headers["X-Aggregate-ID"] = event.AggregateID().String()
	headers["X-Aggregate-Type"] = event.AggregateType()

	return headers
}

// CanHandle returns true if this handler can process the given event type.
// This handler can process multiple event types, so we check against a predefined list.
func (h *WebhookEventHandler) CanHandle(eventType string) bool {
	// Define which events should trigger webhooks
	webhookEvents := map[string]bool{
		"user.registered":   true,
		"user.logged_in":    true,
		"quote.created":     true,
		"quote.calculated":  true,
		"payment.initiated": true,
		"payment.completed": true,
		"payment.failed":    true,
		"policy.created":    true,
		"claim.submitted":   true,
	}

	return webhookEvents[eventType]
}

// HandlerName returns a unique name for this handler.
func (h *WebhookEventHandler) HandlerName() string {
	return "webhook_event_handler"
}
