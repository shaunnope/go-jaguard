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
		slog.Error("Received FollowerInfo", "s", s.Id, "err", "not leader")
		s.Reelect <- true
		return nil, errors.New("not leader")
	}
	s.Zab.FollowerInfoQ <- in
	return &pb.Ping{Data: int64(s.CurrentEpoch)}, nil
}

func (s *Server) ProposeEpoch(ctx context.Context, in *pb.NewEpoch) (*pb.AckEpoch, error) {
	if s.State != FOLLOWING {
		// only follower should receive new epoch
		return nil, errors.New("not follower")
	}

	// TODO: verify if equality is ok
	if int(in.Epoch) >= s.AcceptedEpoch {
		slog.Debug("Accept Epoch", "s", s.Id, "epoch", in.Epoch)
		s.SetAcceptedEpoch(int(in.Epoch))
		res := &pb.AckEpoch{CurrentEpoch: int64(s.AcceptedEpoch), History: s.History.Raw(), LastZxid: s.LastZxid.Raw()}
		return res, nil
	}
	log.Printf("%d rejected new epoch: %d", s.Id, in.Epoch)

	if int(in.Epoch) < s.AcceptedEpoch {
		go func() {
			slog.Info("EPOCH LT", "s", s.Id, "epoch", in.Epoch)
			s.Reelect <- true
		}()
	}

	return nil, errors.New("epoch not accepted")
}

// Send NewLeader to follower. Sync phase
func (s *Server) ProposeLeader(ctx context.Context, in *pb.NewLeader) (*pb.AckLeader, error) {
	if s.State != FOLLOWING {
		return nil, errors.New("not follower")
	}

	switch in.LastZxid {
	case nil:
		// phase 2
		slog.Info("If nil", "s", s.Id, "s.AcceptedEpoch", s.AcceptedEpoch, "epoch", in.Epoch)
		if s.AcceptedEpoch <= int(in.Epoch) {
			// TODO: atomically (what does this mean?)
			s.CurrentEpoch = int(in.Epoch)

			// TODO: accept transactions in order of zxid

			// update history
			// TODO: store to non-volatile memory
			s.ReplaceHistory(in.History)

			slog.Debug("Accept Leader", "s", s.Id, "epoch", in.Epoch)
			res := &pb.AckLeader{}
			return res, nil
		}
		log.Printf("%d rejected new leader: %d", s.Id, in.Epoch)

	default:
		// follower fault
		if in.LastZxid.Extract().Epoch < s.LastZxid.Epoch {
			goto Reelect
		}
		// TODO: implement DIFF, TRUNC. For now, just SNAP
		// copy the snapshot received and commit the changes
		// update history
		// TODO: store to non-volatile memory
		waitHistory = false
		s.ReplaceHistory(in.History)

		if err := s.ZabDeliverAll(); err != nil {
			return nil, err
		}
		slog.Debug("Committed (Rec)", "s", s.Id, "zxid", s.LastZxid)
		slog.Debug("Accept Leader", "s", s.Id, "epoch", in.Epoch)
		s.SetAcceptedEpoch(int(in.Epoch))
		res := &pb.AckLeader{Epoch: int64(s.AcceptedEpoch)}
		return res, nil

	}
Reelect:
	slog.Error("Leader rejected", "s", s.Id, "epoch", in.Epoch)
	slog.Error("in.Epoch", in.Epoch, "smaller", int(in.Epoch) < s.LastZxid.Epoch, "server", s.Id, "Server Last Zxid Epoch", s.LastZxid.Epoch)
	go func() {
		s.Reelect <- true
	}()

	return nil, errors.New("leader not accepted")
}

