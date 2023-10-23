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
	timeout    time.Duration

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
	log.Printf("%d received ping from %d", s.Id, in.From)
	return &pb.Ping{From: int64(s.Id)}, nil
}

// Establish connection to another server
func (s *Server) EstablishConnection(to int) {
	if _, ok := s.Connections[to]; !ok {
		addr := fmt.Sprintf("%s:%d", config.Servers[to].Host, config.Servers[to].Port)
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("%d failed to connect to %d: %v", s.Id, to, err)
		}
		c := pb.NewNodeClient(conn)
		s.Connections[to] = &c
	}
}

func (s *Server) Serve() {
	time.Sleep(1000 * time.Millisecond)
	// vote := s.FastElection(*maxTimeout)
	// log.Printf("%d results: %v", s.Id, vote)

	s.Lock()
	if s.Id == len(config.Servers)-1 {
		s.State = LEADING
		// s.Vote = Vote{0, s.Id}
		s.Vote = &pb.Vote{Id: int64(s.Id)}
		log.Printf("server %d is leader", s.Id)
	} else {
		s.State = FOLLOWING
		s.Vote = &pb.Vote{Id: int64(len(config.Servers) - 1)}
		log.Printf("server %d is following %v", s.Id, s.Vote)
	}
	s.Unlock()

	for {
		state := s.GetState()
		switch state {
		case LEADING:
			for idx := range config.Servers {
				if idx == s.Id {
					continue
				}
				s.EstablishConnection(idx)
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), timeout)
					defer cancel()
					msg := &pb.Ping{From: int64(s.Id)}
					r, err := (*s.Connections[idx]).SendPing(ctx, msg)
					if err != nil {
						log.Printf("%d error sending ping to %d: %v", s.Id, idx, err)
					}
					log.Printf("%d received ack %d", s.Id, r.From)
				}()
				time.Sleep(1000 * time.Millisecond)
			}

		}
	}
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

	// start server routines
	go node.Serve()

	// start grpc server (blocking)
	if err := grpc_s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	flag.Parse()
	timeout = time.Duration(*maxTimeout) * time.Millisecond
	parseConfig(*configPath)
	// Run(*idx)
	for idx := range config.Servers {
		go Run(idx)
	}

	var input string
	fmt.Scanln(&input)
}
