package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BeneficiaryStore defines the interface for beneficiary data operations.
type BeneficiaryStore interface {
	CreateBeneficiary(ctx context.Context, beneficiary *models.Beneficiary) error
	GetBeneficiary(ctx context.Context, id uuid.UUID) (*models.Beneficiary, error)
	ListBeneficiaries(ctx context.Context, policyID *uuid.UUID, limit, offset int) ([]*models.Beneficiary, error)
	UpdateBeneficiary(ctx context.Context, beneficiary *models.Beneficiary) error
	DeleteBeneficiary(ctx context.Context, id uuid.UUID) error
	CountBeneficiaries(ctx context.Context, policyID *uuid.UUID) (int64, error)
}

// beneficiaryStore implements BeneficiaryStore interface.
type beneficiaryStore struct {
	db *gorm.DB
}

// NewBeneficiaryStore creates a new BeneficiaryStore instance.
func NewBeneficiaryStore(db *gorm.DB) BeneficiaryStore {
	return &beneficiaryStore{db: db}
}

// CreateBeneficiary creates a new beneficiary.
func (s *beneficiaryStore) CreateBeneficiary(ctx context.Context, beneficiary *models.Beneficiary) error {
	if err := s.db.WithContext(ctx).Create(beneficiary).Error; err != nil {
		return fmt.Errorf("failed to create beneficiary: %w", err)
	}
	return nil
}

// GetBeneficiary retrieves a beneficiary by ID.
func (s *beneficiaryStore) GetBeneficiary(ctx context.Context, id uuid.UUID) (*models.Beneficiary, error) {
	var beneficiary models.Beneficiary
	if err := s.db.WithContext(ctx).Preload("Policy").First(&beneficiary, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("beneficiary not found")
		}
		return nil, fmt.Errorf("failed to get beneficiary: %w", err)
	}
	return &beneficiary, nil
}

// ListBeneficiaries retrieves a list of beneficiaries with optional filtering.
func (s *beneficiaryStore) ListBeneficiaries(ctx context.Context, policyID *uuid.UUID, limit, offset int) ([]*models.Beneficiary, error) {
	var beneficiaries []*models.Beneficiary
	query := s.db.WithContext(ctx).Model(&models.Beneficiary{}).Preload("Policy")

	if policyID != nil {
		query = query.Where("policy_id = ?", *policyID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&beneficiaries).Error; err != nil {
		return nil, fmt.Errorf("failed to list beneficiaries: %w", err)
	}
	return beneficiaries, nil
}

// UpdateBeneficiary updates an existing beneficiary.
func (s *beneficiaryStore) UpdateBeneficiary(ctx context.Context, beneficiary *models.Beneficiary) error {
	if err := s.db.WithContext(ctx).Save(beneficiary).Error; err != nil {
		return fmt.Errorf("failed to update beneficiary: %w", err)
	}
	return nil
}

// DeleteBeneficiary soft deletes a beneficiary.
func (s *beneficiaryStore) DeleteBeneficiary(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Beneficiary{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete beneficiary: %w", err)
	}
	return nil
}

// CountBeneficiaries returns the total number of beneficiaries with optional filtering.
func (s *beneficiaryStore) CountBeneficiaries(ctx context.Context, policyID *uuid.UUID) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Beneficiary{})

	if policyID != nil {
		query = query.Where("policy_id = ?", *policyID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count beneficiaries: %w", err)
	}
	return count, nil
}
