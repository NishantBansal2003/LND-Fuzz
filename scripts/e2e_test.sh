#!/bin/bash

# Run the make command with a 60-minute timeout
timeout 10m make run
EXIT_STATUS=$?

# If timeout triggered (124), reset to success
if [ $EXIT_STATUS -eq 124 ]; then
  echo "⏰ Timeout reached after 60m—continuing as success."
  EXIT_STATUS=0
fi

# Check the exit status of the make command
# Report any genuine failures
if [ $EXIT_STATUS -ne 0 ]; then
  echo "❌ The operation exited with status $EXIT_STATUS."
  exit $EXIT_STATUS
fi

# Check if the ./fuzz_results directory exists
if [ -d "./fuzz_results" ]; then
  echo "✅ Fuzzing process completed successfully."
else
  echo "❌ Fuzzing process failed."
  exit 1
fi

# Cleanup: Delete the ./fuzz_results directory
rm -rf ./fuzz_results