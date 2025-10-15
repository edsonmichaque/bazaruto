package services

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// ComplianceService handles compliance and regulatory validation for insurance operations.
type ComplianceService struct {
	userStore    store.UserStore
	policyStore  store.PolicyStore
	claimStore   store.ClaimStore
	paymentStore store.PaymentStore
}

// NewComplianceService creates a new ComplianceService instance.
func NewComplianceService(
	userStore store.UserStore,
	policyStore store.PolicyStore,
	claimStore store.ClaimStore,
	paymentStore store.PaymentStore,
) *ComplianceService {
	return &ComplianceService{
		userStore:    userStore,
		policyStore:  policyStore,
		claimStore:   claimStore,
		paymentStore: paymentStore,
	}
}

// ComplianceCheck represents the result of a compliance validation.
type ComplianceCheck struct {
	EntityID        uuid.UUID              `json:"entity_id"`       // ID of the entity being checked
	EntityType      string                 `json:"entity_type"`     // user, policy, claim, payment
	CheckType       string                 `json:"check_type"`      // kyc, aml, regulatory, data_protection
	Status          string                 `json:"status"`          // passed, failed, warning, pending
	Score           float64                `json:"score"`           // Compliance score (0-100)
	Violations      []ComplianceViolation  `json:"violations"`      // List of violations found
	Recommendations []string               `json:"recommendations"` // Recommendations for compliance
	CheckDate       time.Time              `json:"check_date"`      // When the check was performed
	ValidUntil      time.Time              `json:"valid_until"`     // When the check expires
	Metadata        map[string]interface{} `json:"metadata"`
}

// ComplianceViolation represents a specific compliance violation.
type ComplianceViolation struct {
	Code        string                 `json:"code"`        // Violation code
	Severity    string                 `json:"severity"`    // low, medium, high, critical
	Description string                 `json:"description"` // Human-readable description
	Rule        string                 `json:"rule"`        // Regulatory rule or standard
	Remediation string                 `json:"remediation"` // How to fix the violation
	Metadata    map[string]interface{} `json:"metadata"`
}

// ComplianceRule represents a compliance rule or regulation.
type ComplianceRule struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`     // kyc, aml, data_protection, etc.
	Jurisdiction   string                 `json:"jurisdiction"` // Country or region
	Severity       string                 `json:"severity"`     // low, medium, high, critical
	Active         bool                   `json:"active"`
	EffectiveDate  time.Time              `json:"effective_date"`
	ExpirationDate *time.Time             `json:"expiration_date"`
	Conditions     map[string]interface{} `json:"conditions"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ValidateUserCompliance performs comprehensive compliance validation for a user.
func (s *ComplianceService) ValidateUserCompliance(ctx context.Context, userID uuid.UUID) (*ComplianceCheck, error) {
	// Fetch user details
	user, err := s.userStore.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Initialize compliance check
	check := &ComplianceCheck{
		EntityID:   userID,
		EntityType: "user",
		CheckType:  "comprehensive",
		CheckDate:  time.Now(),
		ValidUntil: time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		Metadata:   make(map[string]interface{}),
	}

	// Perform various compliance checks
	violations := []ComplianceViolation{}
	recommendations := []string{}

	// KYC (Know Your Customer) checks
	kycViolations, kycRecommendations := s.performKYCChecks(user)
	violations = append(violations, kycViolations...)
	recommendations = append(recommendations, kycRecommendations...)

	// AML (Anti-Money Laundering) checks
	amlViolations, amlRecommendations := s.performAMLChecks(user)
	violations = append(violations, amlViolations...)
	recommendations = append(recommendations, amlRecommendations...)

	// Data protection checks
	dataViolations, dataRecommendations := s.performDataProtectionChecks(user)
	violations = append(violations, dataViolations...)
	recommendations = append(recommendations, dataRecommendations...)

	// Calculate compliance score
	check.Score = s.calculateComplianceScore(violations)
	check.Violations = violations
	check.Recommendations = recommendations

	// Determine overall status
	check.Status = s.determineComplianceStatus(check.Score, violations)

	// Store metadata
	check.Metadata["user_email"] = user.Email
	check.Metadata["user_status"] = user.Status
	check.Metadata["check_version"] = "1.0"

	return check, nil
}

