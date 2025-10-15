package job

import "time"

// Default configuration constants
const (
	// DefaultQueue is the default queue name for jobs
	DefaultQueue = "default"

	// DefaultMaxRetries is the default maximum number of retry attempts
	DefaultMaxRetries = 3

	// DefaultPriority is the default priority for jobs (0 = normal priority)
	DefaultPriority = 0

	// DefaultTimeout is the default timeout for job execution
	DefaultTimeout = 5 * time.Minute

	// DefaultMaxEvents is the default maximum number of events to store in memory
	DefaultMaxEvents = 10000

	// DefaultBatchSize is the default batch size for processing jobs
	DefaultBatchSize = 100

	// DefaultWorkerCount is the default number of worker goroutines
	DefaultWorkerCount = 5

	// DefaultPollInterval is the default interval for polling for new jobs
	DefaultPollInterval = 1 * time.Second

	// DefaultLockTimeout is the default timeout for job locks
	DefaultLockTimeout = 5 * time.Minute

	// DefaultRetryBackoff is the default backoff duration between retries
	DefaultRetryBackoff = 1 * time.Second

	// DefaultMaxBackoff is the maximum backoff duration
	DefaultMaxBackoff = 5 * time.Minute

	// DefaultBackoffMultiplier is the multiplier for exponential backoff
	DefaultBackoffMultiplier = 2.0
)

// Queue names
const (
	QueueDefault       = "default"
	QueueMailers       = "mailers"
	QueueProcessing    = "processing"
	QueueHeavy         = "heavy"
	QueuePayments      = "payments"
	QueueNotifications = "notifications"
	QueueClaims        = "claims"
	QueueFraud         = "fraud"
	QueueCompliance    = "compliance"
	QueueReports       = "reports"
)

// Job priorities
const (
	PriorityLow      = -10
	PriorityNormal   = 0
	PriorityHigh     = 10
	PriorityCritical = 20
)

// Job statuses
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusRetrying   = "retrying"
	StatusDead       = "dead"
	StatusCancelled  = "cancelled"
)

// Adapter types
const (
	AdapterTypeMemory   = "memory"
	AdapterTypeRedis    = "redis"
	AdapterTypeDatabase = "database"
)

// Error messages
const (
	ErrJobNotFound        = "job not found"
	ErrJobAlreadyExists   = "job already exists"
	ErrJobLocked          = "job is locked"
	ErrJobExpired         = "job has expired"
	ErrJobCancelled       = "job was cancelled"
	ErrMaxRetriesExceeded = "maximum retries exceeded"
	ErrInvalidJobType     = "invalid job type"
	ErrInvalidQueue       = "invalid queue name"
	ErrInvalidPriority    = "invalid priority"
	ErrAdapterNotFound    = "adapter not found"
	ErrAdapterClosed      = "adapter is closed"
)

// Redis configuration defaults
const (
	DefaultRedisAddr     = "localhost:6379"
	DefaultRedisPassword = ""
	DefaultRedisDB       = 0
	DefaultRedisPrefix   = "bazaruto:jobs:"
	DefaultRedisPoolSize = 10
)

// Database configuration defaults
const (
	DefaultDatabaseMaxOpenConns    = 25
	DefaultDatabaseMaxIdleConns    = 5
	DefaultDatabaseConnMaxLifetime = 5 * time.Minute
	DefaultDatabaseConnMaxIdleTime = 1 * time.Minute
)

// Memory adapter configuration defaults
const (
	DefaultMemoryMaxEvents       = 10000
	DefaultMemoryCleanupInterval = 1 * time.Hour
)

// Worker configuration defaults
const (
	DefaultWorkerMaxConcurrency    = 10
	DefaultWorkerShutdownTimeout   = 30 * time.Second
	DefaultWorkerHeartbeatInterval = 30 * time.Second
)

// Job processing limits
const (
	MaxJobPayloadSize = 1024 * 1024 // 1MB
	MaxJobRetries     = 10
	MaxJobTimeout     = 1 * time.Hour
	MinJobTimeout     = 1 * time.Second
)

// Event processing limits
const (
	MaxEventBatchSize = 1000
	MaxEventAge       = 30 * 24 * time.Hour // 30 days
	MaxEventSize      = 64 * 1024           // 64KB
)

// Health check intervals
const (
	HealthCheckInterval = 30 * time.Second
	HealthCheckTimeout  = 5 * time.Second
)

// Metrics collection intervals
const (
	MetricsCollectionInterval = 1 * time.Minute
	MetricsRetentionPeriod    = 7 * 24 * time.Hour // 7 days
)

