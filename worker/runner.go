package worker

import (
	"context"
	"log/slog"
	"os"

	"github.com/NishantBansal2003/LND-Fuzz/config"
	"github.com/NishantBansal2003/LND-Fuzz/fuzz"
	"github.com/NishantBansal2003/LND-Fuzz/git"
)

// Main handles the cloning of repositories and the execution of fuzz testing.
// It ensures that any errors encountered during these processes are logged and
// that the workspace is cleaned up appropriately before the program exits.
func Main(ctx context.Context, logger *slog.Logger, cfg *config.Config,
	doneChan chan struct{}) {

	// Close the channel to indicate that the fuzzing cycle has completed,
	// so that the scheduler can perform cleanup.
	defer close(doneChan)

	// Clone the project and storage repositories based on the loaded
	// configuration.
	if err := git.CloneRepositories(ctx, logger, cfg); err != nil {
		logger.Error("Repository cloning failed", "error", err)

		// Perform workspace cleanup before exiting due to the cloning
		// error.
		config.PerformCleanup(logger, cfg)
		os.Exit(1)
	}

	// Execute fuzz testing on the specified packages.
	if err := fuzz.RunFuzzing(ctx, logger, cfg); err != nil {
		logger.Error("Fuzzing process failed", "error", err)

		// Perform workspace cleanup before exiting due to the fuzzing
		// error.
		config.PerformCleanup(logger, cfg)
		os.Exit(1)
	}
}
