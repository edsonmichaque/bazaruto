package policies

import (
	"context"

	"github.com/edsonmichaque/bazaruto/internal/models"
)

// PolicyPolicy implements authorization policies for insurance policies
type PolicyPolicy struct{}

// Before is called before any other policy method
func (p *PolicyPolicy) Before(ctx context.Context, user *models.User, ability string, arguments ...interface{}) (bool, error) {
	// Admin users can do everything
	if user.Role == "admin" {
		return true, nil
	}

	return false, nil // Continue to specific policy methods
}

// View checks if a user can view a policy
func (p *PolicyPolicy) View(ctx context.Context, user *models.User, policy *models.Policy) bool {
	// Users can view their own policies
	if user.ID == policy.UserID {
		return true
	}

	// Agents can view all policies
	if user.Role == "agent" {
		return true
	}

	return false
}

// Create checks if a user can create a policy
func (p *PolicyPolicy) Create(ctx context.Context, user *models.User, policy *models.Policy) bool {
	// Agents and admins can create policies
	return user.Role == "agent" || user.Role == "admin"
}

// Update checks if a user can update a policy
func (p *PolicyPolicy) Update(ctx context.Context, user *models.User, policy *models.Policy) bool {
	// Agents and admins can update policies
	return user.Role == "agent" || user.Role == "admin"
}

// Delete checks if a user can delete a policy
func (p *PolicyPolicy) Delete(ctx context.Context, user *models.User, policy *models.Policy) bool {
	// Only admins can delete policies
	return user.Role == "admin"
}

// Cancel checks if a user can cancel a policy
func (p *PolicyPolicy) Cancel(ctx context.Context, user *models.User, policy *models.Policy) bool {
	// Users can cancel their own policies
	if user.ID == policy.UserID {
		return true
	}

	// Agents can cancel policies
	if user.Role == "agent" {
		return true
	}

	return false
}

// Renew checks if a user can renew a policy
func (p *PolicyPolicy) Renew(ctx context.Context, user *models.User, policy *models.Policy) bool {
	// Users can renew their own policies
	if user.ID == policy.UserID {
		return true
	}

	// Agents can renew policies
	if user.Role == "agent" {
		return true
	}

	return false
}

// Suspend checks if a user can suspend a policy
func (p *PolicyPolicy) Suspend(ctx context.Context, user *models.User, policy *models.Policy) bool {
	// Only agents and admins can suspend policies
	return user.Role == "agent" || user.Role == "admin"
}

// Activate checks if a user can activate a policy
func (p *PolicyPolicy) Activate(ctx context.Context, user *models.User, policy *models.Policy) bool {
	// Only agents and admins can activate policies
	return user.Role == "agent" || user.Role == "admin"
}
