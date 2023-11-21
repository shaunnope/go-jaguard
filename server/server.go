package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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
	return &pb.Ping{Data: int64(s.Id)}, nil
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

	Checkpoint2(idx, node)

	if *call_watch {
		fmt.Printf("Test watch\n")
		callbackAddr := fmt.Sprintf("%s:%d", "localhost", 50057)
		conn, err := grpc.Dial(callbackAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			slog.Error("Couldn't connect to zkclient", "err", err)
		}
		defer conn.Close()
		client := pb.NewZkCallbackClient(conn)
		client.NotifyWatchTrigger(context.Background(), &pb.WatchNotification{Path: "/test", OperationType: pb.OperationType_DELETE})
	}

	// start grpc service (blocking)
	if err := grpc_s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
