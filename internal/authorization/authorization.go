package authorization

import (
	"context"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/policies"
)

// Service handles authorization operations (Rails/Laravel inspired)
type Service struct {
	rbacService    *RBACService
	policyRegistry *policies.Registry
}

// NewService creates a new authorization service
func NewService() *Service {
	service := &Service{
		rbacService:    NewRBACService(),
		policyRegistry: policies.NewRegistry(),
	}

	return service
}

// RegisterPolicy registers a policy for a resource
func (s *Service) RegisterPolicy(resource string, policy policies.Policy) {
	s.policyRegistry.Register(resource, policy)
}

// GetPolicy returns the policy for a resource
func (s *Service) GetPolicy(resource string) (policies.Policy, error) {
	return s.policyRegistry.Get(resource)
}

// Can checks if a user can perform an action (Laravel-style)
func (s *Service) Can(ctx context.Context, user *models.User, ability string, resource interface{}) (bool, error) {
	// Admin users can do everything
	if user.Role == "admin" {
		return true, nil
	}

	// Get resource type
	resourceType := s.getResourceType(resource)

	// Get the appropriate policy
	policy, err := s.GetPolicy(resourceType)
	if err != nil {
		return false, err
	}

	// Check Before method first
	if beforePolicy, ok := policy.(interface {
		Before(ctx context.Context, user *models.User, ability string, arguments ...interface{}) (bool, error)
	}); ok {
		allowed, err := beforePolicy.Before(ctx, user, ability, resource)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}

	// Call the specific policy method based on ability
	switch resourceType {
	case "user":
		return s.checkUserPolicy(ctx, user, ability, resource)
	case "product":
		return s.checkProductPolicy(ctx, user, ability, resource)
	case "quote":
		return s.checkQuotePolicy(ctx, user, ability, resource)
	case "policy":
		return s.checkPolicyPolicy(ctx, user, ability, resource)
	case "claim":
		return s.checkClaimPolicy(ctx, user, ability, resource)
	default:
		return false, fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

// Cannot checks if a user cannot perform an action (Laravel-style)
func (s *Service) Cannot(ctx context.Context, user *models.User, ability string, resource interface{}) (bool, error) {
	can, err := s.Can(ctx, user, ability, resource)
	return !can, err
}

// checkUserPolicy checks user-specific policies
func (s *Service) checkUserPolicy(ctx context.Context, user *models.User, ability string, resource interface{}) (bool, error) {
	targetUser, ok := resource.(*models.User)
	if !ok {
		return false, fmt.Errorf("resource is not a user")
	}

	policy := &policies.UserPolicy{}

	switch ability {
	case "view":
		return policy.View(ctx, user, targetUser), nil
	case "create":
		return policy.Create(ctx, user, targetUser), nil
	case "update":
		return policy.Update(ctx, user, targetUser), nil
	case "delete":
		return policy.Delete(ctx, user, targetUser), nil
	case "impersonate":
		return policy.Impersonate(ctx, user, targetUser), nil
	case "suspend":
		return policy.Suspend(ctx, user, targetUser), nil
	case "activate":
		return policy.Activate(ctx, user, targetUser), nil
	default:
		return false, fmt.Errorf("unknown ability: %s", ability)
	}
}

// checkProductPolicy checks product-specific policies
func (s *Service) checkProductPolicy(ctx context.Context, user *models.User, ability string, resource interface{}) (bool, error) {
	product, ok := resource.(*models.Product)
	if !ok {
		return false, fmt.Errorf("resource is not a product")
	}

	policy := &policies.ProductPolicy{}

	switch ability {
	case "view":
		return policy.View(ctx, user, product), nil
	case "create":
		return policy.Create(ctx, user, product), nil
	case "update":
		return policy.Update(ctx, user, product), nil
	case "delete":
		return policy.Delete(ctx, user, product), nil
	case "publish":
		return policy.Publish(ctx, user, product), nil
	case "unpublish":
		return policy.Unpublish(ctx, user, product), nil
	case "archive":
		return policy.Archive(ctx, user, product), nil
	default:
		return false, fmt.Errorf("unknown ability: %s", ability)
	}
}

// checkQuotePolicy checks quote-specific policies
func (s *Service) checkQuotePolicy(ctx context.Context, user *models.User, ability string, resource interface{}) (bool, error) {
	quote, ok := resource.(*models.Quote)
	if !ok {
		return false, fmt.Errorf("resource is not a quote")
	}

	policy := &policies.QuotePolicy{}

	switch ability {
	case "view":
		return policy.View(ctx, user, quote), nil
	case "create":
		return policy.Create(ctx, user, quote), nil
	case "update":
		return policy.Update(ctx, user, quote), nil
	case "delete":
		return policy.Delete(ctx, user, quote), nil
	case "accept":
		return policy.Accept(ctx, user, quote), nil
	case "reject":
		return policy.Reject(ctx, user, quote), nil
	case "expire":
		return policy.Expire(ctx, user, quote), nil
	default:
		return false, fmt.Errorf("unknown ability: %s", ability)
	}
}

// checkPolicyPolicy checks insurance policy-specific policies
func (s *Service) checkPolicyPolicy(ctx context.Context, user *models.User, ability string, resource interface{}) (bool, error) {
	policy, ok := resource.(*models.Policy)
	if !ok {
		return false, fmt.Errorf("resource is not a policy")
	}

	policyPolicy := &policies.PolicyPolicy{}

	switch ability {
	case "view":
		return policyPolicy.View(ctx, user, policy), nil
	case "create":
		return policyPolicy.Create(ctx, user, policy), nil
	case "update":
		return policyPolicy.Update(ctx, user, policy), nil
	case "delete":
		return policyPolicy.Delete(ctx, user, policy), nil
	case "cancel":
		return policyPolicy.Cancel(ctx, user, policy), nil
	case "renew":
		return policyPolicy.Renew(ctx, user, policy), nil
	case "suspend":
		return policyPolicy.Suspend(ctx, user, policy), nil
	case "activate":
		return policyPolicy.Activate(ctx, user, policy), nil
	default:
		return false, fmt.Errorf("unknown ability: %s", ability)
	}
}

// checkClaimPolicy checks claim-specific policies
func (s *Service) checkClaimPolicy(ctx context.Context, user *models.User, ability string, resource interface{}) (bool, error) {
	claim, ok := resource.(*models.Claim)
	if !ok {
		return false, fmt.Errorf("resource is not a claim")
	}

	policy := &policies.ClaimPolicy{}

	switch ability {
	case "view":
		return policy.View(ctx, user, claim), nil
	case "create":
		return policy.Create(ctx, user, claim), nil
	case "update":
		return policy.Update(ctx, user, claim), nil
	case "delete":
		return policy.Delete(ctx, user, claim), nil
	case "approve":
		return policy.Approve(ctx, user, claim), nil
	case "reject":
		return policy.Reject(ctx, user, claim), nil
	case "investigate":
		return policy.Investigate(ctx, user, claim), nil
	case "settle":
		return policy.Settle(ctx, user, claim), nil
	case "reopen":
		return policy.Reopen(ctx, user, claim), nil
	default:
		return false, fmt.Errorf("unknown ability: %s", ability)
	}
}

// getResourceType returns the type name of a resource
func (s *Service) getResourceType(resource interface{}) string {
	if resource == nil {
		return "nil"
	}

	// Simple type checking without reflection
	switch resource.(type) {
	case *models.User:
		return "user"
	case *models.Product:
		return "product"
	case *models.Quote:
		return "quote"
	case *models.Policy:
		return "policy"
	case *models.Claim:
		return "claim"
	default:
		return "unknown"
	}
}

// Gate defines a custom authorization gate
type Gate func(ctx context.Context, user *models.User, arguments ...interface{}) bool

// gates stores custom authorization gates
var gates = make(map[string]Gate)

// DefineGate defines a custom authorization gate
func DefineGate(name string, gate Gate) {
	gates[name] = gate
}

// AuthorizeGate checks if a user can pass through a gate
func (s *Service) AuthorizeGate(ctx context.Context, user *models.User, gateName string, arguments ...interface{}) (bool, error) {
	gate, exists := gates[gateName]
	if !exists {
		return false, fmt.Errorf("gate [%s] not defined", gateName)
	}

	return gate(ctx, user, arguments...), nil
}

// ForUser returns an authorization instance for a specific user
func (s *Service) ForUser(user *models.User) *UserAuthorization {
	return &UserAuthorization{
		service: s,
		user:    user,
	}
}

// UserAuthorization provides authorization methods for a specific user
type UserAuthorization struct {
	service *Service
	user    *models.User
}

// Can checks if the user can perform an action
func (ua *UserAuthorization) Can(ctx context.Context, ability string, resource interface{}) (bool, error) {
	return ua.service.Can(ctx, ua.user, ability, resource)
}

// Cannot checks if the user cannot perform an action
func (ua *UserAuthorization) Cannot(ctx context.Context, ability string, resource interface{}) (bool, error) {
	return ua.service.Cannot(ctx, ua.user, ability, resource)
}

// AuthorizeGate checks if the user can pass through a gate
func (ua *UserAuthorization) AuthorizeGate(ctx context.Context, gateName string, arguments ...interface{}) (bool, error) {
	return ua.service.AuthorizeGate(ctx, ua.user, gateName, arguments...)
}
