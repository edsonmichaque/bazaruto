package handlers

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/edsonmichaque/bazaruto/internal/events"
	"github.com/edsonmichaque/bazaruto/internal/jobs"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/pkg/event"
	"github.com/edsonmichaque/bazaruto/pkg/job"
)

// ClaimSubmittedHandler handles claim submission events.
type ClaimSubmittedHandler struct {
	claimService *services.ClaimService
	dispatcher   job.Dispatcher
	logger       *logger.Logger
}

// NewClaimSubmittedHandler creates a new claim submitted event handler.
func NewClaimSubmittedHandler(claimService *services.ClaimService, dispatcher job.Dispatcher, logger *logger.Logger) *ClaimSubmittedHandler {
	return &ClaimSubmittedHandler{
		claimService: claimService,
		dispatcher:   dispatcher,
		logger:       logger,
	}
}

// Handle processes a claim submitted event.
func (h *ClaimSubmittedHandler) Handle(ctx context.Context, event event.Event) error {
	claimEvent, ok := event.(*events.ClaimSubmittedEvent)
	if !ok {
		return fmt.Errorf("expected ClaimSubmittedEvent, got %T", event)
	}

	h.logger.Info("Processing claim submitted event",
		zap.String("claim_id", claimEvent.ClaimID.String()),
		zap.String("user_id", claimEvent.UserID.String()),
		zap.Float64("claim_amount", claimEvent.ClaimAmount))

	// Dispatch fraud detection job
	fraudJob := &jobs.FraudDetectionJob{
		ClaimID: claimEvent.ClaimID,
	}

	if err := h.dispatcher.PerformWithContext(ctx, fraudJob); err != nil {
		h.logger.Error("Failed to dispatch fraud detection job",
			zap.Error(err),
			zap.String("claim_id", claimEvent.ClaimID.String()))
		return fmt.Errorf("failed to dispatch fraud detection job: %w", err)
	}

	h.logger.Info("Fraud detection job dispatched successfully",
		zap.String("claim_id", claimEvent.ClaimID.String()))

	return nil
}

// CanHandle returns true if this handler can process the given event type.
func (h *ClaimSubmittedHandler) CanHandle(eventType string) bool {
	return eventType == "claim.submitted"
}

// HandlerName returns a unique name for this handler.
func (h *ClaimSubmittedHandler) HandlerName() string {
	return "claim_submitted_handler"
}
