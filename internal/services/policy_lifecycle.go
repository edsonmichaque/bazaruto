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

// PolicyLifecycleService handles policy renewal, cancellation, and lifecycle management.
type PolicyLifecycleService struct {
	policyStore       store.PolicyStore
	paymentStore      store.PaymentStore
	subscriptionStore store.SubscriptionStore
	userStore         store.UserStore
	eventService      *EventService
	configManager     *config.Manager
	logger            *logger.Logger
}

// NewPolicyLifecycleService creates a new PolicyLifecycleService instance.
func NewPolicyLifecycleService(
	logger *logger.Logger,
	configManager *config.Manager,
	policyStore store.PolicyStore,
	paymentStore store.PaymentStore,
	subscriptionStore store.SubscriptionStore,
	userStore store.UserStore,
	eventService *EventService,
) *PolicyLifecycleService {
	return &PolicyLifecycleService{
		policyStore:       policyStore,
		paymentStore:      paymentStore,
		subscriptionStore: subscriptionStore,
		userStore:         userStore,
		eventService:      eventService,
		configManager:     configManager,
		logger:            logger,
	}
}

// RenewalResult represents the result of a policy renewal attempt.
type RenewalResult struct {
	Success        bool                   `json:"success"`
	NewPolicyID    *uuid.UUID             `json:"new_policy_id,omitempty"`
	RenewalDate    time.Time              `json:"renewal_date"`
	Premium        float64                `json:"premium"`
	Currency       string                 `json:"currency"`
	Status         string                 `json:"status"` // renewed, failed, pending_payment
	Message        string                 `json:"message"`
	GracePeriodEnd *time.Time             `json:"grace_period_end,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CancellationResult represents the result of a policy cancellation.
type CancellationResult struct {
	Success          bool                   `json:"success"`
	CancellationDate time.Time              `json:"cancellation_date"`
	EffectiveDate    time.Time              `json:"effective_date"`
	RefundAmount     float64                `json:"refund_amount"`
	RefundCurrency   string                 `json:"refund_currency"`
	Status           string                 `json:"status"` // cancelled, pending_refund, failed
	Message          string                 `json:"message"`
	GracePeriodEnd   *time.Time             `json:"grace_period_end,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// RenewPolicy attempts to renew an existing policy.
func (s *PolicyLifecycleService) RenewPolicy(ctx context.Context, policyID uuid.UUID, renewalOptions *RenewalOptions) (*RenewalResult, error) {
	// Fetch existing policy
	policy, err := s.policyStore.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policy: %w", err)
	}

	// Validate policy is eligible for renewal
	if err := s.validateRenewalEligibility(policy); err != nil {
		return &RenewalResult{
			Success: false,
			Status:  "failed",
			Message: err.Error(),
		}, nil
	}

	// Set default renewal options
	if renewalOptions == nil {
		renewalOptions = s.getDefaultRenewalOptions(policy)
	}

	// Calculate new premium
	newPremium, err := s.calculateRenewalPremium(ctx, policy, renewalOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate renewal premium: %w", err)
	}

	// Create new policy
	newPolicy := &models.Policy{
		ProductID:        policy.ProductID,
		UserID:           policy.UserID,
		Premium:          newPremium,
		Currency:         policy.Currency,
		CoverageAmount:   renewalOptions.CoverageAmount,
		Status:           "pending",
		EffectiveDate:    renewalOptions.EffectiveDate,
		ExpirationDate:   renewalOptions.ExpirationDate,
		PaymentFrequency: renewalOptions.PaymentFrequency,
		AutoRenew:        renewalOptions.AutoRenew,
	}

	// Set renewal date if auto-renew is enabled
	if newPolicy.AutoRenew {
		newPolicy.RenewalDate = &newPolicy.ExpirationDate
	}

	// Create new policy in database
	if err := s.policyStore.CreatePolicy(ctx, newPolicy); err != nil {
		return nil, fmt.Errorf("failed to create renewal policy: %w", err)
	}

	// Handle payment for renewal
	result := &RenewalResult{
		NewPolicyID: &newPolicy.ID,
		RenewalDate: time.Now(),
		Premium:     newPremium,
		Currency:    policy.Currency,
		Metadata:    make(map[string]interface{}),
	}

	if renewalOptions.PaymentMethod != "" {
		// Process payment for renewal
		_, err = s.processRenewalPayment(ctx, newPolicy, renewalOptions)
		if err != nil {
			// Payment failed - set grace period
			result.Success = false
			result.Status = "pending_payment"
			result.Message = "Renewal created but payment failed"
			result.GracePeriodEnd = s.calculateGracePeriodEnd()
			result.Metadata["payment_error"] = err.Error()
		} else {
			// Payment successful
			result.Success = true
			result.Status = "renewed"
			result.Message = "Policy renewed successfully"
			newPolicy.Status = "active"
			_ = s.policyStore.UpdatePolicy(ctx, newPolicy)
		}
	} else {
		// No payment method specified - set grace period
		result.Success = false
		result.Status = "pending_payment"
		result.Message = "Renewal created but payment method required"
		result.GracePeriodEnd = s.calculateGracePeriodEnd()
	}

	// Publish renewal event
	if s.eventService != nil {
		policyEvent := events.NewPolicyCreatedEvent(
			newPolicy.ID,
			newPolicy.UserID,
			uuid.Nil, // No quote ID for renewals
			newPolicy.ProductID,
			newPolicy.Premium,
			newPolicy.Currency,
			newPolicy.EffectiveDate,
			newPolicy.ExpirationDate,
			time.Now(),
		)
		if err := s.eventService.PublishEvent(ctx, policyEvent); err != nil {
			s.logger.Error("Failed to publish policy renewal event",
				zap.String("policy_id", newPolicy.ID.String()),
				zap.Error(err))
		}
	}

	// Publish policy renewed event
	if s.eventService != nil {
		policyRenewedEvent := events.NewPolicyRenewedEvent(
			policy.ID,
			newPolicy.ID,
			newPolicy.UserID,
			newPolicy.ProductID,
			newPolicy.Premium,
			newPolicy.Currency,
			newPolicy.EffectiveDate,
			newPolicy.ExpirationDate,
			time.Now(),
		)
		if err := s.eventService.PublishEvent(ctx, policyRenewedEvent); err != nil {
			s.logger.Error("Failed to publish policy renewed event",
				zap.String("old_policy_id", policy.ID.String()),
				zap.String("new_policy_id", newPolicy.ID.String()),
				zap.Error(err))
		}
	}

	return result, nil
}

