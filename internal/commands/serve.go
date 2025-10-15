package commands

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/database"
	"github.com/edsonmichaque/bazaruto/internal/logger"
	"github.com/edsonmichaque/bazaruto/internal/metrics"
	"github.com/edsonmichaque/bazaruto/internal/router"
	"github.com/edsonmichaque/bazaruto/internal/services"
	"github.com/edsonmichaque/bazaruto/internal/store"
	"github.com/edsonmichaque/bazaruto/internal/tracing"
	"go.uber.org/zap"
)

func newServeCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the Bazaruto API server",
		Long: `Start the HTTP server for the Bazaruto API.
The server will listen on the configured address and serve the REST API endpoints.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Initialize observability
			logger := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
			metrics := metrics.NewMetrics()
			var tracer *tracing.Tracer
			if cfg.Tracing.Enabled {
				tracer, err = tracing.NewTracer(cfg.Tracing.ServiceName, cfg.Tracing.Endpoint)
				if err != nil {
					logger.Error("Failed to initialize tracer", zap.Error(err))
					return fmt.Errorf("failed to initialize tracer: %w", err)
				}
			}

			// Connect to database
			db, err := database.Connect(cfg.DB.DSN, database.DBConfig{
				MaxConnections: cfg.DB.MaxConnections,
				MinConnections: cfg.DB.MinConnections,
				ConnectTimeout: cfg.DB.ConnectTimeout,
				AcquireTimeout: cfg.DB.AcquireTimeout,
				MaxLifetime:    cfg.DB.MaxLifetime,
				IdleTimeout:    cfg.DB.IdleTimeout,
			})
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer func() { _ = db.Close() }()

			// Create stores and services
			stores := store.NewStores(db.DB)
			_ = services.NewProductService(stores.Products)
			_ = services.NewQuoteService(stores.Quotes)
			_ = services.NewPolicyService(stores.Policies)
			_ = services.NewClaimService(stores.Claims, stores.Policies)

			// Create router
			r := chi.NewRouter()
			cleanup := router.Register(r, cfg, db, logger, metrics, tracer)
			defer cleanup()

			// Create HTTP server
			srv := &http.Server{
				Addr:              cfg.Server.Addr,
				Handler:           r,
				ReadHeaderTimeout: 10 * time.Second,
				ReadTimeout:       cfg.Server.ReadTimeout,
				WriteTimeout:      cfg.Server.WriteTimeout,
				IdleTimeout:       cfg.Server.IdleTimeout,
			}

			cmd.Printf("Starting server on %s\n", cfg.Server.Addr)

			// Use errgroup for graceful shutdown
			g, gctx := errgroup.WithContext(ctx)

			// Start HTTP server
			g.Go(func() error {
				err := srv.ListenAndServe()
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return err
			})

			// Start metrics server if enabled
			if cfg.MetricsEnabled {
				g.Go(func() error {
					return metrics.StartMetricsServer(gctx, ":9090")
				})
			}

			// Graceful shutdown
			g.Go(func() error {
				<-gctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				cmd.Println("Shutting down gracefully...")
				return srv.Shutdown(shutdownCtx)
			})

			// Wait for all goroutines
			if err := g.Wait(); err != nil {
				return fmt.Errorf("server error: %w", err)
			}

			cmd.Println("Server stopped gracefully.")
			return nil
		},
	}
}
