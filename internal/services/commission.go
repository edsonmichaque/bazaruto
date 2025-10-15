package services

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/google/uuid"
)

// CommissionService handles commission calculation and partner payout logic.
type CommissionService struct {
	partnerStore    store.PartnerStore
	policyStore     store.PolicyStore
	paymentStore    store.PaymentStore
	commissionStore interface{} // Generic store interface
}

// NewCommissionService creates a new CommissionService instance.
func NewCommissionService(
	partnerStore store.PartnerStore,
	policyStore store.PolicyStore,
	paymentStore store.PaymentStore,
	commissionStore interface{},
) *CommissionService {
	return &CommissionService{
		partnerStore:    partnerStore,
		policyStore:     policyStore,
		paymentStore:    paymentStore,
		commissionStore: commissionStore,
	}
}

// CommissionCalculation represents the result of a commission calculation.
type CommissionCalculation struct {
	PolicyID         uuid.UUID              `json:"policy_id"`
	PartnerID        uuid.UUID              `json:"partner_id"`
	CommissionType   string                 `json:"commission_type"`   // initial, renewal, adjustment
	BaseAmount       float64                `json:"base_amount"`       // Base premium amount
	CommissionRate   float64                `json:"commission_rate"`   // Commission rate (percentage)
	CommissionAmount float64                `json:"commission_amount"` // Calculated commission amount
	Currency         string                 `json:"currency"`          // Currency code
	Status           string                 `json:"status"`            // calculated, paid, pending, cancelled
	PaymentDate      *time.Time             `json:"payment_date"`      // When commission was paid
	DueDate          time.Time              `json:"due_date"`          // When commission is due
	CalculationDate  time.Time              `json:"calculation_date"`  // When commission was calculated
	Metadata         map[string]interface{} `json:"metadata"`
}

// CommissionRule represents a commission rule for a partner and product combination.
type CommissionRule struct {
	PartnerID      uuid.UUID              `json:"partner_id"`
	ProductID      uuid.UUID              `json:"product_id"`
	CommissionType string                 `json:"commission_type"` // initial, renewal, adjustment
	Rate           float64                `json:"rate"`            // Commission rate (percentage)
	MinAmount      float64                `json:"min_amount"`      // Minimum commission amount
	MaxAmount      float64                `json:"max_amount"`      // Maximum commission amount
	EffectiveDate  time.Time              `json:"effective_date"`  // When rule becomes effective
	ExpirationDate *time.Time             `json:"expiration_date"` // When rule expires (nil for no expiration)
	Conditions     map[string]interface{} `json:"conditions"`      // Additional conditions
	Metadata       map[string]interface{} `json:"metadata"`
}

// CalculateCommission calculates commission for a policy and partner.
func (s *CommissionService) CalculateCommission(ctx context.Context, policyID uuid.UUID, partnerID uuid.UUID, commissionType string) (*CommissionCalculation, error) {
	// Fetch policy details
	policy, err := s.policyStore.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policy: %w", err)
	}

	// Fetch partner details
	partner, err := s.partnerStore.GetPartner(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch partner: %w", err)
	}

	// Get commission rule for this partner and product
	rule, err := s.getCommissionRule(ctx, partnerID, policy.ProductID, commissionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get commission rule: %w", err)
	}

	// Calculate commission
	calculation := &CommissionCalculation{
		PolicyID:        policyID,
		PartnerID:       partnerID,
		CommissionType:  commissionType,
		BaseAmount:      policy.Premium,
		CommissionRate:  rule.Rate,
		Currency:        policy.Currency,
		Status:          "calculated",
		CalculationDate: time.Now(),
		DueDate:         s.calculateDueDate(commissionType),
		Metadata:        make(map[string]interface{}),
	}

	// Calculate commission amount
	commissionAmount := policy.Premium * (rule.Rate / 100.0)

	// Apply minimum and maximum limits
	if rule.MinAmount > 0 && commissionAmount < rule.MinAmount {
		commissionAmount = rule.MinAmount
	}
	if rule.MaxAmount > 0 && commissionAmount > rule.MaxAmount {
		commissionAmount = rule.MaxAmount
	}

	calculation.CommissionAmount = commissionAmount

	// Store metadata
	calculation.Metadata["partner_name"] = partner.Name
	calculation.Metadata["product_id"] = policy.ProductID.String()
	calculation.Metadata["rule_id"] = rule.PartnerID.String() // Using partner ID as rule identifier
	calculation.Metadata["calculation_version"] = "1.0"

	return calculation, nil
}

