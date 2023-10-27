package main

import (
	"context"
	"math"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

const (
	maxElectionTimeout float64 = 10000
)

type Table map[int]*VoteLog

func (t Table) Put(key int, vote Vote, round int) {
	var version int
	if val, ok := t[key]; ok && val.Round == round {
		version = t[key].Version + 1
	} else {
		version = 1
	}
	t[key] = &VoteLog{vote, round, version}
}

func (t Table) HasQuorum(vote Vote) bool {
	count := 0
	for _, val := range t {
		if val.Vote.Equal(vote) {
			count++
		}
	}
	return count > len(config.Servers)/2
}

func (s *Server) DeduceLeader(id int) {
	if s.Id == id {
		s.State = LEADING
	} else {
		s.State = FOLLOWING
	}
}

// Send an election notification to another server
func (s *Server) ElectNotify(from int) *pb.ElectResponse {
	// don't send to self
	if s.Id == from {
		return nil
	}
	msg := &pb.ElectNotification{
		Id:    int64(s.Id),
		State: int64(s.State),
		Vote: &pb.Vote{
			LastZxid: s.LastZxid.Raw(),
			Id:       int64(s.Id),
		},
		Round: int64(s.Round),
	}

	// send vote notification
	// TODO: adjust timeout value
	r, _ := SendGrpc[*pb.ElectNotification, *pb.ElectResponse](pb.NodeClient.Elect, s, from, msg, *maxTimeout)
	// if err != nil {
	// return nil, err
	// }

	return r
}

// Broadcast election notification to all other servers
func (s *Server) ElectBroadcast() {
	for idx := range config.Servers {
		go s.ElectNotify(idx)
	}
}

// Notification Receiver
func (s *Server) Elect(ctx context.Context, in *pb.ElectNotification) (*pb.ElectResponse, error) {

	msgState := State(in.GetState())
	from := int(in.GetId())

	reply := false

	if s.State == ELECTION {
		s.Queue <- in

		reply = (msgState == ELECTION && int(in.Round) < s.Round)

	} else {
		reply = msgState == ELECTION
	}

	if reply {
		// log.Printf("%d sending vote response to %d", s.Id, from)
		s.ElectNotify(from)
	}

	return &pb.ElectResponse{State: int64(s.State)}, nil
}

func (s *Server) FastElection(t0 int) Vote {
	var timeout int = t0

	// NOTE: might deadlock
	s.Lock()
	defer s.Unlock()

	// set of servers that have responded
	receivedVotes := make(Table)
	outOfElection := make(Table)

	s.State = ELECTION
	s.SetVote(false)
	s.IncRound(false)

	// send vote requests to all other servers
	s.ElectBroadcast()

	for s.State == ELECTION {
		var n VoteMsg
		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			// timeout
			s.ElectBroadcast()
			timeout = int(math.Max(float64(2*timeout), maxElectionTimeout))

		case n = <-s.Queue:
			nVote := n.Vote.Data()

			// log.Printf("%d (%d) received %d (%d): %v", s.Id, s.Round, n.Id, n.Round, nVote)

			if State(n.State) == ELECTION {
				if int(n.Round) > s.Round {
					s.Round = int(n.Round)
					receivedVotes = make(map[int]*VoteLog)

					if nVote.GreaterThan(Vote{LastZxid: s.LastZxid, Id: s.Id}) {
						s.Vote = nVote
					} else {
						s.SetVote(false)
					}
					s.ElectBroadcast()
				} else if int(n.Round) == s.Round && nVote.GreaterThan(s.Vote) {
					s.Vote = nVote
					s.ElectBroadcast()
				} else if int(n.Round) < s.Round {
					continue
				}

				receivedVotes.Put(int(n.Id), nVote, int(n.Round))
				if len(receivedVotes) == len(config.Servers)-1 {
					s.DeduceLeader(s.Vote.Id)
					return s.Vote
				} else if receivedVotes.HasQuorum(s.Vote) && len(s.Queue) > 0 {
					// FIXME: queue should be processed until empty
					time.Sleep(time.Duration(t0) * time.Millisecond)
					if len(s.Queue) == 0 {
						s.DeduceLeader(s.Vote.Id)
						return s.Vote
					}
				}

			} else {
				if int(n.Round) == s.Round {
					receivedVotes.Put(int(n.Id), nVote, int(n.Round))
					if State(n.State) == LEADING {
						s.DeduceLeader(nVote.Id)
						return nVote
					}
					hasQuorum := receivedVotes.HasQuorum(nVote)
					if int(nVote.Id) == s.Id && hasQuorum {
						s.DeduceLeader(nVote.Id)
						return nVote
					} else if _, ok := outOfElection[int(nVote.Id)]; hasQuorum && ok {
						r := s.ElectNotify(nVote.Id)
						if State(r.GetState()) == LEADING {
							s.DeduceLeader(nVote.Id)
							return nVote
						}
					}
				}
				outOfElection.Put(int(n.Id), nVote, int(n.Round))
				hasQuorum := outOfElection.HasQuorum(nVote)
				if int(nVote.Id) == s.Id && hasQuorum {
					s.Round = int(n.Round)
					s.DeduceLeader(nVote.Id)
					return nVote
				} else if _, ok := outOfElection[int(nVote.Id)]; hasQuorum && ok {
					r := s.ElectNotify(nVote.Id)
					if State(r.GetState()) == LEADING {
						s.Round = int(n.Round)
						s.DeduceLeader(nVote.Id)
						return nVote
					}
				}
			}
		}
	}

	return Vote{Id: -1}
}
