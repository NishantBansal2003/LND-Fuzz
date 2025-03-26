package config

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultFuzzTime specifies the default duration for fuzzing if not
	// overridden by environment variables.
	DefaultFuzzTime = "120s"

	// CleanupTimeout is the timeout used for cleanup operations.
	CleanupTimeout = 5 * time.Second

	// CorpusDir is the directory where the fuzzing corpus is stored.
	CorpusDir = "out/corpus"

	// ProjectDir is the directory where the project is located.
	ProjectDir = "out/project"

	// GitUserName is the default git user name used for commits.
	GitUserName = "github-actions[bot]"

	// GitUserEmail is the default git user email used for commits.
	GitUserEmail = "github-actions[bot]@users.noreply.github.com"

	// CommitMessage is the commit message used when updating the fuzz
	// corpus.
	CommitMessage = "Update fuzz corpus"
)

// Config holds the configuration parameters for the fuzzing setup.
type Config struct {
	ProjectSrcPath string
	GitStorageRepo string
	FuzzPkgs       []string
	FuzzTime       string
	NumProcesses   int
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

	fuzzPkgs := os.Getenv("FUZZPKG")
	if fuzzPkgs == "" {
		return nil, errors.New("FUZZPKG environment variable required")
	}
	cfg.FuzzPkgs = strings.Fields(fuzzPkgs)

	return cfg, nil
}