// getCommissionRule retrieves the applicable commission rule for a partner and product.
func (s *CommissionService) getCommissionRule(ctx context.Context, partnerID uuid.UUID, productID uuid.UUID, commissionType string) (*CommissionRule, error) {
	// In a real implementation, this would query commission rules from the database
	// For now, we'll return a default rule based on commission type

	rule := &CommissionRule{
		PartnerID:      partnerID,
		ProductID:      productID,
		CommissionType: commissionType,
		EffectiveDate:  time.Now(),
		Conditions:     make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
	}

	// Set default rates based on commission type
	switch commissionType {
	case "initial":
		rule.Rate = 15.0 // 15% for initial policies
		rule.MinAmount = 50.0
		rule.MaxAmount = 5000.0
	case "renewal":
		rule.Rate = 10.0 // 10% for renewals
		rule.MinAmount = 25.0
		rule.MaxAmount = 3000.0
	case "adjustment":
		rule.Rate = 5.0 // 5% for adjustments
		rule.MinAmount = 10.0
		rule.MaxAmount = 1000.0
	default:
		rule.Rate = 10.0 // Default rate
		rule.MinAmount = 25.0
		rule.MaxAmount = 2000.0
	}

	return rule, nil
}

// calculateDueDate calculates when the commission payment is due.
func (s *CommissionService) calculateDueDate(commissionType string) time.Time {
	// Different commission types have different payment schedules
	switch commissionType {
	case "initial":
		return time.Now().Add(30 * 24 * time.Hour) // 30 days for initial commissions
	case "renewal":
		return time.Now().Add(45 * 24 * time.Hour) // 45 days for renewal commissions
	case "adjustment":
		return time.Now().Add(15 * 24 * time.Hour) // 15 days for adjustments
	default:
		return time.Now().Add(30 * 24 * time.Hour) // Default 30 days
	}
}

