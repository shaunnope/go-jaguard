package main

import (
	"context"
	"log"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func (s *Server) Shutdown(ctx context.Context, req *pb.Ping) (*pb.Ping, error) {

	return &pb.Ping{}, nil
}

func Checkpoint2(idx int, node *Server) {
	if idx == 1 && *multiple_req {
		log.Printf("server %d received request from client", idx)
		go Simulate(node, "/foo")
		go Simulate(node, "/bar")
	}

	if idx == 2 && *multiple_cli {
		log.Printf("server %d received request from client", idx)
		go Simulate(node, "/cli2-1")
	}

	if idx == 3 && *multiple_cli {
		log.Printf("server %d received request from client", idx)
		go Simulate(node, "/cli3-1")
	}
}
