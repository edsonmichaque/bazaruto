package job

import (
	"context"
	"fmt"
)

// Chain represents a sequence of jobs that will be executed one after another
type Chain struct {
	dispatcher *Dispatcher
	jobs       []Job
	callbacks  []func()
}

// NewChain creates a new job chain
func NewChain(dispatcher *Dispatcher, jobs ...Job) (*Chain, error) {
	if len(jobs) == 0 {
		return nil, fmt.Errorf("chain must contain at least one job")
	}

	return &Chain{
		dispatcher: dispatcher,
		jobs:       jobs,
		callbacks:  make([]func(), 0),
	}, nil
}

// Then adds another job to the chain
func (c *Chain) Then(job Job) *Chain {
	c.jobs = append(c.jobs, job)
	return c
}

// OnComplete adds a callback to be executed when the entire chain completes
func (c *Chain) OnComplete(callback func()) *Chain {
	c.callbacks = append(c.callbacks, callback)
	return c
}

// Execute starts the chain execution
func (c *Chain) Execute() error {
	return c.ExecuteWithContext(context.Background())
}

// ExecuteWithContext starts the chain execution with context
func (c *Chain) ExecuteWithContext(ctx context.Context) error {
	// For now, we'll execute jobs sequentially
	// In a more sophisticated implementation, we could use a special chain job
	// that coordinates the execution of the entire chain

	for i, job := range c.jobs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := c.dispatcher.PerformWithContext(ctx, job); err != nil {
				return fmt.Errorf("chain failed at job %d: %w", i, err)
			}
		}
	}

	// Execute completion callbacks
	for _, callback := range c.callbacks {
		callback()
	}

	return nil
}

// Length returns the number of jobs in the chain
func (c *Chain) Length() int {
	return len(c.jobs)
}

// Jobs returns a copy of the jobs in the chain
func (c *Chain) Jobs() []Job {
	jobs := make([]Job, len(c.jobs))
	copy(jobs, c.jobs)
	return jobs
}