// RenewalOptions represents options for policy renewal.
type RenewalOptions struct {
	CoverageAmount   float64   `json:"coverage_amount"`
	EffectiveDate    time.Time `json:"effective_date"`
	ExpirationDate   time.Time `json:"expiration_date"`
	PaymentFrequency string    `json:"payment_frequency"`
	AutoRenew        bool      `json:"auto_renew"`
	PaymentMethod    string    `json:"payment_method"`
}

// validateRenewalEligibility validates if a policy is eligible for renewal.
func (s *PolicyLifecycleService) validateRenewalEligibility(policy *models.Policy) error {
	// Check if policy is active
	if policy.Status != models.PolicyStatusActive {
		return fmt.Errorf("policy is not active and cannot be renewed")
	}

	// Check if policy is not already expired
	if policy.ExpirationDate.Before(time.Now()) {
		return fmt.Errorf("policy has expired and cannot be renewed")
	}

	// Check if policy is not already cancelled
	if policy.Status == models.PolicyStatusCancelled {
		return fmt.Errorf("cancelled policy cannot be renewed")
	}

	// Check if renewal is within allowed timeframe (e.g., 30 days before expiration)
	daysUntilExpiration := time.Until(policy.ExpirationDate).Hours() / 24
	if daysUntilExpiration > 30 {
		return fmt.Errorf("policy cannot be renewed more than 30 days before expiration")
	}

	return nil
}

