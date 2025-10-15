package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event that occurred in the system.
type Event interface {
	// ID returns the unique identifier for this event
	ID() uuid.UUID

	// Type returns the event type name
	Type() string

	// AggregateID returns the ID of the aggregate that generated this event
	AggregateID() uuid.UUID

	// AggregateType returns the type of the aggregate (e.g., "user", "quote", "payment")
	AggregateType() string

	// Version returns the version of the aggregate when this event occurred
	Version() int

	// OccurredAt returns when this event occurred
	OccurredAt() time.Time

	// Data returns the event payload as a map
	Data() map[string]interface{}

	// Metadata returns additional metadata about the event
	Metadata() map[string]interface{}
}

// BaseEvent provides a default implementation of the Event interface.
type BaseEvent struct {
	id            uuid.UUID
	eventType     string
	aggregateID   uuid.UUID
	aggregateType string
	version       int
	occurredAt    time.Time
	data          map[string]interface{}
	metadata      map[string]interface{}
}

// NewBaseEvent creates a new base event.
func NewBaseEvent(eventType, aggregateType string, aggregateID uuid.UUID, version int, data map[string]interface{}) *BaseEvent {
	return &BaseEvent{
		id:            uuid.New(),
		eventType:     eventType,
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		version:       version,
		occurredAt:    time.Now(),
		data:          data,
		metadata:      make(map[string]interface{}),
	}
}

// ID returns the unique identifier for this event.
func (e *BaseEvent) ID() uuid.UUID {
	return e.id
}

// Type returns the event type name.
func (e *BaseEvent) Type() string {
	return e.eventType
}

// AggregateID returns the ID of the aggregate that generated this event.
func (e *BaseEvent) AggregateID() uuid.UUID {
	return e.aggregateID
}

// AggregateType returns the type of the aggregate.
func (e *BaseEvent) AggregateType() string {
	return e.aggregateType
}

// Version returns the version of the aggregate when this event occurred.
func (e *BaseEvent) Version() int {
	return e.version
}

// OccurredAt returns when this event occurred.
func (e *BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// Data returns the event payload as a map.
func (e *BaseEvent) Data() map[string]interface{} {
	return e.data
}

// Metadata returns additional metadata about the event.
func (e *BaseEvent) Metadata() map[string]interface{} {
	return e.metadata
}

// SetMetadata sets metadata for the event.
func (e *BaseEvent) SetMetadata(key string, value interface{}) {
	if e.metadata == nil {
		e.metadata = make(map[string]interface{})
	}
	e.metadata[key] = value
}

// EventHandler defines the interface for handling events.
type EventHandler interface {
	// Handle processes an event
	Handle(ctx context.Context, event Event) error

	// CanHandle returns true if this handler can process the given event type
	CanHandle(eventType string) bool

	// HandlerName returns a unique name for this handler
	HandlerName() string
}

// EventBus defines the interface for publishing and subscribing to events.
type EventBus interface {
	// Publish publishes an event to all registered handlers
	Publish(ctx context.Context, event Event) error

	// Subscribe registers an event handler for specific event types
	Subscribe(handler EventHandler, eventTypes ...string) error

	// Unsubscribe removes an event handler
	Unsubscribe(handlerName string) error

	// Close closes the event bus and cleans up resources
	Close() error
}
