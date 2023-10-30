package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	// log.Printf("PING %d -> %d", in.Data, s.Id)
	return &pb.Ping{Data: int64(s.Id)}, nil
}

// Establish connection to another server
func (s *Server) EstablishConnection(to int, timeout int) (context.Context, context.CancelFunc) {
	if to == s.Id {
		return nil, nil
	}
	if _, ok := s.Connections[to]; !ok {
		addr := fmt.Sprintf("%s:%d", config.Servers[to].Host, config.Servers[to].Port)
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("%d failed to connect to %d: %v", s.Id, to, err)
		}
		c := pb.NewNodeClient(conn)
		s.Connections[to] = &c
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	return ctx, cancel
}

// Start server
//
// Use reference to grpc server to stop it
func (s *Server) Serve(grpc_s *grpc.Server) {
	time.Sleep(200 * time.Millisecond)
	vote := s.FastElection(*maxTimeout)

	s.Setup(vote)
	go s.Heartbeat()
	time.Sleep(200 * time.Millisecond)

	// s.Discovery()
	// log.Printf("%d finished discovery", s.Id)

	// var input string
	// fmt.Scanln(&input)

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
	node := NewNode(idx)
	pb.RegisterNodeServer(grpc_s, node)
	log.Printf("server %d listening at %v", idx, lis.Addr())

	go node.Serve(grpc_s)

	if idx == 1 {
		log.Printf("server %d received request from client", idx)
		go Simulate(node)
	}

	// start grpc service (blocking)
	if err := grpc_s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
