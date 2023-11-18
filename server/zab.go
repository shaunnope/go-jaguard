package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

// grpc calls
// TODO: verify message authors

// Send FollowInfo to leader. Discovery phase
func (s *Server) InformLeader(ctx context.Context, in *pb.FollowerInfo) (*pb.Ping, error) {
	if s.State != LEADING {
		// only leader should receive follower info
		return nil, errors.New("not leader")
	}
	slog.Info("FollowerInfo", "s", s.Id, "from", in.Id, "lastZxid", in.LastZxid.Extract())
	s.Zab.Lock()
	defer s.Zab.Unlock()

	if !s.Zab.HasQuorum {
		// leader has not received quorum. update leader table
		s.Zab.FollowerEpochs[int(in.Id)] = int(in.LastZxid.Epoch)

		jsonData, err := json.MarshalIndent(s.Zab.FollowerEpochs, "", "  ")
		if err != nil {
			log.Fatalf("JSON Marshaling failed: %s", err)
		}
		fmt.Println(string(jsonData))

		if len(s.Zab.FollowerEpochs) > len(config.Servers)/2 {
			slog.Debug("Got quorum", "s", s.Id)
			s.Zab.HasQuorum = true
			s.Zab.QuorumReady <- true
		}

	} else {
		// phase 3
		go func() {
			slog.Info("Waiting for Broadcast to begin", "s", s.Id, "from", in.Id)
			_, ok := <-s.Zab.BroadcastReady
			if ok {
				panic(fmt.Sprintf("%d -> %d: unexpected data on BroadcastReady", in.Id, s.Id))
			}
			slog.Info("Broadcast ready", "s", s.Id, "from", in.Id)
			s.Zab.Lock()
			defer s.Zab.Unlock()
			// TODO: send NEWEPOCH and NEWLEADER to new follower
			SendGrpc[*pb.NewEpoch, *pb.AckEpoch](pb.NodeClient.ProposeEpoch, s, int(in.Id), &pb.NewEpoch{Epoch: int64(s.CurrentEpoch)}, *maxTimeout)

			msg := &pb.NewLeader{Epoch: int64(s.CurrentEpoch), History: s.History.Raw()}
			if r, err := SendGrpc[*pb.NewLeader, *pb.AckLeader](
				pb.NodeClient.ProposeLeader,
				s, int(in.Id), msg,
				*maxTimeout); err == nil {
				SendGrpc[*pb.Ping, *pb.Ping](
					pb.NodeClient.Commit,
					s, int(in.Id), &pb.Ping{},
					*maxTimeout)
				// update leader table
				s.Zab.FollowerEpochs[int(in.Id)] = int(r.Epoch)
			}

		}()
	}

	return &pb.Ping{Data: int64(s.CurrentEpoch)}, nil
}

func (s *Server) ProposeEpoch(ctx context.Context, in *pb.NewEpoch) (*pb.AckEpoch, error) {
	if s.State != FOLLOWING {
		// only follower should receive new epoch
		return nil, errors.New("not follower")
	}

	if int(in.Epoch) > s.AcceptedEpoch {
		log.Printf("%d accepted new epoch: %d", s.Id, in.Epoch)
		// TODO: store to non-volatile memory
		s.AcceptedEpoch = int(in.Epoch)
		res := &pb.AckEpoch{CurrentEpoch: int64(s.AcceptedEpoch), History: s.History.Raw(), LastZxid: s.LastZxid.Raw()}
		return res, nil
	}
	log.Printf("%d rejected new epoch: %d", s.Id, in.Epoch)

	if int(in.Epoch) < s.AcceptedEpoch {
		// s.Lock()
		// s.Unlock()
		s.FastElection(*maxTimeout)
		go func() {
			// FIXME: issue - Election requires lock, but is held by Discovery
			// vote := s.FastElection(*maxTimeout)

		}()
	}

	return nil, errors.New("epoch not accepted")
}

// Send NewLeader to follower. Sync phase
func (s *Server) ProposeLeader(ctx context.Context, in *pb.NewLeader) (*pb.AckLeader, error) {
	if s.State != FOLLOWING {
		return nil, errors.New("not follower")
	}

	if s.AcceptedEpoch == int(in.Epoch) {
		// TODO: atomically
		s.CurrentEpoch = int(in.Epoch)

		// accept transactions in order of zxid, update history
		// TODO: store to non-volatile memory
		s.History = pb.Transactions(in.History).ExtractAll()

		log.Printf("%d accepted new leader: %d", s.Id, in.Epoch)
		s.AcceptedEpoch = int(in.Epoch)
		res := &pb.AckLeader{}
		return res, nil
	}
	log.Printf("%d rejected new leader: %d", s.Id, in.Epoch)

	if int(in.Epoch) < s.AcceptedEpoch {
		// s.Lock()
		// s.Unlock()
		s.FastElection(*maxTimeout)
		go func() {
			// FIXME: issue - Election requires lock, but is held by Discovery
			// vote := s.FastElection(*maxTimeout)

		}()
	}

	return nil, errors.New("leader not accepted")
}

// Commit all outstanding transactions. Sync phase
func (s *Server) Commit(ctx context.Context, in *pb.Ping) (*pb.Ping, error) {
	if s.State != FOLLOWING {
		return nil, errors.New("not follower")
	}

	for _, transaction := range s.History {
		// TODO: deliver
		transaction.Committed = true
	}
	slog.Debug("Committed", "s", s.Id, "epoch", s.CurrentEpoch)

	return &pb.Ping{}, nil
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
		_, err := s.HandleOperation(transaction)

		return &pb.ZabAck{Request: in}, err
	}

	switch in.Transaction.Type {
	case pb.OperationType_WRITE:

	case pb.OperationType_DELETE:
	}

	return nil, errors.New("zab request not accepted")
}

