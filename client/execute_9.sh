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
commands="create /client6 10
create /client7 20
create /client8 30
create /client9 40
create /client10 50
create /client6/a a
create /client7/b b
create /client8/b c
create /client9/d d
create /client10/e e
create /client6/a/1 1
create /client7/b/2 2
create /client8/b/3 3
create /client9/d/4 4
create /client10/e/5 5
create /client6/a/1/last 1
create /client7/b/2/last 2
create /client8/b/3/last 3
create /client9/d/4/last 4
create /client10/e/5/last 5
q"
    ;;
  b)
commands="create /client6 10
create /client7 20
create /client8 30
create /client9 40
create /client10 50
create /client6/a a
create /client7/b b
create /client8/b c
create /client9/d d
create /client10/e e
create /client6/a/1 1
create /client7/b/2 2
create /client8/b/3 3
create /client9/d/4 4
create /client10/e/5 5
create /client6/a/1/last 1
get /client6
get /client7
get /client8
get /client9
q"
    ;;
  c)
commands="create /client6 10
create /client7 20
create /client8 30
create /client9 40
create /client10 50
create /client6/a a
create /client7/b b
create /client8/b c
create /client9/d d
create /client10/e e
create /client6/a/1 1
create /client7/b/2 2
get /client9/d
get /client10/e
get /client6/a/1
get /client7/b/2
get /client6
get /client7
get /client8
get /client9
q"
    ;;
  d)
commands="create /client6 10
create /client7 20
create /client8 30
create /client9 40
get /client6
get /client7
get /client8
get /client9
create /client9/d d
create /client10/e e
create /client6/a/1 1
create /client7/b/2 2
get /client9/d
get /client10/e
get /client6/a/1
get /client7/b/2
get /client6
get /client7
get /client8
get /client9
q"
    ;;
  e)
commands="create /client6 10
create /client7 20
create /client8 30
create /client9 40
get /client6
get /client7
get /client8
get /client9
get /client6
get /client7
get /client8
get /client9
get /client6
get /client7
get /client8
get /client9
get /client6
get /client7
get /client8
get /client9
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