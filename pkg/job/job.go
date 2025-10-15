package job

import (
	"context"
	"encoding/json"
	"time"

	"github.com/edsonmichaque/bazaruto/pkg/job/adapter"
	"github.com/google/uuid"
)

// Job represents a unit of work to be executed asynchronously
type Job interface {
	Perform(ctx context.Context) error
	Queue() string               // Queue name (default, mailers, processing, etc.)
	MaxRetries() int             // Max retry attempts
	RetryBackoff() time.Duration // Base backoff duration
	Priority() int               // Higher = processed first
}

// SerializedJob represents a job that can be persisted and transmitted
type SerializedJob = adapter.SerializedJob

// QueueStats provides statistics about a queue
type QueueStats = adapter.QueueStats

// Adapter defines the interface for job queue backends
type Adapter = adapter.Adapter

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

// AdapterType represents the type of adapter
type AdapterType string

const (
	AdapterMemory   AdapterType = "memory"
	AdapterRedis    AdapterType = "redis"
	AdapterDatabase AdapterType = "database"
)

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

// Config for job execution
type Config struct {
	JobID      string
	Queue      string
	MaxRetries int
	Attempts   int
	Priority   int
	RunAt      time.Time
	Timeout    time.Duration
}
