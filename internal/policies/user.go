package policies

import (
	"context"

	"github.com/edsonmichaque/bazaruto/internal/models"
)

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

// Impersonate checks if a user can impersonate another user
func (p *UserPolicy) Impersonate(ctx context.Context, user *models.User, targetUser *models.User) bool {
	// Only admins can impersonate users
	return user.Role == "admin"
}

// Suspend checks if a user can suspend another user
func (p *UserPolicy) Suspend(ctx context.Context, user *models.User, targetUser *models.User) bool {
	// Only admins can suspend users
	return user.Role == "admin"
}

// Activate checks if a user can activate another user
func (p *UserPolicy) Activate(ctx context.Context, user *models.User, targetUser *models.User) bool {
	// Only admins can activate users
	return user.Role == "admin"
}
