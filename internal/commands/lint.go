package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/edsonmichaque/bazaruto/internal/config"
)

func newLintCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "lint",
		Short: "Validate configuration files",
		Long: `Validate the configuration file or environment settings.
This command loads configuration using the same logic as the application
and verifies that all required fields are present and correctly typed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Validating configuration...")

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("configuration invalid: %w", err)
			}

			// Basic semantic checks
			if cfg.Server.Addr == "" {
				return fmt.Errorf("server.addr must not be empty")
			}
			if cfg.DB.DSN == "" {
				return fmt.Errorf("db.dsn must not be empty")
			}

			// Validate server configuration
			if cfg.Server.ReadTimeout <= 0 {
				return fmt.Errorf("server.read_timeout must be positive")
			}
			if cfg.Server.WriteTimeout <= 0 {
				return fmt.Errorf("server.write_timeout must be positive")
			}
			if cfg.Server.IdleTimeout <= 0 {
				return fmt.Errorf("server.idle_timeout must be positive")
			}

			// Validate database configuration
			if cfg.DB.MaxConnections <= 0 {
				return fmt.Errorf("db.max_connections must be positive")
			}
			if cfg.DB.MinConnections < 0 {
				return fmt.Errorf("db.min_connections must be non-negative")
			}
			if cfg.DB.MinConnections > cfg.DB.MaxConnections {
				return fmt.Errorf("db.min_connections cannot be greater than db.max_connections")
			}

			// Validate rate limiting configuration
			if cfg.Rate.Enabled {
				if cfg.Rate.PerMinute <= 0 {
					return fmt.Errorf("rate.per_minute must be positive when rate limiting is enabled")
				}
				if cfg.Rate.Burst <= 0 {
					return fmt.Errorf("rate.burst must be positive when rate limiting is enabled")
				}
				if cfg.Rate.Provider != "memory" && cfg.Rate.Provider != "redis" {
					return fmt.Errorf("rate.provider must be either 'memory' or 'redis'")
				}
				if cfg.Rate.Provider == "redis" && cfg.Redis.Addr == "" {
					return fmt.Errorf("redis.addr must be set when using Redis rate limiting")
				}
			}

			// Validate observability configuration
			if cfg.LogLevel != "" {
				validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
				valid := false
				for _, level := range validLevels {
					if cfg.LogLevel == level {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("log_level must be one of: %v", validLevels)
				}
			}

			if cfg.LogFormat != "" {
				if cfg.LogFormat != "json" && cfg.LogFormat != "text" {
					return fmt.Errorf("log_format must be either 'json' or 'text'")
				}
			}

			cmd.Println("Configuration loaded successfully.")
			cmd.Printf("Server address: %s\n", cfg.Server.Addr)
			cmd.Printf("Database DSN: %s\n", cfg.DB.DSN)
			cmd.Printf("Rate limiting enabled: %t\n", cfg.Rate.Enabled)
			if cfg.Rate.Enabled {
				cmd.Printf("Rate limiting provider: %s\n", cfg.Rate.Provider)
			}
			cmd.Printf("Logging level: %s\n", cfg.LogLevel)
			cmd.Printf("Logging format: %s\n", cfg.LogFormat)
			cmd.Printf("Metrics enabled: %t\n", cfg.MetricsEnabled)
			cmd.Printf("Tracing enabled: %t\n", cfg.Tracing.Enabled)
			cmd.Println("Configuration validation passed.")
			return nil
		},
	}
}
