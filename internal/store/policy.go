package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PolicyStore defines the interface for policy data operations.
type PolicyStore interface {
	CreatePolicy(ctx context.Context, policy *models.Policy) error
	GetPolicy(ctx context.Context, id uuid.UUID) (*models.Policy, error)
	GetPolicyByNumber(ctx context.Context, policyNumber string) (*models.Policy, error)
	ListPolicies(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string, limit, offset int) ([]*models.Policy, error)
	UpdatePolicy(ctx context.Context, policy *models.Policy) error
	DeletePolicy(ctx context.Context, id uuid.UUID) error
	CountPolicies(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string) (int64, error)
}

// policyStore implements PolicyStore interface.
type policyStore struct {
	db *gorm.DB
}

// NewPolicyStore creates a new PolicyStore instance.
func NewPolicyStore(db *gorm.DB) PolicyStore {
	return &policyStore{db: db}
}

// CreatePolicy creates a new policy.
func (s *policyStore) CreatePolicy(ctx context.Context, policy *models.Policy) error {
	if err := s.db.WithContext(ctx).Create(policy).Error; err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}
	return nil
}

// GetPolicy retrieves a policy by ID.
func (s *policyStore) GetPolicy(ctx context.Context, id uuid.UUID) (*models.Policy, error) {
	var policy models.Policy
	if err := s.db.WithContext(ctx).Preload("Product").Preload("User").Preload("Quote").Preload("Beneficiaries").First(&policy, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("policy not found")
		}
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	return &policy, nil
}

// GetPolicyByNumber retrieves a policy by policy number.
func (s *policyStore) GetPolicyByNumber(ctx context.Context, policyNumber string) (*models.Policy, error) {
	var policy models.Policy
	if err := s.db.WithContext(ctx).Preload("Product").Preload("User").Preload("Quote").Preload("Beneficiaries").First(&policy, "policy_number = ?", policyNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("policy not found")
		}
		return nil, fmt.Errorf("failed to get policy by number: %w", err)
	}
	return &policy, nil
}

// ListPolicies retrieves a list of policies with optional filtering.
func (s *policyStore) ListPolicies(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string, limit, offset int) ([]*models.Policy, error) {
	var policies []*models.Policy
	query := s.db.WithContext(ctx).Model(&models.Policy{}).Preload("Product").Preload("User")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
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

	if err := query.Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	return policies, nil
}

// UpdatePolicy updates an existing policy.
func (s *policyStore) UpdatePolicy(ctx context.Context, policy *models.Policy) error {
	if err := s.db.WithContext(ctx).Save(policy).Error; err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}
	return nil
}

// DeletePolicy soft deletes a policy.
func (s *policyStore) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Policy{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}
	return nil
}

// CountPolicies returns the total number of policies with optional filtering.
func (s *policyStore) CountPolicies(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Policy{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count policies: %w", err)
	}
	return count, nil
}
