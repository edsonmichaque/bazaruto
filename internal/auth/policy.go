package auth

import (
	"context"

	"github.com/edsonmichaque/bazaruto/internal/models"
)

// Policy defines the interface for authorization policies
type Policy interface {
	// Before is called before any other policy method
	Before(ctx context.Context, user *models.User, ability string, arguments ...interface{}) (bool, error)
}

// UserPolicy implements authorization policies for users
type UserPolicy struct{}

// Before is called before any other policy method
func (p *UserPolicy) Before(ctx context.Context, user *models.User, ability string, arguments ...interface{}) (bool, error) {
	// Admin users can do everything
	if user.Role == "admin" {
		return true, nil
	}

	return false, nil // Continue to specific policy methods
}

// View checks if a user can view another user
func (p *UserPolicy) View(ctx context.Context, user *models.User, targetUser *models.User) bool {
	// Users can view their own profile
	if user.ID == targetUser.ID {
		return true
	}

	// Agents can view customers
	if user.Role == "agent" && targetUser.Role == "customer" {
		return true
	}

	return false
}

// Create checks if a user can create another user
func (p *UserPolicy) Create(ctx context.Context, user *models.User, targetUser *models.User) bool {
	// Only admins can create users
	return user.Role == "admin"
}

// Update checks if a user can update another user
func (p *UserPolicy) Update(ctx context.Context, user *models.User, targetUser *models.User) bool {
	// Users can update their own profile
	if user.ID == targetUser.ID {
		return true
	}

	// Agents can update customers
	if user.Role == "agent" && targetUser.Role == "customer" {
		return true
	}

	return false
}

// Delete checks if a user can delete another user
func (p *UserPolicy) Delete(ctx context.Context, user *models.User, targetUser *models.User) bool {
	// Only admins can delete users
	return user.Role == "admin"
}

// ProductPolicy implements authorization policies for products
type ProductPolicy struct{}

// Before is called before any other policy method
func (p *ProductPolicy) Before(ctx context.Context, user *models.User, ability string, arguments ...interface{}) (bool, error) {
	// Admin users can do everything
	if user.Role == "admin" {
		return true, nil
	}

	return false, nil // Continue to specific policy methods
}

// View checks if a user can view a product
func (p *ProductPolicy) View(ctx context.Context, user *models.User, product *models.Product) bool {
	// All authenticated users can view products
	return true
}

// Create checks if a user can create a product
func (p *ProductPolicy) Create(ctx context.Context, user *models.User, product *models.Product) bool {
	// Only admins can create products
	return user.Role == "admin"
}

// Update checks if a user can update a product
func (p *ProductPolicy) Update(ctx context.Context, user *models.User, product *models.Product) bool {
	// Only admins can update products
	return user.Role == "admin"
}

// Delete checks if a user can delete a product
func (p *ProductPolicy) Delete(ctx context.Context, user *models.User, product *models.Product) bool {
	// Only admins can delete products
	return user.Role == "admin"
}

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

// Can checks if a user can perform an action (simple version)
func Can(ctx context.Context, user *models.User, ability string, resource interface{}) (bool, error) {
	// This would use a global policy manager
	// For now, return true for admin users
	if user.Role == "admin" {
		return true, nil
	}

	return false, nil
}

// Cannot checks if a user cannot perform an action
func Cannot(ctx context.Context, user *models.User, ability string, resource interface{}) (bool, error) {
	can, err := Can(ctx, user, ability, resource)
	return !can, err
}
