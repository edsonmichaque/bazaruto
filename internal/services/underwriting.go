package services

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// UnderwritingService handles automated underwriting decisions and policy approval/rejection.
type UnderwritingService struct {
	userStore      store.UserStore
	policyStore    store.PolicyStore
	claimStore     store.ClaimStore
	riskService    *RiskAssessmentService
	fraudService   *FraudDetectionService
	pricingService *PricingEngineService
}

// NewUnderwritingService creates a new UnderwritingService instance.
func NewUnderwritingService(
	userStore store.UserStore,
	policyStore store.PolicyStore,
	claimStore store.ClaimStore,
	riskService *RiskAssessmentService,
	fraudService *FraudDetectionService,
	pricingService *PricingEngineService,
) *UnderwritingService {
	return &UnderwritingService{
		userStore:      userStore,
		policyStore:    policyStore,
		claimStore:     claimStore,
		riskService:    riskService,
		fraudService:   fraudService,
		pricingService: pricingService,
	}
}

// UnderwritingDecision represents the result of an underwriting decision.
type UnderwritingDecision struct {
	Decision        string                  `json:"decision"`        // approved, declined, conditional, pending_review
	Confidence      float64                 `json:"confidence"`      // 0-1, confidence in the decision
	RiskScore       float64                 `json:"risk_score"`      // 0-100, overall risk score
	Premium         float64                 `json:"premium"`         // Calculated premium
	Currency        string                  `json:"currency"`        // Currency code
	Conditions      []UnderwritingCondition `json:"conditions"`      // Conditions for approval
	Reasons         []string                `json:"reasons"`         // Reasons for decision
	Recommendations []string                `json:"recommendations"` // Recommendations
	ValidUntil      time.Time               `json:"valid_until"`     // Decision validity period
	Metadata        map[string]interface{}  `json:"metadata"`
}

// UnderwritingCondition represents a condition that must be met for policy approval.
type UnderwritingCondition struct {
	Type        string                 `json:"type"`        // documentation, payment, inspection, etc.
	Description string                 `json:"description"` // Human-readable description
	Required    bool                   `json:"required"`    // Whether condition is mandatory
	Deadline    *time.Time             `json:"deadline"`    // Deadline for meeting condition
	Metadata    map[string]interface{} `json:"metadata"`
}

// UnderwritingRequest represents a request for underwriting decision.
type UnderwritingRequest struct {
	UserID           uuid.UUID              `json:"user_id"`
	ProductID        uuid.UUID              `json:"product_id"`
	CoverageAmount   float64                `json:"coverage_amount"`
	Currency         string                 `json:"currency"`
	PaymentFrequency string                 `json:"payment_frequency"`
	EffectiveDate    time.Time              `json:"effective_date"`
	ExpirationDate   time.Time              `json:"expiration_date"`
	ApplicationData  map[string]interface{} `json:"application_data"`
	RiskFactors      map[string]interface{} `json:"risk_factors"`
	Options          map[string]interface{} `json:"options"`
}

