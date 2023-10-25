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
			for idx := range config.Servers {
				if idx == s.Id {
					continue
				}
				go func(i int) {
					newCtx, cancel := s.EstablishConnection(i, *maxTimeout)
					defer cancel()

					r, err := (*s.Connections[i]).Create(newCtx, &pb.CreateRequest{
						Path:        in.Path,
						Data:        in.Data,
						Flags:       in.Flags,
						RequestType: pb.RequestType_ANNOUNCEMENT})
					if err != nil || *r.Accept {
						log.Fatalf("err: %v", err)
					}
				}(idx)
			}

		} else {
			// forward to leader
			// TODO: verify version
			newCtx, cancel := s.EstablishConnection(s.Vote.Id, *maxTimeout)
			defer cancel()

			r, err := (*s.Connections[s.Vote.Id]).Create(newCtx, &pb.CreateRequest{
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
				newCtx, cancel := s.EstablishConnection(idx, *maxTimeout)
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
			newCtx, cancel := s.EstablishConnection(s.Vote.Id, *maxTimeout)
			defer cancel()
			r, err := (*s.Connections[s.Vote.Id]).Create(newCtx, &pb.CreateRequest{
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
