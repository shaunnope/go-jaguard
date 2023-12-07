n=$1 # Number of clients

for ((i = 1; i <= n; i++)); do
  osascript -e "tell app \"Terminal\"
    do script \"docker exec -it "go-jaguard-client-$i" bash -c 'go run *.go'\"
  end tell"
done
