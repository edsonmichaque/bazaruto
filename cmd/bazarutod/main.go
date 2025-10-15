package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/edsonmichaque/bazaruto/internal/commands"
)

func main() {
	// Create context that will be cancelled on interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create and execute root command
	root := commands.New(ctx)
	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