// Commit all outstanding transactions. Sync phase
func (s *Server) Commit(ctx context.Context, in *pb.ZabRequest) (*pb.Ping, error) {
	if s.State != FOLLOWING {
		return nil, errors.New("not follower")
	}

	if in.Transaction == nil {
		// Commit all outstanding transactions
		if err := s.ZabDeliverAll(); err != nil {
			return nil, err
		}
		slog.Info("Committed", "s", s.Id, "zxid", s.LastZxid, "history", s.History)

	} else {
		slog.Info("COMMIT", "s", s.Id, "txn", in.Transaction)
		if in.Transaction.Zxid == nil {
			panic("nil zxid in commit")
		}
		go func() {
			// TODO: deliver transaction when ready
		}()
	}

	return &pb.Ping{}, nil
}

// gRPC call for Phase 3: Broadcast. Send a ZAB request to a follower.
func (s *Server) SendZabRequest(ctx context.Context, in *pb.ZabRequest) (*pb.ZabAck, error) {
	if s.GetState() != FOLLOWING {
		return nil, errors.New("not follower")
	}
	// only run when leader is ready
	// FIXME: should wait for leader to be ready, but currently blocks
	// if _, ok := <-s.Zab.BroadcastReady; ok {
	// 	panic(fmt.Sprintf("%d: unexpected data on BroadcastReady", s.Id))
	// }

	slog.Info("Server", s.Id, "Get request", in.RequestType)

	// Handle incoming CreateRequest
	switch in.RequestType {
	case pb.RequestType_PROPOSAL:
		slog.Info("Proposal", "s", s.Id, "txn", in.Transaction.Extract())

		// TODO: verify version
		// send ack to leader
		return &pb.ZabAck{
			Request: in,
			Accept:  int(in.Transaction.Zxid.Epoch) == s.CurrentEpoch,
		}, nil

	case pb.RequestType_ANNOUNCEMENT:
		// TODO: for each commit (announcement), wait until all earlier proposals are committed
		// then, commit
		transaction := in.Transaction.Extract()
		s.Lock()
		defer s.Unlock()
		log.Printf("server %d update local copy", s.Id)
		if transaction.Type != pb.OperationType_SYNC {
			if transaction.Type == pb.OperationType_DELETE || transaction.Type == pb.OperationType_UPDATE {
				watchesTriggered := s.Data.CheckWatchTrigger(&transaction)
				for i := 0; i < len(watchesTriggered); i++ {
					TriggerWatch(watchesTriggered[i], transaction.Type)
				}
			}
			path, err := s.ZabDeliver(transaction)
			transaction.Path = path
			if err != nil {
				slog.Error("ZabDeliver", "s", s.Id, "err", "failed to deliver", "txn", transaction)
				return nil, err

			} else if transaction.Type == pb.OperationType_WRITE {
				watchesTriggered := s.Data.CheckWatchTrigger(&transaction)
				for i := 0; i < len(watchesTriggered); i++ {
					TriggerWatch(watchesTriggered[i], transaction.Type)
				}
			}
		}
		return &pb.ZabAck{Request: in}, nil
	}

	return nil, errors.New("zab request not accepted")
}

// Process a transaction
func (s *Server) HandleOperation(transaction pb.TransactionFragment) (string, error) {
	switch transaction.Type {
	case pb.OperationType_WRITE:
		path, err := s.Data.CreateNode(transaction.Path, transaction.Data, transaction.Flags.IsEphemeral, 1, transaction.Zxid, transaction.Flags.IsSequential)
		if err != nil {
			log.Println(err)
			fmt.Println("Error:", err)
			return "", err
		}
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
		return path, nil

	case pb.OperationType_UPDATE:
		//Check node exists? Done here or in add into setdata method?
		//zxid := pb.ZxidFragment{int(in.Transaction.Zxid.Epoch), int(in.Transaction.Zxid.Counter)}
		//Version??
		_, err := s.StateVector.Data.SetData(transaction.Path, transaction.Data, 0, transaction.Zxid)
		if err != nil {
			log.Println(err)
			return "", err
		}
		log.Printf("node at %s updated", transaction.Path)
		return transaction.Path, nil

	case pb.OperationType_DELETE:
		outcome, err := s.StateVector.Data.DeleteNode(transaction.Path, transaction.Zxid.Raw().Counter)
		if err != nil {
			log.Println(err)
			return "", err
		}
		log.Println(outcome)
		return transaction.Path, nil
	}
	return "", &json.UnsupportedValueError{Str: "Unsupported pb.OperationType"}
}

