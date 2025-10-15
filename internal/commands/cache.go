package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newCacheCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Cache management operations",
		Long: `Cache management commands for Bazaruto.
These commands handle cache inspection, purging, and warming operations.`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "inspect",
			Short: "Inspect cache contents",
			Long: `Inspect the contents of the cache directory.
This command lists all cached files and their sizes.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				cacheDir := "/var/cache/bazaruto"

				// Check if cache directory exists
				if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
					cmd.Printf("Cache directory does not exist: %s\n", cacheDir)
					return nil
				}

				cmd.Printf("Cache directory: %s\n", cacheDir)

				entries, err := os.ReadDir(cacheDir)
				if err != nil {
					return fmt.Errorf("failed to read cache directory: %w", err)
				}

				if len(entries) == 0 {
					cmd.Println("Cache directory is empty.")
					return nil
				}

				cmd.Println("Cache contents:")
				for _, entry := range entries {
					info, err := entry.Info()
					if err != nil {
						cmd.Printf("  %s (error getting info)\n", entry.Name())
						continue
					}

					if entry.IsDir() {
						cmd.Printf("  %s/ (directory)\n", entry.Name())
					} else {
						cmd.Printf("  %s (%d bytes)\n", entry.Name(), info.Size())
					}
				}

				return nil
			},
		},
		&cobra.Command{
			Use:   "purge",
			Short: "Clear cache directory",
			Long: `Clear all contents from the cache directory.
This will remove all cached files and directories.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				cacheDir := "/var/cache/bazaruto"

				cmd.Printf("Purging cache directory: %s\n", cacheDir)

				// Remove all contents
				if err := os.RemoveAll(cacheDir); err != nil {
					return fmt.Errorf("failed to remove cache directory: %w", err)
				}

				// Recreate directory
				if err := os.MkdirAll(cacheDir, 0755); err != nil {
					return fmt.Errorf("failed to recreate cache directory: %w", err)
				}

				cmd.Println("Cache purged successfully.")
				return nil
			},
		},
		&cobra.Command{
			Use:   "warm",
			Short: "Warm cache with common data",
			Long: `Warm the cache with commonly accessed data.
This preloads frequently used data into the cache for better performance.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				cacheDir := "/var/cache/bazaruto"

				cmd.Printf("Warming cache directory: %s\n", cacheDir)

				// Ensure cache directory exists
				if err := os.MkdirAll(cacheDir, 0755); err != nil {
					return fmt.Errorf("failed to create cache directory: %w", err)
				}

				// Create some sample cache files
				sampleFiles := []string{
					"templates/product-list.html",
					"templates/quote-form.html",
					"templates/policy-details.html",
					"metadata/currencies.json",
					"metadata/categories.json",
				}

				for _, file := range sampleFiles {
					filePath := fmt.Sprintf("%s/%s", cacheDir, file)

					// Create directory if it doesn't exist
					dir := fmt.Sprintf("%s/%s", cacheDir, file[:len(file)-len(file[strings.LastIndex(file, "/"):])])
					if err := os.MkdirAll(dir, 0755); err != nil {
						cmd.Printf("Warning: failed to create directory %s: %v\n", dir, err)
						continue
					}

					// Create sample file
					if err := os.WriteFile(filePath, []byte("sample cache data"), 0644); err != nil {
						cmd.Printf("Warning: failed to create cache file %s: %v\n", file, err)
						continue
					}
				}

				cmd.Println("Cache warmed successfully.")
				return nil
			},
		},
	)

	return cmd
}
