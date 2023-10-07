package client

import (
	"flag"
	"fmt"
)

var (
	// flags
	port = flag.Int("port", 50051, "server port")
)

func main() {
	// cli or file of commands to run
	// goroutines
	flag.Parse()
	fmt.Printf("%v", port)
}
