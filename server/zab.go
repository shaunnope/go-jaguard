package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

// grpc calls

func (s *Server) InformLeader(ctx context.Context, in *pb.FollowerInfo) (*pb.Ping, error) {
	if s.State != LEADING {
		return nil, errors.New("not leader")
	}
	log.Printf("%d received follower info from %d", s.Id, in.Id)

	if !s.Leader.HasQuorum {
		s.Leader.Lock()
		defer s.Leader.Unlock()
		// update leader table
		s.Leader.FollowerEpochs[int(in.Id)] = int(in.LastZxid.Epoch)

		jsonData, err := json.MarshalIndent(s.Leader.FollowerEpochs, "", "  ")
		if err != nil {
			log.Fatalf("JSON Marshaling failed: %s", err)
		}
		fmt.Println(string(jsonData))

		if len(s.Leader.FollowerEpochs) > len(config.Servers)/2 {
			fmt.Println("Quoram has been reached")
			s.Leader.HasQuorum = true
			s.Leader.QuorumReady <- true
		}

	} else {
		// TODO: send NEWEPOCH and NEWLEADER to new follower
		_, err := SendGrpc[*pb.NewEpoch, *pb.AckEpoch](pb.NodeClient.ProposeEpoch, s, int(in.Id), &pb.NewEpoch{Epoch: int64(s.CurrentEpoch)}, *maxTimeout)
		if err != nil {
			log.Printf("%d error sending new epoch to %d: %v", s.Id, in.Id, err)
		}
		msg := &pb.NewLeader{Epoch: int64(s.CurrentEpoch), History: s.History.Raw()}
		r, err := SendGrpc[*pb.NewLeader, *pb.AckLeader](
			pb.NodeClient.ProposeLeader,
			s, int(in.Id), msg,
			*maxTimeout)
		if err != nil {
			log.Printf("%d error sending new leader to %d: %v", s.Id, in.Id, err)
		}

		// send COMMIT to new follower
		s.Leader.Lock()
		defer s.Leader.Unlock()
		// update leader table
		s.Leader.FollowerEpochs[int(in.Id)] = int(r.Epoch)

	}

	return &pb.Ping{Data: int64(s.CurrentEpoch)}, nil
}

func (s *Server) ProposeEpoch(ctx context.Context, in *pb.NewEpoch) (*pb.AckEpoch, error) {
	if s.State != FOLLOWING {
		return nil, errors.New("not follower")
	}

	if int(in.Epoch) > s.AcceptedEpoch {
		log.Printf("%d accepted new epoch: %d", s.Id, in.Epoch)
		s.AcceptedEpoch = int(in.Epoch)
		res := &pb.AckEpoch{}
		return res, nil
	}
	log.Printf("%d rejected new epoch: %d", s.Id, in.Epoch)

	if int(in.Epoch) < s.AcceptedEpoch {
		go func() {
			// FIXME: issue - Election requires lock, but is held by Discovery
			// vote := s.FastElection(*maxTimeout)

		}()
	}

	// goto phase 2
	defer s.ZabSync()

	return nil, errors.New("epoch not accepted")
}

// FIXME: incomplete
func (s *Server) ProposeLeader(ctx context.Context, in *pb.NewLeader) (*pb.AckLeader, error) {
	if s.State != FOLLOWING {
		return nil, errors.New("not follower")
	}

	if s.AcceptedEpoch == int(in.Epoch) {
		//atomically
		s.CurrentEpoch = int(in.Epoch)
		transactions := pb.Transactions(in.History).ExtractAll()
		s.History = append(s.History, transactions...)
	}

	if int(in.Epoch) > s.AcceptedEpoch {
		log.Printf("%d accepted new leader: %d", s.Id, in.Epoch)
		s.AcceptedEpoch = int(in.Epoch)
		res := &pb.AckLeader{}
		return res, nil
	}
	log.Printf("%d rejected new leader: %d", s.Id, in.Epoch)

	if int(in.Epoch) < s.AcceptedEpoch {
		go func() {
			// TODO
		}()
	}

	// goto phase 2
	defer s.ZabSync()

	return nil, errors.New("leader not accepted")
}

