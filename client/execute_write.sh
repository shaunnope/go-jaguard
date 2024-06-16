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

commands="create /asd 10
create /epd 20
create /esd 30
create /istd 40
create /dai 50
create /asd/a q
create /epd/b b
create /esd/b c
create /istd/d d
create /dai/e e
create /asd/a/1 1
create /epd/b/2 2
create /esd/b/3 3
create /istd/d/4 4
create /dai/e/5 5
create /asd/a/1/last 1
create /epd/b/2/last 2
create /esd/b/3/last 3
create /istd/d/4/last 4
create /dai/e/5/last 5
q"

execute_commands "$commands"