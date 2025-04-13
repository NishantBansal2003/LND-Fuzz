package fuzz

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/NishantBansal2003/LND-Fuzz/config"
)

// RunFuzzing iterates over the configured fuzz packages and executes
// fuzz targets for each package. It stops if the context is canceled.
func RunFuzzing(ctx context.Context, logger *slog.Logger,
	cfg *config.Config) error {

	for _, pkg := range cfg.FuzzPkgs {
		select {
		case <-ctx.Done():
			return nil
		default:
			targets, err := listFuzzTargets(ctx, logger, pkg)
			if err != nil {
				return fmt.Errorf("failed to list targets for"+
					" package %q: %w", pkg, err)
			}

			for _, target := range targets {
				if err := executeFuzzTarget(ctx, logger, pkg,
					target, cfg); err != nil {
					return fmt.Errorf("fuzzing failed for"+
						" %q/%q: %w", pkg, target, err)
				}
			}
		}
	}
	return nil
}

// listFuzzTargets discovers and returns a list of fuzz targets for the given
// package. It uses "go test -list=^Fuzz" to list the functions and filters
// those that start with "Fuzz".
func listFuzzTargets(ctx context.Context, logger *slog.Logger,
	pkg string) ([]string, error) {

	logger.Info("Discovering fuzz targets", "package", pkg)

	pkgPath := filepath.Join(config.ProjectDir, pkg)
	cmd := exec.CommandContext(ctx, "go", "test", "-list=^Fuzz", ".")
	cmd.Dir = pkgPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go test failed for %q: %w (output: %q)",
			pkg, err, strings.TrimSpace(stderr.String()))
	}

	var targets []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		cleanLine := strings.TrimSpace(line)
		if strings.HasPrefix(cleanLine, "Fuzz") {
			targets = append(targets, cleanLine)
		}
	}

	if len(targets) == 0 {
		logger.Warn("No valid fuzz targets found", "package", pkg)
	}
	return targets, nil
}

// executeFuzzTarget runs the specified fuzz target for a package using the
// "go test" command. It sets up the necessary environment, starts the command,
// and streams its output.
func executeFuzzTarget(ctx context.Context, logger *slog.Logger, pkg string,
	target string, cfg *config.Config) error {

	logger.Info("Executing fuzz target", "package", pkg, "target", target)

	pkgPath := filepath.Join(config.ProjectDir, pkg)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	corpusPath := filepath.Join(cwd, config.CorpusDir, pkg, "testdata",
		"fuzz")

	args := []string{
		"test",
		fmt.Sprintf("-fuzz=^%s$", target),
		fmt.Sprintf("-test.fuzzcachedir=%s", corpusPath),
		fmt.Sprintf("-fuzztime=%s", cfg.FuzzTime),
		fmt.Sprintf("-parallel=%d", cfg.NumProcesses),
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = pkgPath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe failed: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe failed: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("command start failed: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go streamOutput(logger.With("target", target), "stdout", &wg, stdout)
	go streamOutput(logger.With("target", target), "stderr", &wg, stderr)

	wg.Wait()

	if err := cmd.Wait(); err != nil && ctx.Err() == nil {
		return fmt.Errorf("fuzz execution failed: %w", err)
	}

	logger.Info("Fuzz target completed successfully", "package", pkg,
		"target", target,
	)

	return nil
}

// streamOutput reads from the provided reader line by line and logs each line
// using the provided logger. It signals completion via the WaitGroup.
func streamOutput(logger *slog.Logger, stream string, wg *sync.WaitGroup,
	r io.Reader) {

	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logger.Info("Fuzzer output", "stream", stream, "message",
			scanner.Text())
	}
}
