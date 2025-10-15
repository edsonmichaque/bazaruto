package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CoverageStore defines the interface for coverage data operations.
type CoverageStore interface {
	CreateCoverage(ctx context.Context, coverage *models.Coverage) error
	GetCoverage(ctx context.Context, id uuid.UUID) (*models.Coverage, error)
	ListCoverages(ctx context.Context, productID *uuid.UUID, coverageType string, limit, offset int) ([]*models.Coverage, error)
	UpdateCoverage(ctx context.Context, coverage *models.Coverage) error
	DeleteCoverage(ctx context.Context, id uuid.UUID) error
	CountCoverages(ctx context.Context, productID *uuid.UUID, coverageType string) (int64, error)
}

// coverageStore implements CoverageStore interface.
type coverageStore struct {
	db *gorm.DB
}

// NewCoverageStore creates a new CoverageStore instance.
func NewCoverageStore(db *gorm.DB) CoverageStore {
	return &coverageStore{db: db}
}

// CreateCoverage creates a new coverage.
func (s *coverageStore) CreateCoverage(ctx context.Context, coverage *models.Coverage) error {
	if err := s.db.WithContext(ctx).Create(coverage).Error; err != nil {
		return fmt.Errorf("failed to create coverage: %w", err)
	}
	return nil
}

// GetCoverage retrieves a coverage by ID.
func (s *coverageStore) GetCoverage(ctx context.Context, id uuid.UUID) (*models.Coverage, error) {
	var coverage models.Coverage
	if err := s.db.WithContext(ctx).Preload("Product").First(&coverage, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("coverage not found")
		}
		return nil, fmt.Errorf("failed to get coverage: %w", err)
	}
	return &coverage, nil
}

// ListCoverages retrieves a list of coverages with optional filtering.
func (s *coverageStore) ListCoverages(ctx context.Context, productID *uuid.UUID, coverageType string, limit, offset int) ([]*models.Coverage, error) {
	var coverages []*models.Coverage
	query := s.db.WithContext(ctx).Model(&models.Coverage{}).Preload("Product")

	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}
	if coverageType != "" {
		query = query.Where("coverage_type = ?", coverageType)
	}

	query = query.Order("sort_order ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&coverages).Error; err != nil {
		return nil, fmt.Errorf("failed to list coverages: %w", err)
	}
	return coverages, nil
}

// UpdateCoverage updates an existing coverage.
func (s *coverageStore) UpdateCoverage(ctx context.Context, coverage *models.Coverage) error {
	if err := s.db.WithContext(ctx).Save(coverage).Error; err != nil {
		return fmt.Errorf("failed to update coverage: %w", err)
	}
	return nil
}

// DeleteCoverage soft deletes a coverage.
func (s *coverageStore) DeleteCoverage(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Coverage{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete coverage: %w", err)
	}
	return nil
}

// CountCoverages returns the total number of coverages with optional filtering.
func (s *coverageStore) CountCoverages(ctx context.Context, productID *uuid.UUID, coverageType string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Coverage{})

	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}
	if coverageType != "" {
		query = query.Where("coverage_type = ?", coverageType)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count coverages: %w", err)
	}
	return count, nil
}
