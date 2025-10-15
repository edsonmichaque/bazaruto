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

// ClaimService handles business logic for claims.
type ClaimService struct {
	store       store.ClaimStore
	policyStore store.PolicyStore
}

// NewClaimService creates a new ClaimService instance.
func NewClaimService(store store.ClaimStore, policyStore store.PolicyStore) *ClaimService {
	return &ClaimService{
		store:       store,
		policyStore: policyStore,
	}
}

// CreateClaim creates a new claim with business logic validation.
func (s *ClaimService) CreateClaim(ctx context.Context, claim *models.Claim) error {
	// Validate required fields
	if claim.PolicyID == uuid.Nil {
		return fmt.Errorf("policy ID is required")
	}
	if claim.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if claim.Title == "" {
		return fmt.Errorf("claim title is required")
	}
	if claim.Description == "" {
		return fmt.Errorf("claim description is required")
	}
	if claim.ClaimAmount <= 0 {
		return fmt.Errorf("claim amount must be greater than 0")
	}
	if claim.IncidentDate.IsZero() {
		return fmt.Errorf("incident date is required")
	}

	// Set defaults
	if claim.Currency == "" {
		claim.Currency = models.CurrencyUSD
	}
	if claim.Status == "" {
		claim.Status = models.ClaimStatusSubmitted
	}
	if claim.ReportedDate.IsZero() {
		claim.ReportedDate = time.Now()
	}

	// Generate claim number if not provided
	if claim.ClaimNumber == "" {
		claim.ClaimNumber = s.generateClaimNumber()
	}

	// Validate incident date is not in the future
	if claim.IncidentDate.After(time.Now()) {
		return fmt.Errorf("incident date cannot be in the future")
	}

	// Validate reported date is not before incident date
	if claim.ReportedDate.Before(claim.IncidentDate) {
		return fmt.Errorf("reported date cannot be before incident date")
	}

	// Validate claim amount is not negative
	if claim.ClaimAmount < 0 {
		return fmt.Errorf("claim amount cannot be negative")
	}

	// Validate paid amount is not negative
	if claim.PaidAmount < 0 {
		return fmt.Errorf("paid amount cannot be negative")
	}

	// Validate that the policy exists and is active
	policy, err := s.policyStore.GetPolicy(ctx, claim.PolicyID)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	if policy.Status != models.PolicyStatusActive {
		return fmt.Errorf("cannot submit claim for inactive policy")
	}

	// Validate that the incident date is within the policy period
	if claim.IncidentDate.Before(policy.EffectiveDate) || claim.IncidentDate.After(policy.ExpirationDate) {
		return fmt.Errorf("incident date must be within the policy period")
	}

	// Validate that the user owns the policy
	if policy.UserID != claim.UserID {
		return fmt.Errorf("user does not own the policy")
	}

	return s.store.CreateClaim(ctx, claim)
}

// GetClaim retrieves a claim by ID.
func (s *ClaimService) GetClaim(ctx context.Context, id uuid.UUID) (*models.Claim, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("claim ID is required")
	}

	claim, err := s.store.GetClaim(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}

	return claim, nil
}

// GetClaimByNumber retrieves a claim by claim number.
func (s *ClaimService) GetClaimByNumber(ctx context.Context, claimNumber string) (*models.Claim, error) {
	if claimNumber == "" {
		return nil, fmt.Errorf("claim number is required")
	}

	claim, err := s.store.GetClaimByNumber(ctx, claimNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim by number: %w", err)
	}

	return claim, nil
}

// ListClaims retrieves a list of claims with filtering.
func (s *ClaimService) ListClaims(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string, limit, offset int) ([]*models.Claim, error) {
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

	claims, err := s.store.ListClaims(ctx, userID, policyID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list claims: %w", err)
	}

	return claims, nil
}

// UpdateClaim updates an existing claim.
func (s *ClaimService) UpdateClaim(ctx context.Context, claim *models.Claim) error {
	if claim.ID == uuid.Nil {
		return fmt.Errorf("claim ID is required")
	}

	// Get existing claim to validate changes
	existing, err := s.store.GetClaim(ctx, claim.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing claim: %w", err)
	}

	// Prevent changes to critical fields
	if existing.PolicyID != claim.PolicyID {
		return fmt.Errorf("cannot change policy ID")
	}
	if existing.UserID != claim.UserID {
		return fmt.Errorf("cannot change user ID")
	}
	if existing.ClaimNumber != claim.ClaimNumber {
		return fmt.Errorf("cannot change claim number")
	}
	if existing.IncidentDate != claim.IncidentDate {
		return fmt.Errorf("cannot change incident date")
	}
	if existing.ReportedDate != claim.ReportedDate {
		return fmt.Errorf("cannot change reported date")
	}

	// Validate claim amount is not negative
	if claim.ClaimAmount < 0 {
		return fmt.Errorf("claim amount cannot be negative")
	}

	// Validate paid amount is not negative
	if claim.PaidAmount < 0 {
		return fmt.Errorf("paid amount cannot be negative")
	}

	// Validate paid amount does not exceed claim amount
	if claim.PaidAmount > claim.ClaimAmount {
		return fmt.Errorf("paid amount cannot exceed claim amount")
	}

	return s.store.UpdateClaim(ctx, claim)
}

// DeleteClaim soft deletes a claim.
func (s *ClaimService) DeleteClaim(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("claim ID is required")
	}

	// Check if claim exists
	_, err := s.store.GetClaim(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}

	return s.store.DeleteClaim(ctx, id)
}

// CountClaims returns the total number of claims with filtering.
func (s *ClaimService) CountClaims(ctx context.Context, userID *uuid.UUID, policyID *uuid.UUID, status string) (int64, error) {
	return s.store.CountClaims(ctx, userID, policyID, status)
}

// generateClaimNumber generates a unique claim number.
func (s *ClaimService) generateClaimNumber() string {
	// Generate a claim number with timestamp and random component
	timestamp := time.Now().Format("20060102150405")
	random := rand.Intn(9999)
	return fmt.Sprintf("C-%s-%04d", timestamp, random)
}
