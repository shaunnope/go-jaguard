package main

import (
	"context"
	"log"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func (s *Server) Create(ctx context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error) {
	isLeader := s.GetState() == LEADING

	// Handle incoming CreateRequest
	switch in.RequestType {
	case pb.RequestType_PROPOSAL:
		// if leader, send announcement, do nothing with response (rpc)
		// if follower send proposal reply, do nothing with response (rpc)
		if isLeader {
			r, err := s.Create(ctx, &pb.CreateRequest{
				Path:        in.Path,
				Data:        in.Data,
				Flags:       in.Flags,
				RequestType: pb.RequestType_ANNOUNCEMENT})
			if err != nil || *r.Accept {
				log.Fatalf("err: %v", err)
			}
		} else {
			// forward to leader
			// TODO: verify version

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
		if isLeader {
			log.Fatal("leaders shouldnt get announcements")
		} else {
			// todo traverse tree
		}
	case pb.RequestType_CLIENT:
		// if leader send proposal to all followers in for loop (rpc)
		// since its rpc, leader will monitor for responses and decide whether to commit/announce
		// if follower forward to leader, do nothing with response (rpc)
		if isLeader {
			// todo verify version
			// propose to all
			for idx := range config.Servers {
				if idx == s.Id {
					continue
				}
				s.EstablishConnection(idx)
				newCtx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				r, err := (*s.Connections[idx]).Create(newCtx, &pb.CreateRequest{
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
