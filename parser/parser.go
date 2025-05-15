package parser

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/NishantBansal2003/LND-Fuzz/config"
)

var (
	// fuzzFailureRegex matches lines indicating a fuzzing failure or a
	// failing input, capturing the fuzz target name and the corresponding
	// input ID.
	//
	// It matches lines like:
	//   "failure while testing seed corpus entry: FuzzFoo/771e938e4458e983"
	//   "Failing input written to testdata/fuzz/FuzzFoo/771e938e4458e983"
	//
	// Captured groups:
	//   - "target": the fuzz target name (e.g., "FuzzFoo")
	//   - "id": the hexadecimal input ID (e.g., "771e938e4458e983")
	fuzzFailureRegex = regexp.MustCompile(
		`(?:failure while testing seed corpus entry:\s*|Failing ` +
			`input written to\s*testdata/fuzz/)` +
			`(?P<target>[^/]+)/(?P<id>[0-9a-f]+)`,
	)
)

// FuzzProcessor reads a stream of fuzzer output lines, detects failures,
// writes logs, and captures the failing input for later logging.
type FuzzProcessor struct {
	logger     *slog.Logger
	cfg        *config.Config
	corpusPath string
	target     string
	logWriter  LogWriter
	State      *ProcessState
}

// ProcessState holds mutable state information while processing fuzzer output.
type ProcessState struct {
	// SeenFailure indicates whether a failure marker line has been spotted.
	SeenFailure bool

	// ErrorData contains the formatted contents of the failing testcase.
	ErrorData string

	// InputPrinted indicates whether the failing input has already been
	// captured.
	InputPrinted bool
}

// LogWriter defines the interface for writing fuzz failure logs to any sink.
type LogWriter interface {
	// Initialize creates (or truncates, if it already exists) the log file
	// at the specified path, preparing the sink for subsequent log output.
	Initialize(logPath string) error

	// WriteLine writes a single line of fuzzer output (followed by a
	// newline) to the underlying log sink.
	WriteLine(line string) error

	// WriteErrorData writes the formatted contents of the failing testcase
	// (plus newline) to the log file.
	WriteErrorData(data string) error

	// Close writes the trailing formatted contents of the failing testcase
	// and closes the underlying file.
	Close(errorData string) error
}

// FileLogWriter is a LogWriter that writes failure logs into a file.
type FileLogWriter struct {
	file *os.File
}

// Initialize creates (or truncates, if it already exists) the log file at the
// specified path, preparing the sink for subsequent log output.
func (fl *FileLogWriter) Initialize(logPath string) error {
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("Failed to create log file: %w", err)
	}

	fl.file = logFile
	return nil
}

// WriteLine writes a single line of fuzzer output (followed by a newline) to
// the underlying log sink.
func (fl *FileLogWriter) WriteLine(line string) error {
	_, err := fl.file.WriteString(line + "\n")
	if err != nil {
		return fmt.Errorf("Failed to write log file: %w", err)
	}

	return nil
}

// WriteErrorData writes the formatted contents of the failing testcase (plus
// newline) to the log file.
func (fl *FileLogWriter) WriteErrorData(data string) error {
	_, err := fl.file.WriteString(data + "\n")
	if err != nil {
		return fmt.Errorf("Failed to write log file: %w", err)
	}

	return nil
}

// Close writes the trailing formatted contents of the failing testcase and
// closes the underlying file.
func (fl *FileLogWriter) Close(errorData string) error {
	if err := fl.WriteErrorData(errorData); err != nil {
		return fmt.Errorf("error data write failed: %w", err)
	}

	return fl.file.Close()
}

// NewFuzzProcessor constructs a FuzzProcessor for the given logger, config,
// corpus path, and fuzz target name.
func NewFuzzProcessor(logger *slog.Logger, cfg *config.Config,
	corpusPath string, target string) *FuzzProcessor {

	return &FuzzProcessor{
		logger:     logger,
		cfg:        cfg,
		corpusPath: corpusPath,
		target:     target,
		State:      &ProcessState{},
		logWriter:  &FileLogWriter{},
	}
}

