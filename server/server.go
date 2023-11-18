package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedNodeServer
	StateVector
}

func NewNode(idx int) *Server {
	s := &Server{StateVector: NewStateVector(idx)}
	return s
}

func (s *Server) SendPing(ctx context.Context, in *pb.Ping) (*pb.Ping, error) {
	return &pb.Ping{Data: int64(s.Id)}, nil
}

// TODO: SetWatch

func (s *Server) GetExists(ctx context.Context, in *pb.GetExistsRequest) (*pb.GetExistsResponse, error) {
	node, err := s.StateVector.Data.GetNode(in.Path)
	if node == nil {
		return &pb.GetExistsResponse{Exists: false, Zxid: s.LastZxid.Inc().Raw()}, err
	}
	return &pb.GetExistsResponse{Exists: true, Zxid: s.LastZxid.Inc().Raw()}, err
}

func (s *Server) GetData(ctx context.Context, in *pb.GetDataRequest) (*pb.GetDataResponse, error) {
	data, err := s.StateVector.Data.GetData(in.Path)

	return &pb.GetDataResponse{Data: data, Zxid: s.LastZxid.Inc().Raw()}, err
}

func (s *Server) GetChildren(ctx context.Context, in *pb.GetChildrenRequest) (*pb.GetChildrenResponse, error) {
	children, err := s.StateVector.Data.GetNodeChildren(in.Path)
	//Type conversion
	out := make([]string, 0)
	for key := range children {
		out = append(out, key)
	}

	return &pb.GetChildrenResponse{Children: out, Zxid: s.LastZxid.Inc().Raw()}, err
}

// Start server
//
// Use reference to grpc server to stop it
func (s *Server) Serve(grpc_s *grpc.Server) {
	time.Sleep(500 * time.Millisecond)

	if *leader_verbo {
		log.Printf("%d begin election ", s.Id)
	}
	if vote := s.FastElection(*maxTimeout); vote.Id == -1 {
		log.Fatalf("%d failed to elect leader", s.Id)
	}

	go s.Heartbeat()
	time.Sleep(200 * time.Millisecond)

	s.Discovery()
	log.Printf("%d finished discovery", s.Id)

	<-s.Stop
	grpc_s.GracefulStop()
}

func Run(idx int) {
	addr := config.Servers[idx]
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", addr.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpc_s := grpc.NewServer()
	// Server Object that handles gRPC requests
	node := NewNode(idx)
	pb.RegisterNodeServer(grpc_s, node)
	log.Printf("server %d listening at %v", idx, lis.Addr())

	// Run fast election then maintain heartbeat
	go node.Serve(grpc_s)

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

	// start grpc service (blocking)
	if err := grpc_s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
