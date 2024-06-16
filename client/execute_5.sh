# Function to execute commands and return the time taken
execute_commands() {
  start_time=$(date +%s.%N)  # Record start time
  output=$(echo "$1" | go run *.go)  # Use 'go run *.go' to execute the Go program and capture the output
  end_time=$(date +%s.%N)    # Record end time
  elapsed_time=$(echo "$end_time - $start_time" | bc)  # Calculate elapsed time
  echo "$output"  # Print the output
  echo "Time taken: $elapsed_time seconds"
}

# Check if an argument is provided
if [ -z "$1" ]; then
  echo "Usage: $0 [a|b|c|d|e]"
  exit 1
fi

# Determine the set of commands based on the argument
case "$1" in
  a)
commands="create /k 10
create /l 20
create /m 30
create /n 40
create /o 50
create /k/a a
create /l/b b
create /m/b c
create /n/d d
create /o/e e
create /k/a/1 1
create /l/b/2 2
create /m/b/3 3
create /n/d/4 4
create /o/e/5 5
create /k/a/1/last 1
create /l/b/2/last 2
create /m/b/3/last 3
create /n/d/4/last 4
create /o/e/5/last 5
q"
    ;;
  b)
commands="create /k 10
create /l 20
create /m 30
create /n 40
create /o 50
create /k/a a
create /l/b b
create /m/b c
create /n/d d
create /o/e e
create /k/a/1 1
create /l/b/2 2
create /m/b/3 3
create /n/d/4 4
create /o/e/5 5
create /k/a/1/last 1
get /k
get /l
get /m
get /n
q"
    ;;
  c)
commands="create /k 10
create /l 20
create /m 30
create /n 40
create /o 50
create /k/a a
create /l/b b
create /m/b c
create /n/d d
create /o/e e
create /k/a/1 1
create /l/b/2 2
get /n/d
get /o/e
get /k/a/1
get /l/b/2
get /k
get /l
get /m
get /n
q"
    ;;
  d)
commands="create /k 10
create /l 20
create /m 30
create /n 40
get /k
get /l
get /m
get /n
create /n/d d
create /o/e e
create /k/a/1 1
create /l/b/2 2
get /n/d
get /o/e
get /k/a/1
get /l/b/2
get /k
get /l
get /m
get /n
q"
    ;;
  e)
commands="create /k 10
create /l 20
create /m 30
create /n 40
get /k
get /l
get /m
get /n
get /k
get /l
get /m
get /n
get /k
get /l
get /m
get /n
get /k
get /l
get /m
get /n
q"
    ;;
  f)
commands="get /asd
get /epd
get /esd
get /istd
get /dai
get /asd/a
get /epd/b
get /esd/b
get /istd/d
get /dai/e
get /asd/a/1
get /epd/b/2
get /esd/b/3
get /istd/d/4
get /dai/e/5
get /asd/a/1/last
get /epd/b/2/last
get /esd/b/3/last
get /istd/d/4/last
get /dai/e/5/last
q"
    ;;
  # Add cases for other arguments (c, d, e) as needed
  *)
    echo "Invalid argument. Supported arguments: a, b, c, d, e"
    exit 1
    ;;
esac

# Execute the determined set of commands
execute_commands "$commands"