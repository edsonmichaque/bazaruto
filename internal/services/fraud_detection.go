package services

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/events"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FraudDetectionService handles fraud detection and risk assessment for claims and applications.
// This version uses Customer domain model, dynamic configuration, and events.
type FraudDetectionService struct {
	claimStore    store.ClaimStore
	policyStore   store.PolicyStore
	customerStore store.CustomerStore
	configManager *config.ConfigManager
	eventService  *EventService
	logger        *logger.Logger
}

// NewFraudDetectionService creates a new FraudDetectionService instance.
func NewFraudDetectionService(
	claimStore store.ClaimStore,
	policyStore store.PolicyStore,
	customerStore store.CustomerStore,
	configManager *config.ConfigManager,
	eventService *EventService,
	logger *logger.Logger,
) *FraudDetectionService {
	return &FraudDetectionService{
		claimStore:    claimStore,
		policyStore:   policyStore,
		customerStore: customerStore,
		configManager: configManager,
		eventService:  eventService,
		logger:        logger,
	}
}

// FraudScoreV2 represents the result of fraud detection analysis.
type FraudScoreV2 struct {
	Score           float64                `json:"score"`           // 0-100, higher means more likely fraud
	RiskLevel       string                 `json:"risk_level"`      // low, medium, high, critical
	Factors         []FraudFactorV2        `json:"factors"`         // Individual risk factors
	Recommendations []string               `json:"recommendations"` // Recommended actions
	RequiresReview  bool                   `json:"requires_review"` // Whether manual review is needed
	Confidence      float64                `json:"confidence"`      // Confidence in the score (0-1)
	AnalysisDate    time.Time              `json:"analysis_date"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// FraudFactorV2 represents an individual risk factor in fraud detection.
type FraudFactorV2 struct {
	Factor      string  `json:"factor"`      // Name of the risk factor
	Weight      float64 `json:"weight"`      // Weight of this factor (0-1)
	Score       float64 `json:"score"`       // Individual score for this factor (0-100)
	Description string  `json:"description"` // Human-readable description
	Severity    string  `json:"severity"`    // low, medium, high, critical
}

// AnalyzeClaimForFraudV2 performs comprehensive fraud detection analysis on a claim.
func (s *FraudDetectionService) AnalyzeClaimForFraudV2(ctx context.Context, claimID uuid.UUID) (*FraudScoreV2, error) {
	// Get configuration
	config := s.configManager.GetConfig()
	if !config.FraudDetection.Enabled {
		return nil, fmt.Errorf("fraud detection is disabled")
	}
	fraudConfig := config.FraudDetection

	// Fetch claim details
	claim, err := s.claimStore.GetClaim(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch claim: %w", err)
	}

	// Fetch customer details instead of user
	customer, err := s.customerStore.GetByID(ctx, claim.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer: %w", err)
	}

	// Fetch related policy
	policy, err := s.policyStore.GetPolicy(ctx, claim.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policy: %w", err)
	}

	// Perform comprehensive fraud analysis
	score := &FraudScoreV2{
		AnalysisDate: time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// Analyze various fraud indicators using configuration
	factors := []FraudFactorV2{
		s.analyzeClaimTimingV2(claim, policy, &fraudConfig),
		s.analyzeClaimAmountV2(claim, policy, &fraudConfig),
		s.analyzeCustomerHistoryV2(claim, customer, &fraudConfig),
		s.analyzeIncidentPatternsV2(claim, policy, &fraudConfig),
		s.analyzeDocumentationV2(claim, &fraudConfig),
		s.analyzeGeographicRiskV2(claim, customer, &fraudConfig),
		s.analyzeBehavioralPatternsV2(claim, customer, &fraudConfig),
		s.analyzePolicyHistoryV2(claim, policy, &fraudConfig),
	}

	// Calculate weighted fraud score using configuration weights
	totalWeight := 0.0
	weightedScore := 0.0

	for _, factor := range factors {
		if factor.Weight > 0 {
			totalWeight += factor.Weight
			weightedScore += factor.Score * factor.Weight
		}
	}

	if totalWeight > 0 {
		score.Score = weightedScore / totalWeight
	}

	score.Factors = factors
	score.Confidence = s.calculateConfidenceV2(factors, &fraudConfig)
	score.RiskLevel = s.determineRiskLevelV2(score.Score, &fraudConfig)
	score.RequiresReview = s.requiresManualReviewV2(score.Score, factors, &fraudConfig)
	score.Recommendations = s.generateRecommendationsV2(score, factors, &fraudConfig)

	// Store fraud analysis results
	score.Metadata["claim_id"] = claimID.String()
	score.Metadata["policy_id"] = claim.PolicyID.String()
	score.Metadata["customer_id"] = claim.UserID.String()
	score.Metadata["analysis_version"] = fraudConfig.Version

	// Publish fraud analysis completed event
	if s.eventService != nil {
		factorNames := make([]string, len(factors))
		for i, factor := range factors {
			factorNames[i] = factor.Factor
		}

		fraudEvent := events.NewFraudAnalysisCompletedEvent(
			claimID,
			claim.UserID,
			score.Score,
			score.RiskLevel,
			score.RequiresReview,
			score.Confidence,
			factorNames,
			time.Now(),
		)

		if err := s.eventService.PublishEvent(ctx, fraudEvent); err != nil {
			s.logger.Error("Failed to publish fraud analysis completed event",
				zap.Error(err),
				zap.String("claim_id", claimID.String()))
			// Don't fail the analysis if event publishing fails
		}
	}

	return score, nil
}

// analyzeClaimTimingV2 analyzes timing-related fraud indicators using configuration.
func (s *FraudDetectionService) analyzeClaimTimingV2(claim *models.Claim, policy *models.Policy, config *config.FraudDetectionConfig) FraudFactorV2 {
	factor := FraudFactorV2{
		Factor: "claim_timing",
		Weight: config.FactorWeights["claim_timing"],
	}

	// Check if claim is filed very close to policy start
	daysSincePolicyStart := claim.IncidentDate.Sub(policy.EffectiveDate).Hours() / 24
	policyStartThreshold := config.TimingRules.PolicyStartThreshold.Hours() / 24

	if daysSincePolicyStart < policyStartThreshold {
		factor.Score = 80
		factor.Description = fmt.Sprintf("Claim filed within %.0f days of policy start", policyStartThreshold)
		factor.Severity = "high"
	} else if daysSincePolicyStart < policyStartThreshold*4 { // 4x threshold
		factor.Score = 40
		factor.Description = fmt.Sprintf("Claim filed within %.0f days of policy start", policyStartThreshold*4)
		factor.Severity = "medium"
	} else {
		factor.Score = 10
		factor.Description = "Claim filed after policy has been active for a reasonable period"
		factor.Severity = "low"
	}

	// Check reporting delay
	reportingDelay := claim.ReportedDate.Sub(claim.IncidentDate).Hours() / 24
	reportingDelayThreshold := config.TimingRules.ReportingDelayThreshold.Hours() / 24

	if reportingDelay > reportingDelayThreshold {
		factor.Score += 20
		factor.Description += fmt.Sprintf("; Significant delay in reporting (%.0f days)", reportingDelay)
		if factor.Severity == "low" {
			factor.Severity = "medium"
		}
	}

	// Apply weekend multiplier
	weekday := claim.IncidentDate.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		factor.Score *= config.TimingRules.WeekendMultiplier
		factor.Description += "; Incident occurred on weekend"
	}

	// Apply business hours multiplier
	hour := claim.IncidentDate.Hour()
	if hour >= 9 && hour <= 17 {
		factor.Score *= config.TimingRules.BusinessHoursMultiplier
		factor.Description += "; Incident occurred during business hours"
	}

	return factor
}

// analyzeClaimAmountV2 analyzes claim amount-related fraud indicators using configuration.
func (s *FraudDetectionService) analyzeClaimAmountV2(claim *models.Claim, policy *models.Policy, config *config.FraudDetectionConfig) FraudFactorV2 {
	factor := FraudFactorV2{
		Factor: "claim_amount",
		Weight: config.FactorWeights["claim_amount"],
	}

	// Check if claim amount is close to coverage limit
	coverageRatio := claim.ClaimAmount / policy.CoverageAmount
	if coverageRatio > config.AmountRules.CoverageRatioThreshold {
		factor.Score = 70
		factor.Description = "Claim amount is very close to coverage limit"
		factor.Severity = "high"
	} else if coverageRatio > config.AmountRules.CoverageRatioThreshold*0.8 {
		factor.Score = 40
		factor.Description = "Claim amount is high relative to coverage"
		factor.Severity = "medium"
	} else if coverageRatio < 0.1 {
		factor.Score = 5
		factor.Description = "Claim amount is low relative to coverage"
		factor.Severity = "low"
	} else {
		factor.Score = 15
		factor.Description = "Claim amount is within normal range"
		factor.Severity = "low"
	}

	// Check for round numbers (potential red flag)
	if claim.ClaimAmount > 1000 && int(claim.ClaimAmount)%1000 == 0 {
		factor.Score += config.AmountRules.RoundNumberPenalty
		factor.Description += "; Claim amount is a round number"
		if factor.Severity == "low" {
			factor.Severity = "medium"
		}
	}

	// Check against high-value thresholds
	if claim.ClaimAmount > config.AmountRules.VeryHighValueThreshold {
		factor.Score += 30
		factor.Description += "; Very high-value claim"
		if factor.Severity == "medium" {
			factor.Severity = "high"
		}
	} else if claim.ClaimAmount > config.AmountRules.HighValueThreshold {
		factor.Score += 15
		factor.Description += "; High-value claim"
		if factor.Severity == "low" {
			factor.Severity = "medium"
		}
	}

	return factor
}

// analyzeCustomerHistoryV2 analyzes customer's historical patterns for fraud indicators.
func (s *FraudDetectionService) analyzeCustomerHistoryV2(claim *models.Claim, customer *models.Customer, config *config.FraudDetectionConfig) FraudFactorV2 {
	factor := FraudFactorV2{
		Factor: "customer_history",
		Weight: config.FactorWeights["user_history"],
	}

	// Check customer account age
	accountAge := time.Since(customer.CreatedAt)
	newAccountThreshold := config.TimingRules.NewAccountThreshold

	if accountAge < newAccountThreshold {
		factor.Score = 60
		factor.Description = fmt.Sprintf("New customer account (less than %.0f days old)", newAccountThreshold.Hours()/24)
		factor.Severity = "high"
	} else if accountAge < newAccountThreshold*2 {
		factor.Score = 30
		factor.Description = "Relatively new customer account"
		factor.Severity = "medium"
	} else {
		factor.Score = 10
		factor.Description = "Established customer account"
		factor.Severity = "low"
	}

	// Check customer status
	if customer.Status != "active" {
		factor.Score += 30
		factor.Description += "; Customer account is not active"
		factor.Severity = "high"
	}

	// Check KYC status
	if !customer.IsKYCVerified() {
		factor.Score += 20
		factor.Description += "; Customer KYC not verified"
		if factor.Severity == "low" {
			factor.Severity = "medium"
		}
	}

	// Check AML status
	if !customer.IsAMLCleared() {
		factor.Score += 25
		factor.Description += "; Customer AML not cleared"
		if factor.Severity == "low" {
			factor.Severity = "medium"
		}
	}

	// Check risk profile
	if customer.IsHighRisk() {
		factor.Score += 15
		factor.Description += fmt.Sprintf("; Customer has %s risk profile", customer.RiskProfile)
		if factor.Severity == "low" {
			factor.Severity = "medium"
		}
	}

	return factor
}

// analyzeIncidentPatternsV2 analyzes patterns in the incident for fraud indicators.
func (s *FraudDetectionService) analyzeIncidentPatternsV2(claim *models.Claim, policy *models.Policy, config *config.FraudDetectionConfig) FraudFactorV2 {
	factor := FraudFactorV2{
		Factor: "incident_patterns",
		Weight: config.FactorWeights["incident_patterns"],
	}

	// Check if incident occurred on weekend (common fraud pattern)
	weekday := claim.IncidentDate.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		factor.Score = 30
		factor.Description = "Incident occurred on weekend"
		factor.Severity = "medium"
	} else {
		factor.Score = 10
		factor.Description = "Incident occurred on weekday"
		factor.Severity = "low"
	}

	// Check if incident occurred during business hours
	hour := claim.IncidentDate.Hour()
	if hour >= 9 && hour <= 17 {
		factor.Score += 5
		factor.Description += "; Incident occurred during business hours"
	} else {
		factor.Score += 15
		factor.Description += "; Incident occurred outside business hours"
		if factor.Severity == "low" {
			factor.Severity = "medium"
		}
	}

	return factor
}

// analyzeDocumentationV2 analyzes documentation quality for fraud indicators.
func (s *FraudDetectionService) analyzeDocumentationV2(claim *models.Claim, config *config.FraudDetectionConfig) FraudFactorV2 {
	factor := FraudFactorV2{
		Factor: "documentation",
		Weight: config.FactorWeights["documentation"],
	}

	// Check number of documents
	docCount := len(claim.Documents)
	minDocCount := config.DocumentRules.MinDocumentCount

	if docCount == 0 {
		factor.Score = 80
		factor.Description = "No supporting documents provided"
		factor.Severity = "high"
	} else if docCount < minDocCount {
		factor.Score = 40
		factor.Description = fmt.Sprintf("Limited supporting documentation (%d/%d)", docCount, minDocCount)
		factor.Severity = "medium"
	} else {
		factor.Score = 10
		factor.Description = "Adequate supporting documentation"
		factor.Severity = "low"
	}

	// Check document quality (simplified)
	for _, doc := range claim.Documents {
		if doc.FileSize < config.DocumentRules.MinFileSize {
			factor.Score += 10
			factor.Description += "; Some documents appear to be very small"
		}
		if doc.FileSize > config.DocumentRules.MaxFileSize {
			factor.Score += 5
			factor.Description += "; Some documents are very large"
		}
	}

	return factor
}

// analyzeGeographicRiskV2 analyzes geographic risk factors.
func (s *FraudDetectionService) analyzeGeographicRiskV2(claim *models.Claim, customer *models.Customer, config *config.FraudDetectionConfig) FraudFactorV2 {
	factor := FraudFactorV2{
		Factor: "geographic_risk",
		Weight: config.FactorWeights["geographic_risk"],
	}

	// Get customer's primary address
	primaryAddress := customer.GetPrimaryAddress()
	if primaryAddress == nil {
		factor.Score = 30
		factor.Description = "No address information available"
		factor.Severity = "medium"
		return factor
	}

	// Check against high-risk countries
	country := primaryAddress.Country
	if contains(config.GeographicRules.HighRiskCountries, country) {
		factor.Score = 60
		factor.Description = fmt.Sprintf("Customer located in high-risk country: %s", country)
		factor.Severity = "high"
	} else if contains(config.GeographicRules.HighRiskRegions, primaryAddress.State) {
		factor.Score = 40
		factor.Description = fmt.Sprintf("Customer located in high-risk region: %s", primaryAddress.State)
		factor.Severity = "medium"
	} else {
		factor.Score = 20
		factor.Description = "Customer located in standard risk area"
		factor.Severity = "low"
	}

	// Apply country-specific risk multiplier
	if multiplier, exists := config.GeographicRules.CountryRiskMultiplier[country]; exists {
		factor.Score *= multiplier
		factor.Description += fmt.Sprintf("; Applied country risk multiplier: %.2f", multiplier)
	}

	return factor
}

// analyzeBehavioralPatternsV2 analyzes behavioral patterns for fraud indicators.
func (s *FraudDetectionService) analyzeBehavioralPatternsV2(claim *models.Claim, customer *models.Customer, config *config.FraudDetectionConfig) FraudFactorV2 {
	factor := FraudFactorV2{
		Factor: "behavioral_patterns",
		Weight: config.FactorWeights["behavioral_patterns"],
	}

	// Check claim description length and quality
	descLength := len(claim.Description)
	minLength := config.BehavioralRules.MinDescriptionLength
	maxLength := config.BehavioralRules.MaxDescriptionLength

	if descLength < minLength {
		factor.Score = 50
		factor.Description = fmt.Sprintf("Very brief claim description (%d chars)", descLength)
		factor.Severity = "medium"
	} else if descLength > maxLength {
		factor.Score = 30
		factor.Description = fmt.Sprintf("Extremely detailed claim description (%d chars)", descLength)
		factor.Severity = "medium"
	} else {
		factor.Score = 10
		factor.Description = "Appropriate claim description length"
		factor.Severity = "low"
	}

	// Check customer tier (higher tier customers are generally more trustworthy)
	tierLevel := customer.GetCustomerTierLevel()
	if tierLevel >= 3 { // Gold or Platinum
		factor.Score *= 0.8
		factor.Description += fmt.Sprintf("; Customer tier discount applied (%s)", customer.CustomerTier)
	} else if tierLevel == 0 { // No tier assigned
		factor.Score *= 1.2
		factor.Description += "; No customer tier assigned"
	}

	return factor
}

// analyzePolicyHistoryV2 analyzes policy history for fraud indicators.
func (s *FraudDetectionService) analyzePolicyHistoryV2(claim *models.Claim, policy *models.Policy, config *config.FraudDetectionConfig) FraudFactorV2 {
	factor := FraudFactorV2{
		Factor: "policy_history",
		Weight: config.FactorWeights["policy_history"],
	}

	// Check policy age
	policyAge := time.Since(policy.CreatedAt).Hours() / 24 / 365 // years
	if policyAge < 0.25 {                                        // Less than 3 months
		factor.Score = 60
		factor.Description = "Very new policy"
		factor.Severity = "high"
	} else if policyAge < 1 {
		factor.Score = 30
		factor.Description = "Relatively new policy"
		factor.Severity = "medium"
	} else {
		factor.Score = 10
		factor.Description = "Established policy"
		factor.Severity = "low"
	}

	// Check if policy is close to expiration
	daysToExpiration := policy.ExpirationDate.Sub(claim.IncidentDate).Hours() / 24
	if daysToExpiration < 30 {
		factor.Score += 20
		factor.Description += "; Incident occurred near policy expiration"
		if factor.Severity == "low" {
			factor.Severity = "medium"
		}
	}

	return factor
}

// calculateConfidenceV2 calculates the confidence level in the fraud score.
func (s *FraudDetectionService) calculateConfidenceV2(factors []FraudFactorV2, config *config.FraudDetectionConfig) float64 {
	// Confidence is based on the number of factors and their weights
	totalWeight := 0.0
	activeFactors := 0

	for _, factor := range factors {
		if factor.Weight > 0 {
			totalWeight += factor.Weight
			activeFactors++
		}
	}

	// Base confidence on total weight and number of factors
	weightConfidence := totalWeight
	if weightConfidence > 1.0 {
		weightConfidence = 1.0
	}

	factorConfidence := float64(activeFactors) / 8.0 // Assuming 8 total factors
	if factorConfidence > 1.0 {
		factorConfidence = 1.0
	}

	return (weightConfidence + factorConfidence) / 2.0
}

// determineRiskLevelV2 determines the risk level based on the fraud score using configuration.
func (s *FraudDetectionService) determineRiskLevelV2(score float64, config *config.FraudDetectionConfig) string {
	thresholds := config.RiskThresholds

	switch {
	case score >= thresholds.Critical:
		return "critical"
	case score >= thresholds.High:
		return "high"
	case score >= thresholds.Medium:
		return "medium"
	default:
		return "low"
	}
}

// requiresManualReviewV2 determines if manual review is required using configuration.
func (s *FraudDetectionService) requiresManualReviewV2(score float64, factors []FraudFactorV2, config *config.FraudDetectionConfig) bool {
	// Always require review for high scores
	if score >= config.AutoReviewThresholds.ScoreThreshold {
		return true
	}

	// Require review if any factor is critical
	criticalCount := 0
	for _, factor := range factors {
		if factor.Severity == "critical" {
			criticalCount++
		}
	}

	if criticalCount >= config.AutoReviewThresholds.CriticalFactorCount {
		return true
	}

	// Require review if multiple high-severity factors
	highSeverityCount := 0
	for _, factor := range factors {
		if factor.Severity == "high" {
			highSeverityCount++
		}
	}

	return highSeverityCount >= config.AutoReviewThresholds.HighSeverityCount
}

// generateRecommendationsV2 generates recommendations based on the fraud analysis using configuration.
func (s *FraudDetectionService) generateRecommendationsV2(score *FraudScoreV2, factors []FraudFactorV2, config *config.FraudDetectionConfig) []string {
	recommendations := []string{}

	switch score.RiskLevel {
	case "critical":
		recommendations = append(recommendations, "Immediate manual review required")
		recommendations = append(recommendations, "Consider suspending claim processing")
		recommendations = append(recommendations, "Request additional documentation")
		recommendations = append(recommendations, "Consider involving fraud investigation team")

	case "high":
		recommendations = append(recommendations, "Manual review recommended")
		recommendations = append(recommendations, "Request additional supporting documentation")
		recommendations = append(recommendations, "Verify incident details with third parties")

	case "medium":
		recommendations = append(recommendations, "Enhanced verification recommended")
		recommendations = append(recommendations, "Request additional documentation for high-risk factors")

	case "low":
		recommendations = append(recommendations, "Standard processing can proceed")
		recommendations = append(recommendations, "Monitor for any additional risk factors")
	}

	// Add specific recommendations based on high-risk factors
	for _, factor := range factors {
		if factor.Severity == "high" || factor.Severity == "critical" {
			switch factor.Factor {
			case "claim_timing":
				recommendations = append(recommendations, "Verify policy start date and incident timeline")
			case "claim_amount":
				recommendations = append(recommendations, "Obtain independent damage assessment")
			case "customer_history":
				recommendations = append(recommendations, "Verify customer identity and account history")
			case "documentation":
				recommendations = append(recommendations, "Request comprehensive supporting documentation")
			}
		}
	}

	return recommendations
}

// GetFraudScoreV2 retrieves a previously calculated fraud score.
func (s *FraudDetectionService) GetFraudScoreV2(ctx context.Context, claimID uuid.UUID) (*FraudScoreV2, error) {
	// In a real implementation, this would retrieve from a fraud analysis store
	// For now, we'll re-analyze
	return s.AnalyzeClaimForFraudV2(ctx, claimID)
}

// UpdateFraudScoreV2 updates a fraud score with new information.
func (s *FraudDetectionService) UpdateFraudScoreV2(ctx context.Context, claimID uuid.UUID, newScore *FraudScoreV2) error {
	// In a real implementation, this would update the fraud analysis store
	// For now, we'll just validate the input
	if newScore == nil {
		return fmt.Errorf("fraud score cannot be nil")
	}

	if newScore.Score < 0 || newScore.Score > 100 {
		return fmt.Errorf("fraud score must be between 0 and 100")
	}

	return nil
}

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
