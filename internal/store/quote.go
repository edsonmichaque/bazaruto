package store

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QuoteStore defines the interface for quote data operations.
type QuoteStore interface {
	CreateQuote(ctx context.Context, quote *models.Quote) error
	GetQuote(ctx context.Context, id uuid.UUID) (*models.Quote, error)
	GetQuoteByNumber(ctx context.Context, quoteNumber string) (*models.Quote, error)
	ListQuotes(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string, limit, offset int) ([]*models.Quote, error)
	UpdateQuote(ctx context.Context, quote *models.Quote) error
	DeleteQuote(ctx context.Context, id uuid.UUID) error
	ExpireQuote(ctx context.Context, id uuid.UUID) error
	CountQuotes(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string) (int64, error)
}

// quoteStore implements QuoteStore interface.
type quoteStore struct {
	db *gorm.DB
}

// NewQuoteStore creates a new QuoteStore instance.
func NewQuoteStore(db *gorm.DB) QuoteStore {
	return &quoteStore{db: db}
}

// CreateQuote creates a new quote.
func (s *quoteStore) CreateQuote(ctx context.Context, quote *models.Quote) error {
	if err := s.db.WithContext(ctx).Create(quote).Error; err != nil {
		return fmt.Errorf("failed to create quote: %w", err)
	}
	return nil
}

// GetQuote retrieves a quote by ID.
func (s *quoteStore) GetQuote(ctx context.Context, id uuid.UUID) (*models.Quote, error) {
	var quote models.Quote
	if err := s.db.WithContext(ctx).Preload("Product").Preload("User").First(&quote, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("quote not found")
		}
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	return &quote, nil
}

// GetQuoteByNumber retrieves a quote by quote number.
func (s *quoteStore) GetQuoteByNumber(ctx context.Context, quoteNumber string) (*models.Quote, error) {
	var quote models.Quote
	if err := s.db.WithContext(ctx).Preload("Product").Preload("User").First(&quote, "quote_number = ?", quoteNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("quote not found")
		}
		return nil, fmt.Errorf("failed to get quote by number: %w", err)
	}
	return &quote, nil
}

// ListQuotes retrieves a list of quotes with optional filtering.
func (s *quoteStore) ListQuotes(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string, limit, offset int) ([]*models.Quote, error) {
	var quotes []*models.Quote
	query := s.db.WithContext(ctx).Model(&models.Quote{}).Preload("Product").Preload("User")

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

	if err := query.Find(&quotes).Error; err != nil {
		return nil, fmt.Errorf("failed to list quotes: %w", err)
	}
	return quotes, nil
}

// UpdateQuote updates an existing quote.
func (s *quoteStore) UpdateQuote(ctx context.Context, quote *models.Quote) error {
	if err := s.db.WithContext(ctx).Save(quote).Error; err != nil {
		return fmt.Errorf("failed to update quote: %w", err)
	}
	return nil
}

// DeleteQuote soft deletes a quote.
func (s *quoteStore) DeleteQuote(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Quote{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete quote: %w", err)
	}
	return nil
}

// ExpireQuote marks a quote as expired.
func (s *quoteStore) ExpireQuote(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Model(&models.Quote{}).Where("id = ?", id).Update("status", models.QuoteStatusExpired).Error; err != nil {
		return fmt.Errorf("failed to expire quote: %w", err)
	}
	return nil
}

// CountQuotes returns the total number of quotes with optional filtering.
func (s *quoteStore) CountQuotes(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&models.Quote{})

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
		return 0, fmt.Errorf("failed to count quotes: %w", err)
	}
	return count, nil
}
