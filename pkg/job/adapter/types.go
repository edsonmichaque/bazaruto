package adapter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SerializedJob represents a job that can be persisted and transmitted
type SerializedJob struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Type       string          `json:"type" gorm:"not null"`
	Payload    json.RawMessage `json:"payload" gorm:"type:jsonb;not null"`
	Queue      string          `json:"queue" gorm:"not null"`
	Priority   int             `json:"priority" gorm:"not null;default:0"`
	MaxRetries int             `json:"max_retries" gorm:"not null;default:3"`
	Attempts   int             `json:"attempts" gorm:"not null;default:0"`
	RunAt      time.Time       `json:"run_at" gorm:"not null;default:now()"`
	LockedAt   *time.Time      `json:"locked_at" gorm:"index"`
	LockedBy   string          `json:"locked_by"`
	FailedAt   *time.Time      `json:"failed_at"`
	Error      string          `json:"error"`
	CreatedAt  time.Time       `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt  time.Time       `json:"updated_at" gorm:"not null;default:now()"`
}

// TableName returns the table name for GORM
func (SerializedJob) TableName() string {
	return "jobs"
}

// QueueStats provides statistics about a queue
type QueueStats struct {
	Queue      string `json:"queue"`
	Pending    int64  `json:"pending"`
	Processing int64  `json:"processing"`
	Failed     int64  `json:"failed"`
	Completed  int64  `json:"completed"`
}

// Adapter defines the interface for job queue backends
type Adapter interface {
	// Enqueue adds a job to the queue for immediate processing
	Enqueue(ctx context.Context, job *SerializedJob) error

	// EnqueueAt schedules a job to be processed at a specific time
	EnqueueAt(ctx context.Context, job *SerializedJob, at time.Time) error

	// Dequeue retrieves the next job from the queue
	Dequeue(ctx context.Context, queue string) (*SerializedJob, error)

	// Complete marks a job as successfully completed
	Complete(ctx context.Context, jobID uuid.UUID) error

	// Retry marks a job for retry with updated attempt count
	Retry(ctx context.Context, job *SerializedJob) error

	// Dead moves a job to the dead letter queue
	Dead(ctx context.Context, job *SerializedJob) error

	// Stats returns statistics for all queues
	Stats(ctx context.Context) (map[string]*QueueStats, error)

	// Clear removes all jobs from a specific queue
	Clear(ctx context.Context, queue string) error

	// Close closes the adapter and cleans up resources
	Close() error
}

// DeadJob represents a job that has permanently failed
type DeadJob struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primary_key"`
	Type       string          `json:"type" gorm:"not null"`
	Payload    json.RawMessage `json:"payload" gorm:"type:jsonb;not null"`
	Queue      string          `json:"queue" gorm:"not null"`
	Priority   int             `json:"priority" gorm:"not null"`
	MaxRetries int             `json:"max_retries" gorm:"not null"`
	Attempts   int             `json:"attempts" gorm:"not null"`
	RunAt      time.Time       `json:"run_at" gorm:"not null"`
	FailedAt   time.Time       `json:"failed_at" gorm:"not null"`
	Error      string          `json:"error" gorm:"not null"`
	CreatedAt  time.Time       `json:"created_at" gorm:"not null"`
	UpdatedAt  time.Time       `json:"updated_at" gorm:"not null"`
	DeadAt     time.Time       `json:"dead_at" gorm:"not null;default:now()"`
}

// TableName returns the table name for GORM
func (DeadJob) TableName() string {
	return "dead_jobs"
}

// RedisAdapterConfig contains configuration for Redis adapter
type RedisAdapterConfig struct {
	Addr     string
	Password string
	DB       int
	Prefix   string
}

// DatabaseAdapterConfig contains configuration for Database adapter
type DatabaseAdapterConfig struct {
	DSN string
}
