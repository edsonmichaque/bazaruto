package job

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/metrics"
	"github.com/edsonmichaque/bazaruto/internal/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// Handler represents a job execution handler
type Handler func(ctx context.Context, job Job) error

// Middleware represents middleware that can wrap job handlers
type Middleware func(next Handler) Handler

// RetryMiddleware implements exponential backoff retry logic
func RetryMiddleware(maxRetries int) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job Job) error {
			var lastErr error

			for attempt := 0; attempt <= maxRetries; attempt++ {
				if attempt > 0 {
					// Calculate backoff delay
					backoff := time.Duration(attempt*attempt) * time.Second

					// Add jitter to prevent thundering herd
					jitter := time.Duration(attempt) * time.Millisecond * 100
					backoff += jitter

					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(backoff):
					}
				}

				err := next(ctx, job)
				if err == nil {
					return nil
				}

				lastErr = err

				// Don't retry on context cancellation
				if ctx.Err() != nil {
					return ctx.Err()
				}
			}

			return fmt.Errorf("job failed after %d attempts: %w", maxRetries+1, lastErr)
		}
	}
}

// TimeoutMiddleware adds a timeout to job execution
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job Job) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- next(ctx, job)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// LoggingMiddleware adds structured logging to job execution
func LoggingMiddleware(logger *logger.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job Job) error {
			jobID := uuid.New().String()
			jobType := getJobType(job)

			log := logger.With(
				zap.String("job_id", jobID),
				zap.String("job_type", jobType),
				zap.String("queue", job.Queue()),
				zap.Int("priority", job.Priority()))

			start := time.Now()
			log.Info("Job started")

			err := next(ctx, job)

			duration := time.Since(start)

			if err != nil {
				log.Error("Job failed",
					zap.Error(err),
					zap.Duration("duration", duration))
			} else {
				log.Info("Job completed",
					zap.Duration("duration", duration))
			}

			return err
		}
	}
}

// MetricsMiddleware adds Prometheus metrics to job execution
func MetricsMiddleware(metrics *metrics.Metrics) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job Job) error {
			jobType := getJobType(job)
			queue := job.Queue()

			start := time.Now()

			// Increment jobs started counter
			metrics.JobStarted.WithLabelValues(queue, jobType).Inc()

			err := next(ctx, job)

			duration := time.Since(start)

			// Record job duration
			metrics.JobDuration.WithLabelValues(queue, jobType).Observe(duration.Seconds())

			if err != nil {
				// Increment jobs failed counter
				metrics.JobFailed.WithLabelValues(queue, jobType).Inc()
			} else {
				// Increment jobs completed counter
				metrics.JobCompleted.WithLabelValues(queue, jobType).Inc()
			}

			return err
		}
	}
}

// TracingMiddleware adds OpenTelemetry tracing to job execution
func TracingMiddleware(tracer *tracing.Tracer) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job Job) error {
			jobType := getJobType(job)
			spanName := fmt.Sprintf("job.%s", jobType)

			ctx, span := tracer.StartSpan(ctx, spanName)
			defer span.End()

			// Add job attributes to span
			span.SetAttributes(
				attribute.String("job.type", jobType),
				attribute.String("job.queue", job.Queue()),
				attribute.Int("job.priority", job.Priority()),
			)

			err := next(ctx, job)

			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "Job completed successfully")
			}

			return err
		}
	}
}

// RecoveryMiddleware recovers from panics and converts them to errors
func RecoveryMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job Job) (err error) {
			defer func() {
				if r := recover(); r != nil {
					// Get stack trace
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					stack := string(buf[:n])

					err = fmt.Errorf("job panicked: %v\nStack trace:\n%s", r, stack)
				}
			}()

			return next(ctx, job)
		}
	}
}

// getJobType extracts the job type name from a job instance
func getJobType(job Job) string {
	// Use reflection to get the type name
	jobType := fmt.Sprintf("%T", job)

	// Remove package path and pointer prefix
	if len(jobType) > 0 && jobType[0] == '*' {
		jobType = jobType[1:]
	}

	// Find the last dot
	if lastDot := len(jobType) - 1; lastDot >= 0 {
		for i := lastDot; i >= 0; i-- {
			if jobType[i] == '.' {
				jobType = jobType[i+1:]
				break
			}
		}
	}

	return jobType
}
