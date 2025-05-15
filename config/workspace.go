package config

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
)

const HelpText = `Usage: go run main.go [command]

Commands:
  help      Show this help message and exit.

Environment Variables:
  FUZZ_NUM_PROCESSES
          Specifies the number of fuzzing processes to run concurrently.
          Default: Maximum number of CPU cores available on the machine.

  PROJECT_SRC_PATH    (Required)
          The Git repository URL of the project to be fuzzed.
          Formats:
            - Private: https://oauth2:PAT@github.com/OWNER/REPO.git
            - Public:  https://github.com/OWNER/REPO.git

  GIT_STORAGE_REPO    (Required)
          The Git repository where the input corpus is stored.
          Format: https://oauth2:PAT@github.com/OWNER/STORAGEREPO.git

  FUZZ_TIME
          Duration (in seconds) for which the fuzzing engine should run.
          Default: 120 seconds.

  FUZZ_PKG   (Required)
          The specific Go package within the repository to be fuzzed.

  FUZZ_RESULTS_PATH
          Path to store fuzzing results, relative to the current working
	  directory
          Default: Project root directory

Usage Example:
  Set the necessary environment variables, then start fuzzing:
      go run main.go

For more information, please refer to the project documentation.`

// cleanupWorkspace removes the "out" directory to clean up the workspace.
// It uses a context with a timeout (DefaultCleanupTimeout) to limit the
// duration of the cleanup operation. Any errors encountered during the removal
// are logged.
func cleanupWorkspace(logger *slog.Logger) {
	_, cancel := context.WithTimeout(
		context.Background(), DefaultCleanupTimeout,
	)
	defer cancel()

	if err := os.RemoveAll("out"); err != nil {
		logger.Error("Cleanup failed", "error", err)
	}
}

// PerformCleanup handles post-cycle cleanup of the workspace and stores the
// results. If storing of the corpus fails, it logs the error and terminates
// the program.
func PerformCleanup(logger *slog.Logger, cfg *Config) {
	// Ensure that workspace cleanup is performed even if storing fails.
	defer cleanupWorkspace(logger)

	corpusPath := filepath.Join(DefaultCorpusDir, "fuzzing")
	if _, err := os.Stat(corpusPath); os.IsNotExist(err) {
		logger.Info("No corpus directory to output")

		return
	}

	// Copy corpus to the results directory
	if err := copy.Copy(corpusPath, cfg.FuzzResultsPath); err != nil {
		logger.Error("Failed to copy corpus", "error", err)
		os.Exit(1)
	}

	logger.Info("Successfully updated corpus directory")
}
