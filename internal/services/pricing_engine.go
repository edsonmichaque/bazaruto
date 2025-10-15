package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/models"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// PricingEngineService handles comprehensive pricing calculations with dynamic rate adjustments.
type PricingEngineService struct {
	productStore store.ProductStore
	policyStore  store.PolicyStore
	claimStore   store.ClaimStore
	userStore    store.UserStore
}

// NewPricingEngineService creates a new PricingEngineService instance.
func NewPricingEngineService(
	productStore store.ProductStore,
	policyStore store.PolicyStore,
	claimStore store.ClaimStore,
	userStore store.UserStore,
) *PricingEngineService {
	return &PricingEngineService{
		productStore: productStore,
		policyStore:  policyStore,
		claimStore:   claimStore,
		userStore:    userStore,
	}
}

// PricingResult represents the result of a pricing calculation.
type PricingResult struct {
	BasePremium     float64                `json:"base_premium"`
	AdjustedPremium float64                `json:"adjusted_premium"`
	FinalPremium    float64                `json:"final_premium"`
	Currency        string                 `json:"currency"`
	Breakdown       PricingBreakdown       `json:"breakdown"`
	Factors         []PricingFactor        `json:"factors"`
	ValidUntil      time.Time              `json:"valid_until"`
	QuoteID         *uuid.UUID             `json:"quote_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PricingBreakdown provides detailed breakdown of pricing components.
type PricingBreakdown struct {
	BaseRate            float64 `json:"base_rate"`
	CoverageAdjustment  float64 `json:"coverage_adjustment"`
	RiskAdjustment      float64 `json:"risk_adjustment"`
	DiscountAdjustment  float64 `json:"discount_adjustment"`
	TaxAdjustment       float64 `json:"tax_adjustment"`
	FrequencyAdjustment float64 `json:"frequency_adjustment"`
	MarketAdjustment    float64 `json:"market_adjustment"`
	TotalAdjustment     float64 `json:"total_adjustment"`
}

// PricingFactor represents an individual factor affecting pricing.
type PricingFactor struct {
	Factor      string  `json:"factor"`
	Type        string  `json:"type"` // rate, discount, surcharge, tax
	Value       float64 `json:"value"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // positive, negative, neutral
}

// CalculatePremium calculates comprehensive premium pricing for a policy.
func (s *PricingEngineService) CalculatePremium(ctx context.Context, request *PricingRequest) (*PricingResult, error) {
	// Validate pricing request
	if err := s.validatePricingRequest(request); err != nil {
		return nil, fmt.Errorf("invalid pricing request: %w", err)
	}

	// Fetch product details
	product, err := s.productStore.GetProduct(ctx, request.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}

	// Fetch user details for risk assessment
	user, err := s.userStore.FindByID(ctx, request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Initialize pricing result
	result := &PricingResult{
		Currency:   request.Currency,
		ValidUntil: time.Now().Add(24 * time.Hour), // Quote valid for 24 hours
		Metadata:   make(map[string]interface{}),
	}

	// Calculate base premium
	basePremium, err := s.calculateBasePremium(product, request)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate base premium: %w", err)
	}

	result.BasePremium = basePremium
	result.AdjustedPremium = basePremium

	// Apply pricing factors
	factors := []PricingFactor{
		s.calculateCoverageFactor(request),
		s.calculateRiskFactor(ctx, user, request),
		s.calculateDiscountFactor(request),
		s.calculateTaxFactor(request),
		s.calculateFrequencyFactor(request),
		s.calculateMarketFactor(ctx, product, request),
		s.calculateLoyaltyFactor(ctx, user, request),
		s.calculateSeasonalFactor(request),
	}

	// Apply all factors to calculate final premium
	breakdown := PricingBreakdown{
		BaseRate: basePremium,
	}

	for _, factor := range factors {
		switch factor.Type {
		case "rate":
			breakdown.RiskAdjustment += factor.Value
		case "discount":
			breakdown.DiscountAdjustment += factor.Value
		case "surcharge":
			breakdown.RiskAdjustment += factor.Value
		case "tax":
			breakdown.TaxAdjustment += factor.Value
		case "frequency":
			breakdown.FrequencyAdjustment += factor.Value
		case "market":
			breakdown.MarketAdjustment += factor.Value
		}
	}

	// Calculate total adjustment
	breakdown.TotalAdjustment = breakdown.CoverageAdjustment +
		breakdown.RiskAdjustment +
		breakdown.DiscountAdjustment +
		breakdown.TaxAdjustment +
		breakdown.FrequencyAdjustment +
		breakdown.MarketAdjustment

	// Calculate final premium
	result.AdjustedPremium = basePremium + breakdown.TotalAdjustment
	result.FinalPremium = math.Max(result.AdjustedPremium, 0) // Ensure non-negative
	result.Breakdown = breakdown
	result.Factors = factors

	// Store metadata
	result.Metadata["product_id"] = request.ProductID.String()
	result.Metadata["user_id"] = request.UserID.String()
	result.Metadata["coverage_amount"] = request.CoverageAmount
	result.Metadata["calculation_version"] = "2.0"

	return result, nil
}