// gRPC call for Phase 3: Broadcast
func (s *Server) SendZabRequest(ctx context.Context, in *pb.ZabRequest) (*pb.ZabAck, error) {
	isLeader := s.GetState() == LEADING

	// Handle incoming CreateRequest
	switch in.RequestType {
	case pb.RequestType_PROPOSAL:
		if isLeader {
			return nil, errors.New("leaders shouldnt get proposals")
		}
		// log.Printf("%d received proposal: %s", s.Id, in.Transaction.ExtractLogString())

		if int(in.Transaction.Zxid.Epoch) == s.CurrentEpoch {
			// send ack to leader
			return &pb.ZabAck{Request: in}, nil
		}

		return nil, errors.New("proposal not accepted")

	case pb.RequestType_ANNOUNCEMENT:
		if isLeader {
			return nil, errors.New("leaders shouldnt get announcements")
		}
		// Commit the change to history
		s.History = append(s.History, in.Transaction.ExtractLog())
		s.LastZxid = in.Transaction.Extract().Zxid
		// log.Printf("server %d get commit message\n", s.Id)
		// log.Printf("Follower's History: %+v", s.History)

		// TODO: for each commit (announcement), wait until all earlier proposals are committed
		// then, commit

		// @Shi Hiui: Follower commit change on local copy
		// LEADER need to execute request
		transaction := in.Transaction.Extract()
		s.Lock()
		defer s.Unlock()
		log.Printf("server %d update local copy", s.Id)
		s.HandleOperation(transaction)

		return &pb.ZabAck{Request: in}, nil

	case pb.RequestType_CLIENT:
		// if leader send proposal to all followers in for loop (rpc)
		// since its rpc, leader will monitor for responses and decide whether to commit/announce
		// if follower forward to leader, do nothing with response (rpc)
		if isLeader {
			log.Printf("server %d received client request: %v", s.Id, in.Transaction)
			// TODO: verify version
			// propose to all
			s.Lock()
			msg := &pb.ZabRequest{
				Transaction: in.Transaction.WithZxid(s.LastZxid.Inc()),
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

			// jsonData, err := json.MarshalIndent(s.Leader.FollowerEpochs, "", "  ")
			// if err != nil {
			// 	log.Fatalf("JSON Marshaling failed: %s", err)
			// }

			// fmt.Println(string(jsonData))
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
			s.History = append(s.History, in.Transaction.ExtractLog())
			// s.History = append(s.History, in.Transaction.Extract())
			s.LastZxid = msg.Transaction.Extract().Zxid

			// @Shi Hui: Leader commit change on local copy
			// LEADER need to execute request
			transaction := in.Transaction.Extract()
			s.HandleOperation(transaction)

			msg.RequestType = pb.RequestType_ANNOUNCEMENT
			for idx := range s.Leader.FollowerEpochs {
				if idx == s.Id {
					continue
				}

				log.Printf("server %d sending ANNOUNCEMENT TO server %d with message %v", s.Id, idx, msg)
				SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, idx, msg, *maxTimeout)
			}
			defer s.Unlock()
			return &pb.ZabAck{Request: in}, nil
		} else {
			log.Printf("server %d forwarding request to %d", s.Id, s.Vote.Id)
			// todo verify version
			// forward to leader
			SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, s.Vote.Id, in, *maxTimeout)
			return &pb.ZabAck{Request: in}, nil
		}
	}

	switch in.Transaction.Type {
	case pb.OperationType_WRITE:

	case pb.OperationType_DELETE:
	}

	return nil, errors.New("zab request not accepted")
}