// ProcessUnderwriting performs comprehensive underwriting analysis and makes a decision.
func (s *UnderwritingService) ProcessUnderwriting(ctx context.Context, request *UnderwritingRequest) (*UnderwritingDecision, error) {
	// Validate underwriting request
	if err := s.validateUnderwritingRequest(request); err != nil {
		return nil, fmt.Errorf("invalid underwriting request: %w", err)
	}

	// Initialize decision
	decision := &UnderwritingDecision{
		Currency:   request.Currency,
		ValidUntil: time.Now().Add(30 * 24 * time.Hour), // Valid for 30 days
		Metadata:   make(map[string]interface{}),
	}

	// Perform risk assessment
	riskProfile, err := s.riskService.AssessRisk(ctx, request.UserID, request.ProductID, request.CoverageAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to assess risk: %w", err)
	}

	decision.RiskScore = riskProfile.OverallScore

	// Calculate premium
	pricingRequest := &PricingRequest{
		ProductID:        request.ProductID,
		UserID:           request.UserID,
		CoverageAmount:   request.CoverageAmount,
		Currency:         request.Currency,
		PaymentFrequency: request.PaymentFrequency,
		EffectiveDate:    request.EffectiveDate,
		ExpirationDate:   request.ExpirationDate,
		RiskFactors:      request.RiskFactors,
		Options:          request.Options,
	}

	pricingResult, err := s.pricingService.CalculatePremium(ctx, pricingRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate premium: %w", err)
	}

	decision.Premium = pricingResult.FinalPremium

	// Make underwriting decision based on risk profile and other factors
	decision.Decision = s.determineDecision(riskProfile, request)
	decision.Confidence = s.calculateConfidence(riskProfile, request)
	decision.Conditions = s.generateConditions(decision.Decision, riskProfile, request)
	decision.Reasons = s.generateReasons(decision.Decision, riskProfile, request)
	decision.Recommendations = s.generateRecommendations(decision.Decision, riskProfile, request)

	// Store metadata
	decision.Metadata["user_id"] = request.UserID.String()
	decision.Metadata["product_id"] = request.ProductID.String()
	decision.Metadata["coverage_amount"] = request.CoverageAmount
	decision.Metadata["underwriting_version"] = "1.0"

	return decision, nil
}

// validateUnderwritingRequest validates the underwriting request.
func (s *UnderwritingService) validateUnderwritingRequest(request *UnderwritingRequest) error {
	if request.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if request.ProductID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}
	if request.CoverageAmount <= 0 {
		return fmt.Errorf("coverage amount must be greater than 0")
	}
	if request.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if request.PaymentFrequency == "" {
		return fmt.Errorf("payment frequency is required")
	}
	if request.EffectiveDate.IsZero() {
		return fmt.Errorf("effective date is required")
	}
	if request.ExpirationDate.IsZero() {
		return fmt.Errorf("expiration date is required")
	}
	if request.ExpirationDate.Before(request.EffectiveDate) {
		return fmt.Errorf("expiration date must be after effective date")
	}
	return nil
}

// determineDecision determines the underwriting decision based on risk assessment.
func (s *UnderwritingService) determineDecision(riskProfile *RiskProfile, request *UnderwritingRequest) string {
	// Check for critical risk factors
	for _, assessment := range riskProfile.Assessments {
		if assessment.Severity == "critical" {
			return "declined"
		}
	}

	// Check overall risk score
	switch {
	case riskProfile.OverallScore >= 80:
		return "declined"
	case riskProfile.OverallScore >= 60:
		return "conditional"
	case riskProfile.OverallScore >= 40:
		return "pending_review"
	default:
		return "approved"
	}
}

// calculateConfidence calculates the confidence level in the underwriting decision.
func (s *UnderwritingService) calculateConfidence(riskProfile *RiskProfile, request *UnderwritingRequest) float64 {
	// Base confidence on risk assessment (simplified)
	baseConfidence := 0.8 // Default confidence level

	// Adjust confidence based on data quality
	dataQualityScore := 0.0
	for _, assessment := range riskProfile.Assessments {
		switch assessment.DataQuality {
		case "excellent":
			dataQualityScore += 1.0
		case "good":
			dataQualityScore += 0.8
		case "fair":
			dataQualityScore += 0.6
		case "poor":
			dataQualityScore += 0.4
		}
	}

	// Average data quality
	if len(riskProfile.Assessments) > 0 {
		dataQualityScore /= float64(len(riskProfile.Assessments))
	}

	// Combine base confidence with data quality
	finalConfidence := (baseConfidence + dataQualityScore) / 2.0

	// Ensure confidence is within bounds
	if finalConfidence > 1.0 {
		finalConfidence = 1.0
	} else if finalConfidence < 0.0 {
		finalConfidence = 0.0
	}

	return finalConfidence
}

