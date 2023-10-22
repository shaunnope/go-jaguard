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

func (t Table) Put(key int, vote [2]int, round int) {
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
	// TODO: timeout
	// NOTE: not quite sure what cancel() does
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	msg := &pb.ElectNotification{
		Id:    int32(s.Id),
		State: int32(s.State),
		Vote: &pb.Vote{
			LastZxid: int32(s.LastZxid),
			Id:       int32(s.Id),
		},
		Round: int32(s.Round),
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

	msgVote := in.GetVote()
	msgState := State(in.GetState())
	from := int(in.GetId())

	reply := false

	if s.GetState() == ELECTION {
		msg := VoteMsg{
			vote:  [2]int{int(msgVote.LastZxid), int(msgVote.Id)},
			id:    from,
			state: msgState,
			round: int(in.GetRound()),
		}
		s.Queue <- msg

		reply = (msgState == ELECTION && msg.round < s.Round)

	} else {
		reply = msgState == ELECTION
	}

	if reply {
		s.ElectNotify(from)
	}

	return &pb.ElectResponse{State: int32(s.State)}, nil
}

func (s *Server) FastElection(t0 int) Vote {
	var timeout int = t0

	// TODO: might deadlock
	s.Lock()
	defer s.Unlock()

	// set of servers that have responded
	receivedVotes := make(Table)
	outOfElection := make(Table)

	s.State = ELECTION
	s.SetVote(false)
	s.IncRound(false)

	// send vote requests to all other servers
	// TODO: broadcast

	for s.State == ELECTION {
		var n VoteMsg
		select {
		case n = <-s.Queue:
			// received vote
			// receivedVotes[n.id] = VoteLog{n.vote, n.round, 0}

			if n.state == ELECTION {
				if n.round > s.Round {
					s.Round = n.round
					receivedVotes = make(map[int]VoteLog)

					if n.vote.GreaterThan(Vote{s.LastZxid, s.Id}) {
						s.Vote = n.vote
					} else {
						s.SetVote(false)
					}
					s.ElectBroadcast()
				} else if n.round == s.Round && n.vote.GreaterThan(s.Vote) {
					s.Vote = n.vote
					s.ElectBroadcast()
				} else if n.round < s.Round {
					continue
				}

				receivedVotes.Put(n.id, n.vote, n.round)
				if len(receivedVotes) == len(config.Servers) {
					s.DeduceLeader(s.Vote[1])
					return s.Vote
				} else if receivedVotes.HasQuorum(s.Vote) && len(s.Queue) > 0 {
					time.Sleep(time.Duration(t0) * time.Millisecond)
					if len(s.Queue) == 0 {
						s.DeduceLeader(s.Vote[1])
						return s.Vote
					}
				}

			} else {
				if n.round == s.Round {
					receivedVotes.Put(n.id, n.vote, n.round)
					if n.state == LEADING {
						s.DeduceLeader(n.vote[1])
						return n.vote
					}
					hasQuorum := receivedVotes.HasQuorum(n.vote)
					if n.vote[1] == s.Id && hasQuorum {
						s.DeduceLeader(n.vote[1])
						return n.vote
					} else if _, ok := outOfElection[n.vote[1]]; hasQuorum && ok {
						r := s.ElectNotify(n.vote[1])
						if r.GetState() == int32(LEADING) {
							s.DeduceLeader(n.vote[1])
							return n.vote
						}
					}
				}
				outOfElection.Put(n.id, n.vote, n.round)
				hasQuorum := outOfElection.HasQuorum(n.vote)
				if n.vote[1] == s.Id && hasQuorum {
					s.Round = n.round
					s.DeduceLeader(n.vote[1])
					return n.vote
				} else if _, ok := outOfElection[n.vote[1]]; hasQuorum && ok {
					r := s.ElectNotify(n.vote[1])
					if r.GetState() == int32(LEADING) {
						s.Round = n.round
						s.DeduceLeader(n.vote[1])
						return n.vote
					}
				}
			}

		case <-time.After(time.Duration(timeout) * time.Millisecond):
			// timeout
			s.ElectBroadcast()
			timeout = int(math.Max(float64(2*timeout), maxElectionTimeout))
		}
	}

	return Vote{-1, -1}
}
