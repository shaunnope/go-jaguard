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

// TODO: @waishun SetWatch

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
	time.Sleep(500 * time.Millisecond)
	if *leader_verbo {
		log.Printf("server %d begins fast leader election ", s.Id)
	}
	vote := s.FastElection(*maxTimeout)
	if *leader_verbo {
		log.Printf("server %d vote for server %d whose zxid=%v ", s.Id, vote.Id, vote.LastZxid)
	}

	s.Setup(vote)
	s.ElectBroadcast()
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

	if *call_watch {
		fmt.Printf("Test watch\n")
		callbackAddr := fmt.Sprintf("%s:%d", "localhost", 50057)
		conn, err := grpc.Dial(callbackAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			fmt.Println("Couldnt connect to zkclient")
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
