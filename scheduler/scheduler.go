package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/NishantBansal2003/LND-Fuzz/config"
	"github.com/NishantBansal2003/LND-Fuzz/worker"
)

// RunFuzzingCycles starts a continuous loop that triggers fuzzing work for a
// specified duration. It creates a sub-context for each cycle and performs
// cleanup after each cycle before starting a new one. The cycles run
// indefinitely until the parent context is canceled.
func RunFuzzingCycles(ctx context.Context, logger *slog.Logger, cfg *config.
	Config, cycleDuration time.Duration) {

	for {
		// Check if the overall application context has been canceled.
		select {
		case <-ctx.Done():
			logger.Info("Shutdown requested; exiting fuzzing " +
				"cycles.")
			return
		default:
			// Continue with the current cycle.
		}

		// Create a sub-context for the current fuzzing cycle.
		cycleCtx, cancelCycle := context.WithCancel(ctx)

		// Start the fuzzing worker concurrently.
		go runFuzzingWorker(cycleCtx, logger, cfg)

		// Wait for either the cycle duration to elapse or the overall
		// context to cancel.
		select {
		case <-time.After(cycleDuration):
			logger.Info("Cycle duration complete; initiating " +
				"cleanup.")
			// Cancel the current cycle.
			cancelCycle()
			// Give a buffer time for routines to exit gracefully.
			time.Sleep(5 * time.Second)
			config.PerformCleanup(logger, cfg)
		case <-ctx.Done():
			// Overall application context canceled.
			cancelCycle()
			logger.Info("Shutdown initiated during fuzzing " +
				"cycle; performing final cleanup.")
			// Buffer time before cleanup.
			time.Sleep(5 * time.Second)
			config.PerformCleanup(logger, cfg)
			return
		}
	}
}

// runFuzzingWorker executes the fuzzing work until the cycle context is
// canceled. It repeatedly calls the main fuzzing function from the app package.
func runFuzzingWorker(ctx context.Context, logger *slog.Logger, cfg *config.
	Config) {

	logger.Info("Starting fuzzing worker", "startTime", time.Now().
		Format(time.RFC1123))

	// Invoke the main fuzzing operation.
	select {
	case <-ctx.Done():
		logger.Info("Fuzzing worker cycle canceled.")
		return
	default:
		// Execute the main fuzzing operation.
		worker.Main(ctx, logger, cfg)
	}
}
