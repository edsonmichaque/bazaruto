package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductStore defines the interface for product data operations.
type ProductStore interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error)
	ListProducts(ctx context.Context, partnerID *uuid.UUID, category string, limit, offset int) ([]*models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	CountProducts(ctx context.Context, partnerID *uuid.UUID, category string) (int64, error)
}

// productStore implements ProductStore interface.
type productStore struct {
	db *gorm.DB
}

// NewProductStore creates a new ProductStore instance.
func NewProductStore(db *gorm.DB) ProductStore {
	return &productStore{db: db}
}

// CreateProduct creates a new product.
func (s *productStore) CreateProduct(ctx context.Context, product *models.Product) error {
	if err := s.db.WithContext(ctx).Create(product).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

// GetProduct retrieves a product by ID.
func (s *productStore) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var product models.Product
	if err := s.db.WithContext(ctx).Preload("Partner").Preload("Coverages").First(&product, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

// ListProducts retrieves a list of products with optional filtering.
func (s *productStore) ListProducts(ctx context.Context, partnerID *uuid.UUID, category string, limit, offset int) ([]*models.Product, error) {
	var products []*models.Product
	query := s.db.WithContext(ctx).Model(&models.Product{}).Preload("Partner")

	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	return products, nil
}

// UpdateProduct updates an existing product.
func (s *productStore) UpdateProduct(ctx context.Context, product *models.Product) error {
	if err := s.db.WithContext(ctx).Save(product).Error; err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

// DeleteProduct soft deletes a product.
func (s *productStore) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Product{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// CountProducts returns the total number of products with optional filtering.
func (s *productStore) CountProducts(ctx context.Context, partnerID *uuid.UUID, category string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Product{})

	if partnerID != nil {
		query = query.Where("partner_id = ?", *partnerID)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}
	return count, nil
}
