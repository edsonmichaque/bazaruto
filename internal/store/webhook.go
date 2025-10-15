package store

import (
	"context"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebhookStore defines the interface for webhook data operations.
type WebhookStore interface {
	// WebhookConfig operations
	CreateWebhookConfig(ctx context.Context, config *models.WebhookConfig) error
	GetWebhookConfig(ctx context.Context, id uuid.UUID) (*models.WebhookConfig, error)
	UpdateWebhookConfig(ctx context.Context, config *models.WebhookConfig) error
	DeleteWebhookConfig(ctx context.Context, id uuid.UUID) error
	ListWebhookConfigs(ctx context.Context, partnerID *uuid.UUID, isActive *bool, limit, offset int) ([]*models.WebhookConfig, error)
	GetWebhookConfigsForEvent(ctx context.Context, eventType string) ([]*models.WebhookConfig, error)

	// WebhookDelivery operations
	CreateWebhookDelivery(ctx context.Context, delivery *models.WebhookDelivery) error
	GetWebhookDelivery(ctx context.Context, id uuid.UUID) (*models.WebhookDelivery, error)
	UpdateWebhookDelivery(ctx context.Context, delivery *models.WebhookDelivery) error
	GetWebhookDeliveries(ctx context.Context, webhookConfigID *uuid.UUID, status string, limit, offset int) ([]*models.WebhookDelivery, error)
	CountWebhookDeliveries(ctx context.Context, webhookConfigID *uuid.UUID, status string) (int64, error)
}

// webhookStore implements WebhookStore interface.
type webhookStore struct {
	db *gorm.DB
}

// NewWebhookStore creates a new WebhookStore instance.
func NewWebhookStore(db *gorm.DB) WebhookStore {
	return &webhookStore{db: db}
}

// CreateWebhookConfig creates a new webhook configuration.
func (s *webhookStore) CreateWebhookConfig(ctx context.Context, config *models.WebhookConfig) error {
	return s.db.WithContext(ctx).Create(config).Error
}

// GetWebhookConfig retrieves a webhook configuration by ID.
func (s *webhookStore) GetWebhookConfig(ctx context.Context, id uuid.UUID) (*models.WebhookConfig, error) {
	var config models.WebhookConfig
	err := s.db.WithContext(ctx).
		Preload("Partner").
		Where("id = ?", id).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateWebhookConfig updates an existing webhook configuration.
func (s *webhookStore) UpdateWebhookConfig(ctx context.Context, config *models.WebhookConfig) error {
	return s.db.WithContext(ctx).Save(config).Error
}

// DeleteWebhookConfig deletes a webhook configuration.
func (s *webhookStore) DeleteWebhookConfig(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&models.WebhookConfig{}).Error
}

// ListWebhookConfigs retrieves webhook configurations with optional filtering.
func (s *webhookStore) ListWebhookConfigs(ctx context.Context, partnerID *uuid.UUID, isActive *bool, limit, offset int) ([]*models.WebhookConfig, error) {
	var configs []*models.WebhookConfig
	query := s.db.WithContext(ctx).Preload("Partner")

	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	err := query.Limit(limit).Offset(offset).Find(&configs).Error
	return configs, err
}

// GetWebhookConfigsForEvent retrieves all active webhook configurations that should receive a specific event type.
func (s *webhookStore) GetWebhookConfigsForEvent(ctx context.Context, eventType string) ([]*models.WebhookConfig, error) {
	var configs []*models.WebhookConfig

	// Use JSONB contains operator to check if eventType is in the event_types array
	err := s.db.WithContext(ctx).
		Where("is_active = ? AND event_types @> ?", true, `["`+eventType+`"]`).
		Find(&configs).Error

	return configs, err
}

// CreateWebhookDelivery creates a new webhook delivery record.
func (s *webhookStore) CreateWebhookDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	return s.db.WithContext(ctx).Create(delivery).Error
}

// GetWebhookDelivery retrieves a webhook delivery by ID.
func (s *webhookStore) GetWebhookDelivery(ctx context.Context, id uuid.UUID) (*models.WebhookDelivery, error) {
	var delivery models.WebhookDelivery
	err := s.db.WithContext(ctx).
		Preload("WebhookConfig").
		Where("id = ?", id).
		First(&delivery).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

// UpdateWebhookDelivery updates a webhook delivery record.
func (s *webhookStore) UpdateWebhookDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	return s.db.WithContext(ctx).Save(delivery).Error
}

// GetWebhookDeliveries retrieves webhook delivery records with optional filtering.
func (s *webhookStore) GetWebhookDeliveries(ctx context.Context, webhookConfigID *uuid.UUID, status string, limit, offset int) ([]*models.WebhookDelivery, error) {
	var deliveries []*models.WebhookDelivery
	query := s.db.WithContext(ctx).Preload("WebhookConfig")

	if webhookConfigID != nil {
		query = query.Where("webhook_config_id = ?", *webhookConfigID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&deliveries).Error
	return deliveries, err
}

// CountWebhookDeliveries returns the total number of webhook deliveries with optional filtering.
func (s *webhookStore) CountWebhookDeliveries(ctx context.Context, webhookConfigID *uuid.UUID, status string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.WebhookDelivery{})

	if webhookConfigID != nil {
		query = query.Where("webhook_config_id = ?", *webhookConfigID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&count).Error
	return count, err
}
