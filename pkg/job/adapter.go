package job

import (
	"fmt"
)

// AdapterFactory creates adapters based on type
type AdapterFactory interface {
	Create(adapterType AdapterType, config interface{}) (Adapter, error)
}

// DefaultAdapterFactory is the default implementation of AdapterFactory
type DefaultAdapterFactory struct{}

// NewDefaultAdapterFactory creates a new default adapter factory
func NewDefaultAdapterFactory() *DefaultAdapterFactory {
	return &DefaultAdapterFactory{}
}

// Create creates an adapter based on the type and configuration
// This is implemented in the manager to avoid import cycles
func (f *DefaultAdapterFactory) Create(adapterType AdapterType, config interface{}) (Adapter, error) {
	return nil, fmt.Errorf("use Manager.createAdapter instead to avoid import cycles")
}
