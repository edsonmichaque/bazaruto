package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionStore defines the interface for subscription data operations.
type SubscriptionStore interface {
	CreateSubscription(ctx context.Context, subscription *models.Subscription) error
	GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	GetSubscriptionByNumber(ctx context.Context, subscriptionNumber string) (*models.Subscription, error)
	ListSubscriptions(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string, limit, offset int) ([]*models.Subscription, error)
	UpdateSubscription(ctx context.Context, subscription *models.Subscription) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	CountSubscriptions(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string) (int64, error)
}

// subscriptionStore implements SubscriptionStore interface.
type subscriptionStore struct {
	db *gorm.DB
}

// NewSubscriptionStore creates a new SubscriptionStore instance.
func NewSubscriptionStore(db *gorm.DB) SubscriptionStore {
	return &subscriptionStore{db: db}
}

// CreateSubscription creates a new subscription.
func (s *subscriptionStore) CreateSubscription(ctx context.Context, subscription *models.Subscription) error {
	if err := s.db.WithContext(ctx).Create(subscription).Error; err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

// GetSubscription retrieves a subscription by ID.
func (s *subscriptionStore) GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := s.db.WithContext(ctx).Preload("User").Preload("Policy").First(&subscription, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &subscription, nil
}

// GetSubscriptionByNumber retrieves a subscription by subscription number.
func (s *subscriptionStore) GetSubscriptionByNumber(ctx context.Context, subscriptionNumber string) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := s.db.WithContext(ctx).Preload("User").Preload("Policy").First(&subscription, "subscription_number = ?", subscriptionNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription by number: %w", err)
	}
	return &subscription, nil
}

// ListSubscriptions retrieves a list of subscriptions with optional filtering.
func (s *subscriptionStore) ListSubscriptions(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string, limit, offset int) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	query := s.db.WithContext(ctx).Model(&models.Subscription{}).Preload("User").Preload("Policy")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if policyID != nil {
		query = query.Where("policy_id = ?", *policyID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	return subscriptions, nil
}

// UpdateSubscription updates an existing subscription.
func (s *subscriptionStore) UpdateSubscription(ctx context.Context, subscription *models.Subscription) error {
	if err := s.db.WithContext(ctx).Save(subscription).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil
}

// DeleteSubscription soft deletes a subscription.
func (s *subscriptionStore) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

// CountSubscriptions returns the total number of subscriptions with optional filtering.
func (s *subscriptionStore) CountSubscriptions(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Subscription{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if policyID != nil {
		query = query.Where("policy_id = ?", *policyID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}
	return count, nil
}
