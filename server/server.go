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
	port          = flag.Int("port", 50051, "server port")
	leader        = flag.Bool("isLeader", false, "server is leader")
	head          = new(pb.Znode)
	other_servers = []string{"localhost:50052", "localhost:50053", "localhost:50054"}
)

type server struct {
	pb.UnimplementedNodeServer
}

func (s *server) Create(ctx context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error) {
	// Handle incoming CreateRequest
	switch in.RequestType {
	case pb.RequestType_PROPOSAL:
		// if leader, send announcement, do nothing with response (rpc)
		// if follower send proposal reply, do nothing with response (rpc)
		if *leader {
			r, err := s.Create(ctx, &pb.CreateRequest{
				Path:        in.Path,
				Data:        in.Data,
				Flags:       in.Flags,
				RequestType: pb.RequestType_ANNOUNCEMENT})
			if err != nil || *r.Accept {
				log.Fatalf("err: %v", err)
			}
		} else {
			// todo verify version

			r, err := s.Create(ctx, &pb.CreateRequest{
				Path:        in.Path,
				Data:        in.Data,
				Flags:       in.Flags,
				RequestType: pb.RequestType_PROPOSAL})
			if err != nil || *r.Accept {
				log.Fatalf("err: %v", err)
			}
		}
	case pb.RequestType_ANNOUNCEMENT:
		// leaders dont get announcements, panic
		// followers commit locally
		if *leader {
			log.Fatal("leaders shouldnt get announcements")
		} else {
			// todo traverse tree
		}
	case pb.RequestType_CLIENT:
		// if leader send proposal to all followers in for loop (rpc)
		// since its rpc, leader will monitor for responses and decide whether to commit/announce
		// if follower forward to leader, do nothing with response (rpc)
		if *leader {
			// todo verify version
			// propose to all
			for _, element := range other_servers {
				// todo how to send to other IP?
				r, err := s.Create(ctx, &pb.CreateRequest{
					Path:        in.Path,
					Data:        in.Data,
					Flags:       in.Flags,
					RequestType: pb.RequestType_PROPOSAL})
				if err != nil || *r.Accept {
					log.Fatalf("err: %v", err)
				}
			}
		} else {
			// todo verify version

			r, err := s.Create(ctx, &pb.CreateRequest{
				Path:        in.Path,
				Data:        in.Data,
				Flags:       in.Flags,
				RequestType: pb.RequestType_PROPOSAL})
			if err != nil || *r.Accept {
				log.Fatalf("err: %v", err)
			}
		}
	}
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
