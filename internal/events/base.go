package events

import (
	"time"

	"github.com/edsonmichaque/bazaruto/pkg/event"
	"github.com/google/uuid"
)

// BusinessEvent represents a business domain event.
type BusinessEvent interface {
	event.Event
	GetEventType() string
	GetEntityID() uuid.UUID
	GetEntityType() string
	GetTimestamp() time.Time
	GetMetadata() map[string]interface{}
}

// BaseBusinessEvent provides common fields for business events.
type BaseBusinessEvent struct {
	EventID       uuid.UUID              `json:"event_id"`
	EventType     string                 `json:"event_type"`
	EntityID      uuid.UUID              `json:"entity_id"`
	EntityType    string                 `json:"entity_type"`
	Timestamp     time.Time              `json:"timestamp"`
	EventVersion  string                 `json:"version"`
	EventMetadata map[string]interface{} `json:"metadata"`
}

// GetEventID returns the event ID.
func (e *BaseBusinessEvent) GetEventID() uuid.UUID {
	return e.EventID
}

// GetEventType returns the event type.
func (e *BaseBusinessEvent) GetEventType() string {
	return e.EventType
}

// GetEntityID returns the entity ID.
func (e *BaseBusinessEvent) GetEntityID() uuid.UUID {
	return e.EntityID
}

// GetEntityType returns the entity type.
func (e *BaseBusinessEvent) GetEntityType() string {
	return e.EntityType
}

// GetTimestamp returns the event timestamp.
func (e *BaseBusinessEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetVersion returns the event version.
func (e *BaseBusinessEvent) GetVersion() string {
	return e.EventVersion
}

// GetMetadata returns the event metadata.
func (e *BaseBusinessEvent) GetMetadata() map[string]interface{} {
	return e.EventMetadata
}

// ID returns the unique identifier for this event (implements event.Event).
func (e *BaseBusinessEvent) ID() uuid.UUID {
	return e.EventID
}

// Type returns the event type name (implements event.Event).
func (e *BaseBusinessEvent) Type() string {
	return e.EventType
}

// AggregateID returns the ID of the aggregate that generated this event (implements event.Event).
func (e *BaseBusinessEvent) AggregateID() uuid.UUID {
	return e.EntityID
}

// AggregateType returns the type of the aggregate (implements event.Event).
func (e *BaseBusinessEvent) AggregateType() string {
	return e.EntityType
}

// Version returns the version of the aggregate when this event occurred (implements event.Event).
func (e *BaseBusinessEvent) Version() int {
	// For now, return 1 as default version
	return 1
}

// OccurredAt returns when this event occurred (implements event.Event).
func (e *BaseBusinessEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// Data returns the event payload as a map (implements event.Event).
func (e *BaseBusinessEvent) Data() map[string]interface{} {
	// Convert the business event to a generic data map
	data := make(map[string]interface{})

	// Add all the business event fields
	data["event_id"] = e.EventID.String()
	data["event_type"] = e.EventType
	data["entity_id"] = e.EntityID.String()
	data["entity_type"] = e.EntityType
	data["timestamp"] = e.Timestamp
	data["version"] = e.EventVersion

	// Add metadata
	if e.EventMetadata != nil {
		for k, v := range e.EventMetadata {
			data[k] = v
		}
	}

	return data
}

// Metadata returns additional metadata about the event (implements event.Event).
func (e *BaseBusinessEvent) Metadata() map[string]interface{} {
	return e.EventMetadata
}
