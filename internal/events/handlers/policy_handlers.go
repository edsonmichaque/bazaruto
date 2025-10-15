package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/events"
	"github.com/edsonmichaque/bazaruto/internal/jobs"
	"github.com/edsonmichaque/bazaruto/pkg/event"
	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PolicyEventHandler handles policy-related events and dispatches appropriate jobs.
type PolicyEventHandler struct {
	dispatcher job.Dispatcher
	logger     *zap.Logger
}

// NewPolicyEventHandler creates a new policy event handler.
func NewPolicyEventHandler(dispatcher job.Dispatcher, logger *zap.Logger) *PolicyEventHandler {
	return &PolicyEventHandler{
		dispatcher: dispatcher,
		logger:     logger,
	}
}

// HandlePolicyRenewed handles policy renewed events.
func (h *PolicyEventHandler) HandlePolicyRenewed(ctx context.Context, event *events.PolicyRenewedEvent) error {
	h.logger.Info("Handling policy renewed event",
		zap.String("old_policy_id", event.OldPolicyID.String()),
		zap.String("new_policy_id", event.NewPolicyID.String()),
		zap.String("user_id", event.UserID.String()))

	// Dispatch notification job for policy renewal
	notificationJob := &jobs.PushNotificationJob{
		ID:        uuid.New(),
		UserID:    event.UserID,
		Title:     "Policy Renewal Confirmation",
		Body:      fmt.Sprintf("Your policy has been successfully renewed. New policy ID: %s", event.NewPolicyID.String()),
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, notificationJob); err != nil {
		h.logger.Error("Failed to dispatch policy renewal notification job",
			zap.String("policy_id", event.NewPolicyID.String()),
			zap.Error(err))
		return err
	}

	// Dispatch email job for policy renewal
	emailJob := &jobs.SendEmailJob{
		ID:        uuid.New(),
		To:        "", // Would need to fetch user email from database
		Subject:   "Policy Renewal Confirmation",
		Body:      fmt.Sprintf("Your policy has been successfully renewed. New policy ID: %s", event.NewPolicyID.String()),
		Template:  "policy_renewal",
		From:      "noreply@bazaruto.com",
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, emailJob); err != nil {
		h.logger.Error("Failed to dispatch policy renewal email job",
			zap.String("policy_id", event.NewPolicyID.String()),
			zap.Error(err))
		return err
	}

	return nil
}

// HandlePolicyCancelled handles policy cancelled events.
func (h *PolicyEventHandler) HandlePolicyCancelled(ctx context.Context, event *events.PolicyCancelledEvent) error {
	h.logger.Info("Handling policy cancelled event",
		zap.String("policy_id", event.PolicyID.String()),
		zap.String("user_id", event.UserID.String()),
		zap.Float64("refund_amount", event.RefundAmount))

	// Dispatch notification job for policy cancellation
	notificationJob := &jobs.PushNotificationJob{
		ID:        uuid.New(),
		UserID:    event.UserID,
		Title:     "Policy Cancellation Confirmation",
		Body:      fmt.Sprintf("Your policy has been cancelled. Refund amount: %.2f %s", event.RefundAmount, event.Currency),
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, notificationJob); err != nil {
		h.logger.Error("Failed to dispatch policy cancellation notification job",
			zap.String("policy_id", event.PolicyID.String()),
			zap.Error(err))
		return err
	}

	// Dispatch email job for policy cancellation
	emailJob := &jobs.SendEmailJob{
		ID:        uuid.New(),
		To:        "", // Would need to fetch user email from database
		Subject:   "Policy Cancellation Confirmation",
		Body:      fmt.Sprintf("Your policy has been cancelled. Refund amount: %.2f %s", event.RefundAmount, event.Currency),
		Template:  "policy_cancellation",
		From:      "noreply@bazaruto.com",
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, emailJob); err != nil {
		h.logger.Error("Failed to dispatch policy cancellation email job",
			zap.String("policy_id", event.PolicyID.String()),
			zap.Error(err))
		return err
	}

	return nil
}

// HandlePolicyExpired handles policy expired events.
func (h *PolicyEventHandler) HandlePolicyExpired(ctx context.Context, event *events.PolicyExpiredEvent) error {
	h.logger.Info("Handling policy expired event",
		zap.String("policy_id", event.PolicyID.String()),
		zap.String("user_id", event.UserID.String()),
		zap.Time("expiration_date", event.ExpirationDate))

	// Dispatch notification job for policy expiration
	notificationJob := &jobs.PushNotificationJob{
		ID:        uuid.New(),
		UserID:    event.UserID,
		Title:     "Policy Expired",
		Body:      fmt.Sprintf("Your policy %s has expired. Please renew to maintain coverage.", event.PolicyID.String()),
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, notificationJob); err != nil {
		h.logger.Error("Failed to dispatch policy expired notification job",
			zap.String("policy_id", event.PolicyID.String()),
			zap.Error(err))
		return err
	}

	// Dispatch email job for policy expiration
	emailJob := &jobs.SendEmailJob{
		ID:        uuid.New(),
		To:        "", // Would need to fetch user email from database
		Subject:   "Policy Expired - Action Required",
		Body:      fmt.Sprintf("Your policy %s has expired. Please renew to maintain coverage.", event.PolicyID.String()),
		Template:  "policy_expired",
		From:      "noreply@bazaruto.com",
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, emailJob); err != nil {
		h.logger.Error("Failed to dispatch policy expired email job",
			zap.String("policy_id", event.PolicyID.String()),
			zap.Error(err))
		return err
	}

	return nil
}

