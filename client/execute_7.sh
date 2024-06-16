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
commands="create /ff 10
create /gg 20
create /hh 30
create /ii 40
create /jj 50
create /ff/a a
create /gg/b b
create /hh/b c
create /ii/d d
create /jj/e e
create /ff/a/1 1
create /gg/b/2 2
create /hh/b/3 3
create /ii/d/4 4
create /jj/e/5 5
create /ff/a/1/last 1
create /gg/b/2/last 2
create /hh/b/3/last 3
create /ii/d/4/last 4
create /jj/e/5/last 5
q"
    ;;
  b)
commands="create /ff 10
create /gg 20
create /hh 30
create /ii 40
create /jj 50
create /ff/a a
create /gg/b b
create /hh/b c
create /ii/d d
create /jj/e e
create /ff/a/1 1
create /gg/b/2 2
create /hh/b/3 3
create /ii/d/4 4
create /jj/e/5 5
create /ff/a/1/last 1
get /ff
get /gg
get /hh
get /ii
q"
    ;;
  c)
commands="create /ff 10
create /gg 20
create /hh 30
create /ii 40
create /jj 50
create /ff/a a
create /gg/b b
create /hh/b c
create /ii/d d
create /jj/e e
create /ff/a/1 1
create /gg/b/2 2
get /ii/d
get /jj/e
get /ff/a/1
get /gg/b/2
get /ff
get /gg
get /hh
get /ii
q"
    ;;
  d)
commands="create /ff 10
create /gg 20
create /hh 30
create /ii 40
get /ff
get /gg
get /hh
get /ii
create /ii/d d
create /jj/e e
create /ff/a/1 1
create /gg/b/2 2
get /ii/d
get /jj/e
get /ff/a/1
get /gg/b/2
get /ff
get /gg
get /hh
get /ii
q"
    ;;
  e)
commands="create /ff 10
create /gg 20
create /hh 30
create /ii 40
get /ff
get /gg
get /hh
get /ii
get /ff
get /gg
get /hh
get /ii
get /ff
get /gg
get /hh
get /ii
get /ff
get /gg
get /hh
get /ii
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