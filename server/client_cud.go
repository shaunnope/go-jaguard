package main

import (
	"context"
	"log"
	"log/slog"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func (s *Server) HandleClientCUD(ctx context.Context, in *pb.CUDRequest) (*pb.CUDResponse, error) {
	// if follower, forward to leader, do nothing with response (rpc)
	// if leader send proposal to all followers in for loop (rpc)
	// since its rpc, leader will monitor for responses and decide whether to commit/announce
	switch state := s.GetState(); state {
	case FOLLOWING:
		slog.Debug("Client Forward", "s", s.Id, "to", s.Vote.Id, "request", in)
		// TODO: verify version
		// forward to leader
		r, err := SendGrpc[*pb.CUDRequest, *pb.CUDResponse](pb.NodeClient.HandleClientCUD, s, s.Vote.Id, in, *maxTimeout)

		return r, err
	case LEADING:

		log.Printf("server %d received client request: %v", s.Id, in)
		// TODO: verify version
		// propose to all
		s.Lock()
		defer s.Unlock()
		msg := &pb.ZabRequest{
			Transaction: &pb.Transaction{
				Zxid:  s.LastZxid.Inc().Raw(),
				Path:  in.Path,
				Data:  in.Data,
				Type:  in.OperationType,
				Flags: &pb.Flag{IsSequential: in.Flags.IsSequential, IsEphemeral: in.Flags.IsEphemeral},
			},
			RequestType: pb.RequestType_PROPOSAL,
		}

		majoritySize := len(s.Zab.FollowerEpochs)/2 + 1
		slog.Info("Client Propose", "s", s.Id, "majority", majoritySize, "request", msg)
		done := make(chan bool, majoritySize)

		log.Printf("server %d prepare to send message: %s", s.Id, msg.Transaction.ExtractLogString())
		copiedMsg := &pb.ZabRequest{
			Transaction: msg.Transaction,
			RequestType: msg.RequestType,
		}
		for idx := range s.Zab.FollowerEpochs {
			if idx == s.Id {
				continue
			}
			go func(i int) {
				if r, err := SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, i, copiedMsg, *maxTimeout); err == nil && r.Accept {
					done <- true
				}
			}(idx)
		}
		// wait for quorum
		for i := 0; i < majoritySize; i++ {
			<-done
		}
		log.Printf("server %d get quorum", s.Id)

		msg.RequestType = pb.RequestType_ANNOUNCEMENT
		for idx := range s.Zab.FollowerEpochs {
			if idx == s.Id {
				continue
			}
			slog.Info("HandleClientCUD", "s", s.Id, "to", idx, "type", msg.RequestType, "txn", msg.Transaction.Extract())
			SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, idx, msg, *maxTimeout)
		}

		accepted := true
		// commit
		err := s.ZabDeliver(msg.Transaction.Extract())
		return &pb.CUDResponse{Accept: &accepted}, err
	default:
		log.Printf("server %d is in %s state", s.Id, state)
		accepted := false
		return &pb.CUDResponse{Accept: &accepted}, nil
	}

}
