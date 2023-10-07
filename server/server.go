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
	port   = flag.Int("port", 50051, "server port")
	leader = flag.Bool("isLeader", false, "server is leader")
)

type server struct {
	pb.UnimplementedNodeServer
}

// request
func (s *server) CreateRequest_Client(ctx context.Context, in *pb.CreateRequest_Client) (*pb.CreateResponse_Client, error) {
	// if leader, send proposal
	// if follower, forward to leader using pb.RequestType_CLIENT
	return &pb.CreateResponse_Client{}, nil
}

func (s *server) CreateRequest_Proposal(ctx context.Context, in *pb.CreateRequest_Proposal) (*pb.CreateResponse_Announcement, error) {
	// leaders dont get announcements, panic
	// followers commit locally
	return &pb.CreateResponse_Announcement{}, nil
}

func (s *server) CreateRequest_Announcement(ctx context.Context, in *pb.CreateRequest_Announcement) (*pb.CreateResponse_Proposal, error) {
	// if leader, send announcement
	// if follower reply proposal
	return &pb.CreateResponse_Proposal{}, nil
}

// response
func (s *server) CreateResponse_Client(ctx context.Context, in *pb.CreateResponse_Client) error {

}

func (s *server) CreateResponse_Announcement(ctx context.Context, in *pb.CreateResponse_Announcement) (*pb.CreateResponse_Announcement, error) {
	// leaders dont get announcements, panic
	// followers commit locally
	return &pb.CreateResponse_Announcement{}, nil
}

func (s *server) CreateResponse_Proposal(ctx context.Context, in *pb.CreateResponse_Proposal) (*pb.CreateResponse_Proposal, error) {
	// if leader, send announcement
	// if follower reply proposal
	return &pb.CreateResponse_Proposal{}, nil
}

func (s *server) CreateResponse(ctx context.Context, in *pb.CreateResponse) {
	// Handle CreateResponse
	switch in.ResponseType {
	case pb.ResponseType_PROPOSAL:
		// if leader, collate proposals, if enough and commit, send announcement and commit locally
		// followers dont get acknowledgement, panic
	case pb.RequestType_ANNOUNCEMENT:
		// if leader, do nothing ( no need to respond to annoucement acknowledgements)
		// if follower

	}

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
