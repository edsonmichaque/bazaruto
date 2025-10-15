package services

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// WebhookService handles business logic for webhook configurations and deliveries.
type WebhookService struct {
	store        store.WebhookStore
	eventService *EventService
}

// NewWebhookService creates a new WebhookService instance.
func NewWebhookService(store store.WebhookStore, eventService ...*EventService) *WebhookService {
	var evtService *EventService
	if len(eventService) > 0 {
		evtService = eventService[0]
	}
	return &WebhookService{
		store:        store,
		eventService: evtService,
	}
}

// CreateWebhookConfig creates a new webhook configuration.
func (s *WebhookService) CreateWebhookConfig(ctx context.Context, config *models.WebhookConfig) error {
	// Validate required fields
	if config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if len(config.EventTypes) == 0 {
		return fmt.Errorf("at least one event type is required")
	}
	if config.Method == "" {
		config.Method = "POST" // Default to POST
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3 // Default retry count
	}
	if config.Timeout == 0 {
		config.Timeout = 30 // Default timeout in seconds
	}

	// Set defaults
	if config.IsActive {
		config.IsActive = true // Default to active
	}

	return s.store.CreateWebhookConfig(ctx, config)
}

// GetWebhookConfig retrieves a webhook configuration by ID.
func (s *WebhookService) GetWebhookConfig(ctx context.Context, id uuid.UUID) (*models.WebhookConfig, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("webhook config ID is required")
	}

	return s.store.GetWebhookConfig(ctx, id)
}

// UpdateWebhookConfig updates an existing webhook configuration.
func (s *WebhookService) UpdateWebhookConfig(ctx context.Context, config *models.WebhookConfig) error {
	if config.ID == uuid.Nil {
		return fmt.Errorf("webhook config ID is required")
	}

	return s.store.UpdateWebhookConfig(ctx, config)
}

// DeleteWebhookConfig deletes a webhook configuration.
func (s *WebhookService) DeleteWebhookConfig(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("webhook config ID is required")
	}

	return s.store.DeleteWebhookConfig(ctx, id)
}

// ListWebhookConfigs retrieves webhook configurations with optional filtering.
func (s *WebhookService) ListWebhookConfigs(ctx context.Context, opts *models.ListOptions) (*models.ListResponse[*models.WebhookConfig], error) {
	if opts == nil {
		opts = models.DefaultListOptions()
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid list options: %w", err)
	}

	// Extract webhook-specific filters
	var partnerID *uuid.UUID
	var isActive *bool

	if opts.HasFilter("partner_id") {
		if id, ok := opts.GetFilter("partner_id"); ok {
			if uuidVal, ok := id.(uuid.UUID); ok {
				partnerID = &uuidVal
			}
		}
	}
	if opts.HasFilter("is_active") {
		if active, ok := opts.GetFilter("is_active"); ok {
			if boolVal, ok := active.(bool); ok {
				isActive = &boolVal
			}
		}
	}

	configs, err := s.store.ListWebhookConfigs(ctx, partnerID, isActive, opts.GetLimit(), opts.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook configs: %w", err)
	}

	// For now, we'll use a placeholder total count
	// In a real implementation, you'd add a CountWebhookConfigs method to the store
	total := int64(len(configs))

	return models.NewListResponse(configs, total, opts), nil
}

// GetWebhookConfigsForEvent retrieves all active webhook configurations that should receive a specific event type.
func (s *WebhookService) GetWebhookConfigsForEvent(ctx context.Context, eventType string) ([]*models.WebhookConfig, error) {
	return s.store.GetWebhookConfigsForEvent(ctx, eventType)
}

// CreateWebhookDelivery creates a new webhook delivery record.
func (s *WebhookService) CreateWebhookDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	// Validate required fields
	if delivery.WebhookConfigID == uuid.Nil {
		return fmt.Errorf("webhook config ID is required")
	}
	if delivery.EventID == uuid.Nil {
		return fmt.Errorf("event ID is required")
	}
	if delivery.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	if delivery.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	// Set defaults
	if delivery.Status == "" {
		delivery.Status = models.WebhookStatusPending
	}
	if delivery.Method == "" {
		delivery.Method = "POST"
	}

	return s.store.CreateWebhookDelivery(ctx, delivery)
}

// GetWebhookDelivery retrieves a webhook delivery by ID.
func (s *WebhookService) GetWebhookDelivery(ctx context.Context, id uuid.UUID) (*models.WebhookDelivery, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("webhook delivery ID is required")
	}

	return s.store.GetWebhookDelivery(ctx, id)
}

// UpdateWebhookDelivery updates a webhook delivery record.
func (s *WebhookService) UpdateWebhookDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	if delivery.ID == uuid.Nil {
		return fmt.Errorf("webhook delivery ID is required")
	}

	return s.store.UpdateWebhookDelivery(ctx, delivery)
}

// GetWebhookDeliveries retrieves webhook delivery records with optional filtering.
func (s *WebhookService) GetWebhookDeliveries(ctx context.Context, opts *models.WebhookListOptions) (*models.ListResponse[*models.WebhookDelivery], error) {
	if opts == nil {
		opts = models.NewWebhookListOptions()
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid list options: %w", err)
	}

	deliveries, err := s.store.GetWebhookDeliveries(ctx, opts.WebhookConfigID, opts.Status, opts.GetLimit(), opts.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook deliveries: %w", err)
	}

	total, err := s.store.CountWebhookDeliveries(ctx, opts.WebhookConfigID, opts.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to count webhook deliveries: %w", err)
	}

	return models.NewListResponse(deliveries, total, opts.ListOptions), nil
}

// CountWebhookDeliveries returns the total number of webhook deliveries with optional filtering.
func (s *WebhookService) CountWebhookDeliveries(ctx context.Context, webhookConfigID *uuid.UUID, status string) (int64, error) {
	return s.store.CountWebhookDeliveries(ctx, webhookConfigID, status)
}
