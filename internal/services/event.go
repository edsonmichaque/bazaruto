package services

import (
	"context"

	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/pkg/event"
	"go.uber.org/zap"
)

// EventService handles event publishing and subscription management.
type EventService struct {
	eventBus event.EventBus
	logger   *logger.Logger
}

// NewEventService creates a new event service.
func NewEventService(eventBus event.EventBus, logger *logger.Logger) *EventService {
	return &EventService{
		eventBus: eventBus,
		logger:   logger,
	}
}

// PublishEvent publishes an event to the event bus asynchronously.
func (s *EventService) PublishEvent(ctx context.Context, event event.Event) error {
	s.logger.Info("Publishing event",
		zap.String("event_type", event.Type()),
		zap.String("event_id", event.ID().String()),
		zap.String("aggregate_id", event.AggregateID().String()))

	// Publish asynchronously - the event bus handles this
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("Failed to publish event",
			zap.Error(err),
			zap.String("event_type", event.Type()),
			zap.String("event_id", event.ID().String()))
		return err
	}

	return nil
}

// SubscribeHandler subscribes an event handler to specific event types.
func (s *EventService) SubscribeHandler(handler event.EventHandler, eventTypes ...string) error {
	if err := s.eventBus.Subscribe(handler, eventTypes...); err != nil {
		s.logger.Error("Failed to subscribe event handler",
			zap.Error(err),
			zap.String("handler_name", handler.HandlerName()),
			zap.Strings("event_types", eventTypes))
		return err
	}

	s.logger.Info("Event handler subscribed",
		zap.String("handler_name", handler.HandlerName()),
		zap.Strings("event_types", eventTypes))

	return nil
}

// UnsubscribeHandler unsubscribes an event handler.
func (s *EventService) UnsubscribeHandler(handlerName string) error {
	if err := s.eventBus.Unsubscribe(handlerName); err != nil {
		s.logger.Error("Failed to unsubscribe event handler",
			zap.Error(err),
			zap.String("handler_name", handlerName))
		return err
	}

	s.logger.Info("Event handler unsubscribed",
		zap.String("handler_name", handlerName))

	return nil
}

// Close closes the event service and cleans up resources.
func (s *EventService) Close() error {
	if err := s.eventBus.Close(); err != nil {
		s.logger.Error("Failed to close event bus",
			zap.Error(err))
		return err
	}

	s.logger.Info("Event service closed")
	return nil
}
