#!/bin/bash

# Set environment variables
ENV_FILE_NAME="./sample.env"
VOLUME_MOUNTS_CMD="-v ./temp:/app/fuzz_results"

# Create the temp directory if it doesn't exist
mkdir -p ./temp

# Start the make command in the background
make docker-run-file ENV_FILE="$ENV_FILE_NAME" VOLUME_MOUNTS="$VOLUME_MOUNTS_CMD" &
MAKE_PID=$!

# After 60 minutes, send SIGINT (Ctrl+C) to the make process
(
  sleep 60m
  echo "⏰ 60 minutes elapsed, sending SIGINT to make (PID $MAKE_PID)..."
  kill -INT $MAKE_PID
) &

WAITER_PID=$!

# Wait for the make command to finish
wait $MAKE_PID
EXIT_STATUS=$?

# Kill the waiter if make finished early
kill $WAITER_PID 2>/dev/null

# Check the exit status of the make command
if [ $EXIT_STATUS -ne 0 ]; then
  echo "❌ The operation exited with status $EXIT_STATUS."
  exit $EXIT_STATUS
fi

# Check if the ./temp/fuzz_results directory exists
if [ -d "./temp/fuzz_results" ]; then
  echo "✅ Fuzzing process completed successfully."
else
  echo "❌ Fuzzing process failed."
  exit 1
fi

# Cleanup: Delete the ./temp directory
rm -rf ./temp