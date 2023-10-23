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

type Table map[int]VoteLog

func (t Table) Put(key int, vote Vote, round int) {
	var version int
	if val, ok := t[key]; ok && val.round == round {
		version = t[key].version + 1
	} else {
		version = 1
	}
	t[key] = VoteLog{vote, round, version}
}

func (t Table) HasQuorum(vote Vote) bool {
	count := 0
	for _, val := range t {
		if val.vote.Equal(vote) {
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
	// establish connection
	s.EstablishConnection(from)

	// send vote notif
	// TODO: adjust timeout value
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	msg := &pb.ElectNotification{
		Id:    int64(s.Id),
		State: int64(s.State),
		Vote: &pb.Vote{
			LastZxid: s.LastZxid.Raw(),
			Id:       int64(s.Id),
		},
		Round: int64(s.Round),
	}
	r, err := (*s.Connections[from]).Elect(ctx, msg)
	if err != nil {
		log.Printf("%d error sending vote notif to %d: %v", s.Id, from, err)
		// return nil, err
	}

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

	if s.GetState() == ELECTION {
		msg := VoteMsgFrom(in)
		s.Queue <- msg

		reply = (msgState == ELECTION && msg.round < s.Round)

	} else {
		reply = msgState == ELECTION
	}

	if reply {
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
		var n *VoteMsg
		select {
		case n = <-s.Queue:
			// received vote
			// receivedVotes[n.id] = VoteLog{n.vote, n.round, 0}

			if n.state == ELECTION {
				if n.round > s.Round {
					s.Round = n.round
					receivedVotes = make(map[int]VoteLog)

					if n.vote.GreaterThan(Vote{s.LastZxid, s.Id}) {
						s.Vote = *n.vote
					} else {
						s.SetVote(false)
					}
					s.ElectBroadcast()
				} else if n.round == s.Round && n.vote.GreaterThan(s.Vote) {
					s.Vote = *n.vote
					s.ElectBroadcast()
				} else if n.round < s.Round {
					continue
				}

				receivedVotes.Put(n.id, *n.vote, n.round)
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
				if n.round == s.Round {
					receivedVotes.Put(n.id, *n.vote, n.round)
					if n.state == LEADING {
						s.DeduceLeader(n.vote.Id)
						return *n.vote
					}
					hasQuorum := receivedVotes.HasQuorum(*n.vote)
					if n.vote.Id == s.Id && hasQuorum {
						s.DeduceLeader(n.vote.Id)
						return *n.vote
					} else if _, ok := outOfElection[n.vote.Id]; hasQuorum && ok {
						r := s.ElectNotify(n.vote.Id)
						if State(r.GetState()) == LEADING {
							s.DeduceLeader(n.vote.Id)
							return *n.vote
						}
					}
				}
				outOfElection.Put(n.id, *n.vote, n.round)
				hasQuorum := outOfElection.HasQuorum(*n.vote)
				if n.vote.Id == s.Id && hasQuorum {
					s.Round = n.round
					s.DeduceLeader(n.vote.Id)
					return *n.vote
				} else if _, ok := outOfElection[n.vote.Id]; hasQuorum && ok {
					r := s.ElectNotify(n.vote.Id)
					if State(r.GetState()) == LEADING {
						s.Round = n.round
						s.DeduceLeader(n.vote.Id)
						return *n.vote
					}
				}
			}

		case <-time.After(time.Duration(timeout) * time.Millisecond):
			// timeout
			s.ElectBroadcast()
			timeout = int(math.Max(float64(2*timeout), maxElectionTimeout))
		}
	}

	return Vote{Id: -1}
}