// performKYCChecks performs Know Your Customer compliance checks.
func (s *ComplianceService) performKYCChecks(user *models.User) ([]ComplianceViolation, []string) {
	violations := []ComplianceViolation{}
	recommendations := []string{}

	// Check if user has provided required information
	if user.FullName == "" {
		violations = append(violations, ComplianceViolation{
			Code:        "KYC_001",
			Severity:    "high",
			Description: "Full name is required for KYC compliance",
			Rule:        "Customer Identification Program",
			Remediation: "Provide full legal name",
		})
	}

	if user.Email == "" {
		violations = append(violations, ComplianceViolation{
			Code:        "KYC_002",
			Severity:    "critical",
			Description: "Email address is required for KYC compliance",
			Rule:        "Customer Identification Program",
			Remediation: "Provide valid email address",
		})
	}

	// Check user status
	if user.Status != models.StatusActive {
		violations = append(violations, ComplianceViolation{
			Code:        "KYC_003",
			Severity:    "medium",
			Description: "User account is not in active status",
			Rule:        "Customer Due Diligence",
			Remediation: "Activate user account or provide justification",
		})
	}

	// Generate recommendations
	if len(violations) == 0 {
		recommendations = append(recommendations, "User meets KYC requirements")
	} else {
		recommendations = append(recommendations, "Address KYC violations before proceeding")
		recommendations = append(recommendations, "Consider enhanced due diligence for high-risk users")
	}

	return violations, recommendations
}

// performAMLChecks performs Anti-Money Laundering compliance checks.
func (s *ComplianceService) performAMLChecks(user *models.User) ([]ComplianceViolation, []string) {
	violations := []ComplianceViolation{}
	recommendations := []string{}

	// Check for suspicious patterns (simplified)
	accountAge := time.Since(user.CreatedAt).Hours() / 24 / 365 // years

	if accountAge < 0.1 { // Less than 1 month
		violations = append(violations, ComplianceViolation{
			Code:        "AML_001",
			Severity:    "medium",
			Description: "New account requires enhanced monitoring",
			Rule:        "Suspicious Activity Reporting",
			Remediation: "Implement enhanced monitoring procedures",
		})
	}

	// Check for high-risk indicators
	if user.Email != "" && len(user.Email) < 5 {
		violations = append(violations, ComplianceViolation{
			Code:        "AML_002",
			Severity:    "low",
			Description: "Email address appears suspicious",
			Rule:        "Customer Due Diligence",
			Remediation: "Verify email address authenticity",
		})
	}

	// Generate recommendations
	if len(violations) == 0 {
		recommendations = append(recommendations, "No AML concerns identified")
	} else {
		recommendations = append(recommendations, "Implement enhanced AML monitoring")
		recommendations = append(recommendations, "Consider filing suspicious activity report if warranted")
	}

	return violations, recommendations
}

// performDataProtectionChecks performs data protection compliance checks.
func (s *ComplianceService) performDataProtectionChecks(user *models.User) ([]ComplianceViolation, []string) {
	violations := []ComplianceViolation{}
	recommendations := []string{}

	// Check for data protection compliance
	if user.Email == "" {
		violations = append(violations, ComplianceViolation{
			Code:        "DP_001",
			Severity:    "high",
			Description: "Email address required for data protection compliance",
			Rule:        "GDPR Article 6",
			Remediation: "Provide valid email address for consent management",
		})
	}

	// Check for data minimization
	if user.FullName != "" && len(user.FullName) > 100 {
		violations = append(violations, ComplianceViolation{
			Code:        "DP_002",
			Severity:    "low",
			Description: "Full name exceeds reasonable length",
			Rule:        "GDPR Article 5",
			Remediation: "Verify name length and necessity",
		})
	}

	// Generate recommendations
	if len(violations) == 0 {
		recommendations = append(recommendations, "Data protection requirements met")
	} else {
		recommendations = append(recommendations, "Address data protection violations")
		recommendations = append(recommendations, "Ensure proper consent mechanisms are in place")
	}

	return violations, recommendations
}

// calculateComplianceScore calculates an overall compliance score based on violations.
func (s *ComplianceService) calculateComplianceScore(violations []ComplianceViolation) float64 {
	if len(violations) == 0 {
		return 100.0 // Perfect score
	}

	// Calculate score based on violation severity
	totalPenalty := 0.0
	for _, violation := range violations {
		switch violation.Severity {
		case "critical":
			totalPenalty += 25.0
		case "high":
			totalPenalty += 15.0
		case "medium":
			totalPenalty += 10.0
		case "low":
			totalPenalty += 5.0
		}
	}

	// Ensure score doesn't go below 0
	score := 100.0 - totalPenalty
	if score < 0 {
		score = 0
	}

	return score
}