func (s *Server) HandleOperation(transaction pb.TransactionFragment) {

	switch transaction.Type {
	case pb.OperationType_WRITE:
		//zxid := pb.ZxidFragment{int(in.Transaction.Zxid.Epoch), int(in.Transaction.Zxid.Counter)}
		//ephemeral owner??
		_, err := s.StateVector.Data.CreateNode(transaction.Path, transaction.Data, false, 1, transaction.Zxid, false)
		if err != nil {
			log.Println(err)
		}
		// log.Printf("server %d created znode @ PATH: %s", s.Id, path)
		fileName := fmt.Sprintf("server%d.txt", s.Id)
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			// Handle the error
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()
		fmt.Fprintln(file, s.Id)
		currentTime := time.Now().Format("2006/01/02 15:04:05")
		fmt.Fprintf(file, "%s Updated Tree\n", currentTime)
		printTree(s.StateVector.Data, file, "/", "")
		fmt.Fprintln(file, "")

	// case pb.OperationType_UPDATE:
	// 	//Check node exists? Done here or in add into setdata method?
	// 	//zxid := pb.ZxidFragment{int(in.Transaction.Zxid.Epoch), int(in.Transaction.Zxid.Counter)}
	// 	//Version??
	// 	s.StateVector.Data.SetData(transaction.Path, transaction.Data, 0, transaction.Zxid)
	// 	log.Printf("node at %s updated", transaction.Path)

	case pb.OperationType_DELETE:
		outcome, err := s.StateVector.Data.DeleteNode(transaction.Path, 0)
		if err != nil {
			log.Println(err)
		}
		log.Println(outcome)

	}

}

// end grpc calls

// Phase 1 of ZAB
func (s *Server) Discovery() {
	// why lock here?
	// state: read
	// vote: read
	// history?: write

	s.Lock()
	defer s.Unlock()

	switch s.State {
	case FOLLOWING:
		msg := &pb.FollowerInfo{Id: int64(s.Id), LastZxid: &pb.Zxid{Epoch: int64(s.AcceptedEpoch), Counter: -1}}
		_, err := SendGrpc[*pb.FollowerInfo, *pb.Ping](pb.NodeClient.InformLeader, s, s.Vote.Id, msg, *maxTimeout)

		if err != nil {
			log.Printf("%d error sending follower info to %d: %v", s.Id, s.Vote.Id, err)
		} else {
			log.Printf("%d sent follower info to %d", s.Id, s.Vote.Id)
		}

	case LEADING:
		s.Leader.Reset()
		<-s.Leader.QuorumReady
		maxEpoch := -1
		for _, epoch := range s.Leader.FollowerEpochs {
			if epoch > maxEpoch {
				maxEpoch = epoch
			}
		}
		log.Printf("%d max epoch %d", s.Id, maxEpoch)

		mostRecent := &pb.AckEpoch{CurrentEpoch: -1, History: nil, LastZxid: &pb.Zxid{Epoch: -1, Counter: -1}}
		for idx := range s.Leader.FollowerEpochs {
			msg := &pb.NewEpoch{Epoch: int64(maxEpoch + 1)}
			r, err := SendGrpc[*pb.NewEpoch, *pb.AckEpoch](pb.NodeClient.ProposeEpoch, s, idx, msg, *maxTimeout)
			if err != nil {
				log.Printf("%d error sending new epoch to %d: %v", s.Id, idx, err)
			}
			// follower rejected
			if r == nil {
				return
			}
			if r.CurrentEpoch > mostRecent.CurrentEpoch {
				// if r.CurrentEpoch > mostRecent.CurrentEpoch || (r.CurrentEpoch == mostRecent.CurrentEpoch && !(r.LastZxid.Extract().LessThan(mostRecent.LastZxid.Extract()))) {
				mostRecent = r
			}
		}
		s.History = pb.Transactions(mostRecent.History).ExtractAll()
		// goto phase 2
		defer s.ZabSync()
	}
}

// Phase 2 of ZAB
func (s *Server) ZabSync() {
	// why lock here?
	// state: read
	// vote: read
	// history?: write

	s.Lock()
	defer s.Unlock()

	switch s.State {
	case FOLLOWING:

	}
}

// Phase 3 utils

func (s *Server) ZabCommit() {

	switch s.State {
	case FOLLOWING:

	}
}
