package services

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/events"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// QuoteService handles business logic for quotes.
type QuoteService struct {
	store        store.QuoteStore
	eventService *EventService
}

// NewQuoteService creates a new QuoteService instance.
func NewQuoteService(store store.QuoteStore, eventService ...*EventService) *QuoteService {
	var evtService *EventService
	if len(eventService) > 0 {
		evtService = eventService[0]
	}
	return &QuoteService{
		store:        store,
		eventService: evtService,
	}
}

// CreateQuote creates a new quote with business logic validation.
func (s *QuoteService) CreateQuote(ctx context.Context, quote *models.Quote) error {
	// Validate required fields
	if quote.ProductID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}
	if quote.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if quote.BasePrice <= 0 {
		return fmt.Errorf("base price must be greater than 0")
	}
	if quote.FinalPrice <= 0 {
		return fmt.Errorf("final price must be greater than 0")
	}

	// Set defaults
	if quote.Currency == "" {
		quote.Currency = models.CurrencyUSD
	}
	if quote.Status == "" {
		quote.Status = models.QuoteStatusPending
	}
	if quote.ValidUntil.IsZero() {
		quote.ValidUntil = time.Now().Add(24 * time.Hour) // Default 24 hours validity
	}

	// Generate quote number if not provided
	if quote.QuoteNumber == "" {
		quote.QuoteNumber = s.generateQuoteNumber()
	}

	// Validate final price is not negative
	if quote.FinalPrice < 0 {
		return fmt.Errorf("final price cannot be negative")
	}

	// Validate discount is not negative
	if quote.Discount < 0 {
		return fmt.Errorf("discount cannot be negative")
	}

	// Validate tax is not negative
	if quote.Tax < 0 {
		return fmt.Errorf("tax cannot be negative")
	}

	// Validate validity period
	if quote.ValidUntil.Before(time.Now()) {
		return fmt.Errorf("valid until date cannot be in the past")
	}

	if err := s.store.CreateQuote(ctx, quote); err != nil {
		return fmt.Errorf("failed to create quote: %w", err)
	}

	// Publish quote created event
	if s.eventService != nil {
		event := events.NewQuoteCreatedEvent(quote.ID, quote.UserID, quote.ProductID, quote.BasePrice, quote.Currency, quote.ValidUntil)
		if err := s.eventService.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the quote creation
		}
	}

	return nil
}

// GetQuote retrieves a quote by ID.
func (s *QuoteService) GetQuote(ctx context.Context, id uuid.UUID) (*models.Quote, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("quote ID is required")
	}

	quote, err := s.store.GetQuote(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// Check if quote has expired
	if quote.ValidUntil.Before(time.Now()) && quote.Status == models.QuoteStatusPending {
		// Auto-expire the quote
		if err := s.store.ExpireQuote(ctx, id); err != nil {
			return nil, fmt.Errorf("failed to expire quote: %w", err)
		}
		quote.Status = models.QuoteStatusExpired
	}

	return quote, nil
}

// GetQuoteByNumber retrieves a quote by quote number.
func (s *QuoteService) GetQuoteByNumber(ctx context.Context, quoteNumber string) (*models.Quote, error) {
	if quoteNumber == "" {
		return nil, fmt.Errorf("quote number is required")
	}

	quote, err := s.store.GetQuoteByNumber(ctx, quoteNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote by number: %w", err)
	}

	// Check if quote has expired
	if quote.ValidUntil.Before(time.Now()) && quote.Status == models.QuoteStatusPending {
		// Auto-expire the quote
		if err := s.store.ExpireQuote(ctx, quote.ID); err != nil {
			return nil, fmt.Errorf("failed to expire quote: %w", err)
		}
		quote.Status = models.QuoteStatusExpired
	}

	return quote, nil
}

// ListQuotes retrieves a list of quotes with filtering.
func (s *QuoteService) ListQuotes(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string, limit, offset int) ([]*models.Quote, error) {
	// Validate pagination parameters
	if limit < 0 {
		limit = 0
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	quotes, err := s.store.ListQuotes(ctx, userID, productID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list quotes: %w", err)
	}

	// Auto-expire pending quotes that have passed their validity period
	now := time.Now()
	for _, quote := range quotes {
		if quote.Status == models.QuoteStatusPending && quote.ValidUntil.Before(now) {
			if err := s.store.ExpireQuote(ctx, quote.ID); err != nil {
				// Log error but continue processing
				continue
			}
			quote.Status = models.QuoteStatusExpired
		}
	}

	return quotes, nil
}

// UpdateQuote updates an existing quote.
func (s *QuoteService) UpdateQuote(ctx context.Context, quote *models.Quote) error {
	if quote.ID == uuid.Nil {
		return fmt.Errorf("quote ID is required")
	}

	// Get existing quote to validate changes
	existing, err := s.store.GetQuote(ctx, quote.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing quote: %w", err)
	}

	// Prevent changes to critical fields
	if existing.ProductID != quote.ProductID {
		return fmt.Errorf("cannot change product ID")
	}
	if existing.UserID != quote.UserID {
		return fmt.Errorf("cannot change user ID")
	}
	if existing.QuoteNumber != quote.QuoteNumber {
		return fmt.Errorf("cannot change quote number")
	}

	// Validate final price is not negative
	if quote.FinalPrice < 0 {
		return fmt.Errorf("final price cannot be negative")
	}

	return s.store.UpdateQuote(ctx, quote)
}

// DeleteQuote soft deletes a quote.
func (s *QuoteService) DeleteQuote(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("quote ID is required")
	}

	// Check if quote exists
	_, err := s.store.GetQuote(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}

	return s.store.DeleteQuote(ctx, id)
}

