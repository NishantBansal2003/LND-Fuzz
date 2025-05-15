package config

import (
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateProcessCount(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedResult int
	}{
		{
			name:           "process exceeds CPU cores",
			envValue:       strconv.Itoa(runtime.NumCPU() + 1),
			expectedResult: runtime.NumCPU(),
		},
		{
			name:           "process within CPU cores",
			envValue:       "1",
			expectedResult: 1,
		},
		{
			name:           "negative process value",
			envValue:       "-1",
			expectedResult: runtime.NumCPU(),
		},
		{
			name:           "non-numeric process value",
			envValue:       "five",
			expectedResult: runtime.NumCPU(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the FUZZ_NUM_PROCESSES environment variable for
			// the test case.
			err := os.Setenv("FUZZ_NUM_PROCESSES", tt.envValue)
			assert.NoError(t, err)

			// Call the function under test.
			actualResult := calculateProcessCount()

			assert.Equal(t, tt.expectedResult, actualResult,
				"calculated process count does not match")
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// tests := []struct {
	// 	name           string
	// 	projectSrcPath string
	// 	gitStorageRepo string
	// 	fuzzPkgs       string
	// 	fuzzTime       string
	// 	numProcesses   string
	// }{}
	// for _, tt := range tests {
	// }
}