func (s *Server) HandleOperation(transaction pb.TransactionFragment) (string, error) {

	switch transaction.Type {
	case pb.OperationType_WRITE:
		//zxid := pb.ZxidFragment{int(in.Transaction.Zxid.Epoch), int(in.Transaction.Zxid.Counter)}
		//ephemeral owner??
		_, err := s.StateVector.Data.CreateNode(transaction.Path, transaction.Data, transaction.Flags.IsEphemeral, 1, transaction.Zxid, transaction.Flags.IsSequential)
		if err != nil {
			log.Println(err)
			fmt.Println("Error:", err)
			return "", err
		}
		// log.Printf("server %d created znode @ PATH: %s", s.Id, path)
		fileName := fmt.Sprintf("server%d.txt", s.Id)
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			// Handle the error
			fmt.Println("Error opening file:", err)
			return "", err
		}
		defer file.Close()
		fmt.Fprintln(file, s.Id)
		currentTime := time.Now().Format("2006/01/02 15:04:05")
		fmt.Fprintf(file, "%s Updated Tree\n", currentTime)
		printTree(s.StateVector.Data, file, "/", "")
		fmt.Fprintln(file, "")
		return "Success", nil

	case pb.OperationType_UPDATE:
		//Check node exists? Done here or in add into setdata method?
		//zxid := pb.ZxidFragment{int(in.Transaction.Zxid.Epoch), int(in.Transaction.Zxid.Counter)}
		//Version??
		s.StateVector.Data.SetData(transaction.Path, transaction.Data, 0, transaction.Zxid)
		log.Printf("node at %s updated", transaction.Path)
		return "Success", nil

	case pb.OperationType_DELETE:
		outcome, err := s.StateVector.Data.DeleteNode(transaction.Path, transaction.Zxid.Raw().Counter)
		if err != nil {
			log.Println(err)
			return "", err
		}
		log.Println(outcome)
		return "Success", nil
	}
	return "", &json.UnsupportedValueError{Str: "Unsupported pb.OperationType"}
}

// end grpc calls

// Routine to maintain Zab Session
func (s *Server) ZabStart() {
	if vote := s.FastElection(*maxTimeout); vote.Id == -1 {
		slog.Error("Election failed", "s", s.Id)
	} else {
		s.Discovery()
	}

}

// Phase 1 of ZAB
func (s *Server) Discovery() {
	s.Lock()

	switch s.State {
	case FOLLOWING:
		defer s.Unlock()
		msg := &pb.FollowerInfo{Id: int64(s.Id), LastZxid: &pb.Zxid{Epoch: int64(s.AcceptedEpoch), Counter: -1}}
		SendGrpc[*pb.FollowerInfo, *pb.Ping](pb.NodeClient.InformLeader, s, s.Vote.Id, msg, *maxTimeout)

	case LEADING:
		s.Zab.Reset()
		<-s.Zab.QuorumReady
		s.Zab.Lock()
		maxEpoch := -1
		for _, epoch := range s.Zab.FollowerEpochs {
			if epoch > maxEpoch {
				maxEpoch = epoch
			}
		}
		slog.Info("Max epoch", "s", s.Id, "epoch", maxEpoch)
		s.CurrentEpoch = maxEpoch + 1

		mostRecent := &pb.AckEpoch{CurrentEpoch: -1, History: nil, LastZxid: &pb.Zxid{Epoch: -1, Counter: -1}}
		for idx := range s.Zab.FollowerEpochs {
			msg := &pb.NewEpoch{Epoch: int64(s.CurrentEpoch)}
			r, _ := SendGrpc[*pb.NewEpoch, *pb.AckEpoch](pb.NodeClient.ProposeEpoch, s, idx, msg, *maxTimeout)
			// follower rejected
			if r == nil {
				slog.Info("follower rejected", "s", s.Id, "f", idx)
				return
			}
			// TODO: extract comparison to function
			if r.CurrentEpoch > mostRecent.CurrentEpoch || (r.CurrentEpoch == mostRecent.CurrentEpoch && !(r.LastZxid.Extract().LessThan(mostRecent.LastZxid.Extract()))) {
				mostRecent = r
			}
		}

		log.Printf("%d most recent: %d", s.Id, mostRecent.CurrentEpoch)
		// TODO: store to non-volatile memory
		s.History = pb.Transactions(mostRecent.History).ExtractAll()
		s.Zab.Unlock()

		// goto phase 2
		defer s.ZabSync()
	}
}

// Phase 2 of ZAB
func (s *Server) ZabSync() {
	if s.State != LEADING {
		// only leader runs ZabSync()
		// followers wait for leader to send NewLeader/Commit
		return
	}
	defer s.Unlock()

	// Send newleader to followers
	majority := len(s.Zab.FollowerEpochs)/2 + 1
	ready := make(chan bool, majority)
	for i := range s.Zab.FollowerEpochs {
		go func(idx int) {
			msg := &pb.NewLeader{Epoch: int64(s.CurrentEpoch), History: s.History.Raw()}
			SendGrpc[*pb.NewLeader, *pb.AckLeader](
				pb.NodeClient.ProposeLeader,
				s, idx, msg,
				*maxTimeout)
			ready <- true
		}(i)

	}

	// wait for quorum
	for i := 0; i < majority; i++ {
		<-ready
	}
	// send commit to all followers
	for i := range s.Zab.FollowerEpochs {
		SendGrpc[*pb.Ping, *pb.Ping](
			pb.NodeClient.Commit,
			s, i, &pb.Ping{},
			*maxTimeout)
	}

	// phase 3
	close(s.Zab.BroadcastReady)
}

// Phase 3 utils

func (s *Server) ZabCommit() {

	switch s.State {
	case FOLLOWING:

	}
}