// getDefaultRenewalOptions returns default renewal options for a policy.
func (s *PolicyLifecycleService) getDefaultRenewalOptions(policy *models.Policy) *RenewalOptions {
	// Calculate new effective and expiration dates (typically 1 year from current expiration)
	newEffectiveDate := policy.ExpirationDate
	newExpirationDate := policy.ExpirationDate.AddDate(1, 0, 0)

	return &RenewalOptions{
		CoverageAmount:   policy.CoverageAmount,
		EffectiveDate:    newEffectiveDate,
		ExpirationDate:   newExpirationDate,
		PaymentFrequency: policy.PaymentFrequency,
		AutoRenew:        policy.AutoRenew,
	}
}

// calculateRenewalPremium calculates the premium for policy renewal.
func (s *PolicyLifecycleService) calculateRenewalPremium(ctx context.Context, policy *models.Policy, options *RenewalOptions) (float64, error) {
	// Base premium from existing policy
	basePremium := policy.Premium

	// Apply coverage amount adjustment
	if options.CoverageAmount != policy.CoverageAmount {
		coverageRatio := options.CoverageAmount / policy.CoverageAmount
		basePremium *= coverageRatio
	}

	// Apply annual rate increase (e.g., 3% per year)
	rateIncrease := 1.03 // 3% increase
	basePremium *= rateIncrease

	// Apply payment frequency adjustment
	switch options.PaymentFrequency {
	case "annually":
		basePremium *= 0.95 // 5% discount for annual payment
	case "quarterly":
		basePremium *= 1.02 // 2% surcharge for quarterly payment
	case "monthly":
		basePremium *= 1.05 // 5% surcharge for monthly payment
	}

	return basePremium, nil
}

// processRenewalPayment processes payment for policy renewal.
func (s *PolicyLifecycleService) processRenewalPayment(ctx context.Context, policy *models.Policy, options *RenewalOptions) (*PaymentResult, error) {
	// Create payment record
	payment := &models.Payment{
		UserID:          policy.UserID,
		PolicyID:        &policy.ID,
		Amount:          policy.Premium,
		Currency:        policy.Currency,
		Status:          models.PaymentStatusPending,
		PaymentMethod:   options.PaymentMethod,
		PaymentProvider: "internal", // In real implementation, this would be the actual provider
	}

	// Create payment in database
	if err := s.paymentStore.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Process payment (simplified - in real implementation, this would integrate with payment gateway)
	payment.Status = models.PaymentStatusCompleted
	now := time.Now()
	payment.ProcessedAt = &now
	payment.TransactionID = fmt.Sprintf("renewal_%d_%s", time.Now().Unix(), policy.ID.String()[:8])

	// Update payment in database
	if err := s.paymentStore.UpdatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return &PaymentResult{
		Success:       true,
		TransactionID: payment.TransactionID,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
	}, nil
}

