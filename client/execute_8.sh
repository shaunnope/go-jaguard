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
commands="create /client1 10
create /client2 20
create /client3 30
create /client4 40
create /client5 50
create /client1/a a
create /client2/b b
create /client3/b c
create /client4/d d
create /client5/e e
create /client1/a/1 1
create /client2/b/2 2
create /client3/b/3 3
create /client4/d/4 4
create /client5/e/5 5
create /client1/a/1/last 1
create /client2/b/2/last 2
create /client3/b/3/last 3
create /client4/d/4/last 4
create /client5/e/5/last 5
q"
    ;;
  b)
commands="create /client1 10
create /client2 20
create /client3 30
create /client4 40
create /client5 50
create /client1/a a
create /client2/b b
create /client3/b c
create /client4/d d
create /client5/e e
create /client1/a/1 1
create /client2/b/2 2
create /client3/b/3 3
create /client4/d/4 4
create /client5/e/5 5
create /client1/a/1/last 1
get /client1
get /client2
get /client3
get /client4
q"
    ;;
  c)
commands="create /client1 10
create /client2 20
create /client3 30
create /client4 40
create /client5 50
create /client1/a a
create /client2/b b
create /client3/b c
create /client4/d d
create /client5/e e
create /client1/a/1 1
create /client2/b/2 2
get /client4/d
get /client5/e
get /client1/a/1
get /client2/b/2
get /client1
get /client2
get /client3
get /client4
q"
    ;;
  d)
commands="create /client1 10
create /client2 20
create /client3 30
create /client4 40
get /client1
get /client2
get /client3
get /client4
create /client4/d d
create /client5/e e
create /client1/a/1 1
create /client2/b/2 2
get /client4/d
get /client5/e
get /client1/a/1
get /client2/b/2
get /client1
get /client2
get /client3
get /client4
q"
    ;;
  e)
commands="create /client1 10
create /client2 20
create /client3 30
create /client4 40
get /client1
get /client2
get /client3
get /client4
get /client1
get /client2
get /client3
get /client4
get /client1
get /client2
get /client3
get /client4
get /client1
get /client2
get /client3
get /client4
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