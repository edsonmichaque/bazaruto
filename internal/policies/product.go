package policies

import (
	"context"

	"github.com/edsonmichaque/bazaruto/internal/models"
)

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

// Publish checks if a user can publish a product
func (p *ProductPolicy) Publish(ctx context.Context, user *models.User, product *models.Product) bool {
	// Only admins can publish products
	return user.Role == "admin"
}

// Unpublish checks if a user can unpublish a product
func (p *ProductPolicy) Unpublish(ctx context.Context, user *models.User, product *models.Product) bool {
	// Only admins can unpublish products
	return user.Role == "admin"
}

// Archive checks if a user can archive a product
func (p *ProductPolicy) Archive(ctx context.Context, user *models.User, product *models.Product) bool {
	// Only admins can archive products
	return user.Role == "admin"
}
