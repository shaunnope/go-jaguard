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
		r, err := SendGrpc[*pb.CUDSRequest, *pb.CUDSResponse](pb.NodeClient.HandleClientCUDS, s, s.Vote.Id, in, *maxTimeout)

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

		log.Printf("server %d prepare to send message: %s", s.Id, msg.Transaction.LogString())
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

		transaction := msg.Transaction.Extract()
		var err error
		if in.OperationType != pb.OperationType_SYNC {
			if in.OperationType == pb.OperationType_DELETE || in.OperationType == pb.OperationType_UPDATE {
				transactionFrag := msg.Transaction.ExtractLog()
				watchesTriggered := s.Data.CheckWatchTrigger(&transactionFrag)
				for i := 0; i < len(watchesTriggered); i++ {
					TriggerWatch(watchesTriggered[i], in.OperationType)
				}
			}
			if transaction.Type == pb.OperationType_UPDATE {
				fmt.Printf("Transaction's Data update with %s", transaction.Data)
			}
			err = s.ZabDeliver(msg.Transaction.Extract())
			if in.OperationType == pb.OperationType_WRITE {
				transactionFrag := msg.Transaction.ExtractLog()
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
			slog.Info("HandleClientCUD", "s", s.Id, "to", idx, "type", msg.RequestType, "txn", msg.Transaction.Extract())
			SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, idx, msg, *maxTimeout)
		}

		accepted := true
		// commit
		return &pb.CUDSResponse{Accept: &accepted}, err
	default:
		log.Printf("server %d is in %s state", s.Id, state)
		accepted := false
		return &pb.CUDSResponse{Accept: &accepted}, nil
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

	fmt.Printf("Host:%s, Port:%s\n", in.ClientHost, in.ClientPort)
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
