package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	app "github.com/edsonmichaque/bazaruto/internal/application"
	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/router"
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

			// Wire the entire application
			application, err := app.NewApplication(ctx, cfg)
			if err != nil {
				return fmt.Errorf("failed to wire application: %w", err)
			}
			defer func() { _ = application.Close() }()

			// Create router with wired application
			r := chi.NewRouter()
			cleanup := router.RegisterWithApp(r, application)
			defer cleanup()

			// Set the handler for the application server
			application.SetHandler(r)

			cmd.Printf("Starting server on %s\n", application.Config.Server.Addr)

			// Use errgroup for graceful shutdown
			g, gctx := errgroup.WithContext(ctx)

			// Start HTTP server using Application method
			g.Go(func() error {
				if err := application.StartServer(gctx); err != nil {
					return fmt.Errorf("failed to start server: %w", err)
				}
				return nil
			})

			// Start workers using Application method
			g.Go(func() error {
				if err := application.StartWorkers(gctx); err != nil {
					return fmt.Errorf("failed to start workers: %w", err)
				}
				return nil
			})

			// Start metrics server if enabled
			if application.Config.MetricsEnabled {
				g.Go(func() error {
					return application.Metrics.StartMetricsServer(gctx, ":9090")
				})
			}

			// Graceful shutdown
			g.Go(func() error {
				<-gctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				cmd.Println("Shutting down gracefully...")

				// Stop server and workers
				if err := application.StopServer(shutdownCtx); err != nil {
					cmd.Printf("Warning: failed to stop server: %v\n", err)
				}
				if err := application.StopWorkers(shutdownCtx); err != nil {
					cmd.Printf("Warning: failed to stop workers: %v\n", err)
				}

				return nil
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