// ProcessCommissionPayment processes payment of a commission.
func (s *CommissionService) ProcessCommissionPayment(ctx context.Context, calculationID uuid.UUID, paymentMethod string) (*CommissionPayment, error) {
	// Fetch commission calculation
	// In a real implementation, this would fetch from the commission store
	calculation := &CommissionCalculation{
		PolicyID:         uuid.New(),
		PartnerID:        uuid.New(),
		Status:           "calculated",
		CommissionAmount: 1000.0, // Simulated amount
		Currency:         "USD",
	}

	// Validate commission can be paid
	if calculation.Status != "calculated" {
		return nil, fmt.Errorf("commission is not in calculated status")
	}

	// Create commission payment record
	payment := &CommissionPayment{
		CommissionID:  calculationID,
		PartnerID:     calculation.PartnerID,
		Amount:        calculation.CommissionAmount,
		Currency:      calculation.Currency,
		PaymentMethod: paymentMethod,
		Status:        "pending",
		PaymentDate:   time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// Process payment (simplified - in real implementation, this would integrate with payment gateway)
	payment.Status = "completed"
	payment.TransactionID = fmt.Sprintf("comm_%d_%s", time.Now().Unix(), calculationID.String()[:8])

	// Update commission status
	calculation.Status = "paid"
	now := time.Now()
	calculation.PaymentDate = &now

	// Store payment and update commission
	// In a real implementation, this would store in the commission store
	// For now, we'll simulate successful storage

	return payment, nil
}

// CommissionPayment represents a commission payment record.
type CommissionPayment struct {
	ID            uuid.UUID              `json:"id"`
	CommissionID  uuid.UUID              `json:"commission_id"`
	PartnerID     uuid.UUID              `json:"partner_id"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	PaymentMethod string                 `json:"payment_method"`
	Status        string                 `json:"status"` // pending, completed, failed
	TransactionID string                 `json:"transaction_id"`
	PaymentDate   time.Time              `json:"payment_date"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// GetCommissionHistory retrieves commission history for a partner.
func (s *CommissionService) GetCommissionHistory(ctx context.Context, partnerID uuid.UUID, startDate, endDate time.Time) ([]CommissionCalculation, error) {
	// In a real implementation, this would query commission history from the database
	// For now, we'll return an empty slice
	return []CommissionCalculation{}, nil
}

// GetPendingCommissions retrieves all pending commissions.
func (s *CommissionService) GetPendingCommissions(ctx context.Context) ([]CommissionCalculation, error) {
	// In a real implementation, this would query pending commissions from the database
	// For now, we'll return an empty slice
	return []CommissionCalculation{}, nil
}

// CalculatePartnerEarnings calculates total earnings for a partner in a given period.
func (s *CommissionService) CalculatePartnerEarnings(ctx context.Context, partnerID uuid.UUID, startDate, endDate time.Time) (*PartnerEarnings, error) {
	// In a real implementation, this would calculate actual earnings from commission data
	// For now, we'll return a simulated result

	earnings := &PartnerEarnings{
		PartnerID:     partnerID,
		StartDate:     startDate,
		EndDate:       endDate,
		TotalEarnings: 0.0,
		Currency:      "USD",
		Breakdown:     make(map[string]float64),
		Metadata:      make(map[string]interface{}),
	}

	// Simulate earnings breakdown by commission type
	earnings.Breakdown["initial"] = 2500.0
	earnings.Breakdown["renewal"] = 1200.0
	earnings.Breakdown["adjustment"] = 300.0
	earnings.TotalEarnings = 4000.0

	earnings.Metadata["calculation_date"] = time.Now()
	earnings.Metadata["calculation_version"] = "1.0"

	return earnings, nil
}

// PartnerEarnings represents total earnings for a partner in a given period.
type PartnerEarnings struct {
	PartnerID     uuid.UUID              `json:"partner_id"`
	StartDate     time.Time              `json:"start_date"`
	EndDate       time.Time              `json:"end_date"`
	TotalEarnings float64                `json:"total_earnings"`
	Currency      string                 `json:"currency"`
	Breakdown     map[string]float64     `json:"breakdown"` // Breakdown by commission type
	Metadata      map[string]interface{} `json:"metadata"`
}

// UpdateCommissionRule updates a commission rule for a partner and product.
func (s *CommissionService) UpdateCommissionRule(ctx context.Context, rule *CommissionRule) error {
	// Validate commission rule
	if err := s.validateCommissionRule(rule); err != nil {
		return fmt.Errorf("invalid commission rule: %w", err)
	}

	// In a real implementation, this would update the rule in the database
	// For now, we'll just validate the input
	return nil
}

// validateCommissionRule validates a commission rule.
func (s *CommissionService) validateCommissionRule(rule *CommissionRule) error {
	if rule == nil {
		return fmt.Errorf("commission rule cannot be nil")
	}

	if rule.PartnerID == uuid.Nil {
		return fmt.Errorf("partner ID is required")
	}

	if rule.ProductID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}

	if rule.CommissionType == "" {
		return fmt.Errorf("commission type is required")
	}

	if rule.Rate < 0 || rule.Rate > 100 {
		return fmt.Errorf("commission rate must be between 0 and 100")
	}

	if rule.MinAmount < 0 {
		return fmt.Errorf("minimum amount cannot be negative")
	}

	if rule.MaxAmount < 0 {
		return fmt.Errorf("maximum amount cannot be negative")
	}

	if rule.MaxAmount > 0 && rule.MinAmount > rule.MaxAmount {
		return fmt.Errorf("minimum amount cannot be greater than maximum amount")
	}

	if rule.EffectiveDate.IsZero() {
		return fmt.Errorf("effective date is required")
	}

	if rule.ExpirationDate != nil && rule.ExpirationDate.Before(rule.EffectiveDate) {
		return fmt.Errorf("expiration date must be after effective date")
	}

	return nil
}

// GetCommissionRules retrieves commission rules for a partner.
func (s *CommissionService) GetCommissionRules(ctx context.Context, partnerID uuid.UUID) ([]CommissionRule, error) {
	// In a real implementation, this would query commission rules from the database
	// For now, we'll return an empty slice
	return []CommissionRule{}, nil
}

// ProcessBulkCommissionPayments processes multiple commission payments in batch.
func (s *CommissionService) ProcessBulkCommissionPayments(ctx context.Context, paymentRequests []CommissionPaymentRequest) ([]CommissionPayment, error) {
	payments := []CommissionPayment{}

	for _, request := range paymentRequests {
		payment, err := s.ProcessCommissionPayment(ctx, request.CommissionID, request.PaymentMethod)
		if err != nil {
			// Log error but continue processing other payments
			continue
		}
		payments = append(payments, *payment)
	}

	return payments, nil
}

// CommissionPaymentRequest represents a request to process a commission payment.
type CommissionPaymentRequest struct {
	CommissionID  uuid.UUID `json:"commission_id"`
	PaymentMethod string    `json:"payment_method"`
}

// ValidateCommissionCalculation validates the integrity of a commission calculation.
func (s *CommissionService) ValidateCommissionCalculation(calculation *CommissionCalculation) error {
	if calculation == nil {
		return fmt.Errorf("commission calculation cannot be nil")
	}

	if calculation.PolicyID == uuid.Nil {
		return fmt.Errorf("policy ID is required")
	}

	if calculation.PartnerID == uuid.Nil {
		return fmt.Errorf("partner ID is required")
	}

	if calculation.CommissionType == "" {
		return fmt.Errorf("commission type is required")
	}

	if calculation.BaseAmount < 0 {
		return fmt.Errorf("base amount cannot be negative")
	}

	if calculation.CommissionRate < 0 || calculation.CommissionRate > 100 {
		return fmt.Errorf("commission rate must be between 0 and 100")
	}

	if calculation.CommissionAmount < 0 {
		return fmt.Errorf("commission amount cannot be negative")
	}

	if calculation.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if calculation.Status == "" {
		return fmt.Errorf("status is required")
	}

	validStatuses := []string{"calculated", "paid", "pending", "cancelled"}
	valid := false
	for _, validStatus := range validStatuses {
		if calculation.Status == validStatus {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status: %s", calculation.Status)
	}

	if calculation.CalculationDate.IsZero() {
		return fmt.Errorf("calculation date is required")
	}

	if calculation.DueDate.IsZero() {
		return fmt.Errorf("due date is required")
	}

	if calculation.DueDate.Before(calculation.CalculationDate) {
		return fmt.Errorf("due date must be after calculation date")
	}

	return nil
}