// determineComplianceStatus determines the overall compliance status based on score and violations.
func (s *ComplianceService) determineComplianceStatus(score float64, violations []ComplianceViolation) string {
	// Check for critical violations
	for _, violation := range violations {
		if violation.Severity == "critical" {
			return "failed"
		}
	}

	// Determine status based on score
	switch {
	case score >= 90:
		return "passed"
	case score >= 70:
		return "warning"
	case score >= 50:
		return "pending"
	default:
		return "failed"
	}
}

// ValidatePolicyCompliance performs compliance validation for a policy.
func (s *ComplianceService) ValidatePolicyCompliance(ctx context.Context, policyID uuid.UUID) (*ComplianceCheck, error) {
	// Fetch policy details
	policy, err := s.policyStore.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policy: %w", err)
	}

	// Initialize compliance check
	check := &ComplianceCheck{
		EntityID:   policyID,
		EntityType: "policy",
		CheckType:  "regulatory",
		CheckDate:  time.Now(),
		ValidUntil: time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		Metadata:   make(map[string]interface{}),
	}

	// Perform policy-specific compliance checks
	violations := []ComplianceViolation{}
	recommendations := []string{}

	// Check policy validity
	if policy.ExpirationDate.Before(time.Now()) {
		violations = append(violations, ComplianceViolation{
			Code:        "POL_001",
			Severity:    "high",
			Description: "Policy has expired",
			Rule:        "Insurance Regulatory Requirements",
			Remediation: "Renew or cancel expired policy",
		})
	}

	// Check coverage amount limits
	if policy.CoverageAmount > 10000000 { // $10M limit
		violations = append(violations, ComplianceViolation{
			Code:        "POL_002",
			Severity:    "medium",
			Description: "Coverage amount exceeds regulatory limits",
			Rule:        "Maximum Coverage Limits",
			Remediation: "Reduce coverage amount or obtain special approval",
		})
	}

	// Check premium reasonableness
	if policy.Premium <= 0 {
		violations = append(violations, ComplianceViolation{
			Code:        "POL_003",
			Severity:    "critical",
			Description: "Premium amount is invalid",
			Rule:        "Premium Calculation Standards",
			Remediation: "Recalculate premium with valid amount",
		})
	}

	// Calculate compliance score
	check.Score = s.calculateComplianceScore(violations)
	check.Violations = violations
	check.Recommendations = recommendations
	check.Status = s.determineComplianceStatus(check.Score, violations)

	// Store metadata
	check.Metadata["policy_number"] = policy.PolicyNumber
	check.Metadata["coverage_amount"] = policy.CoverageAmount
	check.Metadata["premium"] = policy.Premium
	check.Metadata["check_version"] = "1.0"

	return check, nil
}

// ValidateClaimCompliance performs compliance validation for a claim.
func (s *ComplianceService) ValidateClaimCompliance(ctx context.Context, claimID uuid.UUID) (*ComplianceCheck, error) {
	// Fetch claim details
	claim, err := s.claimStore.GetClaim(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch claim: %w", err)
	}

	// Initialize compliance check
	check := &ComplianceCheck{
		EntityID:   claimID,
		EntityType: "claim",
		CheckType:  "fraud_prevention",
		CheckDate:  time.Now(),
		ValidUntil: time.Now().Add(180 * 24 * time.Hour), // Valid for 6 months
		Metadata:   make(map[string]interface{}),
	}

	// Perform claim-specific compliance checks
	violations := []ComplianceViolation{}
	recommendations := []string{}

	// Check claim amount reasonableness
	if claim.ClaimAmount <= 0 {
		violations = append(violations, ComplianceViolation{
			Code:        "CLM_001",
			Severity:    "critical",
			Description: "Claim amount is invalid",
			Rule:        "Claim Validation Standards",
			Remediation: "Provide valid claim amount",
		})
	}

	// Check incident date reasonableness
	if claim.IncidentDate.After(time.Now()) {
		violations = append(violations, ComplianceViolation{
			Code:        "CLM_002",
			Severity:    "high",
			Description: "Incident date is in the future",
			Rule:        "Claim Timeline Validation",
			Remediation: "Provide valid incident date",
		})
	}

	// Check reporting delay
	reportingDelay := claim.ReportedDate.Sub(claim.IncidentDate).Hours() / 24
	if reportingDelay > 365 { // More than 1 year
		violations = append(violations, ComplianceViolation{
			Code:        "CLM_003",
			Severity:    "medium",
			Description: "Excessive delay in claim reporting",
			Rule:        "Timely Reporting Requirements",
			Remediation: "Provide justification for reporting delay",
		})
	}

	// Calculate compliance score
	check.Score = s.calculateComplianceScore(violations)
	check.Violations = violations
	check.Recommendations = recommendations
	check.Status = s.determineComplianceStatus(check.Score, violations)

	// Store metadata
	check.Metadata["claim_number"] = claim.ClaimNumber
	check.Metadata["claim_amount"] = claim.ClaimAmount
	check.Metadata["incident_date"] = claim.IncidentDate
	check.Metadata["check_version"] = "1.0"

	return check, nil
}