// ExpireQuote marks a quote as expired.
func (s *QuoteService) ExpireQuote(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("quote ID is required")
	}

	return s.store.ExpireQuote(ctx, id)
}

// CountQuotes returns the total number of quotes with filtering.
func (s *QuoteService) CountQuotes(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string) (int64, error) {
	return s.store.CountQuotes(ctx, userID, productID, status)
}

// CalculatePremium calculates the premium for a quote using risk assessment algorithms.
func (s *QuoteService) CalculatePremium(ctx context.Context, quoteID uuid.UUID) (*models.Quote, error) {
	// Fetch quote details from database
	quote, err := s.GetQuote(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote: %w", err)
	}

	if quote == nil {
		return nil, fmt.Errorf("quote not found: %s", quoteID.String())
	}

	// Calculate premium using risk assessment algorithms
	premium, err := s.calculatePremium(ctx, quote)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate premium: %w", err)
	}

	// Update quote with calculated premium
	quote.FinalPrice = premium
	quote.Status = models.QuoteStatusActive

	if err := s.UpdateQuote(ctx, quote); err != nil {
		return nil, fmt.Errorf("failed to update quote: %w", err)
	}

	// Publish quote calculated event
	if s.eventService != nil {
		event := events.NewQuoteCalculatedEvent(quote.ID, quote.UserID, quote.FinalPrice, quote.BasePrice, quote.Currency, time.Now())
		if err := s.eventService.PublishEvent(ctx, event); err != nil {
			// Log error but don't fail the calculation
		}
	}

	return quote, nil
}

// calculatePremium performs the actual premium calculation using risk assessment
func (s *QuoteService) calculatePremium(ctx context.Context, quote *models.Quote) (float64, error) {
	// Base premium calculation
	basePremium := s.calculateBasePremium(quote)

	// Apply risk factors
	riskMultiplier := s.calculateRiskMultiplier(quote)

	// Apply coverage factors
	coverageMultiplier := s.calculateCoverageMultiplier(quote)

	// Calculate final premium
	finalPremium := basePremium * riskMultiplier * coverageMultiplier

	// Round to 2 decimal places
	return math.Round(finalPremium*100) / 100, nil
}

// calculateBasePremium calculates the base premium based on product type and coverage amount
func (s *QuoteService) calculateBasePremium(quote *models.Quote) float64 {
	// Use the base price from the quote as starting point
	basePrice := quote.BasePrice

	// Apply product type multipliers
	productMultipliers := map[string]float64{
		"auto":     1.0, // Standard rate
		"home":     1.2, // 20% higher
		"life":     0.8, // 20% lower
		"health":   1.5, // 50% higher
		"business": 1.8, // 80% higher
	}

	// Get product category from the related product
	productType := "auto" // Default
	if quote.Product.Category != "" {
		productType = quote.Product.Category
	}

	multiplier, exists := productMultipliers[productType]
	if !exists {
		multiplier = 1.0 // Default multiplier
	}

	return basePrice * multiplier
}

// calculateRiskMultiplier applies risk factors to the premium
func (s *QuoteService) calculateRiskMultiplier(quote *models.Quote) float64 {
	multiplier := 1.0

	// Apply risk factors from the quote
	for _, riskFactor := range quote.RiskFactors {
		multiplier *= riskFactor.Impact
	}

	// If no risk factors are present, apply default risk assessment
	if len(quote.RiskFactors) == 0 {
		// Default risk multiplier based on quote age
		quoteAge := time.Since(quote.CreatedAt)
		if quoteAge > 24*time.Hour {
			multiplier *= 1.1 // Slightly higher risk for older quotes
		}
	}

	return multiplier
}

// calculateCoverageMultiplier applies coverage-specific factors
func (s *QuoteService) calculateCoverageMultiplier(quote *models.Quote) float64 {
	multiplier := 1.0

	// Apply discount if present
	if quote.Discount > 0 {
		multiplier *= (1.0 - quote.Discount/100.0) // Convert percentage to decimal
	}

	// Apply tax if present
	if quote.Tax > 0 {
		multiplier *= (1.0 + quote.Tax/100.0) // Convert percentage to decimal
	}

	// Currency factor (simplified)
	if quote.Currency != "USD" {
		// In a real implementation, you would use current exchange rates
		multiplier *= 1.05 // 5% adjustment for non-USD currencies
	}

	return multiplier
}

// generateQuoteNumber generates a unique quote number.
func (s *QuoteService) generateQuoteNumber() string {
	// Generate a quote number with timestamp and random component
	timestamp := time.Now().Format("20060102150405")
	random := rand.Intn(9999)
	return fmt.Sprintf("Q-%s-%04d", timestamp, random)
}
