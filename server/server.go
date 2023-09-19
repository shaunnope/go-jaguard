package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	// flags
	port = flag.Int("port", 50051, "server port")
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

}
