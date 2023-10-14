package client

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// flags
	port = flag.Int("port", 50051, "server port")
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	// cli or file of commands to run
	// goroutines
	flag.Parse()
	fmt.Printf("%v", port)

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewNodeClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Create(ctx, &pb.CreateRequest{Path: "/home/folder1", Data: []byte{1, 1, 1, 0}, Flags: "flag", RequestType: pb.RequestType_CLIENT})
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	log.Printf("Greeting: %b", r.Accept)
}
