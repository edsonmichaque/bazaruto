package commands

import (
	"encoding/json"
	"fmt"

	"github.com/edsonmichaque/bazaruto/internal/version"
	"github.com/spf13/cobra"
)

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display version information including build details and git commit hash",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Get()

			// Check if JSON output is requested
			if cmd.Flags().Changed("json") {
				jsonOutput, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal version info: %w", err)
				}
				cmd.Println(string(jsonOutput))
				return nil
			}

			// Default text output
			cmd.Println(version.String())
			return nil
		},
	}

	// Add JSON output flag
	cmd.Flags().Bool("json", false, "Output version information in JSON format")

	return cmd
}