// PaymentResult represents the result of a payment processing attempt.
type PaymentResult struct {
	Success       bool    `json:"success"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Error         string  `json:"error,omitempty"`
}

// calculateGracePeriodEnd calculates when the grace period ends.
func (s *PolicyLifecycleService) calculateGracePeriodEnd() *time.Time {
	gracePeriodEnd := time.Now().Add(15 * 24 * time.Hour) // 15 days grace period
	return &gracePeriodEnd
}

// CancelPolicy cancels an existing policy.
func (s *PolicyLifecycleService) CancelPolicy(ctx context.Context, policyID uuid.UUID, cancellationOptions *CancellationOptions) (*CancellationResult, error) {
	// Fetch existing policy
	policy, err := s.policyStore.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policy: %w", err)
	}

	// Validate policy can be cancelled
	if err := s.validateCancellationEligibility(policy); err != nil {
		return &CancellationResult{
			Success: false,
			Status:  "failed",
			Message: err.Error(),
		}, nil
	}

	// Set default cancellation options
	if cancellationOptions == nil {
		cancellationOptions = s.getDefaultCancellationOptions(policy)
	}

	// Calculate refund amount
	refundAmount, err := s.calculateRefundAmount(policy, cancellationOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate refund amount: %w", err)
	}

	// Update policy status
	policy.Status = "cancelled"
	now := time.Now()
	policy.UpdatedAt = now

	if err := s.policyStore.UpdatePolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to update policy status: %w", err)
	}

	// Create cancellation result
	result := &CancellationResult{
		Success:          true,
		CancellationDate: now,
		EffectiveDate:    cancellationOptions.EffectiveDate,
		RefundAmount:     refundAmount,
		RefundCurrency:   policy.Currency,
		Status:           "cancelled",
		Message:          "Policy cancelled successfully",
		Metadata:         make(map[string]interface{}),
	}

	// Process refund if applicable
	if refundAmount > 0 {
		refundResult, err := s.processRefund(ctx, policy, refundAmount, cancellationOptions)
		if err != nil {
			result.Status = "pending_refund"
			result.Message = "Policy cancelled but refund processing failed"
			result.Metadata["refund_error"] = err.Error()
		} else {
			result.Status = "cancelled"
			result.Message = "Policy cancelled and refund processed"
			result.Metadata["refund_transaction_id"] = refundResult.TransactionID
		}
	}

	// Publish cancellation event
	if s.eventService != nil {
		// Note: We would need a PolicyCancelledEvent in the events package
		// For now, we'll log the cancellation
		s.logger.Info("Policy cancelled",
			zap.String("policy_id", policy.ID.String()),
			zap.String("user_id", policy.UserID.String()),
			zap.Float64("refund_amount", refundAmount),
			zap.String("reason", cancellationOptions.Reason))
	}

	// Publish policy cancelled event
	if s.eventService != nil {
		policyCancelledEvent := events.NewPolicyCancelledEvent(
			policy.ID,
			policy.UserID,
			policy.ProductID,
			refundAmount,
			policy.Currency,
			now,
			cancellationOptions.EffectiveDate,
			cancellationOptions.Reason,
		)
		if err := s.eventService.PublishEvent(ctx, policyCancelledEvent); err != nil {
			s.logger.Error("Failed to publish policy cancelled event",
				zap.String("policy_id", policy.ID.String()),
				zap.Error(err))
		}
	}

	return result, nil
}

// CancellationOptions represents options for policy cancellation.
type CancellationOptions struct {
	EffectiveDate time.Time `json:"effective_date"`
	Reason        string    `json:"reason"`
	RefundMethod  string    `json:"refund_method"`
}

// validateCancellationEligibility validates if a policy can be cancelled.
func (s *PolicyLifecycleService) validateCancellationEligibility(policy *models.Policy) error {
	// Check if policy is active
	if policy.Status != "active" {
		return fmt.Errorf("only active policies can be cancelled")
	}

	// Check if policy has not already been cancelled
	if policy.Status == "cancelled" {
		return fmt.Errorf("policy has already been cancelled")
	}

	return nil
}

// getDefaultCancellationOptions returns default cancellation options for a policy.
func (s *PolicyLifecycleService) getDefaultCancellationOptions(policy *models.Policy) *CancellationOptions {
	return &CancellationOptions{
		EffectiveDate: time.Now(),
		Reason:        "Customer request",
		RefundMethod:  "original_payment_method",
	}
}

// calculateRefundAmount calculates the refund amount for policy cancellation.
func (s *PolicyLifecycleService) calculateRefundAmount(policy *models.Policy, options *CancellationOptions) (float64, error) {
	// Calculate unused premium
	policyDuration := policy.ExpirationDate.Sub(policy.EffectiveDate).Hours() / 24 / 365 // years
	usedDuration := options.EffectiveDate.Sub(policy.EffectiveDate).Hours() / 24 / 365   // years

	if usedDuration < 0 {
		usedDuration = 0
	}

	if usedDuration >= policyDuration {
		return 0, nil // No refund if policy has been used for full duration
	}

	// Calculate pro-rated refund
	unusedRatio := (policyDuration - usedDuration) / policyDuration
	refundAmount := policy.Premium * unusedRatio

	// Apply cancellation fee (e.g., 10% of refund amount)
	cancellationFee := refundAmount * 0.10
	refundAmount -= cancellationFee

	// Ensure refund is not negative
	if refundAmount < 0 {
		refundAmount = 0
	}

	return refundAmount, nil
}

// processRefund processes the refund for policy cancellation.
func (s *PolicyLifecycleService) processRefund(ctx context.Context, policy *models.Policy, refundAmount float64, options *CancellationOptions) (*PaymentResult, error) {
	// Create refund payment record
	now := time.Now()
	refund := &models.Payment{
		UserID:          policy.UserID,
		PolicyID:        &policy.ID,
		Amount:          -refundAmount, // Negative amount for refund
		Currency:        policy.Currency,
		Status:          "completed",
		PaymentMethod:   options.RefundMethod,
		PaymentProvider: "internal",
		RefundAmount:    refundAmount,
		RefundedAt:      &now,
	}

	// Create refund in database
	if err := s.paymentStore.CreatePayment(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return &PaymentResult{
		Success:       true,
		TransactionID: refund.TransactionID,
		Amount:        refundAmount,
		Currency:      policy.Currency,
	}, nil
}

// ProcessExpiredPolicies processes policies that have expired.
func (s *PolicyLifecycleService) ProcessExpiredPolicies(ctx context.Context) error {
	s.logger.Info("Processing expired policies")

	// Query for expired policies
	// Note: GetExpiredPolicies method needs to be implemented in PolicyStore
	// For now, we'll simulate with an empty slice
	expiredPolicies := []*models.Policy{}
	// expiredPolicies, err := s.policyStore.GetExpiredPolicies(ctx)
	// if err != nil {
	// 	return fmt.Errorf("failed to fetch expired policies: %w", err)
	// }

	processedCount := 0
	for _, policy := range expiredPolicies {
		// Update policy status to expired
		policy.Status = "expired"
		policy.UpdatedAt = time.Now()

		if err := s.policyStore.UpdatePolicy(ctx, policy); err != nil {
			s.logger.Error("Failed to update expired policy status",
				zap.String("policy_id", policy.ID.String()),
				zap.Error(err))
			continue
		}

		// Publish policy expired event
		if s.eventService != nil {
			policyExpiredEvent := events.NewPolicyExpiredEvent(
				policy.ID,
				policy.UserID,
				policy.ProductID,
				policy.ExpirationDate,
				time.Now(),
			)
			if err := s.eventService.PublishEvent(ctx, policyExpiredEvent); err != nil {
				s.logger.Error("Failed to publish policy expired event",
					zap.String("policy_id", policy.ID.String()),
					zap.Error(err))
			}
		}

		processedCount++
	}

	s.logger.Info("Processed expired policies",
		zap.Int("count", processedCount))

	return nil
}

// ProcessGracePeriodExpirations processes policies whose grace periods have expired.
func (s *PolicyLifecycleService) ProcessGracePeriodExpirations(ctx context.Context) error {
	s.logger.Info("Processing grace period expirations")

	// Query for policies with expired grace periods
	// Note: GetPoliciesWithExpiredGracePeriod method needs to be implemented in PolicyStore
	// For now, we'll simulate with an empty slice
	gracePeriodExpiredPolicies := []*models.Policy{}
	// gracePeriodExpiredPolicies, err := s.policyStore.GetPoliciesWithExpiredGracePeriod(ctx)
	// if err != nil {
	// 	return fmt.Errorf("failed to fetch policies with expired grace periods: %w", err)
	// }

	processedCount := 0
	for _, policy := range gracePeriodExpiredPolicies {
		// Cancel policy if grace period has expired
		cancellationOptions := &CancellationOptions{
			EffectiveDate: time.Now(),
			Reason:        "Grace period expired",
			RefundMethod:  "original_payment_method",
		}

		_, err := s.CancelPolicy(ctx, policy.ID, cancellationOptions)
		if err != nil {
			s.logger.Error("Failed to cancel policy after grace period expiration",
				zap.String("policy_id", policy.ID.String()),
				zap.Error(err))
			continue
		}

		// Publish grace period expiration event
		if s.eventService != nil {
			gracePeriodExpiredEvent := events.NewGracePeriodExpiredEvent(
				policy.ID,
				policy.UserID,
				policy.ProductID,
				time.Now().Add(-15*24*time.Hour), // Assume 15 days grace period
				time.Now(),
			)
			if err := s.eventService.PublishEvent(ctx, gracePeriodExpiredEvent); err != nil {
				s.logger.Error("Failed to publish grace period expired event",
					zap.String("policy_id", policy.ID.String()),
					zap.Error(err))
			}
		}

		processedCount++
	}

	s.logger.Info("Processed grace period expirations",
		zap.Int("count", processedCount))

	return nil
}

// GetPolicyStatus retrieves the current status of a policy including renewal/cancellation information.
func (s *PolicyLifecycleService) GetPolicyStatus(ctx context.Context, policyID uuid.UUID) (*PolicyStatus, error) {
	// Fetch policy details
	policy, err := s.policyStore.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policy: %w", err)
	}

	// Calculate status information
	status := &PolicyStatus{
		PolicyID:        policy.ID,
		Status:          policy.Status,
		EffectiveDate:   policy.EffectiveDate,
		ExpirationDate:  policy.ExpirationDate,
		DaysUntilExpiry: int(time.Until(policy.ExpirationDate).Hours() / 24),
		CanRenew:        s.canRenew(policy),
		CanCancel:       s.canCancel(policy),
		AutoRenew:       policy.AutoRenew,
		GracePeriodEnd:  s.calculateGracePeriodEnd(),
	}

	return status, nil
}

// PolicyStatus represents the current status of a policy.
type PolicyStatus struct {
	PolicyID        uuid.UUID  `json:"policy_id"`
	Status          string     `json:"status"`
	EffectiveDate   time.Time  `json:"effective_date"`
	ExpirationDate  time.Time  `json:"expiration_date"`
	DaysUntilExpiry int        `json:"days_until_expiry"`
	CanRenew        bool       `json:"can_renew"`
	CanCancel       bool       `json:"can_cancel"`
	AutoRenew       bool       `json:"auto_renew"`
	GracePeriodEnd  *time.Time `json:"grace_period_end,omitempty"`
}

// canRenew checks if a policy can be renewed.
func (s *PolicyLifecycleService) canRenew(policy *models.Policy) bool {
	return policy.Status == "active" &&
		policy.ExpirationDate.After(time.Now()) &&
		time.Until(policy.ExpirationDate).Hours()/24 <= 30
}

// canCancel checks if a policy can be cancelled.
func (s *PolicyLifecycleService) canCancel(policy *models.Policy) bool {
	return policy.Status == "active"
}

// ProcessAutoRenewals processes policies that are eligible for auto-renewal.
func (s *PolicyLifecycleService) ProcessAutoRenewals(ctx context.Context) error {
	s.logger.Info("Processing auto-renewals")

	// Query for policies eligible for auto-renewal
	// Note: GetPoliciesEligibleForAutoRenewal method needs to be implemented in PolicyStore
	// For now, we'll simulate with an empty slice
	autoRenewalPolicies := []*models.Policy{}
	// autoRenewalPolicies, err := s.policyStore.GetPoliciesEligibleForAutoRenewal(ctx)
	// if err != nil {
	// 	return fmt.Errorf("failed to fetch policies eligible for auto-renewal: %w", err)
	// }

	processedCount := 0
	for _, policy := range autoRenewalPolicies {
		// Attempt auto-renewal
		renewalOptions := s.getDefaultRenewalOptions(policy)
		renewalOptions.AutoRenew = true

		// Get user's default payment method
		// Note: GetUser method needs to be implemented in UserStore
		// For now, we'll skip the payment method lookup
		// user, err := s.userStore.GetUser(ctx, policy.UserID)
		// if err != nil {
		// 	s.logger.Error("Failed to fetch user for auto-renewal",
		// 		zap.String("policy_id", policy.ID.String()),
		// 		zap.String("user_id", policy.UserID.String()),
		// 		zap.Error(err))
		// 	continue
		// }

		// Set default payment method if available
		// if user.DefaultPaymentMethod != "" {
		// 	renewalOptions.PaymentMethod = user.DefaultPaymentMethod
		// }

		result, err := s.RenewPolicy(ctx, policy.ID, renewalOptions)
		if err != nil {
			s.logger.Error("Failed to auto-renew policy",
				zap.String("policy_id", policy.ID.String()),
				zap.Error(err))
			continue
		}

		if result.Success {
			s.logger.Info("Policy auto-renewed successfully",
				zap.String("old_policy_id", policy.ID.String()),
				zap.String("new_policy_id", result.NewPolicyID.String()))
		} else {
			s.logger.Warn("Policy auto-renewal failed",
				zap.String("policy_id", policy.ID.String()),
				zap.String("status", result.Status),
				zap.String("message", result.Message))
		}

		processedCount++
	}

	s.logger.Info("Processed auto-renewals",
		zap.Int("count", processedCount))

	return nil
}

// GetUpcomingRenewals retrieves policies that are coming up for renewal.
func (s *PolicyLifecycleService) GetUpcomingRenewals(ctx context.Context, daysAhead int) ([]*models.Policy, error) {
	// Query for policies expiring within the specified number of days
	// Note: GetPoliciesExpiringWithin method needs to be implemented in PolicyStore
	// For now, we'll simulate with an empty slice
	upcomingRenewals := []*models.Policy{}
	// upcomingRenewals, err := s.policyStore.GetPoliciesExpiringWithin(ctx, daysAhead)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to fetch upcoming renewals: %w", err)
	// }

	return upcomingRenewals, nil
}

// SendRenewalReminders sends renewal reminder notifications for policies expiring soon.
func (s *PolicyLifecycleService) SendRenewalReminders(ctx context.Context, daysAhead int) error {
	s.logger.Info("Sending renewal reminders", zap.Int("days_ahead", daysAhead))

	upcomingRenewals, err := s.GetUpcomingRenewals(ctx, daysAhead)
	if err != nil {
		return fmt.Errorf("failed to get upcoming renewals: %w", err)
	}

	sentCount := 0
	for _, policy := range upcomingRenewals {
		// Publish renewal reminder event
		if s.eventService != nil {
			renewalReminderEvent := events.NewRenewalReminderEvent(
				policy.ID,
				policy.UserID,
				policy.ProductID,
				daysAhead,
				policy.ExpirationDate,
				time.Now(),
			)
			if err := s.eventService.PublishEvent(ctx, renewalReminderEvent); err != nil {
				s.logger.Error("Failed to publish renewal reminder event",
					zap.String("policy_id", policy.ID.String()),
					zap.Error(err))
				continue
			}
		}

		sentCount++
	}

	s.logger.Info("Sent renewal reminders",
		zap.Int("count", sentCount))

	return nil
}