// PricingRequest represents a request for premium calculation.
type PricingRequest struct {
	ProductID        uuid.UUID              `json:"product_id"`
	UserID           uuid.UUID              `json:"user_id"`
	CoverageAmount   float64                `json:"coverage_amount"`
	Currency         string                 `json:"currency"`
	PaymentFrequency string                 `json:"payment_frequency"`
	EffectiveDate    time.Time              `json:"effective_date"`
	ExpirationDate   time.Time              `json:"expiration_date"`
	RiskFactors      map[string]interface{} `json:"risk_factors"`
	Discounts        []string               `json:"discounts"`
	Options          map[string]interface{} `json:"options"`
}

// validatePricingRequest validates the pricing request.
func (s *PricingEngineService) validatePricingRequest(request *PricingRequest) error {
	if request.ProductID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}
	if request.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
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

// calculateBasePremium calculates the base premium for a product.
func (s *PricingEngineService) calculateBasePremium(product *models.Product, request *PricingRequest) (float64, error) {
	// Base rate per $1000 of coverage
	baseRate := 0.0

	// Product-specific base rates (in real implementation, these would come from product configuration)
	switch product.Category {
	case "auto":
		baseRate = 15.0 // $15 per $1000 coverage
	case "home":
		baseRate = 8.0 // $8 per $1000 coverage
	case "life":
		baseRate = 5.0 // $5 per $1000 coverage
	case "health":
		baseRate = 25.0 // $25 per $1000 coverage
	case "business":
		baseRate = 20.0 // $20 per $1000 coverage
	default:
		baseRate = 10.0 // Default rate
	}

	// Calculate base premium
	coverageInThousands := request.CoverageAmount / 1000
	basePremium := baseRate * coverageInThousands

	// Apply policy duration factor
	policyDuration := request.ExpirationDate.Sub(request.EffectiveDate).Hours() / 24 / 365 // years
	if policyDuration < 1 {
		// Pro-rate for policies less than 1 year
		basePremium *= policyDuration
	}

	return basePremium, nil
}

// calculateCoverageFactor calculates the coverage amount adjustment factor.
func (s *PricingEngineService) calculateCoverageFactor(request *PricingRequest) PricingFactor {
	factor := PricingFactor{
		Factor: "coverage_amount",
		Type:   "rate",
	}

	// Higher coverage amounts may have different rate structures
	if request.CoverageAmount > 1000000 { // Over $1M
		factor.Value = request.CoverageAmount * 0.001 // 0.1% surcharge
		factor.Description = "High-value coverage surcharge"
		factor.Impact = "positive"
	} else if request.CoverageAmount > 500000 { // Over $500K
		factor.Value = request.CoverageAmount * 0.0005 // 0.05% surcharge
		factor.Description = "Moderate-value coverage adjustment"
		factor.Impact = "positive"
	} else {
		factor.Value = 0
		factor.Description = "Standard coverage amount"
		factor.Impact = "neutral"
	}

	return factor
}

// calculateRiskFactor calculates risk-based pricing adjustments.
func (s *PricingEngineService) calculateRiskFactor(ctx context.Context, user *models.User, request *PricingRequest) PricingFactor {
	factor := PricingFactor{
		Factor: "risk_assessment",
		Type:   "rate",
	}

	// Account age factor
	accountAge := time.Since(user.CreatedAt).Hours() / 24 / 365 // years
	if accountAge < 0.5 {                                       // Less than 6 months
		factor.Value = request.CoverageAmount * 0.02 // 2% surcharge
		factor.Description = "New customer risk surcharge"
		factor.Impact = "positive"
	} else if accountAge < 2 { // Less than 2 years
		factor.Value = request.CoverageAmount * 0.01 // 1% surcharge
		factor.Description = "Moderate customer history"
		factor.Impact = "positive"
	} else {
		factor.Value = request.CoverageAmount * -0.005 // 0.5% discount
		factor.Description = "Established customer discount"
		factor.Impact = "negative"
	}

	// Apply custom risk factors from request
	if riskFactors, ok := request.RiskFactors["custom_factors"]; ok {
		if customFactors, ok := riskFactors.(map[string]interface{}); ok {
			for riskType, riskValue := range customFactors {
				if value, ok := riskValue.(float64); ok {
					factor.Value += value
					factor.Description += fmt.Sprintf("; %s: %.2f", riskType, value)
				}
			}
		}
	}

	return factor
}

