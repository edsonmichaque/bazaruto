package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ClaimStore defines the interface for claim data operations.
type ClaimStore interface {
	CreateClaim(ctx context.Context, claim *models.Claim) error
	GetClaim(ctx context.Context, id uuid.UUID) (*models.Claim, error)
	GetClaimByNumber(ctx context.Context, claimNumber string) (*models.Claim, error)
	ListClaims(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string, limit, offset int) ([]*models.Claim, error)
	UpdateClaim(ctx context.Context, claim *models.Claim) error
	DeleteClaim(ctx context.Context, id uuid.UUID) error
	CountClaims(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string) (int64, error)
}

// claimStore implements ClaimStore interface.
type claimStore struct {
	db *gorm.DB
}

// NewClaimStore creates a new ClaimStore instance.
func NewClaimStore(db *gorm.DB) ClaimStore {
	return &claimStore{db: db}
}

// CreateClaim creates a new claim.
func (s *claimStore) CreateClaim(ctx context.Context, claim *models.Claim) error {
	if err := s.db.WithContext(ctx).Create(claim).Error; err != nil {
		return fmt.Errorf("failed to create claim: %w", err)
	}
	return nil
}

// GetClaim retrieves a claim by ID.
func (s *claimStore) GetClaim(ctx context.Context, id uuid.UUID) (*models.Claim, error) {
	var claim models.Claim
	if err := s.db.WithContext(ctx).Preload("Policy").Preload("User").First(&claim, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("claim not found")
		}
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}
	return &claim, nil
}

// GetClaimByNumber retrieves a claim by claim number.
func (s *claimStore) GetClaimByNumber(ctx context.Context, claimNumber string) (*models.Claim, error) {
	var claim models.Claim
	if err := s.db.WithContext(ctx).Preload("Policy").Preload("User").First(&claim, "claim_number = ?", claimNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("claim not found")
		}
		return nil, fmt.Errorf("failed to get claim by number: %w", err)
	}
	return &claim, nil
}

// ListClaims retrieves a list of claims with optional filtering.
func (s *claimStore) ListClaims(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string, limit, offset int) ([]*models.Claim, error) {
	var claims []*models.Claim
	query := s.db.WithContext(ctx).Model(&models.Claim{}).Preload("Policy").Preload("User")

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

	if err := query.Find(&claims).Error; err != nil {
		return nil, fmt.Errorf("failed to list claims: %w", err)
	}
	return claims, nil
}

// UpdateClaim updates an existing claim.
func (s *claimStore) UpdateClaim(ctx context.Context, claim *models.Claim) error {
	if err := s.db.WithContext(ctx).Save(claim).Error; err != nil {
		return fmt.Errorf("failed to update claim: %w", err)
	}
	return nil
}

// DeleteClaim soft deletes a claim.
func (s *claimStore) DeleteClaim(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Claim{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete claim: %w", err)
	}
	return nil
}

// CountClaims returns the total number of claims with optional filtering.
func (s *claimStore) CountClaims(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Claim{})

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
		return 0, fmt.Errorf("failed to count claims: %w", err)
	}
	return count, nil
}
