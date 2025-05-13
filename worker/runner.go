package worker

import (
	"context"
	"log/slog"
	"os"

	"github.com/NishantBansal2003/LND-Fuzz/config"
	"github.com/NishantBansal2003/LND-Fuzz/fuzz"
	"github.com/NishantBansal2003/LND-Fuzz/git"
)

func Main(ctx context.Context, logger *slog.Logger, cfg *config.Config) {
	// Clone the project and storage repositories based on the loaded
	// configuration.
	if err := git.CloneRepositories(ctx, logger, cfg); err != nil {
		logger.Error("Repository cloning failed", "error", err)

		// Perform workspace cleanup before exiting due to the cloning
		// error.
		config.CleanupWorkspace(logger)
		os.Exit(1)
	}

	// Execute fuzz testing on the specified packages.
	if err := fuzz.RunFuzzing(ctx, logger, cfg); err != nil {
		logger.Error("Fuzzing process failed", "error", err)

		// Perform workspace cleanup before exiting due to the fuzzing error.
		config.CleanupWorkspace(logger)
		os.Exit(1)
	}
}
