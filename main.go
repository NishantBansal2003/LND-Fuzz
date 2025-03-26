package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/NishantBansal2003/LND-Fuzz/config"
	"github.com/NishantBansal2003/LND-Fuzz/fuzz"
	"github.com/NishantBansal2003/LND-Fuzz/git"
)

func main() {
	// Create a cancellable context to manage the lifetime of operations.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize a structured logger that outputs logs in text format to
	// stdout.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Load configuration settings from environment variables.
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	// Ensure that the workspace is cleaned up after execution.
	defer config.CleanupWorkspace(logger)

	// Clone the project and storage repositories based on the loaded
	// configuration.
	if err := git.CloneRepositories(ctx, logger, cfg); err != nil {
		logger.Error("Repository cloning failed", "error", err)
		os.Exit(1)
	}

	// Execute fuzz testing on the specified packages.
	if err := fuzz.RunFuzzing(ctx, logger, cfg); err != nil {
		logger.Error("Fuzzing process failed", "error", err)
		os.Exit(1)
	}

	// Commit any changes in the corpus repository and push the commit to
	// the remote repository.
	if err := git.CommitAndPushResults(logger); err != nil {
		logger.Error("Failed to commit/push results", "error", err)
		os.Exit(1)
	}
}
