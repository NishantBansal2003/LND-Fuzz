package app

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
		os.Exit(1)
	}

	// Execute fuzz testing on the specified packages.
	if err := fuzz.RunFuzzing(ctx, logger, cfg); err != nil {
		logger.Error("Fuzzing process failed", "error", err)
		os.Exit(1)
	}

}
