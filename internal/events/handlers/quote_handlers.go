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

// QuoteCreatedHandler handles quote creation events.
type QuoteCreatedHandler struct {
	quoteService *services.QuoteService
	dispatcher   job.Dispatcher
	logger       *logger.Logger
}

// NewQuoteCreatedHandler creates a new quote created event handler.
func NewQuoteCreatedHandler(quoteService *services.QuoteService, dispatcher job.Dispatcher, logger *logger.Logger) *QuoteCreatedHandler {
	return &QuoteCreatedHandler{
		quoteService: quoteService,
		dispatcher:   dispatcher,
		logger:       logger,
	}
}

// Handle processes a quote created event.
func (h *QuoteCreatedHandler) Handle(ctx context.Context, event event.Event) error {
	quoteEvent, ok := event.(*events.QuoteCreatedEvent)
	if !ok {
		return fmt.Errorf("expected QuoteCreatedEvent, got %T", event)
	}

	h.logger.Info("Processing quote created event",
		zap.String("quote_id", quoteEvent.QuoteID.String()),
		zap.String("user_id", quoteEvent.UserID.String()),
		zap.Float64("coverage_amount", quoteEvent.CoverageAmount))

	// Dispatch premium calculation job
	calculateJob := &jobs.CalculatePremiumJob{
		QuoteID:      quoteEvent.QuoteID,
		QuoteService: h.quoteService,
	}

	if err := h.dispatcher.PerformWithContext(ctx, calculateJob); err != nil {
		h.logger.Error("Failed to dispatch premium calculation job",
			zap.Error(err),
			zap.String("quote_id", quoteEvent.QuoteID.String()))
		return fmt.Errorf("failed to dispatch premium calculation job: %w", err)
	}

	h.logger.Info("Premium calculation job dispatched successfully",
		zap.String("quote_id", quoteEvent.QuoteID.String()))

	return nil
}

// CanHandle returns true if this handler can process the given event type.
func (h *QuoteCreatedHandler) CanHandle(eventType string) bool {
	return eventType == "quote.created"
}

// HandlerName returns a unique name for this handler.
func (h *QuoteCreatedHandler) HandlerName() string {
	return "quote_created_handler"
}

// QuoteCalculatedHandler handles quote calculation events.
type QuoteCalculatedHandler struct {
	dispatcher job.Dispatcher
	logger     *logger.Logger
}

// NewQuoteCalculatedHandler creates a new quote calculated event handler.
func NewQuoteCalculatedHandler(dispatcher job.Dispatcher, logger *logger.Logger) *QuoteCalculatedHandler {
	return &QuoteCalculatedHandler{
		dispatcher: dispatcher,
		logger:     logger,
	}
}

// Handle processes a quote calculated event.
func (h *QuoteCalculatedHandler) Handle(ctx context.Context, event event.Event) error {
	quoteEvent, ok := event.(*events.QuoteCalculatedEvent)
	if !ok {
		return fmt.Errorf("expected QuoteCalculatedEvent, got %T", event)
	}

	h.logger.Info("Processing quote calculated event",
		zap.String("quote_id", quoteEvent.QuoteID.String()),
		zap.String("user_id", quoteEvent.UserID.String()),
		zap.Float64("final_premium", quoteEvent.FinalPremium))

	// Dispatch quote PDF generation job
	pdfJob := &jobs.GenerateQuotePDFJob{
		QuoteID: quoteEvent.QuoteID,
	}

	if err := h.dispatcher.PerformWithContext(ctx, pdfJob); err != nil {
		h.logger.Error("Failed to dispatch quote PDF generation job",
			zap.Error(err),
			zap.String("quote_id", quoteEvent.QuoteID.String()))
		return fmt.Errorf("failed to dispatch quote PDF generation job: %w", err)
	}

	h.logger.Info("Quote PDF generation job dispatched successfully",
		zap.String("quote_id", quoteEvent.QuoteID.String()))

	return nil
}

// CanHandle returns true if this handler can process the given event type.
func (h *QuoteCalculatedHandler) CanHandle(eventType string) bool {
	return eventType == "quote.calculated"
}

// HandlerName returns a unique name for this handler.
func (h *QuoteCalculatedHandler) HandlerName() string {
	return "quote_calculated_handler"
}
