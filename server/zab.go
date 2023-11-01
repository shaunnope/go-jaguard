package main

import (
	"context"
	"errors"
	pb "github.com/shaunnope/go-jaguard/zouk"
	"log"
	"sync"
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

		if len(s.Leader.FollowerEpochs) > len(config.Servers)/2 {
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
		}()
	}

	// goto phase 2
	defer s.ZabSync()

	return nil, errors.New("leader not accepted")
}

func (s *Server) SendZabRequest(ctx context.Context, in *pb.ZabRequest) (*pb.ZabAck, error) {
	isLeader := s.GetState() == LEADING

	// Handle incoming CreateRequest
	switch in.RequestType {
	case pb.RequestType_PROPOSAL:
		if isLeader {
			return nil, errors.New("leaders shouldnt get proposals")
		}
		log.Printf("%d received proposal: %v", s.Id, in.Transaction)

		if int(in.Transaction.Zxid.Epoch) == s.CurrentEpoch {
			// accept proposal
			s.History = append(s.History, in.Transaction.Extract())
			s.LastZxid = in.Transaction.Extract().Zxid
			// send ack to leader
			return &pb.ZabAck{Request: in}, nil
		}

		return nil, errors.New("proposal not accepted")

	case pb.RequestType_ANNOUNCEMENT:
		if isLeader {
			return nil, errors.New("leaders shouldnt get announcements")
		}

		// TODO: for each commit (announcement), wait until all earlier proposals are committed
		// then, commit

		// use priority queue because transactions may not arrive in order
		// Check that transaction is at the start of the queue
		// FOLLOWER need to execute request

		/*
			//transactionFragment := in.Transaction.Extract()
			//s.CommitQueue.Push(transactionFragment)

			//s.CommitQueue.Update(transactionFragment) //???

			transaction := in.Transaction.Extract()

			//wait until transaction is at the start of the queue

			s.Unlock()
			switch in.Transaction.Type {
			case pb.OperationType_CREATE:
				zxid := pb.ZxidFragment{int(in.Transaction.Zxid.Epoch), int(in.Transaction.Zxid.Counter)}
				//ephemeral owner??
				path, err := s.StateVector.Data.CreateNode(transaction.Path, transaction.Data, transaction.Flags&pb.EPHEMERAL != 0, 0, zxid, transaction.Flags&pb.SEQUENTIAL != 0)
				if err != nil {
					return nil, err
				}
				log.Printf("node created at %s", path)

			case pb.OperationType_UPDATE:
				//Check node exists? Done here or in add into setdata method?
				zxid := pb.ZxidFragment{int(in.Transaction.Zxid.Epoch), int(in.Transaction.Zxid.Counter)}
				//Version??
				s.StateVector.Data.SetData(transaction.Path, transaction.Data, 0, zxid)
				log.Printf("node at ")
			case pb.OperationType_DELETE:
				outcome, err := s.StateVector.Data.DeleteNode(in.Transaction.Path, 0)
				if err != nil {
					return nil, err
				}
				log.Println(outcome)

			}
			s.Lock()

			//Remove transaction from CommitQueue
			//Add transaction into History
			s.History = append(s.History, in.Transaction.Extract())

		*/

	case pb.RequestType_CLIENT:
		// if leader send proposal to all followers in for loop (rpc)
		// since its rpc, leader will monitor for responses and decide whether to commit/announce
		// if follower forward to leader, do nothing with response (rpc)
		if isLeader {
			log.Printf("%d received client request: %v", s.Id, in.Transaction)
			// TODO: verify version
			// propose to all

			msg := &pb.ZabRequest{
				Transaction: in.Transaction.WithZxid(s.LastZxid.Inc()),
				RequestType: pb.RequestType_PROPOSAL,
			}

			wg := sync.WaitGroup{}
			wg.Add(len(s.Leader.FollowerEpochs)/2 + 1)

			for idx := range s.Leader.FollowerEpochs {
				if idx == s.Id {
					continue
				}

				go func(i int) {
					_, err := SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, i, msg, *maxTimeout)
					if err == nil {
						wg.Done()
					}
				}(idx)
			}
			// wait for quorum
			wg.Wait()

			// commit
			s.History = append(s.History, in.Transaction.Extract())
			s.LastZxid = msg.Transaction.Extract().Zxid

			msg.RequestType = pb.RequestType_ANNOUNCEMENT
			for idx := range s.Leader.FollowerEpochs {
				if idx == s.Id {
					continue
				}

				go SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, idx, msg, *maxTimeout)
			}

			// LEADER need to execute request

		} else {
			log.Printf("%d forwarding request to %d", s.Id, s.Vote.Id)
			// todo verify version
			// forward to leader
			SendGrpc[*pb.ZabRequest, *pb.ZabAck](pb.NodeClient.SendZabRequest, s, s.Vote.Id, in, *maxTimeout)

		}
	}

	return nil, errors.New("zab request not accepted")
}

// end grpc calls

//func (s *Server) HandleOperation() {

//}

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

// Phase 3 of ZAB
// the main phase for non-faulty operation

func (s *Server) ZabBroadcast() {

	// if leading, invoke ready (?)

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
