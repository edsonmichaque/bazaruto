package services

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// RiskAssessmentService handles comprehensive risk assessment for insurance applications and policies.
type RiskAssessmentService struct {
	userStore   store.UserStore
	policyStore store.PolicyStore
	claimStore  store.ClaimStore
}

// NewRiskAssessmentService creates a new RiskAssessmentService instance.
func NewRiskAssessmentService(userStore store.UserStore, policyStore store.PolicyStore, claimStore store.ClaimStore) *RiskAssessmentService {
	return &RiskAssessmentService{
		userStore:   userStore,
		policyStore: policyStore,
		claimStore:  claimStore,
	}
}

// RiskProfile represents a comprehensive risk assessment result.
type RiskProfile struct {
	OverallScore      float64                `json:"overall_score"`      // 0-100, higher means higher risk
	RiskLevel         string                 `json:"risk_level"`         // low, medium, high, very_high
	RiskCategory      string                 `json:"risk_category"`      // personal, commercial, specialty
	Assessments       []RiskAssessment       `json:"assessments"`        // Individual risk assessments
	Recommendations   []string               `json:"recommendations"`    // Risk mitigation recommendations
	PremiumAdjustment float64                `json:"premium_adjustment"` // Percentage adjustment to base premium
	ApprovalStatus    string                 `json:"approval_status"`    // approved, declined, conditional
	Conditions        []string               `json:"conditions"`         // Conditions for approval
	AssessmentDate    time.Time              `json:"assessment_date"`
	ValidUntil        time.Time              `json:"valid_until"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// RiskAssessment represents an individual risk factor assessment.
type RiskAssessment struct {
	Factor      string    `json:"factor"`       // Name of the risk factor
	Category    string    `json:"category"`     // Category of risk (demographic, behavioral, etc.)
	Score       float64   `json:"score"`        // Risk score for this factor (0-100)
	Weight      float64   `json:"weight"`       // Weight of this factor in overall assessment
	Impact      float64   `json:"impact"`       // Impact on premium (multiplier)
	Description string    `json:"description"`  // Human-readable description
	Severity    string    `json:"severity"`     // low, medium, high, critical
	Mitigation  string    `json:"mitigation"`   // Suggested mitigation strategy
	DataQuality string    `json:"data_quality"` // Quality of data used (excellent, good, fair, poor)
	LastUpdated time.Time `json:"last_updated"`
}

// AssessRisk performs comprehensive risk assessment for a user and product combination.
func (s *RiskAssessmentService) AssessRisk(ctx context.Context, userID uuid.UUID, productID uuid.UUID, coverageAmount float64) (*RiskProfile, error) {
	// Fetch user details
	user, err := s.userStore.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Perform comprehensive risk assessments
	assessments := []RiskAssessment{
		s.assessDemographicRisk(user),
		s.assessBehavioralRisk(ctx, user),
		s.assessFinancialRisk(ctx, user),
		s.assessGeographicRisk(user),
		s.assessProductSpecificRisk(productID, coverageAmount),
		s.assessHistoricalRisk(ctx, user),
		s.assessLifestyleRisk(user),
		s.assessComplianceRisk(user),
	}

	// Calculate overall risk score
	profile := &RiskProfile{
		AssessmentDate: time.Now(),
		ValidUntil:     time.Now().Add(90 * 24 * time.Hour), // Valid for 90 days
		Metadata:       make(map[string]interface{}),
	}

	// Calculate weighted overall score
	totalWeight := 0.0
	weightedScore := 0.0
	totalImpact := 1.0

	for _, assessment := range assessments {
		if assessment.Weight > 0 {
			totalWeight += assessment.Weight
			weightedScore += assessment.Score * assessment.Weight
			totalImpact *= assessment.Impact
		}
	}

	if totalWeight > 0 {
		profile.OverallScore = weightedScore / totalWeight
	}

	profile.Assessments = assessments
	profile.RiskLevel = s.determineRiskLevel(profile.OverallScore)
	profile.RiskCategory = s.determineRiskCategory(productID)
	profile.PremiumAdjustment = s.calculatePremiumAdjustment(profile.OverallScore, totalImpact)
	profile.ApprovalStatus = s.determineApprovalStatus(profile.OverallScore, assessments)
	profile.Conditions = s.generateApprovalConditions(profile.OverallScore, assessments)
	profile.Recommendations = s.generateRiskRecommendations(profile, assessments)

	// Store metadata
	profile.Metadata["user_id"] = userID.String()
	profile.Metadata["product_id"] = productID.String()
	profile.Metadata["coverage_amount"] = coverageAmount
	profile.Metadata["assessment_version"] = "2.0"

	return profile, nil
}

// assessDemographicRisk assesses demographic risk factors.
func (s *RiskAssessmentService) assessDemographicRisk(user *models.User) RiskAssessment {
	assessment := RiskAssessment{
		Factor:      "demographic_risk",
		Category:    "demographic",
		Weight:      0.20,
		LastUpdated: time.Now(),
	}

	// Age-based risk assessment (simplified)
	// In a real implementation, this would use actual age data
	assessment.Score = 30 // Default moderate risk
	assessment.Description = "Demographic risk assessment based on user profile"
	assessment.Severity = "medium"
	assessment.Impact = 1.0
	assessment.DataQuality = "good"
	assessment.Mitigation = "Standard demographic risk factors applied"

	return assessment
}

// assessBehavioralRisk assesses behavioral risk factors.
func (s *RiskAssessmentService) assessBehavioralRisk(ctx context.Context, user *models.User) RiskAssessment {
	assessment := RiskAssessment{
		Factor:      "behavioral_risk",
		Category:    "behavioral",
		Weight:      0.25,
		LastUpdated: time.Now(),
	}

	// Check user's account activity and behavior patterns
	accountAge := time.Since(user.CreatedAt).Hours() / 24 / 365 // years

	if accountAge < 0.5 { // Less than 6 months
		assessment.Score = 60
		assessment.Description = "New user with limited behavioral history"
		assessment.Severity = "high"
		assessment.Impact = 1.3
	} else if accountAge < 2 {
		assessment.Score = 40
		assessment.Description = "User with moderate behavioral history"
		assessment.Severity = "medium"
		assessment.Impact = 1.1
	} else {
		assessment.Score = 20
		assessment.Description = "Established user with good behavioral history"
		assessment.Severity = "low"
		assessment.Impact = 0.9
	}

	assessment.DataQuality = "good"
	assessment.Mitigation = "Monitor user behavior patterns and account activity"

	return assessment
}

// assessFinancialRisk assesses financial risk factors.
func (s *RiskAssessmentService) assessFinancialRisk(ctx context.Context, user *models.User) RiskAssessment {
	assessment := RiskAssessment{
		Factor:      "financial_risk",
		Category:    "financial",
		Weight:      0.30,
		LastUpdated: time.Now(),
	}

	// In a real implementation, this would integrate with credit bureaus,
	// financial data providers, and payment history
	assessment.Score = 25 // Default low-moderate risk
	assessment.Description = "Financial risk assessment based on available data"
	assessment.Severity = "low"
	assessment.Impact = 1.0
	assessment.DataQuality = "fair" // Limited financial data available
	assessment.Mitigation = "Request additional financial documentation for high-value policies"

	return assessment
}

// assessGeographicRisk assesses geographic risk factors.
func (s *RiskAssessmentService) assessGeographicRisk(user *models.User) RiskAssessment {
	assessment := RiskAssessment{
		Factor:      "geographic_risk",
		Category:    "geographic",
		Weight:      0.15,
		LastUpdated: time.Now(),
	}

	// In a real implementation, this would use geographic risk databases
	// and historical claim data by region
	assessment.Score = 35 // Default moderate risk
	assessment.Description = "Geographic risk assessment based on location data"
	assessment.Severity = "medium"
	assessment.Impact = 1.1
	assessment.DataQuality = "good"
	assessment.Mitigation = "Consider location-specific risk factors and coverage options"

	return assessment
}

// assessProductSpecificRisk assesses product-specific risk factors.
func (s *RiskAssessmentService) assessProductSpecificRisk(productID uuid.UUID, coverageAmount float64) RiskAssessment {
	assessment := RiskAssessment{
		Factor:      "product_specific_risk",
		Category:    "product",
		Weight:      0.20,
		LastUpdated: time.Now(),
	}

	// Assess risk based on coverage amount
	if coverageAmount > 1000000 { // Over $1M
		assessment.Score = 70
		assessment.Description = "High-value coverage requiring enhanced risk assessment"
		assessment.Severity = "high"
		assessment.Impact = 1.5
	} else if coverageAmount > 500000 { // Over $500K
		assessment.Score = 50
		assessment.Description = "Moderate to high-value coverage"
		assessment.Severity = "medium"
		assessment.Impact = 1.2
	} else if coverageAmount > 100000 { // Over $100K
		assessment.Score = 30
		assessment.Description = "Standard coverage amount"
		assessment.Severity = "low"
		assessment.Impact = 1.0
	} else {
		assessment.Score = 20
		assessment.Description = "Low-value coverage"
		assessment.Severity = "low"
		assessment.Impact = 0.9
	}

	assessment.DataQuality = "excellent"
	assessment.Mitigation = "Adjust coverage limits and deductibles based on risk profile"

	return assessment
}

// assessHistoricalRisk assesses historical risk factors.
func (s *RiskAssessmentService) assessHistoricalRisk(ctx context.Context, user *models.User) RiskAssessment {
	assessment := RiskAssessment{
		Factor:      "historical_risk",
		Category:    "historical",
		Weight:      0.25,
		LastUpdated: time.Now(),
	}

	// Check user's claim history
	// In a real implementation, this would query actual claim data
	assessment.Score = 20 // Default low risk
	assessment.Description = "Historical risk assessment based on claim history"
	assessment.Severity = "low"
	assessment.Impact = 0.95
	assessment.DataQuality = "good"
	assessment.Mitigation = "Continue monitoring claim patterns and frequency"

	return assessment
}

// assessLifestyleRisk assesses lifestyle risk factors.
func (s *RiskAssessmentService) assessLifestyleRisk(user *models.User) RiskAssessment {
	assessment := RiskAssessment{
		Factor:      "lifestyle_risk",
		Category:    "lifestyle",
		Weight:      0.10,
		LastUpdated: time.Now(),
	}

	// In a real implementation, this would use lifestyle questionnaires,
	// social media analysis, and other data sources
	assessment.Score = 30 // Default moderate risk
	assessment.Description = "Lifestyle risk assessment based on available data"
	assessment.Severity = "medium"
	assessment.Impact = 1.05
	assessment.DataQuality = "fair"
	assessment.Mitigation = "Request lifestyle questionnaire for comprehensive assessment"

	return assessment
}

// assessComplianceRisk assesses compliance and regulatory risk factors.
func (s *RiskAssessmentService) assessComplianceRisk(user *models.User) RiskAssessment {
	assessment := RiskAssessment{
		Factor:      "compliance_risk",
		Category:    "compliance",
		Weight:      0.05,
		LastUpdated: time.Now(),
	}

	// Check user status and compliance indicators
	if user.Status != models.StatusActive {
		assessment.Score = 80
		assessment.Description = "User account is not in active status"
		assessment.Severity = "critical"
		assessment.Impact = 2.0
	} else {
		assessment.Score = 10
		assessment.Description = "User account is in good standing"
		assessment.Severity = "low"
		assessment.Impact = 1.0
	}

	assessment.DataQuality = "excellent"
	assessment.Mitigation = "Ensure user account is in good standing before policy issuance"

	return assessment
}

// determineRiskLevel determines the overall risk level based on the score.
func (s *RiskAssessmentService) determineRiskLevel(score float64) string {
	switch {
	case score >= 80:
		return "very_high"
	case score >= 60:
		return "high"
	case score >= 40:
		return "medium"
	default:
		return "low"
	}
}

// determineRiskCategory determines the risk category based on the product.
func (s *RiskAssessmentService) determineRiskCategory(productID uuid.UUID) string {
	// In a real implementation, this would query the product details
	// For now, we'll use a simplified approach
	return "personal" // Default to personal insurance
}

// calculatePremiumAdjustment calculates the premium adjustment based on risk.
func (s *RiskAssessmentService) calculatePremiumAdjustment(score float64, totalImpact float64) float64 {
	// Base adjustment from overall score
	baseAdjustment := (score - 50) * 0.02 // 2% per point above/below 50

	// Apply impact multiplier
	impactAdjustment := (totalImpact - 1.0) * 100

	// Combine adjustments
	totalAdjustment := baseAdjustment + impactAdjustment

	// Cap adjustments at reasonable limits
	if totalAdjustment > 200 { // Max 200% increase
		totalAdjustment = 200
	} else if totalAdjustment < -50 { // Max 50% decrease
		totalAdjustment = -50
	}

	return totalAdjustment
}

// determineApprovalStatus determines the approval status based on risk assessment.
func (s *RiskAssessmentService) determineApprovalStatus(score float64, assessments []RiskAssessment) string {
	// Check for critical risk factors
	for _, assessment := range assessments {
		if assessment.Severity == "critical" {
			return "declined"
		}
	}

	// Check overall score
	switch {
	case score >= 80:
		return "declined"
	case score >= 60:
		return "conditional"
	default:
		return "approved"
	}
}

// generateApprovalConditions generates conditions for conditional approval.
func (s *RiskAssessmentService) generateApprovalConditions(score float64, assessments []RiskAssessment) []string {
	conditions := []string{}

	if score >= 60 {
		conditions = append(conditions, "Enhanced monitoring required")
		conditions = append(conditions, "Quarterly risk review mandatory")
	}

	// Add conditions based on high-risk factors
	for _, assessment := range assessments {
		if assessment.Severity == "high" {
			switch assessment.Factor {
			case "financial_risk":
				conditions = append(conditions, "Additional financial documentation required")
			case "behavioral_risk":
				conditions = append(conditions, "Extended probationary period")
			case "product_specific_risk":
				conditions = append(conditions, "Reduced coverage limits or higher deductibles")
			}
		}
	}

	return conditions
}

// generateRiskRecommendations generates risk mitigation recommendations.
func (s *RiskAssessmentService) generateRiskRecommendations(profile *RiskProfile, assessments []RiskAssessment) []string {
	recommendations := []string{}

	// General recommendations based on risk level
	switch profile.RiskLevel {
	case "very_high":
		recommendations = append(recommendations, "Consider declining application or requiring significant risk mitigation")
		recommendations = append(recommendations, "Implement enhanced monitoring and reporting")
		recommendations = append(recommendations, "Require additional security measures")

	case "high":
		recommendations = append(recommendations, "Implement enhanced risk monitoring")
		recommendations = append(recommendations, "Require additional documentation and verification")
		recommendations = append(recommendations, "Consider higher deductibles or reduced coverage")

	case "medium":
		recommendations = append(recommendations, "Standard risk monitoring procedures")
		recommendations = append(recommendations, "Regular risk assessment reviews")
		recommendations = append(recommendations, "Consider risk mitigation strategies")

	case "low":
		recommendations = append(recommendations, "Standard processing and monitoring")
		recommendations = append(recommendations, "Consider premium discounts for low-risk profile")
	}

	// Specific recommendations based on risk factors
	for _, assessment := range assessments {
		if assessment.Severity == "high" || assessment.Severity == "critical" {
			recommendations = append(recommendations, assessment.Mitigation)
		}
	}

	return recommendations
}

// UpdateRiskAssessment updates an existing risk assessment with new data.
func (s *RiskAssessmentService) UpdateRiskAssessment(ctx context.Context, userID uuid.UUID, productID uuid.UUID, newData map[string]interface{}) (*RiskProfile, error) {
	// In a real implementation, this would update the risk assessment with new data
	// For now, we'll re-assess with the new information
	coverageAmount := 100000.0 // Default coverage amount
	if amount, ok := newData["coverage_amount"].(float64); ok {
		coverageAmount = amount
	}

	return s.AssessRisk(ctx, userID, productID, coverageAmount)
}

// GetRiskAssessment retrieves a previously calculated risk assessment.
func (s *RiskAssessmentService) GetRiskAssessment(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*RiskProfile, error) {
	// In a real implementation, this would retrieve from a risk assessment store
	// For now, we'll re-assess
	return s.AssessRisk(ctx, userID, productID, 100000.0) // Default coverage amount
}

// ValidateRiskAssessment validates the integrity of a risk assessment.
func (s *RiskAssessmentService) ValidateRiskAssessment(profile *RiskProfile) error {
	if profile == nil {
		return fmt.Errorf("risk profile cannot be nil")
	}

	if profile.OverallScore < 0 || profile.OverallScore > 100 {
		return fmt.Errorf("overall score must be between 0 and 100")
	}

	if profile.AssessmentDate.IsZero() {
		return fmt.Errorf("assessment date is required")
	}

	if profile.ValidUntil.IsZero() {
		return fmt.Errorf("valid until date is required")
	}

	if profile.ValidUntil.Before(profile.AssessmentDate) {
		return fmt.Errorf("valid until date must be after assessment date")
	}

	// Validate individual assessments
	for _, assessment := range profile.Assessments {
		if assessment.Score < 0 || assessment.Score > 100 {
			return fmt.Errorf("assessment score must be between 0 and 100")
		}
		if assessment.Weight < 0 || assessment.Weight > 1 {
			return fmt.Errorf("assessment weight must be between 0 and 1")
		}
	}

	return nil
}
