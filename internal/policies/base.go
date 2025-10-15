package policies

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
)

// Policy defines the interface for authorization policies
type Policy interface {
	// Before is called before any other policy method
	Before(ctx context.Context, user *models.User, ability string, arguments ...interface{}) (bool, error)
}

// Registry manages policy registration and retrieval
type Registry struct {
	policies map[string]Policy
}

// NewRegistry creates a new policy registry
func NewRegistry() *Registry {
	registry := &Registry{
		policies: make(map[string]Policy),
	}

	// Register default policies
	registry.registerDefaultPolicies()

	return registry
}

// Register registers a policy for a resource
func (r *Registry) Register(resource string, policy Policy) {
	r.policies[resource] = policy
}

// Get returns the policy for a resource
func (r *Registry) Get(resource string) (Policy, error) {
	policy, exists := r.policies[resource]
	if !exists {
		return nil, fmt.Errorf("policy not found for resource: %s", resource)
	}
	return policy, nil
}

// List returns all registered policies
func (r *Registry) List() map[string]Policy {
	return r.policies
}

// registerDefaultPolicies registers the default policies
func (r *Registry) registerDefaultPolicies() {
	r.policies["user"] = &UserPolicy{}
	r.policies["product"] = &ProductPolicy{}
	r.policies["quote"] = &QuotePolicy{}
	r.policies["policy"] = &PolicyPolicy{}
	r.policies["claim"] = &ClaimPolicy{}
}
