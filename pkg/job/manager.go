package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/metrics"
	"github.com/edsonmichaque/bazaruto/internal/tracing"
	"github.com/edsonmichaque/bazaruto/pkg/job/adapter"
	"go.uber.org/zap"
)

// Manager manages job processing with both in-process and separate worker support
type Manager struct {
	adapter    Adapter
	registry   *Registry
	worker     *Worker
	dispatcher *Dispatcher
	started    bool
	mu         sync.RWMutex
	logger     *logger.Logger
	metrics    *metrics.Metrics
	tracer     *tracing.Tracer
}

// ManagerConfig contains configuration for the job manager
type ManagerConfig struct {
	Adapter      string
	Queues       []string
	Concurrency  int
	PollInterval int64
	MaxRetries   int
	Timeout      int64

	// Redis-specific
	Redis RedisAdapterConfig

	// Database-specific
	Database DatabaseAdapterConfig
}

// NewManager creates a new job manager
func NewManager(config ManagerConfig, logger *logger.Logger, metrics *metrics.Metrics, tracer *tracing.Tracer) (*Manager, error) {
	// Create adapter based on configuration
	adapter, err := createAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create job adapter: %w", err)
	}

	// Create job registry
	registry := NewRegistry()

	// Create dispatcher
	dispatcher := NewDispatcher(adapter, registry)

	// Create worker
	workerConfig := WorkerConfig{
		Queues:       config.Queues,
		Concurrency:  config.Concurrency,
		PollInterval: time.Duration(config.PollInterval) * time.Second,
	}

	worker := NewWorker(adapter, registry, workerConfig, logger)

	// Add middleware
	worker.Use(RecoveryMiddleware())
	worker.Use(LoggingMiddleware(logger))
	worker.Use(MetricsMiddleware(metrics))
	if tracer != nil {
		worker.Use(TracingMiddleware(tracer))
	}
	worker.Use(TimeoutMiddleware(time.Duration(config.Timeout) * time.Second))
	worker.Use(RetryMiddleware(config.MaxRetries))

	return &Manager{
		adapter:    adapter,
		registry:   registry,
		worker:     worker,
		dispatcher: dispatcher,
		logger:     logger,
		metrics:    metrics,
		tracer:     tracer,
	}, nil
}

// Start begins processing jobs in the same process
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("manager is already started")
	}

	m.logger.Info("Starting in-process job workers",
		zap.Strings("queues", m.worker.queues),
		zap.Int("concurrency", m.worker.concurrency))

	// Start worker in a goroutine
	go func() {
		if err := m.worker.Start(ctx); err != nil {
			m.logger.Error("Worker failed", zap.Error(err))
		}
	}()

	m.started = true
	return nil
}

// Stop gracefully stops the manager
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return
	}

	m.logger.Info("Stopping job manager")
	m.worker.Stop()
	m.started = false
}

// Dispatcher returns the job dispatcher for enqueuing jobs
func (m *Manager) Dispatcher() *Dispatcher {
	return m.dispatcher
}

// Registry returns the job registry for registering job types
func (m *Manager) Registry() *Registry {
	return m.registry
}

// Stats returns statistics for all queues
func (m *Manager) Stats(ctx context.Context) (map[string]*QueueStats, error) {
	return m.adapter.Stats(ctx)
}

// Close closes the manager and its underlying adapter
func (m *Manager) Close() error {
	m.Stop()
	return m.adapter.Close()
}

// createAdapter creates the appropriate job adapter based on configuration
func createAdapter(config ManagerConfig) (Adapter, error) {
	switch config.Adapter {
	case "memory":
		return adapter.NewMemoryAdapter(), nil
	case "redis":
		redisConfig := adapter.RedisAdapterConfig{
			Addr:     config.Redis.Addr,
			Password: config.Redis.Password,
			DB:       config.Redis.DB,
			Prefix:   "bazaruto:jobs",
		}
		return adapter.NewRedisAdapter(redisConfig)
	case "database":
		databaseConfig := adapter.DatabaseAdapterConfig{
			DSN: config.Database.DSN,
		}
		return adapter.NewDatabaseAdapter(databaseConfig)
	default:
		return nil, fmt.Errorf("unsupported job adapter: %s", config.Adapter)
	}
}