// calculateDiscountFactor calculates discount adjustments.
func (s *PricingEngineService) calculateDiscountFactor(request *PricingRequest) PricingFactor {
	factor := PricingFactor{
		Factor: "discounts",
		Type:   "discount",
	}

	// Apply available discounts
	for _, discount := range request.Discounts {
		switch discount {
		case "multi_policy":
			factor.Value += request.CoverageAmount * -0.10 // 10% discount
			factor.Description += "Multi-policy discount; "
		case "loyalty":
			factor.Value += request.CoverageAmount * -0.05 // 5% discount
			factor.Description += "Loyalty discount; "
		case "early_payment":
			factor.Value += request.CoverageAmount * -0.03 // 3% discount
			factor.Description += "Early payment discount; "
		case "safe_driver":
			factor.Value += request.CoverageAmount * -0.08 // 8% discount
			factor.Description += "Safe driver discount; "
		case "security_system":
			factor.Value += request.CoverageAmount * -0.06 // 6% discount
			factor.Description += "Security system discount; "
		}
	}

	if factor.Description == "" {
		factor.Description = "No applicable discounts"
		factor.Impact = "neutral"
	} else {
		factor.Impact = "negative" // Discounts reduce premium
	}

	return factor
}

// calculateTaxFactor calculates tax adjustments.
func (s *PricingEngineService) calculateTaxFactor(request *PricingRequest) PricingFactor {
	factor := PricingFactor{
		Factor: "taxes",
		Type:   "tax",
	}

	// Tax rates vary by jurisdiction and product type
	// In a real implementation, this would use actual tax tables
	taxRate := 0.08 // 8% default tax rate
	factor.Value = request.CoverageAmount * taxRate
	factor.Description = fmt.Sprintf("Insurance tax (%.1f%%)", taxRate*100)
	factor.Impact = "positive"

	return factor
}

// calculateFrequencyFactor calculates payment frequency adjustments.
func (s *PricingEngineService) calculateFrequencyFactor(request *PricingRequest) PricingFactor {
	factor := PricingFactor{
		Factor: "payment_frequency",
		Type:   "frequency",
	}

	switch request.PaymentFrequency {
	case "annually":
		factor.Value = request.CoverageAmount * -0.05 // 5% discount
		factor.Description = "Annual payment discount"
		factor.Impact = "negative"
	case "quarterly":
		factor.Value = request.CoverageAmount * 0.02 // 2% surcharge
		factor.Description = "Quarterly payment surcharge"
		factor.Impact = "positive"
	case "monthly":
		factor.Value = request.CoverageAmount * 0.05 // 5% surcharge
		factor.Description = "Monthly payment surcharge"
		factor.Impact = "positive"
	default:
		factor.Value = 0
		factor.Description = "Standard payment frequency"
		factor.Impact = "neutral"
	}

	return factor
}

// calculateMarketFactor calculates market-based pricing adjustments.
func (s *PricingEngineService) calculateMarketFactor(ctx context.Context, product *models.Product, request *PricingRequest) PricingFactor {
	factor := PricingFactor{
		Factor: "market_conditions",
		Type:   "market",
	}

	// Market adjustments based on current conditions
	// In a real implementation, this would use market data and economic indicators
	marketAdjustment := 0.03 // 3% market adjustment
	factor.Value = request.CoverageAmount * marketAdjustment
	factor.Description = "Current market conditions adjustment"
	factor.Impact = "positive"

	return factor
}

