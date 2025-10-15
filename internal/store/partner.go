package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PartnerStore defines the interface for partner data operations.
type PartnerStore interface {
	CreatePartner(ctx context.Context, partner *models.Partner) error
	GetPartner(ctx context.Context, id uuid.UUID) (*models.Partner, error)
	GetPartnerByLicense(ctx context.Context, licenseNumber string) (*models.Partner, error)
	ListPartners(ctx context.Context, limit, offset int) ([]*models.Partner, error)
	UpdatePartner(ctx context.Context, partner *models.Partner) error
	DeletePartner(ctx context.Context, id uuid.UUID) error
	CountPartners(ctx context.Context) (int64, error)
}

// partnerStore implements PartnerStore interface.
type partnerStore struct {
	db *gorm.DB
}

// NewPartnerStore creates a new PartnerStore instance.
func NewPartnerStore(db *gorm.DB) PartnerStore {
	return &partnerStore{db: db}
}

// CreatePartner creates a new partner.
func (s *partnerStore) CreatePartner(ctx context.Context, partner *models.Partner) error {
	if err := s.db.WithContext(ctx).Create(partner).Error; err != nil {
		return fmt.Errorf("failed to create partner: %w", err)
	}
	return nil
}

// GetPartner retrieves a partner by ID.
func (s *partnerStore) GetPartner(ctx context.Context, id uuid.UUID) (*models.Partner, error) {
	var partner models.Partner
	if err := s.db.WithContext(ctx).First(&partner, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("partner not found")
		}
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}
	return &partner, nil
}

// GetPartnerByLicense retrieves a partner by license number.
func (s *partnerStore) GetPartnerByLicense(ctx context.Context, licenseNumber string) (*models.Partner, error) {
	var partner models.Partner
	if err := s.db.WithContext(ctx).First(&partner, "license_number = ?", licenseNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("partner not found")
		}
		return nil, fmt.Errorf("failed to get partner by license: %w", err)
	}
	return &partner, nil
}

// ListPartners retrieves a list of partners with pagination.
func (s *partnerStore) ListPartners(ctx context.Context, limit, offset int) ([]*models.Partner, error) {
	var partners []*models.Partner
	query := s.db.WithContext(ctx).Model(&models.Partner{})

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&partners).Error; err != nil {
		return nil, fmt.Errorf("failed to list partners: %w", err)
	}
	return partners, nil
}

// UpdatePartner updates an existing partner.
func (s *partnerStore) UpdatePartner(ctx context.Context, partner *models.Partner) error {
	if err := s.db.WithContext(ctx).Save(partner).Error; err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}
	return nil
}

// DeletePartner soft deletes a partner.
func (s *partnerStore) DeletePartner(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Partner{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete partner: %w", err)
	}
	return nil
}

// CountPartners returns the total number of partners.
func (s *partnerStore) CountPartners(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Partner{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count partners: %w", err)
	}
	return count, nil
}
