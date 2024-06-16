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
commands="create /h 10
create /he 20
create /hel 30
create /hell 40
create /hello 50
create /h/a a
create /he/b b
create /hel/b c
create /hell/d d
create /hello/e e
create /h/a/1 1
create /he/b/2 2
create /hel/b/3 3
create /hell/d/4 4
create /hello/e/5 5
create /h/a/1/last 1
create /he/b/2/last 2
create /hel/b/3/last 3
create /hell/d/4/last 4
create /hello/e/5/last 5
q"
    ;;
  b)
commands="create /h 10
create /he 20
create /hel 30
create /hell 40
create /hello 50
create /h/a a
create /he/b b
create /hel/b c
create /hell/d d
create /hello/e e
create /h/a/1 1
create /he/b/2 2
create /hel/b/3 3
create /hell/d/4 4
create /hello/e/5 5
create /h/a/1/last 1
get /h
get /he
get /hel
get /hell
q"
    ;;
  c)
commands="create /h 10
create /he 20
create /hel 30
create /hell 40
create /hello 50
create /h/a a
create /he/b b
create /hel/b c
create /hell/d d
create /hello/e e
create /h/a/1 1
create /he/b/2 2
get /hell/d
get /hello/e
get /h/a/1
get /he/b/2
get /h
get /he
get /hel
get /hell
q"
    ;;
  d)
commands="create /h 10
create /he 20
create /hel 30
create /hell 40
get /h
get /he
get /hel
get /hell
create /hell/d d
create /hello/e e
create /h/a/1 1
create /he/b/2 2
get /hell/d
get /hello/e
get /h/a/1
get /he/b/2
get /h
get /he
get /hel
get /hell
q"
    ;;
  e)
commands="create /h 10
create /he 20
create /hel 30
create /hell 40
get /h
get /he
get /hel
get /hell
get /h
get /he
get /hel
get /hell
get /h
get /he
get /hel
get /hell
get /h
get /he
get /hel
get /hell
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