// HandleGracePeriodExpired handles grace period expired events.
func (h *PolicyEventHandler) HandleGracePeriodExpired(ctx context.Context, event *events.GracePeriodExpiredEvent) error {
	h.logger.Info("Handling grace period expired event",
		zap.String("policy_id", event.PolicyID.String()),
		zap.String("user_id", event.UserID.String()),
		zap.Time("grace_period_end_date", event.GracePeriodEndDate))

	// Dispatch notification job for grace period expiration
	notificationJob := &jobs.PushNotificationJob{
		ID:        uuid.New(),
		UserID:    event.UserID,
		Title:     "Policy Cancelled - Grace Period Expired",
		Body:      fmt.Sprintf("Your policy %s has been cancelled due to expired grace period.", event.PolicyID.String()),
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, notificationJob); err != nil {
		h.logger.Error("Failed to dispatch grace period expired notification job",
			zap.String("policy_id", event.PolicyID.String()),
			zap.Error(err))
		return err
	}

	// Dispatch email job for grace period expiration
	emailJob := &jobs.SendEmailJob{
		ID:        uuid.New(),
		To:        "", // Would need to fetch user email from database
		Subject:   "Policy Cancelled - Grace Period Expired",
		Body:      fmt.Sprintf("Your policy %s has been cancelled due to expired grace period.", event.PolicyID.String()),
		Template:  "grace_period_expired",
		From:      "noreply@bazaruto.com",
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, emailJob); err != nil {
		h.logger.Error("Failed to dispatch grace period expired email job",
			zap.String("policy_id", event.PolicyID.String()),
			zap.Error(err))
		return err
	}

	return nil
}

// HandleRenewalReminder handles renewal reminder events.
func (h *PolicyEventHandler) HandleRenewalReminder(ctx context.Context, event *events.RenewalReminderEvent) error {
	h.logger.Info("Handling renewal reminder event",
		zap.String("policy_id", event.PolicyID.String()),
		zap.String("user_id", event.UserID.String()),
		zap.Int("days_until_expiry", event.DaysUntilExpiry))

	// Dispatch notification job for renewal reminder
	notificationJob := &jobs.PushNotificationJob{
		ID:        uuid.New(),
		UserID:    event.UserID,
		Title:     "Policy Renewal Reminder",
		Body:      fmt.Sprintf("Your policy %s expires in %d days. Please renew to maintain coverage.", event.PolicyID.String(), event.DaysUntilExpiry),
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, notificationJob); err != nil {
		h.logger.Error("Failed to dispatch renewal reminder notification job",
			zap.String("policy_id", event.PolicyID.String()),
			zap.Error(err))
		return err
	}

	// Dispatch email job for renewal reminder
	emailJob := &jobs.SendEmailJob{
		ID:        uuid.New(),
		To:        "", // Would need to fetch user email from database
		Subject:   "Policy Renewal Reminder",
		Body:      fmt.Sprintf("Your policy %s expires in %d days. Please renew to maintain coverage.", event.PolicyID.String(), event.DaysUntilExpiry),
		Template:  "renewal_reminder",
		From:      "noreply@bazaruto.com",
		Attempts:  0,
		RunAtTime: time.Now(),
	}

	if err := h.dispatcher.PerformWithContext(ctx, emailJob); err != nil {
		h.logger.Error("Failed to dispatch renewal reminder email job",
			zap.String("policy_id", event.PolicyID.String()),
			zap.Error(err))
		return err
	}

	return nil
}

// Handle processes events and dispatches appropriate jobs.
func (h *PolicyEventHandler) Handle(ctx context.Context, event event.Event) error {
	switch e := event.(type) {
	case *events.PolicyRenewedEvent:
		return h.HandlePolicyRenewed(ctx, e)
	case *events.PolicyCancelledEvent:
		return h.HandlePolicyCancelled(ctx, e)
	case *events.PolicyExpiredEvent:
		return h.HandlePolicyExpired(ctx, e)
	case *events.GracePeriodExpiredEvent:
		return h.HandleGracePeriodExpired(ctx, e)
	case *events.RenewalReminderEvent:
		return h.HandleRenewalReminder(ctx, e)
	default:
		return fmt.Errorf("unsupported event type: %T", event)
	}
}

// CanHandle returns true if this handler can process the given event type.
func (h *PolicyEventHandler) CanHandle(eventType string) bool {
	supportedTypes := map[string]bool{
		"policy.renewed":       true,
		"policy.cancelled":     true,
		"policy.expired":       true,
		"grace_period.expired": true,
		"renewal.reminder":     true,
	}
	return supportedTypes[eventType]
}

// HandlerName returns a unique name for this handler.
func (h *PolicyEventHandler) HandlerName() string {
	return "policy_event_handler"
}