// generateConditions generates conditions for conditional approval.
func (s *UnderwritingService) generateConditions(decision string, riskProfile *RiskProfile, request *UnderwritingRequest) []UnderwritingCondition {
	conditions := []UnderwritingCondition{}

	if decision != "conditional" {
		return conditions
	}

	// Generate conditions based on high-risk factors
	for _, assessment := range riskProfile.Assessments {
		if assessment.Severity == "high" {
			switch assessment.Factor {
			case "financial_risk":
				conditions = append(conditions, UnderwritingCondition{
					Type:        "documentation",
					Description: "Provide additional financial documentation",
					Required:    true,
					Deadline:    s.calculateDeadline(30), // 30 days
					Metadata:    map[string]interface{}{"category": "financial"},
				})
			case "behavioral_risk":
				conditions = append(conditions, UnderwritingCondition{
					Type:        "monitoring",
					Description: "Enhanced monitoring period required",
					Required:    true,
					Deadline:    s.calculateDeadline(90), // 90 days
					Metadata:    map[string]interface{}{"category": "behavioral"},
				})
			case "product_specific_risk":
				conditions = append(conditions, UnderwritingCondition{
					Type:        "inspection",
					Description: "Professional inspection required",
					Required:    true,
					Deadline:    s.calculateDeadline(14), // 14 days
					Metadata:    map[string]interface{}{"category": "inspection"},
				})
			}
		}
	}

	// Add general conditions for high-value coverage
	if request.CoverageAmount > 500000 {
		conditions = append(conditions, UnderwritingCondition{
			Type:        "payment",
			Description: "Advance payment required",
			Required:    true,
			Deadline:    s.calculateDeadline(7), // 7 days
			Metadata:    map[string]interface{}{"category": "payment"},
		})
	}

	return conditions
}

// generateReasons generates reasons for the underwriting decision.
func (s *UnderwritingService) generateReasons(decision string, riskProfile *RiskProfile, request *UnderwritingRequest) []string {
	reasons := []string{}

	switch decision {
	case "approved":
		reasons = append(reasons, "Risk assessment indicates acceptable risk level")
		reasons = append(reasons, "All underwriting criteria met")
		reasons = append(reasons, "Premium calculated within acceptable range")

	case "conditional":
		reasons = append(reasons, "Risk assessment indicates elevated risk requiring additional conditions")
		reasons = append(reasons, "Some underwriting criteria require additional verification")

		// Add specific reasons based on high-risk factors
		for _, assessment := range riskProfile.Assessments {
			if assessment.Severity == "high" {
				reasons = append(reasons, fmt.Sprintf("High risk in %s category", assessment.Factor))
			}
		}

	case "pending_review":
		reasons = append(reasons, "Risk assessment requires manual review")
		reasons = append(reasons, "Automated decision not possible with current data")
		reasons = append(reasons, "Additional information required for final decision")

	case "declined":
		reasons = append(reasons, "Risk assessment indicates unacceptable risk level")
		reasons = append(reasons, "Underwriting criteria not met")

		// Add specific reasons based on critical risk factors
		for _, assessment := range riskProfile.Assessments {
			if assessment.Severity == "critical" {
				reasons = append(reasons, fmt.Sprintf("Critical risk in %s category", assessment.Factor))
			}
		}
	}

	return reasons
}

// generateRecommendations generates recommendations based on the underwriting decision.
func (s *UnderwritingService) generateRecommendations(decision string, riskProfile *RiskProfile, request *UnderwritingRequest) []string {
	recommendations := []string{}

	switch decision {
	case "approved":
		recommendations = append(recommendations, "Policy can be issued immediately")
		recommendations = append(recommendations, "Standard monitoring procedures apply")
		recommendations = append(recommendations, "Consider additional coverage options")

	case "conditional":
		recommendations = append(recommendations, "Meet all specified conditions before policy issuance")
		recommendations = append(recommendations, "Enhanced monitoring will be required")
		recommendations = append(recommendations, "Consider risk mitigation strategies")

	case "pending_review":
		recommendations = append(recommendations, "Provide additional documentation for review")
		recommendations = append(recommendations, "Manual underwriting review will be conducted")
		recommendations = append(recommendations, "Decision will be communicated within 5 business days")

	case "declined":
		recommendations = append(recommendations, "Consider alternative coverage options")
		recommendations = append(recommendations, "Address risk factors before reapplication")
		recommendations = append(recommendations, "Reapply after 6 months with improved risk profile")
	}

	// Add specific recommendations based on risk factors
	for _, assessment := range riskProfile.Assessments {
		if assessment.Severity == "high" || assessment.Severity == "critical" {
			recommendations = append(recommendations, assessment.Mitigation)
		}
	}

	return recommendations
}

