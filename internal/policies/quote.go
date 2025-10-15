package policies

import (
	"context"

	"github.com/edsonmichaque/bazaruto/internal/models"
)

// QuotePolicy implements authorization policies for quotes
type QuotePolicy struct{}

// Before is called before any other policy method
func (p *QuotePolicy) Before(ctx context.Context, user *models.User, ability string, arguments ...interface{}) (bool, error) {
	// Admin users can do everything
	if user.Role == "admin" {
		return true, nil
	}

	return false, nil // Continue to specific policy methods
}

// View checks if a user can view a quote
func (p *QuotePolicy) View(ctx context.Context, user *models.User, quote *models.Quote) bool {
	// Users can view their own quotes
	if user.ID == quote.UserID {
		return true
	}

	// Agents can view all quotes
	if user.Role == "agent" {
		return true
	}

	return false
}

// Create checks if a user can create a quote
func (p *QuotePolicy) Create(ctx context.Context, user *models.User, quote *models.Quote) bool {
	// All authenticated users can create quotes
	return true
}

// Update checks if a user can update a quote
func (p *QuotePolicy) Update(ctx context.Context, user *models.User, quote *models.Quote) bool {
	// Users can update their own quotes
	if user.ID == quote.UserID {
		return true
	}

	// Agents can update all quotes
	if user.Role == "agent" {
		return true
	}

	return false
}

// Delete checks if a user can delete a quote
func (p *QuotePolicy) Delete(ctx context.Context, user *models.User, quote *models.Quote) bool {
	// Users can delete their own quotes
	if user.ID == quote.UserID {
		return true
	}

	// Agents can delete all quotes
	if user.Role == "agent" {
		return true
	}

	return false
}

// Accept checks if a user can accept a quote
func (p *QuotePolicy) Accept(ctx context.Context, user *models.User, quote *models.Quote) bool {
	// Users can accept their own quotes
	if user.ID == quote.UserID {
		return true
	}

	// Agents can accept quotes on behalf of users
	if user.Role == "agent" {
		return true
	}

	return false
}

// Reject checks if a user can reject a quote
func (p *QuotePolicy) Reject(ctx context.Context, user *models.User, quote *models.Quote) bool {
	// Users can reject their own quotes
	if user.ID == quote.UserID {
		return true
	}

	// Agents can reject quotes on behalf of users
	if user.Role == "agent" {
		return true
	}

	return false
}

// Expire checks if a user can expire a quote
func (p *QuotePolicy) Expire(ctx context.Context, user *models.User, quote *models.Quote) bool {
	// Only agents and admins can expire quotes
	return user.Role == "agent" || user.Role == "admin"
}
