package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/jobs"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/metrics"
	"github.com/edsonmichaque/bazaruto/internal/tracing"
	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/edsonmichaque/bazaruto/pkg/job/factory"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewWorkerCommand creates a new worker command
func NewWorkerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Start background job worker",
		Long:  "Start a background job worker to process jobs from the configured queues",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Initialize logger
			logger := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)

			// Initialize metrics
			metrics := metrics.NewMetrics()

			// Initialize tracer
			var tracer *tracing.Tracer
			if cfg.Tracing.Enabled {
				tracer, err = tracing.NewTracer(cfg.Tracing.ServiceName, cfg.Tracing.Endpoint)
				if err != nil {
					logger.Error("Failed to initialize tracer", zap.Error(err))
					return fmt.Errorf("failed to initialize tracer: %w", err)
				}
			}

			// Create job registry and register all job types
			registry := job.NewRegistry()
			registerJobTypes(registry)

			// Create adapter based on configuration
			adapter, err := createAdapter(cfg)
			if err != nil {
				return fmt.Errorf("failed to create job adapter: %w", err)
			}
			defer adapter.Close()

			// Create worker
			workerConfig := job.WorkerConfig{
				Queues:       cfg.Jobs.Queues,
				Concurrency:  cfg.Jobs.Concurrency,
				PollInterval: cfg.Jobs.PollInterval,
			}

			worker := job.NewWorker(adapter, registry, workerConfig, logger)

			// Add middleware
			worker.Use(job.RecoveryMiddleware())
			worker.Use(job.LoggingMiddleware(logger))
			worker.Use(job.MetricsMiddleware(metrics))
			if tracer != nil {
				worker.Use(job.TracingMiddleware(tracer))
			}
			worker.Use(job.TimeoutMiddleware(cfg.Jobs.Timeout))
			worker.Use(job.RetryMiddleware(cfg.Jobs.MaxRetries))

			// Create context with cancellation
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle shutdown signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigChan
				logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
				cancel()
				worker.Stop()
			}()

			// Start worker
			logger.Info("Starting job worker...")
			if err := worker.Start(ctx); err != nil {
				return fmt.Errorf("worker failed: %w", err)
			}

			logger.Info("Job worker stopped")
			return nil
		},
	}
}

// registerJobTypes registers all available job types with the registry
func registerJobTypes(registry *job.Registry) {
	// Email jobs
	registry.RegisterJob(&jobs.SendEmailJob{})
	registry.RegisterJob(&jobs.WelcomeEmailJob{})

	// PDF jobs
	registry.RegisterJob(&jobs.GenerateQuotePDFJob{})

	// Payment jobs
	registry.RegisterJob(&jobs.ProcessPaymentJob{})

	// Processing jobs
	registry.RegisterJob(&jobs.CalculatePremiumJob{})
	registry.RegisterJob(&jobs.FraudDetectionJob{})

	// Notification jobs
	registry.RegisterJob(&jobs.PushNotificationJob{})
}

// createAdapter creates the appropriate job adapter based on configuration
func createAdapter(cfg *config.Config) (job.Adapter, error) {
	switch cfg.Jobs.Adapter {
	case "memory":
		return factory.CreateAdapter(job.AdapterMemory, nil)
	case "redis":
		redisConfig := job.RedisAdapterConfig{
			Addr:     cfg.Jobs.Redis.Addr,
			Password: cfg.Jobs.Redis.Password,
			DB:       cfg.Jobs.Redis.DB,
		}
		return factory.CreateAdapter(job.AdapterRedis, redisConfig)
	case "database":
		databaseConfig := job.DatabaseAdapterConfig{
			DSN: cfg.Jobs.Database.DSN,
		}
		return factory.CreateAdapter(job.AdapterDatabase, databaseConfig)
	default:
		return nil, fmt.Errorf("unsupported job adapter: %s", cfg.Jobs.Adapter)
	}
}
