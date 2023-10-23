package main

import (
	"context"
	"log"
	"math"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

const (
	maxElectionTimeout float64 = 10000
)

type Table map[int]*VoteLog

func (t Table) Put(key int, vote *pb.Vote, round int) {
	var version int
	if val, ok := t[key]; ok && val.Round == round {
		version = t[key].Version + 1
	} else {
		version = 1
	}
	t[key] = &VoteLog{vote, round, version}
}

func (t Table) HasQuorum(vote *pb.Vote) bool {
	count := 0
	for _, val := range t {
		if (*val.Vote).Equal(vote) {
			count++
		}
	}
	return count > len(config.Servers)/2
}

func (s *Server) DeduceLeader(id int64) {
	if s.Id == int(id) {
		s.State = LEADING
	} else {
		s.State = FOLLOWING
	}
}

// Send an election notification to another server
func (s *Server) ElectNotify(from int64) *pb.ElectResponse {
	// don't send to self
	if s.Id == int(from) {
		return nil
	}
	// establish connection
	s.EstablishConnection(int(from))

	// send vote notif
	// TODO: adjust timeout value
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	msg := &pb.ElectNotification{
		Id:    int64(s.Id),
		State: int64(s.State),
		Vote: &pb.Vote{
			LastZxid: s.LastZxid,
			Id:       int64(s.Id),
		},
		Round: int64(s.Round),
	}
	r, err := (*s.Connections[int(from)]).Elect(ctx, msg)
	if err != nil {
		log.Printf("%d error sending vote notif to %d: %v", s.Id, from, err)
		// return nil, err
	}

	return r
}

// Broadcast election notification to all other servers
func (s *Server) ElectBroadcast() {
	for idx := range config.Servers {
		go s.ElectNotify(int64(idx))
	}
}

// Notification Receiver
func (s *Server) Elect(ctx context.Context, in *pb.ElectNotification) (*pb.ElectResponse, error) {

	msgState := State(in.GetState())
	from := in.GetId()

	reply := false

	if s.GetState() == ELECTION {
		var msg VoteMsg = in
		s.Queue <- msg

		reply = (msgState == ELECTION && int(msg.Round) < s.Round)

	} else {
		reply = msgState == ELECTION
	}

	if reply {
		s.ElectNotify(from)
	}

	return &pb.ElectResponse{State: int64(s.State)}, nil
}

func (s *Server) FastElection(t0 int) *pb.Vote {
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
		case n = <-s.Queue:
			// received vote
			// receivedVotes[n.id] = VoteLog{n.Vote, n.round, 0}

			if State(n.State) == ELECTION {
				if int(n.Round) > s.Round {
					s.Round = int(n.Round)
					receivedVotes = make(map[int]*VoteLog)

					if n.Vote.GreaterThan(&pb.Vote{LastZxid: s.LastZxid, Id: int64(s.Id)}) {
						s.Vote = n.Vote
					} else {
						s.SetVote(false)
					}
					s.ElectBroadcast()
				} else if int(n.Round) == s.Round && n.Vote.GreaterThan(s.Vote) {
					s.Vote = n.Vote
					s.ElectBroadcast()
				} else if int(n.Round) < s.Round {
					continue
				}

				receivedVotes.Put(int(n.Id), n.Vote, int(n.Round))
				if len(receivedVotes) == len(config.Servers) {
					s.DeduceLeader(s.Vote.Id)
					return s.Vote
				} else if receivedVotes.HasQuorum(s.Vote) && len(s.Queue) > 0 {
					time.Sleep(time.Duration(t0) * time.Millisecond)
					if len(s.Queue) == 0 {
						s.DeduceLeader(s.Vote.Id)
						return s.Vote
					}
				}

			} else {
				if int(n.Round) == s.Round {
					receivedVotes.Put(int(n.Id), n.Vote, int(n.Round))
					if State(n.State) == LEADING {
						s.DeduceLeader(n.Vote.Id)
						return n.Vote
					}
					hasQuorum := receivedVotes.HasQuorum(n.Vote)
					if int(n.Vote.Id) == s.Id && hasQuorum {
						s.DeduceLeader(n.Vote.Id)
						return n.Vote
					} else if _, ok := outOfElection[int(n.Vote.Id)]; hasQuorum && ok {
						r := s.ElectNotify(n.Vote.Id)
						if State(r.GetState()) == LEADING {
							s.DeduceLeader(n.Vote.Id)
							return n.Vote
						}
					}
				}
				outOfElection.Put(int(n.Id), n.Vote, int(n.Round))
				hasQuorum := outOfElection.HasQuorum(n.Vote)
				if int(n.Vote.Id) == s.Id && hasQuorum {
					s.Round = int(n.Round)
					s.DeduceLeader(n.Vote.Id)
					return n.Vote
				} else if _, ok := outOfElection[int(n.Vote.Id)]; hasQuorum && ok {
					r := s.ElectNotify(n.Vote.Id)
					if State(r.GetState()) == LEADING {
						s.Round = int(n.Round)
						s.DeduceLeader(n.Vote.Id)
						return n.Vote
					}
				}
			}

		case <-time.After(time.Duration(timeout) * time.Millisecond):
			// timeout
			s.ElectBroadcast()
			timeout = int(math.Max(float64(2*timeout), maxElectionTimeout))
		}
	}

	return &pb.Vote{Id: -1}
}
