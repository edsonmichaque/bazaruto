package job

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// Batch represents a collection of jobs that will be executed in parallel
type Batch struct {
	id         uuid.UUID
	dispatcher *Dispatcher
	jobs       []Job
	callbacks  []func()
	completed  int64
	total      int64
	mu         sync.RWMutex
	errors     []error
}

// NewBatch creates a new job batch
func NewBatch(dispatcher *Dispatcher, jobs ...Job) (*Batch, error) {
	if len(jobs) == 0 {
		return nil, fmt.Errorf("batch must contain at least one job")
	}

	return &Batch{
		id:         uuid.New(),
		dispatcher: dispatcher,
		jobs:       jobs,
		callbacks:  make([]func(), 0),
		total:      int64(len(jobs)),
		errors:     make([]error, 0),
	}, nil
}

// OnComplete adds a callback to be executed when the entire batch completes
func (b *Batch) OnComplete(callback func()) *Batch {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.callbacks = append(b.callbacks, callback)
	return b
}

// Execute starts the batch execution
func (b *Batch) Execute() error {
	return b.ExecuteWithContext(context.Background())
}

// ExecuteWithContext starts the batch execution with context
func (b *Batch) ExecuteWithContext(ctx context.Context) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Start all jobs in parallel
	for i, job := range b.jobs {
		wg.Add(1)
		go func(index int, j Job) {
			defer wg.Done()

			err := b.dispatcher.PerformWithContext(ctx, j)

			// Update completion count
			atomic.AddInt64(&b.completed, 1)

			// Store error if any
			if err != nil {
				mu.Lock()
				b.errors = append(b.errors, fmt.Errorf("job %d failed: %w", index, err))
				mu.Unlock()
			}
		}(i, job)
	}

	// Wait for all jobs to complete
	wg.Wait()

	// Execute completion callbacks
	b.mu.RLock()
	callbacks := make([]func(), len(b.callbacks))
	copy(callbacks, b.callbacks)
	b.mu.RUnlock()

	for _, callback := range callbacks {
		callback()
	}

	// Return error if any jobs failed
	if len(b.errors) > 0 {
		return fmt.Errorf("batch completed with %d errors: %v", len(b.errors), b.errors)
	}

	return nil
}

// Progress returns the completion progress of the batch
func (b *Batch) Progress() (completed, total int) {
	completed = int(atomic.LoadInt64(&b.completed))
	total = int(atomic.LoadInt64(&b.total))
	return completed, total
}

// IsComplete returns true if all jobs in the batch have completed
func (b *Batch) IsComplete() bool {
	return atomic.LoadInt64(&b.completed) >= atomic.LoadInt64(&b.total)
}

// Errors returns any errors that occurred during batch execution
func (b *Batch) Errors() []error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	errors := make([]error, len(b.errors))
	copy(errors, b.errors)
	return errors
}

// ID returns the unique identifier for this batch
func (b *Batch) ID() uuid.UUID {
	return b.id
}

// Length returns the number of jobs in the batch
func (b *Batch) Length() int {
	return len(b.jobs)
}

// Jobs returns a copy of the jobs in the batch
func (b *Batch) Jobs() []Job {
	jobs := make([]Job, len(b.jobs))
	copy(jobs, b.jobs)
	return jobs
}
