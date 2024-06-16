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
commands="create /aa 10
create /bb 20
create /cc 30
create /dd 40
create /ee 50
create /aa/a a
create /bb/b b
create /cc/b c
create /dd/d d
create /ee/e e
create /aa/a/1 1
create /bb/b/2 2
create /cc/b/3 3
create /dd/d/4 4
create /ee/e/5 5
create /aa/a/1/last 1
create /bb/b/2/last 2
create /cc/b/3/last 3
create /dd/d/4/last 4
create /ee/e/5/last 5
q"
    ;;
  b)
commands="create /aa 10
create /bb 20
create /cc 30
create /dd 40
create /ee 50
create /aa/a a
create /bb/b b
create /cc/b c
create /dd/d d
create /ee/e e
create /aa/a/1 1
create /bb/b/2 2
create /cc/b/3 3
create /dd/d/4 4
create /ee/e/5 5
create /aa/a/1/last 1
get /aa
get /bb
get /cc
get /dd
q"
    ;;
  c)
commands="create /aa 10
create /bb 20
create /cc 30
create /dd 40
create /ee 50
create /aa/a a
create /bb/b b
create /cc/b c
create /dd/d d
create /ee/e e
create /aa/a/1 1
create /bb/b/2 2
get /dd/d
get /ee/e
get /aa/a/1
get /bb/b/2
get /aa
get /bb
get /cc
get /dd
q"
    ;;
  d)
commands="create /aa 10
create /bb 20
create /cc 30
create /dd 40
get /aa
get /bb
get /cc
get /dd
create /dd/d d
create /ee/e e
create /aa/a/1 1
create /bb/b/2 2
get /dd/d
get /ee/e
get /aa/a/1
get /bb/b/2
get /aa
get /bb
get /cc
get /dd
q"
    ;;
  e)
commands="create /aa 10
create /bb 20
create /cc 30
create /dd 40
get /aa
get /bb
get /cc
get /dd
cget /aa
get /bb
get /cc
get /dd
get /aa
get /bb
get /cc
get /dd
get /aa
get /bb
get /cc
get /dd
q"    ;;
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