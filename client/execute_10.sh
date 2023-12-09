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
commands="create /client11 10
create /client12 20
create /client13 30
create /client14 40
create /client15 50
create /client11/a a
create /client12/b b
create /client13/b c
create /client14/d d
create /client15/e e
create /client11/a/1 1
create /client12/b/2 2
create /client13/b/3 3
create /client14/d/4 4
create /client15/e/5 5
create /client11/a/1/last 1
create /client12/b/2/last 2
create /client13/b/3/last 3
create /client14/d/4/last 4
create /client15/e/5/last 5
q"
    ;;
  b)
commands="create /client11 10
create /client12 20
create /client13 30
create /client14 40
create /client15 50
create /client11/a a
create /client12/b b
create /client13/b c
create /client14/d d
create /client15/e e
create /client11/a/1 1
create /client12/b/2 2
create /client13/b/3 3
create /client14/d/4 4
create /client15/e/5 5
create /client11/a/1/last 1
get /client11
get /client12
get /client13
get /client14
q"
    ;;
  c)
commands="create /client11 10
create /client12 20
create /client13 30
create /client14 40
create /client15 50
create /client11/a a
create /client12/b b
create /client13/b c
create /client14/d d
create /client15/e e
create /client11/a/1 1
create /client12/b/2 2
get /client14/d
get /client15/e
get /client11/a/1
get /client12/b/2
get /client11
get /client12
get /client13
get /client14
q"
    ;;
  d)
commands="create /client11 10
create /client12 20
create /client13 30
create /client14 40
get /client11
get /client12
get /client13
get /client14
create /client14/d d
create /client15/e e
create /client11/a/1 1
create /client12/b/2 2
get /client14/d
get /client15/e
get /client11/a/1
get /client12/b/2
get /client11
get /client12
get /client13
get /client14
q"
    ;;
  e)
commands="create /client11 10
create /client12 20
create /client13 30
create /client14 40
get /client11
get /client12
get /client13
get /client14
get /client11
get /client12
get /client13
get /client14
get /client11
get /client12
get /client13
get /client14
get /client11
get /client12
get /client13
get /client14
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