// end grpc calls

// Async listener to handle Phase 1 FollowerInfo
func (s *Server) ProcessFollowerInfo() {
	for {
		select {
		case _, ok := <-s.Zab.Abort:
			if ok {
				panic(fmt.Sprintf("%d: unexpected data on Abort", s.Id))
			}
			return
		case in := <-s.Zab.FollowerInfoQ:
			s.Zab.Lock()
			slog.Info("FollowerInfo", "s", s.Id, "from", in.Id, "lastZxid", in.LastZxid.Extract(), "counter", int(in.LastZxid.Counter))
			switch int(in.LastZxid.Counter) {
			case -1:
				// phase 1
				if !s.Zab.HasQuorum {
					// leader has not received quorum. update leader table
					s.Zab.FollowerEpochs[int(in.Id)] = int(in.LastZxid.Epoch)

					jsonData, err := json.MarshalIndent(s.Zab.FollowerEpochs, "", "  ")
					if err != nil {
						log.Fatalf("JSON Marshaling failed: %s", err)
					}
					fmt.Println(string(jsonData))

					if len(s.Zab.FollowerEpochs) > len(config.Servers)/2 {
						slog.Info("Got quorum", "s", s.Id)
						s.Zab.HasQuorum = true
						s.Zab.QuorumReady <- true
					}

				} else {
					// phase 3
					go func() {
						slog.Debug("Waiting for Broadcast", "s", s.Id, "from", in.Id)
						s.WaitForBroadcast()
						slog.Info("Broadcast ready", "s", s.Id, "from", in.Id)

						s.Zab.Lock()
						defer s.Zab.Unlock()
						to := int(in.Id)
						// send NEWEPOCH and NEWLEADER to new follower
						SendGrpc(pb.NodeClient.ProposeEpoch, s, to, &pb.NewEpoch{Epoch: int64(s.CurrentEpoch)}, *maxTimeout)

						msg := &pb.NewLeader{Epoch: int64(s.LastZxid.Epoch) + 1, History: s.History.Raw()}
						slog.Info("Phase 3")
						if r, err := SendGrpc(pb.NodeClient.ProposeLeader,
							s, to, msg, *maxTimeout,
						); err == nil {
							SendGrpc(pb.NodeClient.Commit,
								s, to, &pb.ZabRequest{}, *maxTimeout,
							)
							// update leader table
							s.Zab.FollowerEpochs[to] = int(r.Epoch)
						}

					}()
				}

			default:
				to := int(in.Id)
				// follower recovery
				// 1. Send NEWLEADER to follower + history (SNAP)
				// slog.Info("New Node Client", s.Connections[to])
				msg := &pb.NewLeader{LastZxid: s.LastZxid.Raw(), History: s.History.Raw()}
				slog.Info("Process Follower Info")
				if r, err := SendGrpc(pb.NodeClient.ProposeLeader,
					s, to, msg, *maxTimeout*10); err == nil {
					// update leader table (follower alr accepted lastzxid epoch)
					// s.Zab.FollowerEpochs[int(in.Id)] = s.LastZxid.Epoch
					s.Zab.FollowerEpochs[to] = int(r.Epoch)
					slog.Info("F (Rec)", "s", s.Id, "F", in.Id, "epoch", r.Epoch, "history", s.History)
				}
				// 2. Send one of SNAP, DIFF, TRUNC to follower
				// TODO: will just send SNAP for now (above)
			}
			s.Zab.Unlock()
		}
	}
}

