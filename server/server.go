package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// flags
	configPath = flag.String("config", "config.json", "path to config file")
	config     Config

	idx = flag.Int("idx", 0, "server index")

	maxTimeout = flag.Int("maxTimeout", 1000, "max timeout for election")

	// port = flag.Int("port", 50051, "server port")
	// leader = flag.Bool("isLeader", false, "server is leader")
	// head = new(pb.Znode)
	// other_servers = []string{"localhost:50052", "localhost:50053", "localhost:50054"}
)

func parseConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}
}

type Server struct {
	pb.UnimplementedNodeServer
	StateVector
}

func (s *Server) SendPing(ctx context.Context, in *pb.Ping) (*pb.Ping, error) {
	log.Printf("PING %d -> %d", in.Data, s.Id)
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
	// vote := s.FastElection(*maxTimeout)
	// log.Printf("%d results: %v", s.Id, vote)

	s.Setup()
	go s.Heartbeat()

	// s.Discovery()
	// log.Printf("%d finished discovery", s.Id)

	// var input string
	// fmt.Scanln(&input)

	<-s.Stop
	grpc_s.GracefulStop()
}

func newNode(idx int) *Server {
	s := &Server{StateVector: newStateVector(idx)}
	return s
}

func Run(idx int) {
	addr := config.Servers[idx]
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", addr.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpc_s := grpc.NewServer()
	node := newNode(idx)
	pb.RegisterNodeServer(grpc_s, node)
	log.Printf("server %d listening at %v", idx, lis.Addr())

	go node.Serve(grpc_s)

	// start grpc service (blocking)
	if err := grpc_s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	flag.Parse()
	parseConfig(*configPath)
	// Run(*idx)
	for idx := range config.Servers {
		go Run(idx)
	}

	var input string
	fmt.Scanln(&input)
}
