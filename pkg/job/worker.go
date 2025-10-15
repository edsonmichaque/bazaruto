package job

import (
	"context"
	"sync"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/logger"
	"go.uber.org/zap"
)

// Worker processes jobs from queues using a pool of goroutines
type Worker struct {
	adapter     Adapter
	registry    *Registry
	queues      []string
	concurrency int
	middleware  []Middleware
	shutdown    chan struct{}
	waitGroup   sync.WaitGroup
	logger      *logger.Logger
}

// WorkerConfig contains configuration for the worker
type WorkerConfig struct {
	Queues       []string
	Concurrency  int
	PollInterval time.Duration
}

// NewWorker creates a new worker instance
func NewWorker(adapter Adapter, registry *Registry, config WorkerConfig, logger *logger.Logger) *Worker {
	if config.Concurrency <= 0 {
		config.Concurrency = 1
	}

	if config.PollInterval <= 0 {
		config.PollInterval = time.Second
	}

	return &Worker{
		adapter:     adapter,
		registry:    registry,
		queues:      config.Queues,
		concurrency: config.Concurrency,
		middleware:  make([]Middleware, 0),
		shutdown:    make(chan struct{}),
		logger:      logger,
	}
}

// Use adds middleware to the worker
func (w *Worker) Use(middleware Middleware) {
	w.middleware = append(w.middleware, middleware)
}

// Start begins processing jobs from the configured queues
func (w *Worker) Start(ctx context.Context) error {
	w.logger.Info("Starting job worker",
		zap.Strings("queues", w.queues),
		zap.Int("concurrency", w.concurrency))

	// Start worker goroutines
	for i := 0; i < w.concurrency; i++ {
		w.waitGroup.Add(1)
		go w.workerLoop(ctx, i)
	}

	// Wait for shutdown signal
	<-w.shutdown
	w.logger.Info("Worker shutdown requested")

	// Wait for all workers to finish
	w.waitGroup.Wait()
	w.logger.Info("All workers stopped")

	return nil
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	close(w.shutdown)
}

// workerLoop is the main loop for each worker goroutine
func (w *Worker) workerLoop(ctx context.Context, workerID int) {
	defer w.waitGroup.Done()

	log := &logger.Logger{Logger: w.logger.With(zap.Int("worker_id", workerID))}
	log.Info("Worker started")

	ticker := time.NewTicker(time.Second) // Poll interval
	defer ticker.Stop()

	for {
		select {
		case <-w.shutdown:
			log.Info("Worker stopping")
			return
		case <-ctx.Done():
			log.Info("Worker context cancelled", zap.Error(ctx.Err()))
			return
		case <-ticker.C:
			w.processJobs(ctx, workerID, log)
		}
	}
}

// processJobs attempts to process jobs from all configured queues
func (w *Worker) processJobs(ctx context.Context, workerID int, log *logger.Logger) {
	for _, queue := range w.queues {
		select {
		case <-w.shutdown:
			return
		case <-ctx.Done():
			return
		default:
			if err := w.processQueue(ctx, queue, workerID, log); err != nil {
				// Log error but continue processing other queues
				log.Error("Error processing queue", zap.Error(err), zap.String("queue", queue))
			}
		}
	}
}

// processQueue processes a single job from the specified queue
func (w *Worker) processQueue(ctx context.Context, queue string, workerID int, log *logger.Logger) error {
	// Try to dequeue a job
	serializedJob, err := w.adapter.Dequeue(ctx, queue)
	if err != nil {
		// No jobs available, this is normal
		return nil
	}

	// Deserialize the job
	job, err := w.registry.Deserialize(serializedJob)
	if err != nil {
		log.Error("Failed to deserialize job", zap.Error(err), zap.String("job_id", serializedJob.ID.String()))
		// Move to dead letter queue
		_ = w.adapter.Dead(ctx, serializedJob)
		return err
	}

	// Create execution context with job ID
	jobCtx := context.WithValue(ctx, "job_id", serializedJob.ID)
	jobCtx = context.WithValue(jobCtx, "worker_id", workerID)

	// Build middleware chain
	handler := w.buildHandler(job)

	// Execute the job
	err = handler(jobCtx, job)

	// Handle result
	if err != nil {
		log.Error("Job failed", zap.Error(err), zap.String("job_id", serializedJob.ID.String()))

		// Check if we should retry
		if serializedJob.Attempts < serializedJob.MaxRetries {
			if retryErr := w.adapter.Retry(ctx, serializedJob); retryErr != nil {
				log.Error("Failed to retry job", zap.Error(retryErr), zap.String("job_id", serializedJob.ID.String()))
			} else {
				log.Info("Job scheduled for retry", zap.String("job_id", serializedJob.ID.String()), zap.Int("attempt", serializedJob.Attempts+1))
			}
		} else {
			// Max retries exceeded, move to dead letter queue
			if deadErr := w.adapter.Dead(ctx, serializedJob); deadErr != nil {
				log.Error("Failed to move job to dead letter queue", zap.Error(deadErr), zap.String("job_id", serializedJob.ID.String()))
			} else {
				log.Info("Job moved to dead letter queue", zap.String("job_id", serializedJob.ID.String()))
			}
		}
	} else {
		// Job completed successfully
		if completeErr := w.adapter.Complete(ctx, serializedJob.ID); completeErr != nil {
			log.Error("Failed to mark job as complete", zap.Error(completeErr), zap.String("job_id", serializedJob.ID.String()))
		} else {
			log.Info("Job completed successfully", zap.String("job_id", serializedJob.ID.String()))
		}
	}

	return nil
}

// buildHandler builds the middleware chain for job execution
func (w *Worker) buildHandler(job Job) Handler {
	handler := func(ctx context.Context, j Job) error {
		return j.Perform(ctx)
	}

	// Apply middleware in reverse order (last added is outermost)
	for i := len(w.middleware) - 1; i >= 0; i-- {
		handler = w.middleware[i](handler)
	}

	return handler
}
