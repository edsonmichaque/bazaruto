package policies

import (
	"context"

	"github.com/edsonmichaque/bazaruto/internal/models"
)

// ClaimPolicy implements authorization policies for claims
type ClaimPolicy struct{}

// Before is called before any other policy method
func (p *ClaimPolicy) Before(ctx context.Context, user *models.User, ability string, arguments ...interface{}) (bool, error) {
	// Admin users can do everything
	if user.Role == "admin" {
		return true, nil
	}

	return false, nil // Continue to specific policy methods
}

// View checks if a user can view a claim
func (p *ClaimPolicy) View(ctx context.Context, user *models.User, claim *models.Claim) bool {
	// Users can view their own claims
	if user.ID == claim.UserID {
		return true
	}

	// Agents can view all claims
	if user.Role == "agent" {
		return true
	}

	return false
}

// Create checks if a user can create a claim
func (p *ClaimPolicy) Create(ctx context.Context, user *models.User, claim *models.Claim) bool {
	// Users can create claims for their own policies
	if user.ID == claim.UserID {
		return true
	}

	// Agents can create claims
	if user.Role == "agent" {
		return true
	}

	return false
}

// Update checks if a user can update a claim
func (p *ClaimPolicy) Update(ctx context.Context, user *models.User, claim *models.Claim) bool {
	// Users can update their own claims (if not approved)
	if user.ID == claim.UserID && claim.Status != "approved" {
		return true
	}

	// Agents can update all claims
	if user.Role == "agent" {
		return true
	}

	return false
}

// Delete checks if a user can delete a claim
func (p *ClaimPolicy) Delete(ctx context.Context, user *models.User, claim *models.Claim) bool {
	// Users can delete their own claims (if not approved)
	if user.ID == claim.UserID && claim.Status != "approved" {
		return true
	}

	// Agents can delete claims
	if user.Role == "agent" {
		return true
	}

	return false
}

// Approve checks if a user can approve a claim
func (p *ClaimPolicy) Approve(ctx context.Context, user *models.User, claim *models.Claim) bool {
	// Only agents and admins can approve claims
	return user.Role == "agent" || user.Role == "admin"
}

// Reject checks if a user can reject a claim
func (p *ClaimPolicy) Reject(ctx context.Context, user *models.User, claim *models.Claim) bool {
	// Only agents and admins can reject claims
	return user.Role == "agent" || user.Role == "admin"
}

// Investigate checks if a user can investigate a claim
func (p *ClaimPolicy) Investigate(ctx context.Context, user *models.User, claim *models.Claim) bool {
	// Only agents and admins can investigate claims
	return user.Role == "agent" || user.Role == "admin"
}

// Settle checks if a user can settle a claim
func (p *ClaimPolicy) Settle(ctx context.Context, user *models.User, claim *models.Claim) bool {
	// Only agents and admins can settle claims
	return user.Role == "agent" || user.Role == "admin"
}

// Reopen checks if a user can reopen a claim
func (p *ClaimPolicy) Reopen(ctx context.Context, user *models.User, claim *models.Claim) bool {
	// Only agents and admins can reopen claims
	return user.Role == "agent" || user.Role == "admin"
}