var (
	waitHistory = false
)

func (s *Server) Startup() {
	slog.Info("Startup", "s", s.Id)
	s.WaitForLive()
	go s.Heartbeat()

	s.Discovery()
	for waitHistory {
		s.ZabRecover()
		// msg := &pb.FollowerInfo{Id: int64(s.Id), LastZxid: s.LastZxid.Raw()}
		// SendGrpc(pb.NodeClient.InformLeader, s, s.Vote.Id, msg, *maxTimeout)
		time.Sleep(time.Second)
	}
	slog.Info("Startup complete", "s", s.Id)
}

// Routine to start Zab Session
func (s *Server) ZabStart(t0 int) error {
	// time.Sleep(time.Duration(10000) * time.Millisecond)
	if s.ZabRecover() == nil {
		slog.Info("Recovered", "s", s.Id)
	} else if vote := s.FastElection(t0); vote.Id == -1 {
		slog.Error("Election failed", "s", s.Id)
		return errors.New("failed to elect leader")
	} else {
		slog.Info("Elected", "s", s.Id, "L", vote.Id)
	}

	s.Startup()
	return nil
}

func (s *Server) ZabRecover() error {
	s.Lock()
	defer s.Unlock()
	if err := s.LoadStates(); err != nil {
		return err
	} else if s.State == LEADING {
		// Load successfully
		// Ensure that reload cannot
		s.State = FOLLOWING
		s.Vote.Id = 5
	}
	slog.Info("Recovery", "s", s.Id, "State", s.State)

	switch s.State {
	case LEADING:
		msg := &pb.Ping{Data: int64(s.LastZxid.Epoch)}
		for idx := range config.Servers {
			if idx == s.Id {
				continue
			}
			if r, err := SendGrpc(pb.NodeClient.GetLeaderInfo,
				s, idx, msg, *maxTimeout,
			); err == nil {
				if int(r.Id) == -1 || int(r.Id) == s.Id {
					return errors.New("leader recovery")
				}
				vote := &pb.VoteFragment{Id: int(r.Id), LastZxid: r.LastZxid.Extract()}
				if err := s.SaveState(data_STATE, vote.Marshal()); err != nil {
					return err
				} else {
					s.State = FOLLOWING
					s.Vote = *vote
				}
				return nil
			}
		}
		return errors.New("leader recovery")
	case FOLLOWING:
		msg := &pb.FollowerInfo{Id: int64(s.Id), LastZxid: s.LastZxid.Raw()}
		if _, err := SendGrpc(pb.NodeClient.InformLeader,
			s, s.Vote.Id, msg, *maxTimeout,
		); err != nil {
			slog.Error("Connection Denied", "s", s.Id, "L", s.Vote.Id, "err", err)
			return err
		}
		waitHistory = true
		slog.Debug("Connected", "s", s.Id, "L", s.Vote.Id)
	default:
		return fmt.Errorf("invalid state: %d", s.State)
	}
	return nil
}

func (s *Server) GetLeaderInfo(ctx context.Context, in *pb.Ping) (*pb.FollowerInfo, error) {
	if int(in.Data) == s.AcceptedEpoch {
		go func() {
			s.Reelect <- true
		}()
		return &pb.FollowerInfo{Id: -1, LastZxid: s.Vote.LastZxid.Raw()}, nil
	}
	if s.State == ELECTION {
		return &pb.FollowerInfo{Id: -1, LastZxid: s.Vote.LastZxid.Raw()}, nil
	}
	return &pb.FollowerInfo{Id: int64(s.Vote.Id), LastZxid: s.Vote.LastZxid.Raw()}, nil
}

