package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func (s *Server) HandleClientCUDS(ctx context.Context, in *pb.CUDSRequest) (*pb.CUDSResponse, error) {
	// if follower, forward to leader, do nothing with response (rpc)
	// if leader send proposal to all followers in for loop (rpc)
	// since its rpc, leader will monitor for responses and decide whether to commit/announce
	switch state := s.State; state {
	case FOLLOWING:
		slog.Debug("Client Forward", "s", s.Id, "to", s.Vote.Id, "request", in)
		// TODO: verify version
		// forward to leader
		r, err := SendGrpc(pb.NodeClient.HandleClientCUDS, s, s.Vote.Id, in, *maxTimeout)

		return r, err
	case LEADING:

		log.Printf("server %d received client request: %v", s.Id, in)
		// TODO: verify version
		// propose to all
		s.Zab.Lock()
		defer s.Zab.Unlock()
		nextZxid := s.LastZxid.Inc()
		if nextZxid.Epoch != s.CurrentEpoch {
			nextZxid = pb.ZxidFragment{Epoch: s.CurrentEpoch, Counter: 1}
		}

		msg := &pb.ZabRequest{
			Transaction: &pb.Transaction{
				Zxid:  nextZxid.Raw(),
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

		log.Printf("server %d prepare to send message: %s", s.Id, msg.Transaction.Extract())
		copiedMsg := &pb.ZabRequest{
			Transaction: msg.Transaction,
			RequestType: msg.RequestType,
		}
		for idx := range s.Zab.FollowerEpochs {
			go func(i int) {
				log.Printf("server %d send ZabRequest to %d", s.Id, i)
				if r, err := SendGrpc(pb.NodeClient.SendZabRequest, s, i, copiedMsg, *maxTimeout*10); err == nil && r.Accept {
					done <- true
				} else {
					fmt.Printf("server %d send ZabRequest to %d failed: %v\n", s.Id, i, err)
				}
			}(idx)
		}
		// wait for quorum
		for i := 0; i < majoritySize; i++ {
			<-done
		}
		log.Printf("server %d get quorum", s.Id)

		transaction := msg.Transaction.Extract()
		executedPath := transaction.Path
		var err error
		if in.OperationType != pb.OperationType_SYNC {
			transactionFrag := msg.Transaction.ExtractLog()
			if in.OperationType == pb.OperationType_DELETE || in.OperationType == pb.OperationType_UPDATE {
				watchesTriggered := s.Data.CheckWatchTrigger(&transactionFrag)
				for _, watch := range watchesTriggered {
					TriggerWatch(watch, in.OperationType)
				}
			}
			if transaction.Type == pb.OperationType_UPDATE {
				fmt.Printf("Transaction's Data update with %s\n", transaction.Data)
			}
			transactionFrag.Path, err = s.ZabDeliver(msg.Transaction.Extract())
			executedPath = transactionFrag.Path
			if in.OperationType == pb.OperationType_WRITE {
				watchesTriggered := s.Data.CheckWatchTrigger(&transactionFrag)
				for i := 0; i < len(watchesTriggered); i++ {
					TriggerWatch(watchesTriggered[i], in.OperationType)
				}
			}
		}

		msg.RequestType = pb.RequestType_ANNOUNCEMENT
		for idx := range s.Zab.FollowerEpochs {
			if idx == s.Id {
				continue
			}

			if _, serr := SendGrpc(pb.NodeClient.SendZabRequest, s, idx, msg, *maxTimeout); serr == nil {
				slog.Info("HandleClientCUD", "s", s.Id, "to", idx, "type", msg.RequestType, "txn", msg.Transaction.Extract())
			}
		}

		accepted := true
		// commit
		return &pb.CUDSResponse{Accept: &accepted, Path: &executedPath}, err
	default:
		log.Printf("server %d is in %s state", s.Id, state)
		accepted := false
		return &pb.CUDSResponse{Accept: &accepted, Path: &in.Path}, nil
	}

}

func (s *Server) GetExists(ctx context.Context, in *pb.GetExistsRequest) (*pb.GetExistsResponse, error) {
	node, err := s.StateVector.Data.GetNode(in.Path)
	if node == nil {
		return &pb.GetExistsResponse{Exists: false, Zxid: s.LastZxid.Inc().Raw()}, nil
	}
	if in.SetWatch {
		s.StateVector.Data.AddWatchToNode(in.Path, &pb.Watch{
			Path:       in.Path,
			Type:       pb.Exists,
			ClientAddr: pb.Addr{Host: in.ClientHost, Port: in.ClientPort},
		})
	}

	return &pb.GetExistsResponse{Exists: true, Zxid: s.LastZxid.Inc().Raw()}, err
}

func (s *Server) GetData(ctx context.Context, in *pb.GetDataRequest) (*pb.GetDataResponse, error) {
	data, err := s.StateVector.Data.GetData(in.Path)
	if in.SetWatch {
		s.StateVector.Data.AddWatchToNode(in.Path, &pb.Watch{
			Path:       in.Path,
			Type:       pb.GetData,
			ClientAddr: pb.Addr{Host: in.ClientHost, Port: in.ClientPort},
		})
	}

	return &pb.GetDataResponse{Data: data, Zxid: s.LastZxid.Inc().Raw()}, err
}

func (s *Server) GetChildren(ctx context.Context, in *pb.GetChildrenRequest) (*pb.GetChildrenResponse, error) {
	children, err := s.StateVector.Data.GetNodeChildren(in.Path)
	slog.Info("CONNECTED", "s", s.Id, "host", in.ClientHost, "port", in.ClientPort)
	if in.SetWatch {
		s.StateVector.Data.AddWatchToNode(in.Path, &pb.Watch{
			Path:       in.Path,
			Type:       pb.GetChildren,
			ClientAddr: pb.Addr{Host: in.ClientHost, Port: in.ClientPort},
		})
	}
	//Type conversion
	out := make([]string, 0)
	for key := range children {
		out = append(out, key)
	}

	return &pb.GetChildrenResponse{Children: out, Zxid: s.LastZxid.Inc().Raw()}, err
}
