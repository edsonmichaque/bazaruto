package commands

import (
	"context"

	"github.com/spf13/cobra"
)

// New creates the root command for bazarutod.
func New(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bazarutod",
		Short: "Bazaruto backend service",
		Long: `Bazaruto is a production-ready Go backend service for an insurance product marketplace.
It provides RESTful APIs for managing products, quotes, policies, and claims.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Register subcommands explicitly
	cmd.AddCommand(
		newServeCmd(ctx),
		newDBCmd(ctx),
		newAdminCmd(ctx),
		newCacheCmd(ctx),
		newLintCmd(ctx),
		newVersionCmd(),
		NewWorkerCommand(),
	)

	return cmd
}
