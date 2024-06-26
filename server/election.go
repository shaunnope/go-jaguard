package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

const (
	maxElectionTimeout float64 = 100000
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

func (t Table) Quorum(vote Vote) int {
	count := 0
	for _, val := range t {
		if val.Vote.Equal(vote) {
			count++
		}
	}
	return count
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

func (s *Server) DeduceLeader(vote Vote) {
	if s.Id == vote.Id {
		s.State = LEADING
	} else {
		s.State = FOLLOWING
	}
	s.Vote = vote
}

// Send an election notification to another server
func (s *Server) ElectNotify(from int) *pb.ElectResponse {
	// don't send to self
	if s.Id == from {
		return nil
	}
	s.Lock()
	defer s.Unlock()

	msg := &pb.ElectNotification{
		Id:    int64(s.Id),
		State: int64(s.State),
		Vote: &pb.Vote{
			LastZxid: s.LastZxid.Raw(),
			Id:       int64(s.Vote.Id),
		},
		Round: int64(s.Round),
	}

	// send vote notification
	// TODO: adjust timeout value
	// TODO: Handle error?
	r, _ := SendGrpc(pb.NodeClient.Elect, s, from, msg, *maxTimeout)

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
		s.ElectNotify(from)
	}

	return &pb.ElectResponse{State: int64(s.State)}, nil
}

// Perform Fast Leader Election
func (s *Server) FastElection(t0 int) Vote {
	var timeout int = t0

	// NOTE: might deadlock

	// set of servers that have responded
	receivedVotes := make(Table)
	outOfElection := make(Table)

	s.SetStateAndVote(ELECTION, nil)

	fileName := fmt.Sprintf("server%d.txt", s.Id)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}
	fmt.Fprintln(file, s.Id)
	currentTime := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(file, "Fast Election started on %s \n", currentTime)
	defer file.Close()

	// send vote requests to all other servers
	if *leader_verbo {
		fmt.Fprintf(file, "server %d begins election\n", s.Id)
	}
	s.ElectBroadcast()
	var quorumVote *Vote = nil
	for s.GetState() == ELECTION {
		var n VoteMsg
		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			if quorumVote != nil {
				// log.Printf("s%d has quorum and no other notification and elect %d as the leader\n", s.Id, quorumVote.Id)
				s.DeduceLeader(*quorumVote)
				return *quorumVote
			}
			// no reply increase timeout
			s.ElectBroadcast()
			timeout = int(math.Min(float64(2*timeout), maxElectionTimeout))

		case n = <-s.Queue:
			nVote := n.Vote.Data()

			if *leader_verbo {
				fmt.Fprintf(file, "%d (%d) received %d (%d): %v", s.Id, s.Round, n.Id, n.Round, nVote)
			}

			if State(n.State) == ELECTION {
				if int(n.Round) > s.Round {
					s.Round = int(n.Round)
					receivedVotes = make(map[int]*VoteLog)
					var newVote *Vote = nil
					if nVote.GreaterThan(Vote{LastZxid: s.LastZxid, Id: s.Id}) {
						newVote = &nVote
					}
					s.SetVote(newVote)
					s.ElectBroadcast()
				} else if int(n.Round) == s.Round && nVote.GreaterThan(s.Vote) {
					if *leader_verbo {
						fmt.Fprintf(file, "%d (%d) agree %d (%d): %v", s.Id, s.Round, n.Id, n.Round, nVote)
					}
					s.SetVote(&nVote)
					s.ElectBroadcast()
				} else if int(n.Round) < s.Round {
					continue
				}

				receivedVotes.Put(int(n.Id), nVote, int(n.Round))
				if len(receivedVotes) == len(config.Servers)-1 {
					slog.Debug("ELECTION All votes received", "s", s.Id)
					if *leader_verbo {
						fmt.Fprintf(file, "%d (%d) voting: %v", s.Id, s.Round, s.Vote)
					}
					s.DeduceLeader(s.Vote)
					return s.Vote
				} else if receivedVotes.HasQuorum(s.Vote) && len(s.Queue) > 0 {
					log.Printf("s%d vote received quorum: %v\n", s.Id, s.Vote)
					quorumVote = &s.Vote
				}

			} else {
				if int(n.Round) == s.Round {
					receivedVotes.Put(int(n.Id), nVote, int(n.Round))

					if State(n.State) == LEADING {
						log.Printf("s%d received %d as leader\n", s.Id, n.Id)
						if *leader_verbo {
							fmt.Fprintf(file, "%d found leader %d: %v\n", s.Id, n.Id, nVote)
						}
						s.DeduceLeader(nVote)
						return nVote
					}
					hasQuorum := receivedVotes.HasQuorum(nVote)
					if int(nVote.Id) == s.Id && hasQuorum {
						if *leader_verbo {
							fmt.Fprintf(file, "server %d sees that it is being followed by %d and has achieve quorum so it can prepare to vote for itself\n", s.Id, n.Id)
						}
						s.DeduceLeader(nVote)
						return nVote
					} else if _, ok := outOfElection[int(nVote.Id)]; hasQuorum && ok {
						r := s.ElectNotify(nVote.Id)
						if State(r.GetState()) == LEADING {
							if *leader_verbo {
								fmt.Fprintf(file, "server %d sees the quorum and the new leader %d that %d has voted for so follow the same\n", s.Id, nVote.Id, n.Id)
							}
							s.DeduceLeader(nVote)
							return nVote
						}
					}
					if *leader_verbo {
						if !hasQuorum {
							fmt.Fprintf(file, "server %d see that server %d whose Zxid=%v is FOLLOWING but cannot conclude yet because quorum is %d\n", s.Id, n.Id, nVote.LastZxid, receivedVotes.Quorum(nVote))
						} else {
							fmt.Fprintf(file, "server %d see that server %d whose Zxid=%v is FOLLOWING but cannot conclude yet because did not see %d out of election\n", s.Id, n.Id, nVote.LastZxid, nVote.Id)
						}
					}
				}
				outOfElection.Put(int(n.Id), nVote, int(n.Round))
				hasQuorum := outOfElection.HasQuorum(nVote)
				if int(nVote.Id) == s.Id && hasQuorum {
					s.Round = int(n.Round)
					if *leader_verbo {
						fmt.Fprintf(file, "server %d sees that it is being followed by %d and has achieve quorum so it can prepare to vote for itself\n", s.Id, n.Id)
					}
					s.DeduceLeader(nVote)
					return nVote
				} else if _, ok := outOfElection[int(nVote.Id)]; hasQuorum && ok {
					r := s.ElectNotify(nVote.Id)
					if State(r.GetState()) == LEADING {
						s.Round = int(n.Round)
						if *leader_verbo {
							fmt.Fprintf(file, "server %d sees that %d is LEADING and will follow\n", s.Id, nVote.Id)
						}
						s.DeduceLeader(nVote)
						return nVote
					}
				}
			}

		}
	}

	return Vote{Id: -1}
}
