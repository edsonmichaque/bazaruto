package services

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// ProductService handles business logic for products.
type ProductService struct {
	store store.ProductStore
}

// NewProductService creates a new ProductService instance.
func NewProductService(store store.ProductStore) *ProductService {
	return &ProductService{
		store: store,
	}
}

// CreateProduct creates a new product with business logic validation.
func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	// Validate required fields
	if product.Name == "" {
		return fmt.Errorf("product name is required")
	}
	if product.Category == "" {
		return fmt.Errorf("product category is required")
	}
	if product.PartnerID == uuid.Nil {
		return fmt.Errorf("partner ID is required")
	}
	if product.BasePrice <= 0 {
		return fmt.Errorf("base price must be greater than 0")
	}
	if product.CoverageAmount <= 0 {
		return fmt.Errorf("coverage amount must be greater than 0")
	}
	if product.CoveragePeriod <= 0 {
		return fmt.Errorf("coverage period must be greater than 0")
	}

	// Set defaults
	if product.Currency == "" {
		product.Currency = models.CurrencyUSD
	}
	if product.Status == "" {
		product.Status = models.StatusActive
	}
	if product.EffectiveDate.IsZero() {
		product.EffectiveDate = time.Now()
	}

	// Validate effective date is not in the past
	if product.EffectiveDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return fmt.Errorf("effective date cannot be in the past")
	}

	// Validate expiration date if provided
	if product.ExpirationDate != nil && product.ExpirationDate.Before(product.EffectiveDate) {
		return fmt.Errorf("expiration date must be after effective date")
	}

	return s.store.CreateProduct(ctx, product)
}

// GetProduct retrieves a product by ID.
func (s *ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("product ID is required")
	}

	product, err := s.store.GetProduct(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Check if product is active
	if product.Status != models.StatusActive {
		return nil, fmt.Errorf("product is not active")
	}

	// Check if product has expired
	if product.ExpirationDate != nil && product.ExpirationDate.Before(time.Now()) {
		return nil, fmt.Errorf("product has expired")
	}

	return product, nil
}

// ListProducts retrieves a list of products with filtering.
func (s *ProductService) ListProducts(ctx context.Context, opts *models.ProductListOptions) ([]*models.Product, error) {
	if opts == nil {
		opts = models.NewProductListOptions()
	}

	// Validate and set defaults
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PerPage < 1 {
		opts.PerPage = 20
	}
	if opts.PerPage > 100 {
		opts.PerPage = 100
	}

	// Calculate offset from page and per_page
	offset := (opts.Page - 1) * opts.PerPage

	products, err := s.store.ListProducts(ctx, opts.PartnerID, opts.Category, opts.PerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Filter out inactive and expired products
	var activeProducts []*models.Product
	now := time.Now()
	for _, product := range products {
		if product.Status == models.StatusActive {
			if product.ExpirationDate == nil || product.ExpirationDate.After(now) {
				activeProducts = append(activeProducts, product)
			}
		}
	}

	return activeProducts, nil
}

// UpdateProduct updates an existing product.
func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
	if product.ID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}

	// Validate required fields
	if product.Name == "" {
		return fmt.Errorf("product name is required")
	}
	if product.Category == "" {
		return fmt.Errorf("product category is required")
	}
	if product.BasePrice <= 0 {
		return fmt.Errorf("base price must be greater than 0")
	}

	// Get existing product to validate changes
	existing, err := s.store.GetProduct(ctx, product.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing product: %w", err)
	}

	// Prevent changes to critical fields
	if existing.PartnerID != product.PartnerID {
		return fmt.Errorf("cannot change partner ID")
	}

	return s.store.UpdateProduct(ctx, product)
}

// DeleteProduct soft deletes a product.
func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}

	// Check if product exists
	_, err := s.store.GetProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	return s.store.DeleteProduct(ctx, id)
}

// CountProducts returns the total number of products with filtering.
func (s *ProductService) CountProducts(ctx context.Context, opts *models.ProductListOptions) (int64, error) {
	if opts == nil {
		opts = models.NewProductListOptions()
	}
	return s.store.CountProducts(ctx, opts.PartnerID, opts.Category)
}
