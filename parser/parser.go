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

type FuzzProcessor struct {
	logger     *slog.Logger
	cfg        *config.Config
	corpusPath string
	target     string
	logWriter  LogWriter
	State      *ProcessState
}

type ProcessState struct {
	SeenFailure  bool
	ErrorData    string
	InputPrinted bool
}

type LogWriter interface {
	Initialize(logPath string) error
	WriteLine(line string) error
	WriteErrorData(data string) error
	Close(errorData string) error
}

type FileLogWriter struct {
	file *os.File
}

func (fl *FileLogWriter) Initialize(logPath string) error {
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("Failed to create log file: %w", err)
	}

	fl.file = logFile
	return nil
}

func (fl *FileLogWriter) WriteLine(line string) error {
	_, err := fl.file.WriteString(line + "\n")
	if err != nil {
		return fmt.Errorf("Failed to write log file: %w", err)
	}

	return nil
}

func (fl *FileLogWriter) WriteErrorData(data string) error {
	_, err := fl.file.WriteString(data + "\n")
	if err != nil {
		return fmt.Errorf("Failed to write log file: %w", err)
	}

	return nil
}

func (fl *FileLogWriter) Close(errorData string) error {
	if err := fl.WriteErrorData(errorData); err != nil {
		return fmt.Errorf("error data write failed: %w", err)
	}

	return fl.file.Close()
}

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

func (fp *FuzzProcessor) ProcessStream(stream io.Reader) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		if err := fp.processLine(scanner.Text()); err != nil {
			fp.logger.Error("Error processing line", "error", err)
		}
	}

	defer func() { _ = fp.logWriter.Close(fp.State.ErrorData) }()
}

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
