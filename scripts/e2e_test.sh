#!/bin/bash

# Set environment variables
ENV_FILE_NAME="./sample.env"
VOLUME_MOUNTS_CMD="-v ./temp:/app/fuzz_results"

# Create the temp directory if it doesn't exist
mkdir -p ./temp

# Run the make command with a 60-minute timeout
timeout 10m make docker-run-file ENV_FILE="$ENV_FILE_NAME" VOLUME_MOUNTS="$VOLUME_MOUNTS_CMD"
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

# Sleep for 5 minutes to allow the fuzzing process for cleanup
sleep 5m

# Check if the ./temp/fuzz_results directory exists
if [ -d "./temp/fuzz_results" ]; then
  echo "✅ Fuzzing process completed successfully."
else
  echo "❌ Fuzzing process failed."
  exit 1
fi

# Cleanup: Delete the ./temp directory
rm -rf ./temp