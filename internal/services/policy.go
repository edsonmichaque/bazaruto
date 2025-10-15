package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// PolicyService handles business logic for policies.
type PolicyService struct {
	store store.PolicyStore
}

// NewPolicyService creates a new PolicyService instance.
func NewPolicyService(store store.PolicyStore) *PolicyService {
	return &PolicyService{
		store: store,
	}
}

// CreatePolicy creates a new policy with business logic validation.
func (s *PolicyService) CreatePolicy(ctx context.Context, policy *models.Policy) error {
	// Validate required fields
	if policy.ProductID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}
	if policy.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if policy.Premium <= 0 {
		return fmt.Errorf("premium must be greater than 0")
	}
	if policy.CoverageAmount <= 0 {
		return fmt.Errorf("coverage amount must be greater than 0")
	}
	if policy.EffectiveDate.IsZero() {
		return fmt.Errorf("effective date is required")
	}
	if policy.ExpirationDate.IsZero() {
		return fmt.Errorf("expiration date is required")
	}

	// Set defaults
	if policy.Currency == "" {
		policy.Currency = models.CurrencyUSD
	}
	if policy.Status == "" {
		policy.Status = models.PolicyStatusActive
	}
	if policy.PaymentFrequency == "" {
		policy.PaymentFrequency = "monthly"
	}

	// Generate policy number if not provided
	if policy.PolicyNumber == "" {
		policy.PolicyNumber = s.generatePolicyNumber()
	}

	// Validate effective date is not in the past
	if policy.EffectiveDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return fmt.Errorf("effective date cannot be in the past")
	}

	// Validate expiration date is after effective date
	if policy.ExpirationDate.Before(policy.EffectiveDate) {
		return fmt.Errorf("expiration date must be after effective date")
	}

	// Validate payment frequency
	validFrequencies := []string{"monthly", "quarterly", "annually"}
	if !contains(validFrequencies, policy.PaymentFrequency) {
		return fmt.Errorf("invalid payment frequency: %s", policy.PaymentFrequency)
	}

	// Set renewal date if auto-renew is enabled
	if policy.AutoRenew && policy.RenewalDate == nil {
		policy.RenewalDate = &policy.ExpirationDate
	}

	return s.store.CreatePolicy(ctx, policy)
}

// GetPolicy retrieves a policy by ID.
func (s *PolicyService) GetPolicy(ctx context.Context, id uuid.UUID) (*models.Policy, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("policy ID is required")
	}

	policy, err := s.store.GetPolicy(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	// Check if policy has expired
	if policy.ExpirationDate.Before(time.Now()) && policy.Status == models.PolicyStatusActive {
		// Auto-expire the policy
		policy.Status = models.PolicyStatusExpired
		if err := s.store.UpdatePolicy(ctx, policy); err != nil {
			return nil, fmt.Errorf("failed to update expired policy: %w", err)
		}
	}

	return policy, nil
}

// GetPolicyByNumber retrieves a policy by policy number.
func (s *PolicyService) GetPolicyByNumber(ctx context.Context, policyNumber string) (*models.Policy, error) {
	if policyNumber == "" {
		return nil, fmt.Errorf("policy number is required")
	}

	policy, err := s.store.GetPolicyByNumber(ctx, policyNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy by number: %w", err)
	}

	// Check if policy has expired
	if policy.ExpirationDate.Before(time.Now()) && policy.Status == models.PolicyStatusActive {
		// Auto-expire the policy
		policy.Status = models.PolicyStatusExpired
		if err := s.store.UpdatePolicy(ctx, policy); err != nil {
			return nil, fmt.Errorf("failed to update expired policy: %w", err)
		}
	}

	return policy, nil
}

// ListPolicies retrieves a list of policies with filtering.
func (s *PolicyService) ListPolicies(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string, limit, offset int) ([]*models.Policy, error) {
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

	policies, err := s.store.ListPolicies(ctx, userID, productID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	// Auto-expire active policies that have passed their expiration date
	now := time.Now()
	for _, policy := range policies {
		if policy.Status == models.PolicyStatusActive && policy.ExpirationDate.Before(now) {
			policy.Status = models.PolicyStatusExpired
			if err := s.store.UpdatePolicy(ctx, policy); err != nil {
				// Log error but continue processing
				continue
			}
		}
	}

	return policies, nil
}

// UpdatePolicy updates an existing policy.
func (s *PolicyService) UpdatePolicy(ctx context.Context, policy *models.Policy) error {
	if policy.ID == uuid.Nil {
		return fmt.Errorf("policy ID is required")
	}

	// Get existing policy to validate changes
	existing, err := s.store.GetPolicy(ctx, policy.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing policy: %w", err)
	}

	// Prevent changes to critical fields
	if existing.ProductID != policy.ProductID {
		return fmt.Errorf("cannot change product ID")
	}
	if existing.UserID != policy.UserID {
		return fmt.Errorf("cannot change user ID")
	}
	if existing.PolicyNumber != policy.PolicyNumber {
		return fmt.Errorf("cannot change policy number")
	}

	// Validate premium is not negative
	if policy.Premium < 0 {
		return fmt.Errorf("premium cannot be negative")
	}

	// Validate coverage amount is not negative
	if policy.CoverageAmount < 0 {
		return fmt.Errorf("coverage amount cannot be negative")
	}

	return s.store.UpdatePolicy(ctx, policy)
}

// DeletePolicy soft deletes a policy.
func (s *PolicyService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("policy ID is required")
	}

	// Check if policy exists
	_, err := s.store.GetPolicy(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	return s.store.DeletePolicy(ctx, id)
}

// CountPolicies returns the total number of policies with filtering.
func (s *PolicyService) CountPolicies(ctx context.Context, userID *uuid.UUID, productID *uuid.UUID, status string) (int64, error) {
	return s.store.CountPolicies(ctx, userID, productID, status)
}

// generatePolicyNumber generates a unique policy number.
func (s *PolicyService) generatePolicyNumber() string {
	// Generate a policy number with timestamp and random component
	timestamp := time.Now().Format("20060102150405")
	random := rand.Intn(9999)
	return fmt.Sprintf("P-%s-%04d", timestamp, random)
}
