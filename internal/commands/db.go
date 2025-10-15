package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/edsonmichaque/bazaruto/internal/config"
	"github.com/edsonmichaque/bazaruto/internal/database"
)

func newDBCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management operations",
		Long: `Database management commands for Bazaruto.
These commands handle database migrations, resets, and information.`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "migrate",
			Short: "Run database migrations",
			Long: `Run database migrations to create or update the database schema.
This command will create all necessary tables and indexes.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

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

				cmd.Println("Running database migrations...")
				if err := database.RunMigrations(db.DB); err != nil {
					return fmt.Errorf("failed to run migrations: %w", err)
				}

				cmd.Println("Migrations completed successfully.")
				return nil
			},
		},
		&cobra.Command{
			Use:   "reset",
			Short: "Drop and recreate database schema",
			Long: `Drop all tables and recreate the database schema.
WARNING: This will delete all data in the database!`,
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

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

				cmd.Println("Dropping all tables...")
				if err := database.DropAll(db.DB); err != nil {
					return fmt.Errorf("failed to drop tables: %w", err)
				}

				cmd.Println("Running database migrations...")
				if err := database.RunMigrations(db.DB); err != nil {
					return fmt.Errorf("failed to run migrations: %w", err)
				}

				cmd.Println("Database reset completed successfully.")
				return nil
			},
		},
		&cobra.Command{
			Use:   "info",
			Short: "Show database connection information",
			Long: `Display database connection information and statistics.
This includes connection details and current database status.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

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

				cmd.Printf("Database DSN: %s\n", cfg.DB.DSN)
				cmd.Printf("Max Connections: %d\n", cfg.DB.MaxConnections)
				cmd.Printf("Min Connections: %d\n", cfg.DB.MinConnections)
				cmd.Printf("Connect Timeout: %s\n", cfg.DB.ConnectTimeout)
				cmd.Printf("Acquire Timeout: %s\n", cfg.DB.AcquireTimeout)
				cmd.Printf("Max Lifetime: %s\n", cfg.DB.MaxLifetime)
				cmd.Printf("Idle Timeout: %s\n", cfg.DB.IdleTimeout)

				// Get database statistics
				stats, err := db.Stats()
				if err != nil {
					cmd.Printf("Failed to get database statistics: %v\n", err)
					return nil
				}

				cmd.Println("\nDatabase Statistics:")
				for key, value := range stats {
					cmd.Printf("  %s: %v\n", key, value)
				}

				return nil
			},
		},
	)

	return cmd
}