// ProcessStream reads each line from the fuzzing output, processes it (logging
// every line and capturing any failure details), and when complete closes the
// log writer, flushing any accumulated error data.
func (fp *FuzzProcessor) ProcessStream(stream io.Reader) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		if err := fp.processLine(scanner.Text()); err != nil {
			fp.logger.Error("Error processing line", "error", err)
		}
	}

	// Ensure we flush and close the log writer with the final error data.
	defer func() { _ = fp.logWriter.Close(fp.State.ErrorData) }()
}

// processLine handles one line of fuzz output: it logs it, checks for failure
// markers, and if in failure mode, writes lines and captures failing input.
func (fp *FuzzProcessor) processLine(line string) error {
	fp.logger.Info("Fuzzer output", "message", line)

	if !fp.State.SeenFailure {
		if err := fp.handleFailureDetection(line); err != nil {
			return fmt.Errorf("failure detection failed: %w", err)
		}
	}

	if fp.State.SeenFailure {
		if err := fp.handleFailureLine(line); err != nil {
			return fmt.Errorf("failure line handling failed: %w",
				err)
		}
	}

	return nil
}

// handleFailureDetection looks for the first "--- FAIL:" marker. When found,
// it initializes the failure log file for subsequent lines.
func (fp *FuzzProcessor) handleFailureDetection(line string) error {
	if strings.Contains(line, "--- FAIL:") {
		fp.State.SeenFailure = true
		logFileName := fmt.Sprintf("%s_failure.log", fp.target)
		logPath := filepath.Join(fp.cfg.FuzzResultsPath, logFileName)

		if err := fp.logWriter.Initialize(logPath); err != nil {
			return fmt.Errorf("log writer initialization failed: "+
				"%w", err)
		}

		fp.logger.Info("Failure log initialized", "path", logPath)
	}
	return nil
}

// handleFailureLine writes the line to the log, then on the first occurrence
// of a failure-input marker extracts the testcase and reads its contents.
func (fp *FuzzProcessor) handleFailureLine(line string) error {
	if err := fp.logWriter.WriteLine(line); err != nil {
		return fmt.Errorf("failed to write log line: %w", err)
	}

	if fp.State.InputPrinted {
		return nil
	}

	target, id, err := parseFailureLine(line)
	if err != nil {
		return fmt.Errorf("failure line parsing failed: %w", err)
	}
	if target == "" || id == "" {
		return nil
	}

	errorData, err := fp.readInputData(target, id)
	if err != nil {
		return fmt.Errorf("input data read failed: %w", err)
	}

	fp.State.ErrorData = errorData
	fp.State.InputPrinted = true
	return nil
}

// parseFailureLine returns the fuzz target name and input ID if the line
// matches fuzzFailureRegex; otherwise returns empty strings.
func parseFailureLine(line string) (string, string, error) {
	matches := fuzzFailureRegex.FindStringSubmatch(line)
	if matches == nil {
		return "", "", nil
	}

	var target, id string
	for i, name := range fuzzFailureRegex.SubexpNames() {
		switch name {
		case "target":
			target = matches[i]
		case "id":
			id = matches[i]
		}
	}
	return target, id, nil
}

// readInputData attempts to read the failing input file from the corpus and
// returns either its contents or an error placeholder string.
func (fp *FuzzProcessor) readInputData(target, id string) (string, error) {
	failingInputPath := filepath.Join(target, id)
	inputPath := filepath.Join(fp.corpusPath, failingInputPath)
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Sprintf("\n<< failed to read %s: %v >>\n",
			failingInputPath, err), nil
	}
	return fmt.Sprintf("\n\n=== Failing testcase (%s) ===\n%s",
		failingInputPath, data), nil
}