// calculateLoyaltyFactor calculates loyalty-based pricing adjustments.
func (s *PricingEngineService) calculateLoyaltyFactor(ctx context.Context, user *models.User, request *PricingRequest) PricingFactor {
	factor := PricingFactor{
		Factor: "loyalty",
		Type:   "discount",
	}

	// Check user's policy history for loyalty discounts
	// In a real implementation, this would query actual policy history
	accountAge := time.Since(user.CreatedAt).Hours() / 24 / 365 // years

	if accountAge > 5 {
		factor.Value = request.CoverageAmount * -0.08 // 8% loyalty discount
		factor.Description = "Long-term customer loyalty discount"
		factor.Impact = "negative"
	} else if accountAge > 2 {
		factor.Value = request.CoverageAmount * -0.03 // 3% loyalty discount
		factor.Description = "Customer loyalty discount"
		factor.Impact = "negative"
	} else {
		factor.Value = 0
		factor.Description = "No loyalty discount applicable"
		factor.Impact = "neutral"
	}

	return factor
}

// calculateSeasonalFactor calculates seasonal pricing adjustments.
func (s *PricingEngineService) calculateSeasonalFactor(request *PricingRequest) PricingFactor {
	factor := PricingFactor{
		Factor: "seasonal",
		Type:   "rate",
	}

	// Seasonal adjustments based on effective date
	month := request.EffectiveDate.Month()

	switch month {
	case 12, 1, 2: // Winter months
		factor.Value = request.CoverageAmount * 0.02 // 2% winter surcharge
		factor.Description = "Winter season adjustment"
		factor.Impact = "positive"
	case 6, 7, 8: // Summer months
		factor.Value = request.CoverageAmount * 0.01 // 1% summer surcharge
		factor.Description = "Summer season adjustment"
		factor.Impact = "positive"
	default: // Spring/Fall
		factor.Value = 0
		factor.Description = "Standard seasonal rate"
		factor.Impact = "neutral"
	}

	return factor
}

// UpdatePricingFactors updates pricing factors based on new data or market conditions.
func (s *PricingEngineService) UpdatePricingFactors(ctx context.Context, productID uuid.UUID, factors map[string]interface{}) error {
	// In a real implementation, this would update pricing factors in the database
	// For now, we'll validate the input
	if productID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}

	if factors == nil {
		return fmt.Errorf("pricing factors cannot be nil")
	}

	// Validate factor values
	for factorName, factorValue := range factors {
		if factorValue == nil {
			return fmt.Errorf("factor %s cannot be nil", factorName)
		}
	}

	return nil
}

// GetPricingHistory retrieves pricing history for analysis.
func (s *PricingEngineService) GetPricingHistory(ctx context.Context, productID uuid.UUID, startDate, endDate time.Time) ([]PricingResult, error) {
	// In a real implementation, this would query pricing history from the database
	// For now, we'll return an empty slice
	return []PricingResult{}, nil
}

// ComparePricing compares pricing across different scenarios.
func (s *PricingEngineService) ComparePricing(ctx context.Context, baseRequest *PricingRequest, scenarios []PricingRequest) ([]PricingComparison, error) {
	// Calculate base pricing
	baseResult, err := s.CalculatePremium(ctx, baseRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate base premium: %w", err)
	}

	// Calculate scenario pricing
	comparisons := []PricingComparison{
		{
			Scenario: "Base",
			Result:   baseResult,
		},
	}

	for i, scenario := range scenarios {
		result, err := s.CalculatePremium(ctx, &scenario)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate scenario %d premium: %w", i, err)
		}

		comparison := PricingComparison{
			Scenario: fmt.Sprintf("Scenario %d", i+1),
			Result:   result,
		}

		// Calculate difference from base
		comparison.Difference = result.FinalPremium - baseResult.FinalPremium
		comparison.PercentageChange = (comparison.Difference / baseResult.FinalPremium) * 100

		comparisons = append(comparisons, comparison)
	}

	return comparisons, nil
}

// PricingComparison represents a comparison between different pricing scenarios.
type PricingComparison struct {
	Scenario         string         `json:"scenario"`
	Result           *PricingResult `json:"result"`
	Difference       float64        `json:"difference"`
	PercentageChange float64        `json:"percentage_change"`
}

// ValidatePricingResult validates the integrity of a pricing result.
func (s *PricingEngineService) ValidatePricingResult(result *PricingResult) error {
	if result == nil {
		return fmt.Errorf("pricing result cannot be nil")
	}

	if result.BasePremium < 0 {
		return fmt.Errorf("base premium cannot be negative")
	}

	if result.FinalPremium < 0 {
		return fmt.Errorf("final premium cannot be negative")
	}

	if result.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if result.ValidUntil.IsZero() {
		return fmt.Errorf("valid until date is required")
	}

	if result.ValidUntil.Before(time.Now()) {
		return fmt.Errorf("valid until date must be in the future")
	}

	return nil
}
