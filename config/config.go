package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	// DefaultFuzzTime specifies the default duration for fuzzing if not
	// overridden by environment variables.
	DefaultFuzzTime = "120s"

	// DefaultCleanupTimeout is the timeout used for cleanup operations.
	DefaultCleanupTimeout = 5 * time.Second

	// DefaultCorpusDir is the directory where the fuzzing corpus is stored.
	DefaultCorpusDir = "out/corpus"

	// DefaultProjectDir is the directory where the project is located.
	DefaultProjectDir = "out/project"

	// DefaultGitUserName is the default git user name used for commits.
	DefaultGitUserName = "github-actions[bot]"

	// DefaultGitUserEmail is the default git user email used for commits.
	DefaultGitUserEmail = "github-actions[bot]@users.noreply.github.com"

	// DefaultCommitMessage is the commit message used when updating the
	// fuzz corpus.
	DefaultCommitMessage = "Update fuzz corpus"

	// DefaultReportName is the directory name where fuzzing results are
	// stored.
	DefaultReportName = "fuzz_results"
)

// Config holds the configuration parameters for the fuzzing setup.
type Config struct {
	ProjectSrcPath  string
	GitStorageRepo  string
	FuzzResultsPath string
	FuzzPkgs        []string
	FuzzTime        string
	NumProcesses    int
}

// LoadEnv loads environment variables from a .env file in the current
// directory. It should be called before accessing any environment variables in
// the application. If the file is not found or fails to load, the application
// will return the error.
func LoadEnv() error {
	// If '.env' does not exist, load environment variables from the
	// process's environment
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return nil
	}

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	return nil
}

// calculateProcessCount determines the number of processes to use based on the
// FUZZ_NUM_PROCESSES environment variable. If the variable is set to a valid
// number,it will return that value (capped by the number of available CPUs).
// Otherwise, it returns the number of CPUs available.
func calculateProcessCount() int {
	if envVal := os.Getenv("FUZZ_NUM_PROCESSES"); envVal != "" {
		num, err := strconv.Atoi(envVal)
		if err == nil && num > 0 {
			maxProcs := runtime.NumCPU()
			if num > maxProcs {
				return maxProcs
			}
			return num
		}
	}

	return runtime.NumCPU()
}

// LoadConfig loads the fuzzing configuration from environment variables.
// It sets default values where applicable and returns an error if required
// environment variables are missing or invalid.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		ProjectSrcPath: os.Getenv("PROJECT_SRC_PATH"),
		GitStorageRepo: os.Getenv("GIT_STORAGE_REPO"),
		FuzzTime:       DefaultFuzzTime,
	}

	if cfg.ProjectSrcPath == "" {
		return nil, errors.New("PROJECT_SRC_PATH environment variable" +
			" required")
	}
	if cfg.GitStorageRepo == "" {
		return nil, errors.New("GIT_STORAGE_REPO environment variable" +
			" required")
	}

	cfg.FuzzResultsPath = filepath.Join(
		os.Getenv("FUZZ_RESULTS_PATH"), DefaultReportName,
	)
	err := os.MkdirAll(cfg.FuzzResultsPath, 0755)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("Failed to create directory: %w", err)
	}

	if fuzzTimeStr := os.Getenv("FUZZ_TIME"); fuzzTimeStr != "" {
		seconds, err := strconv.Atoi(fuzzTimeStr)
		if err != nil {
			return nil, fmt.Errorf("FUZZ_TIME environment "+
				"variable must be a number, got %q",
				fuzzTimeStr)
		}

		cfg.FuzzTime = fmt.Sprintf("%ds", seconds)
	}

	cfg.NumProcesses = calculateProcessCount()

	fuzzPkgs := os.Getenv("FUZZ_PKG")
	if fuzzPkgs == "" {
		return nil, errors.New("FUZZ_PKG environment variable required")
	}
	cfg.FuzzPkgs = strings.Fields(fuzzPkgs)

	return cfg, nil
}
