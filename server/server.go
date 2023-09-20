package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
)

var (
	// flags
	port = flag.Int("port", 50051, "server port")
)

type server struct {
	pb.UnimplementedNodeServer
}

func (s *server) Create(ctx context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error) {
	return &pb.CreateResponse{}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterNodeServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
