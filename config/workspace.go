package config

import (
	"context"
	"log/slog"
	"os"
)

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
