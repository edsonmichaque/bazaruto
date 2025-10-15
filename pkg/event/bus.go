package event

import (
	"context"
	"fmt"
	"sync"
)

// Bus implements an in-memory event bus with async publishing.
type Bus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// NewBus creates a new in-memory event bus.
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish publishes an event to all registered handlers asynchronously.
func (b *Bus) Publish(ctx context.Context, event Event) error {
	b.mutex.RLock()
	handlers, exists := b.handlers[event.Type()]
	b.mutex.RUnlock()

	if !exists {
		// No handlers registered for this event type - this is not an error
		return nil
	}

	// Publish asynchronously - don't wait for handlers to complete
	go func() {
		// Process handlers concurrently
		var wg sync.WaitGroup

		for _, handler := range handlers {
			wg.Add(1)
			go func(h EventHandler) {
				defer wg.Done()
				// Create a new context for each handler to avoid cancellation
				handlerCtx := context.Background()
				if err := h.Handle(handlerCtx, event); err != nil {
					// Log error but don't fail the entire operation
					// In a real implementation, you might want to use a logger here
					_ = err // Suppress unused variable warning
				}
			}(handler)
		}

		wg.Wait()
	}()

	return nil
}

// Subscribe registers an event handler for specific event types.
func (b *Bus) Subscribe(handler EventHandler, eventTypes ...string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, eventType := range eventTypes {
		if b.handlers[eventType] == nil {
			b.handlers[eventType] = make([]EventHandler, 0)
		}

		// Check if handler is already registered for this event type
		for _, existingHandler := range b.handlers[eventType] {
			if existingHandler.HandlerName() == handler.HandlerName() {
				return fmt.Errorf("handler %s is already registered for event type %s",
					handler.HandlerName(), eventType)
			}
		}

		b.handlers[eventType] = append(b.handlers[eventType], handler)
	}

	return nil
}

// Unsubscribe removes an event handler.
func (b *Bus) Unsubscribe(handlerName string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	removed := false
	for eventType, handlers := range b.handlers {
		newHandlers := make([]EventHandler, 0, len(handlers))
		for _, handler := range handlers {
			if handler.HandlerName() != handlerName {
				newHandlers = append(newHandlers, handler)
			} else {
				removed = true
			}
		}
		b.handlers[eventType] = newHandlers
	}

	if !removed {
		return fmt.Errorf("handler %s not found", handlerName)
	}

	return nil
}

// Close closes the event bus and cleans up resources.
func (b *Bus) Close() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.handlers = make(map[string][]EventHandler)
	return nil
}

// GetHandlerCount returns the number of handlers for a specific event type.
func (b *Bus) GetHandlerCount(eventType string) int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	handlers, exists := b.handlers[eventType]
	if !exists {
		return 0
	}
	return len(handlers)
}

// GetRegisteredEventTypes returns all registered event types.
func (b *Bus) GetRegisteredEventTypes() []string {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	eventTypes := make([]string, 0, len(b.handlers))
	for eventType := range b.handlers {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}
