package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/NishantBansal2003/LND-Fuzz/config"
	"github.com/NishantBansal2003/LND-Fuzz/git"
	"github.com/NishantBansal2003/LND-Fuzz/worker"
)

// runFuzzingCycles starts a continuous loop that triggers fuzzing work for a
// specified duration. It creates a sub-context for each cycle and performs
// cleanup after each cycle before starting a new one. The cycles run
// indefinitely until the parent context is canceled.
func runFuzzingCycles(ctx context.Context, logger *slog.Logger, cfg *config.
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
			performCleanup(logger)
		case <-ctx.Done():
			// Overall application context canceled.
			cancelCycle()
			logger.Info("Shutdown initiated during fuzzing " +
				"cycle; performing final cleanup.")
			// Buffer time before cleanup.
			time.Sleep(5 * time.Second)
			performCleanup(logger)
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

// performCleanup handles post-cycle cleanup of the workspace and commits/pushes
// the results. If committing or pushing fails, it logs the error and terminates
// the program.
//
// Note: cleanupWorkspace is deferred within this function.
func performCleanup(logger *slog.Logger) {
	// Ensure that workspace cleanup is performed even if
	// CommitAndPushResults fails.
	defer config.CleanupWorkspace(logger)

	// Commit and push results; if an error occurs, log it and exit.
	if err := git.CommitAndPushResults(logger); err != nil {
		logger.Error("Failed to commit/push results", "error", err)
		os.Exit(1)
	}
}

// main is the entry point of the application. It sets up signal handling for
// graceful shutdown, loads configuration, and starts the continuous fuzzing
// cycles.
func main() {
	// Display help text if "help" argument is provided.
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Println(config.HelpText)
		os.Exit(0)
	}

	// Create a cancellable context to manage the application's lifecycle.
	appCtx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	// Set up signal handling for graceful shutdown on SIGINT and SIGTERM.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Received interrupt signal; shutting down " +
			"gracefully...")
		cancelApp()
	}()

	// Initialize a structured logger that outputs logs in text format.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Load environment variables from a .env file.
	if err := config.LoadEnv(); err != nil {
		logger.Error("Failed to load environment variables", "error",
			err)
		os.Exit(1)
	}

	// Load configuration settings from environment variables.
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Parse the fuzzing cycle duration from configuration (e.g., "20s").
	cycleDuration, err := time.ParseDuration(cfg.FuzzTime)
	if err != nil {
		logger.Error("Error parsing cycle duration", "durationString",
			cfg.FuzzTime, "error", err)
		os.Exit(1)
	}

	// Start the continuous fuzzing cycles.
	runFuzzingCycles(appCtx, logger, cfg, cycleDuration)

	fmt.Println("Program exited.")
}
