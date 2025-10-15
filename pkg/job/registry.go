package job

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Registry manages job type registration and serialization
type Registry struct {
	jobs map[string]func() Job
}

// NewRegistry creates a new job registry
func NewRegistry() *Registry {
	return &Registry{
		jobs: make(map[string]func() Job),
	}
}

// Register registers a job type with the registry
func (r *Registry) Register(name string, factory func() Job) {
	r.jobs[name] = factory
}

// RegisterJob registers a job type using reflection to get the type name
func (r *Registry) RegisterJob(job Job) {
	typeName := r.getTypeName(job)
	r.jobs[typeName] = func() Job {
		// Create a new instance of the same type
		return reflect.New(reflect.TypeOf(job).Elem()).Interface().(Job)
	}
}

// Create creates a new job instance by type name
func (r *Registry) Create(name string) (Job, error) {
	factory, exists := r.jobs[name]
	if !exists {
		return nil, fmt.Errorf("job type %s not registered", name)
	}
	return factory(), nil
}

// Serialize serializes a job to a SerializedJob
func (r *Registry) Serialize(job Job) (*SerializedJob, error) {
	typeName := r.getTypeName(job)

	payload, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job payload: %w", err)
	}

	return &SerializedJob{
		Type:       typeName,
		Payload:    payload,
		Queue:      job.Queue(),
		Priority:   job.Priority(),
		MaxRetries: job.MaxRetries(),
	}, nil
}

// Deserialize deserializes a SerializedJob back to a Job
func (r *Registry) Deserialize(sj *SerializedJob) (Job, error) {
	job, err := r.Create(sj.Type)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(sj.Payload, job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job payload: %w", err)
	}

	return job, nil
}

// getTypeName extracts the type name from a job instance
func (r *Registry) getTypeName(job Job) string {
	t := reflect.TypeOf(job)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Convert to snake_case for consistency
	name := t.Name()
	return strings.ToLower(name)
}

// ListRegistered returns a list of all registered job types
func (r *Registry) ListRegistered() []string {
	types := make([]string, 0, len(r.jobs))
	for name := range r.jobs {
		types = append(types, name)
	}
	return types
}
