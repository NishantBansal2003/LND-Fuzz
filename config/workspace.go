package config

import (
	"context"
	"log/slog"
	"os"
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

Usage Example:
  Set the necessary environment variables, then start fuzzing:
      go run main.go

For more information, please refer to the project documentation.`

// CleanupWorkspace removes the "out" directory to clean up the workspace.
// It uses a context with a timeout (CleanupTimeout) to limit the duration of
// the cleanup operation. Any errors encountered during the removal are logged.
func CleanupWorkspace(logger *slog.Logger) {
	_, cancel := context.WithTimeout(context.Background(), CleanupTimeout)
	defer cancel()

	if err := os.RemoveAll("out"); err != nil {
		logger.Error("Cleanup failed", "error", err)
	}
}