// calculateDeadline calculates a deadline based on days from now.
func (s *UnderwritingService) calculateDeadline(days int) *time.Time {
	deadline := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	return &deadline
}

// ReviewUnderwritingDecision allows manual review and override of automated underwriting decisions.
func (s *UnderwritingService) ReviewUnderwritingDecision(ctx context.Context, decisionID string, review *UnderwritingReview) (*UnderwritingDecision, error) {
	// In a real implementation, this would retrieve the original decision and apply the review
	// For now, we'll simulate the review process

	// Validate review
	if review == nil {
		return nil, fmt.Errorf("underwriting review cannot be nil")
	}

	if review.ReviewerID == uuid.Nil {
		return nil, fmt.Errorf("reviewer ID is required")
	}

	if review.Decision == "" {
		return nil, fmt.Errorf("review decision is required")
	}

	// Create updated decision based on review
	updatedDecision := &UnderwritingDecision{
		Decision:   review.Decision,
		Confidence: 1.0, // Manual review has high confidence
		Reasons:    []string{review.Reason},
		ValidUntil: time.Now().Add(30 * 24 * time.Hour),
		Metadata: map[string]interface{}{
			"reviewer_id":          review.ReviewerID.String(),
			"review_date":          time.Now(),
			"original_decision_id": decisionID,
		},
	}

	return updatedDecision, nil
}

// UnderwritingReview represents a manual review of an underwriting decision.
type UnderwritingReview struct {
	ReviewerID uuid.UUID `json:"reviewer_id"`
	Decision   string    `json:"decision"` // approved, declined, conditional
	Reason     string    `json:"reason"`   // Reason for the review decision
	Comments   string    `json:"comments"` // Additional comments
}

// GetUnderwritingHistory retrieves underwriting history for a user.
func (s *UnderwritingService) GetUnderwritingHistory(ctx context.Context, userID uuid.UUID) ([]UnderwritingDecision, error) {
	// In a real implementation, this would query underwriting history from the database
	// For now, we'll return an empty slice
	return []UnderwritingDecision{}, nil
}

// UpdateUnderwritingRules updates underwriting rules and criteria.
func (s *UnderwritingService) UpdateUnderwritingRules(ctx context.Context, rules *UnderwritingRules) error {
	// In a real implementation, this would update underwriting rules in the database
	// For now, we'll validate the input
	if rules == nil {
		return fmt.Errorf("underwriting rules cannot be nil")
	}

	if rules.ProductID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}

	if rules.Rules == nil {
		return fmt.Errorf("rules cannot be nil")
	}

	return nil
}

// UnderwritingRules represents underwriting rules for a product.
type UnderwritingRules struct {
	ProductID uuid.UUID              `json:"product_id"`
	Rules     map[string]interface{} `json:"rules"`
	Version   string                 `json:"version"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ValidateUnderwritingDecision validates the integrity of an underwriting decision.
func (s *UnderwritingService) ValidateUnderwritingDecision(decision *UnderwritingDecision) error {
	if decision == nil {
		return fmt.Errorf("underwriting decision cannot be nil")
	}

	if decision.Decision == "" {
		return fmt.Errorf("decision is required")
	}

	validDecisions := []string{"approved", "declined", "conditional", "pending_review"}
	valid := false
	for _, validDecision := range validDecisions {
		if decision.Decision == validDecision {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid decision: %s", decision.Decision)
	}

	if decision.Confidence < 0 || decision.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	if decision.RiskScore < 0 || decision.RiskScore > 100 {
		return fmt.Errorf("risk score must be between 0 and 100")
	}

	if decision.Premium < 0 {
		return fmt.Errorf("premium cannot be negative")
	}

	if decision.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if decision.ValidUntil.IsZero() {
		return fmt.Errorf("valid until date is required")
	}

	if decision.ValidUntil.Before(time.Now()) {
		return fmt.Errorf("valid until date must be in the future")
	}

	return nil
}