// Phase 1 of ZAB
func (s *Server) Discovery() {
	s.Lock()
	slog.Info("Discovery", "s", s.Id, "State", s.State)
	switch s.State {
	case FOLLOWING:
		defer s.Unlock()
		msg := &pb.FollowerInfo{
			Id: int64(s.Id), LastZxid: &pb.Zxid{Epoch: int64(s.AcceptedEpoch), Counter: -1},
		}
		if _, err := SendGrpc(pb.NodeClient.InformLeader,
			s, s.Vote.Id, msg, *maxTimeout); err != nil {
			slog.Error("Connection Denied", "s", s.Id, "L", s.Vote.Id, "err", err)
		}

	case LEADING:
		s.Zab.Reset()
		go s.ProcessFollowerInfo()
		<-s.Zab.QuorumReady
		s.Zab.Lock()
		maxEpoch := -1
		for _, epoch := range s.Zab.FollowerEpochs {
			if epoch > maxEpoch {
				maxEpoch = epoch
			}
		}
		slog.Info("Max epoch", "s", s.Id, "epoch", maxEpoch)
		s.SetEpochs(maxEpoch + 1)

		mostRecent := &pb.AckEpoch{CurrentEpoch: -1, History: nil, LastZxid: &pb.Zxid{Epoch: -1, Counter: -1}}
		for idx := range s.Zab.FollowerEpochs {
			msg := &pb.NewEpoch{Epoch: int64(s.CurrentEpoch)}
			r, _ := SendGrpc(pb.NodeClient.ProposeEpoch, s, idx, msg, *maxTimeout)
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
		s.ReplaceHistory(mostRecent.History)
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
	msg := &pb.NewLeader{Epoch: int64(s.CurrentEpoch), History: s.History.Raw()}
	slog.Info("ZabSync")
	for i := range s.Zab.FollowerEpochs {
		go func(idx int) {
			if _, err := SendGrpc(pb.NodeClient.ProposeLeader,
				s, idx, msg, *maxTimeout); err == nil {
				ready <- true
			}
		}(i)

	}

	// wait for quorum
	for i := 0; i < majority; i++ {
		<-ready
	}
	// send commit to all followers
	for i := range s.Zab.FollowerEpochs {
		SendGrpc(pb.NodeClient.Commit, s, i, &pb.ZabRequest{}, *maxTimeout)
	}

	// phase 3
	if err := s.ZabDeliverAll(); err != nil {
		slog.Error("ZabSync", "s", s.Id, "err", err)
		// panic("failed to deliver")
	}
	if s.History.Len() == 0 {
		s.SetLastZxid(pb.ZxidFragment{Epoch: s.CurrentEpoch})
	}

	close(s.Zab.BroadcastReady)
}

// Phase 3 utils

// Wait for BroadcastReady
func (s *Server) WaitForBroadcast() {
	if _, ok := <-s.Zab.BroadcastReady; ok {
		panic(fmt.Sprintf("%d: unexpected data on BroadcastReady", s.Id))
	}
}

// Commit a transaction locally.
//
// Note: This should be run synchronously
func (s *Server) ZabDeliver(t pb.TransactionFragment) (string, error) {
	// TODO: non-volatile memory
	path, err := s.HandleOperation(t)
	if err != nil {
		return "", err
	}

	t.Committed = true
	s.History.Transactions = append(s.History.Transactions, t)
	s.SetLastZxid(t.Zxid)
	slog.Info("Commit", "s", s.Id, "txn", t)

	return path, nil
}

// Commit all outstanding transactions
//
// Note: This should be run synchronously
func (s *Server) ZabDeliverAll() error {
	for i := s.History.LastCommitId + 1; i < s.History.Len(); i++ {
		t := s.History.Transactions[i]
		// TODO: non-volatile memory
		if _, err := s.HandleOperation(t); err != nil {
			return err
		}

		t.Committed = true
		s.History.LastCommitId = i
	}
	s.SetLastZxid(s.History.LastCommitZxid())
	slog.Info("CommitAll", "s", s.Id, "zxid", s.LastZxid)

	return nil
}
