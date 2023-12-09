#!/bin/bash

# Function to execute commands and return the time taken
execute_commands() {
  start_time=$(date +%s.%N)  # Record start time
  output=$(echo "$1" | go run *.go)  # Use 'go run *.go' to execute the Go program and capture the output
  end_time=$(date +%s.%N)    # Record end time
  elapsed_time=$(echo "$end_time - $start_time" | bc)  # Calculate elapsed time
  echo "$output"  # Print the output
  echo "Time taken: $elapsed_time seconds"
}

# Execute the provided commands and capture the time and output
commands="$1"
execute_commands "$commands"