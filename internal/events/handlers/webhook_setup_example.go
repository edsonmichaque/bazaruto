package handlers

import (
	"context"

	"go.uber.org/zap"

	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/pkg/job"
)

// SetupWebhookSystem demonstrates how to set up the webhook system
func SetupWebhookSystem(
	eventService *services.EventService,
	webhookService *services.WebhookService,
	dispatcher job.Dispatcher,
	logger *logger.Logger,
) error {
	// Create the webhook event handler
	webhookHandler := NewWebhookEventHandler(webhookService, dispatcher, logger)

	// Subscribe the webhook handler to multiple event types
	eventTypes := []string{
		"user.registered",
		"user.logged_in",
		"quote.created",
		"quote.calculated",
		"payment.initiated",
		"payment.completed",
		"payment.failed",
		"policy.created",
		"claim.submitted",
	}

	// Subscribe to all event types
	if err := eventService.SubscribeHandler(webhookHandler, eventTypes...); err != nil {
		return err
	}

	logger.Info("Webhook event handler subscribed to events",
		zap.Strings("event_types", eventTypes))

	return nil
}

// CreateExampleWebhookConfigs demonstrates how to create webhook configurations
func CreateExampleWebhookConfigs(ctx context.Context, webhookService *services.WebhookService) error {
	// Example 1: CRM Integration - Listen to user events
	crmWebhook := &models.WebhookConfig{
		URL:        "https://crm.example.com/webhooks/bazaruto",
		Method:     "POST",
		EventTypes: []string{"user.registered", "user.logged_in"},
		Headers: map[string]string{
			"Authorization": "Bearer crm-api-token",
			"X-Source":      "bazaruto",
		},
		Secret:      "crm-webhook-secret",
		IsActive:    true,
		RetryCount:  3,
		Timeout:     30,
		Description: "CRM integration for user events",
		Tags:        "crm,user,integration",
	}

	if err := webhookService.CreateWebhookConfig(ctx, crmWebhook); err != nil {
		return err
	}

	// Example 2: Analytics Integration - Listen to payment events
	analyticsWebhook := &models.WebhookConfig{
		URL:        "https://analytics.example.com/webhooks/payments",
		Method:     "POST",
		EventTypes: []string{"payment.completed", "payment.failed"},
		Headers: map[string]string{
			"X-API-Key": "analytics-api-key",
		},
		IsActive:    true,
		RetryCount:  5,
		Timeout:     45,
		Description: "Analytics tracking for payment events",
		Tags:        "analytics,payments,tracking",
	}

	if err := webhookService.CreateWebhookConfig(ctx, analyticsWebhook); err != nil {
		return err
	}

	// Example 3: Partner Integration - Listen to quote events
	partnerWebhook := &models.WebhookConfig{
		URL:        "https://partner.example.com/api/webhooks/quotes",
		Method:     "POST",
		EventTypes: []string{"quote.created", "quote.calculated"},
		Headers: map[string]string{
			"Authorization": "Bearer partner-token",
			"Content-Type":  "application/json",
		},
		Secret:      "partner-webhook-secret",
		IsActive:    true,
		RetryCount:  3,
		Timeout:     30,
		Description: "Partner integration for quote events",
		Tags:        "partner,quotes,integration",
	}

	if err := webhookService.CreateWebhookConfig(ctx, partnerWebhook); err != nil {
		return err
	}

	return nil
}

// ExampleWebhookPayloads shows what the webhook payloads look like
func ExampleWebhookPayloads() {
	// Example payload for user.registered event
	userRegisteredPayload := map[string]interface{}{
		"event": map[string]interface{}{
			"id":             "123e4567-e89b-12d3-a456-426614174000",
			"type":           "user.registered",
			"aggregate_id":   "123e4567-e89b-12d3-a456-426614174001",
			"aggregate_type": "user",
			"version":        1,
			"occurred_at":    "2024-01-15T10:30:00Z",
			"data": map[string]interface{}{
				"user_id":   "123e4567-e89b-12d3-a456-426614174001",
				"email":     "user@example.com",
				"full_name": "John Doe",
				"role":      "customer",
			},
		},
		"user": map[string]interface{}{
			"id":        "123e4567-e89b-12d3-a456-426614174001",
			"email":     "user@example.com",
			"full_name": "John Doe",
			"role":      "customer",
		},
		"timestamp": "2024-01-15T10:30:00Z",
	}

	// Example payload for payment.completed event
	paymentCompletedPayload := map[string]interface{}{
		"event": map[string]interface{}{
			"id":             "123e4567-e89b-12d3-a456-426614174002",
			"type":           "payment.completed",
			"aggregate_id":   "123e4567-e89b-12d3-a456-426614174003",
			"aggregate_type": "payment",
			"version":        2,
			"occurred_at":    "2024-01-15T10:35:00Z",
			"data": map[string]interface{}{
				"payment_id":     "123e4567-e89b-12d3-a456-426614174003",
				"user_id":        "123e4567-e89b-12d3-a456-426614174001",
				"amount":         99.99,
				"currency":       "USD",
				"transaction_id": "txn_1234567890",
				"completed_at":   "2024-01-15T10:35:00Z",
			},
		},
		"payment": map[string]interface{}{
			"id":             "123e4567-e89b-12d3-a456-426614174003",
			"user_id":        "123e4567-e89b-12d3-a456-426614174001",
			"amount":         99.99,
			"currency":       "USD",
			"transaction_id": "txn_1234567890",
			"completed_at":   "2024-01-15T10:35:00Z",
		},
		"timestamp": "2024-01-15T10:35:00Z",
	}

	// These payloads would be sent to the configured webhook URLs
	_ = userRegisteredPayload
	_ = paymentCompletedPayload
}