// GetComplianceRules retrieves active compliance rules for a jurisdiction.
func (s *ComplianceService) GetComplianceRules(ctx context.Context, jurisdiction string) ([]ComplianceRule, error) {
	// In a real implementation, this would query compliance rules from the database
	// For now, we'll return default rules
	rules := []ComplianceRule{
		{
			ID:            "KYC_001",
			Name:          "Customer Identification Program",
			Description:   "Requires collection of customer identification information",
			Category:      "kyc",
			Jurisdiction:  jurisdiction,
			Severity:      "high",
			Active:        true,
			EffectiveDate: time.Now().AddDate(-1, 0, 0), // 1 year ago
			Conditions:    make(map[string]interface{}),
			Metadata:      make(map[string]interface{}),
		},
		{
			ID:            "AML_001",
			Name:          "Anti-Money Laundering",
			Description:   "Requires monitoring for suspicious activities",
			Category:      "aml",
			Jurisdiction:  jurisdiction,
			Severity:      "critical",
			Active:        true,
			EffectiveDate: time.Now().AddDate(-1, 0, 0), // 1 year ago
			Conditions:    make(map[string]interface{}),
			Metadata:      make(map[string]interface{}),
		},
	}

	return rules, nil
}

// UpdateComplianceRule updates a compliance rule.
func (s *ComplianceService) UpdateComplianceRule(ctx context.Context, rule *ComplianceRule) error {
	// Validate compliance rule
	if err := s.validateComplianceRule(rule); err != nil {
		return fmt.Errorf("invalid compliance rule: %w", err)
	}

	// In a real implementation, this would update the rule in the database
	// For now, we'll just validate the input
	return nil
}

// validateComplianceRule validates a compliance rule.
func (s *ComplianceService) validateComplianceRule(rule *ComplianceRule) error {
	if rule == nil {
		return fmt.Errorf("compliance rule cannot be nil")
	}

	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.Category == "" {
		return fmt.Errorf("rule category is required")
	}

	if rule.Jurisdiction == "" {
		return fmt.Errorf("jurisdiction is required")
	}

	if rule.Severity == "" {
		return fmt.Errorf("severity is required")
	}

	validSeverities := []string{"low", "medium", "high", "critical"}
	valid := false
	for _, validSeverity := range validSeverities {
		if rule.Severity == validSeverity {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid severity: %s", rule.Severity)
	}

	if rule.EffectiveDate.IsZero() {
		return fmt.Errorf("effective date is required")
	}

	if rule.ExpirationDate != nil && rule.ExpirationDate.Before(rule.EffectiveDate) {
		return fmt.Errorf("expiration date must be after effective date")
	}

	return nil
}

// ValidateComplianceCheck validates the integrity of a compliance check.
func (s *ComplianceService) ValidateComplianceCheck(check *ComplianceCheck) error {
	if check == nil {
		return fmt.Errorf("compliance check cannot be nil")
	}

	if check.EntityID == uuid.Nil {
		return fmt.Errorf("entity ID is required")
	}

	if check.EntityType == "" {
		return fmt.Errorf("entity type is required")
	}

	if check.CheckType == "" {
		return fmt.Errorf("check type is required")
	}

	if check.Status == "" {
		return fmt.Errorf("status is required")
	}

	validStatuses := []string{"passed", "failed", "warning", "pending"}
	valid := false
	for _, validStatus := range validStatuses {
		if check.Status == validStatus {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status: %s", check.Status)
	}

	if check.Score < 0 || check.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100")
	}

	if check.CheckDate.IsZero() {
		return fmt.Errorf("check date is required")
	}

	if check.ValidUntil.IsZero() {
		return fmt.Errorf("valid until date is required")
	}

	if check.ValidUntil.Before(check.CheckDate) {
		return fmt.Errorf("valid until date must be after check date")
	}

	return nil
}

