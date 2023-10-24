package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

type Transaction = pb.TransactionFragment

func (s *Server) InformLeader(ctx context.Context, in *pb.FollowerInfo) (*pb.Ping, error) {
	if s.State != LEADING {
		panic(fmt.Sprintf("%d is not leader", s.Id))
	}
	log.Printf("%d received follower info from %d", s.Id, in.Id)

	// update leader table
	s.Leader.FollowerEpochs[int(in.Id)] = int(in.LastZxid.Epoch)

	if len(s.Leader.FollowerEpochs) > len(config.Servers)/2 {
		s.Leader.HasQuorum <- true
	}

	return &pb.Ping{Data: int64(s.CurrentEpoch)}, nil
}

func (s *Server) ProposeEpoch(ctx context.Context, in *pb.NewEpoch) (*pb.AckEpoch, error) {
	if s.GetState() != FOLLOWING {
		panic(fmt.Sprintf("%d is not follower", s.Id))
	}
	log.Printf("%d received new epoch from %d", s.Id, in.Epoch)

	if int(in.Epoch) > s.AcceptedEpoch {
		s.AcceptedEpoch = int(in.Epoch)
		res := &pb.AckEpoch{}
		return res, nil
	}

	defer s.FastElection(int(timeout))

	// goto phase 2

	return nil, errors.New("epoch not accepted")
}

func (s *Server) Discovery() {
	s.Lock()
	defer s.Unlock()

	switch s.State {
	case FOLLOWING:
		s.EstablishConnection(s.Vote.Id)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		msg := &pb.FollowerInfo{Id: int64(s.Id), LastZxid: &pb.Zxid{Epoch: int64(s.AcceptedEpoch), Counter: -1}}
		_, err := (*s.Connections[s.Vote.Id]).InformLeader(ctx, msg)
		if err != nil {
			log.Printf("%d error sending follower info to %d: %v", s.Id, idx, err)
		}

	case LEADING:
		<-s.Leader.HasQuorum
		maxEpoch := -1
		for _, epoch := range s.Leader.FollowerEpochs {
			if epoch > maxEpoch {
				maxEpoch = epoch
			}
		}
		mostRecent := &pb.AckEpoch{CurrentEpoch: -1, History: nil, LastZxid: &pb.Zxid{Epoch: -1, Counter: -1}}
		for idx := range s.Leader.FollowerEpochs {
			s.EstablishConnection(idx)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			msg := &pb.NewEpoch{Epoch: int64(maxEpoch + 1)}
			r, err := (*s.Connections[idx]).ProposeEpoch(ctx, msg)
			if err != nil {
				log.Printf("%d error sending new epoch to %d: %v", s.Id, idx, err)
			}
			// follower rejected
			if r == nil {
				return
			}
			if r.CurrentEpoch > mostRecent.CurrentEpoch || (r.CurrentEpoch == mostRecent.CurrentEpoch && !(r.LastZxid.Extract().LessThan(mostRecent.LastZxid.Extract()))) {
				mostRecent = r
			}
		}
		s.History = pb.Transactions(mostRecent.History).ExtractAll()
		// goto phase 2
	}
}
