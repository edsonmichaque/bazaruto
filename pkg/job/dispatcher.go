package job

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Dispatcher provides a high-level API for dispatching jobs
type Dispatcher struct {
	adapter  Adapter
	registry *Registry
}

// NewDispatcher creates a new job dispatcher
func NewDispatcher(adapter Adapter, registry *Registry) *Dispatcher {
	return &Dispatcher{
		adapter:  adapter,
		registry: registry,
	}
}

// Perform dispatches a job for immediate execution
func (d *Dispatcher) Perform(job Job) error {
	return d.PerformWithContext(context.Background(), job)
}

// PerformWithContext dispatches a job for immediate execution with context
func (d *Dispatcher) PerformWithContext(ctx context.Context, job Job) error {
	serializedJob, err := d.registry.Serialize(job)
	if err != nil {
		return fmt.Errorf("failed to serialize job: %w", err)
	}

	// Set job ID
	serializedJob.ID = uuid.New()
	serializedJob.CreatedAt = time.Now()
	serializedJob.UpdatedAt = time.Now()

	return d.adapter.Enqueue(ctx, serializedJob)
}

// PerformLater dispatches a job for execution in the background
func (d *Dispatcher) PerformLater(job Job) error {
	return d.PerformLaterWithContext(context.Background(), job)
}

// PerformLaterWithContext dispatches a job for execution in the background with context
func (d *Dispatcher) PerformLaterWithContext(ctx context.Context, job Job) error {
	// Same as Perform for now, but could be optimized for background processing
	return d.PerformWithContext(ctx, job)
}

// PerformAt dispatches a job to be executed at a specific time
func (d *Dispatcher) PerformAt(job Job, at time.Time) error {
	return d.PerformAtWithContext(context.Background(), job, at)
}

// PerformAtWithContext dispatches a job to be executed at a specific time with context
func (d *Dispatcher) PerformAtWithContext(ctx context.Context, job Job, at time.Time) error {
	serializedJob, err := d.registry.Serialize(job)
	if err != nil {
		return fmt.Errorf("failed to serialize job: %w", err)
	}

	// Set job ID and timing
	serializedJob.ID = uuid.New()
	serializedJob.RunAt = at
	serializedJob.CreatedAt = time.Now()
	serializedJob.UpdatedAt = time.Now()

	return d.adapter.EnqueueAt(ctx, serializedJob, at)
}

// PerformIn dispatches a job to be executed after a specific duration
func (d *Dispatcher) PerformIn(job Job, duration time.Duration) error {
	return d.PerformInWithContext(context.Background(), job, duration)
}

// PerformInWithContext dispatches a job to be executed after a specific duration with context
func (d *Dispatcher) PerformInWithContext(ctx context.Context, job Job, duration time.Duration) error {
	at := time.Now().Add(duration)
	return d.PerformAtWithContext(ctx, job, at)
}

// Chain creates a chain of jobs that will be executed sequentially
func (d *Dispatcher) Chain(jobs ...Job) (*Chain, error) {
	return NewChain(d, jobs...)
}

// Batch creates a batch of jobs that will be executed in parallel
func (d *Dispatcher) Batch(jobs ...Job) (*Batch, error) {
	return NewBatch(d, jobs...)
}

// Stats returns statistics for all queues
func (d *Dispatcher) Stats(ctx context.Context) (map[string]*QueueStats, error) {
	return d.adapter.Stats(ctx)
}

// Clear removes all jobs from a specific queue
func (d *Dispatcher) Clear(ctx context.Context, queue string) error {
	return d.adapter.Clear(ctx, queue)
}

// Close closes the dispatcher and its underlying adapter
func (d *Dispatcher) Close() error {
	return d.adapter.Close()
}
