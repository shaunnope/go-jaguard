n=$1 # Number of clients
z=$2 # Type of operations

for ((i = 1; i <= n; i++)); do
  osascript -e "tell app \"Terminal\"
    do script \"docker exec -it go-jaguard-client-$i bash -c './execute_$i.sh $z'\"
  end tell"
done