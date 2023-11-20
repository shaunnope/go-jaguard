package main

import (
	"context"
	"log"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func (s *Server) HandleClientCUDS(ctx context.Context, in *pb.CUDSRequest) (*pb.CUDSResponse, error) {
	isLeader := s.GetState() == LEADING

	// if leader send proposal to all followers in for loop (rpc)
	// since its rpc, leader will monitor for responses and decide whether to commit/announce
	// if follower forward to leader, do nothing with response (rpc)
	if isLeader {
		log.Printf("server %d received client request: %v", s.Id, in)
		// TODO: verify version
		// propose to all
		s.Lock()
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

		s.Leader.FollowerEpochs = map[int]int{
			0: 0,
			1: 0,
			2: 0,
			3: 0,
			4: 0,
		}

		majoritySize := len(s.Leader.FollowerEpochs)/2 + 1
		log.Printf("server %d need %v to reach majority", s.Id, majoritySize)
		done := make(chan bool, majoritySize)

		successfulSends := 0

		log.Printf("server %d prepare to send message: %s", s.Id, msg.Transaction.ExtractLogString())
		copiedMsg := &pb.ZabRequest{
			Transaction: msg.Transaction,
			RequestType: msg.RequestType,
		}
		for idx := range s.Leader.FollowerEpochs {
			if idx == s.Id {
				continue
			}
			go func(i int) {
				_, err := SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, i, copiedMsg, *maxTimeout)
				if err == nil {
					// NOTE: might have race condition on successfulSends
					successfulSends++
					// log.Printf("server %d gotten %d acknowledgement", s.Id, successfulSends)
					if successfulSends <= majoritySize {
						done <- true
					}
				}
			}(idx)
		}
		// wait for quorum
		for len(done) < majoritySize {
		}
		log.Printf("server %d get quoram", s.Id)

		// commit
		s.History = append(s.History, msg.Transaction.ExtractLog())
		s.LastZxid = msg.Transaction.Extract().Zxid

		// @Shi Hui: Leader commit change on local copy
		// LEADER need to execute request
		transaction := msg.Transaction.Extract()
		var err error
		if in.OperationType != pb.OperationType_SYNC {
			_, err = s.HandleOperation(transaction)
		}

		msg.RequestType = pb.RequestType_ANNOUNCEMENT
		for idx := range s.Leader.FollowerEpochs {
			if idx == s.Id {
				continue
			}

			log.Printf("server %d sending ANNOUNCEMENT TO server %d with message %v", s.Id, idx, msg)
			SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, idx, msg, *maxTimeout)
		}
		defer s.Unlock()
		accepted := true
		return &pb.CUDSResponse{Accept: &accepted}, err
	} else {
		log.Printf("server %d forwarding request to %d", s.Id, s.Vote.Id)
		// todo verify version
		// forward to leader
		r, err := SendGrpc[*pb.CUDSRequest, *pb.CUDSResponse](pb.NodeClient.HandleClientCUDS, s, s.Vote.Id, in, *maxTimeout)

		return r, err
	}

}
