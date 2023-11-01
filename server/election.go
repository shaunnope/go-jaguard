package main

import (
	"context"
	"fmt"
	"math"
	"os"
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
	s.Lock()
	defer s.Unlock()
	if s.Id == from {
		return nil
	}
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

	// set of servers that have responded
	receivedVotes := make(Table)
	outOfElection := make(Table)

	s.State = ELECTION
	s.SetVote(false)
	s.IncRound(false)

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
		fmt.Fprintf(file, "server %d begins election by checking who are alive\n", s.Id)
	}
	s.ElectBroadcast()

	for s.State == ELECTION {
		var n VoteMsg
		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			// no reply increase timeout
			s.ElectBroadcast()
			timeout = int(math.Max(float64(2*timeout), maxElectionTimeout))

		case n = <-s.Queue:
			nVote := n.Vote.Data()

			if *leader_verbo {
				// log.Printf("%d (%d) received %d (%d): %v", s.Id, s.Round, n.Id, n.Round, nVote)
				fmt.Fprintf(file, "server %d received response from %d is voting for %d \n", s.Id, n.Id, nVote.Id)
			}

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
					if *leader_verbo {
						fmt.Fprintf(file, "server %d sees server %d 's vote and agree %d with zxid %v as a better candidate\n", s.Id, n.Id, nVote.LastZxid, nVote.Id)
					}
					s.Vote = nVote
					s.ElectBroadcast()
				} else if int(n.Round) < s.Round {
					continue
				}

				receivedVotes.Put(int(n.Id), nVote, int(n.Round))
				if len(receivedVotes) == len(config.Servers)-1 {
					if *leader_verbo {
						fmt.Fprintf(file, "server %d has everyone's responses and prepares to elect %d as the leader\n", s.Id, s.Vote.Id)
					}
					s.DeduceLeader(s.Vote.Id)
					return s.Vote
				} else if receivedVotes.HasQuorum(s.Vote) && len(s.Queue) > 0 {
					// FIXME: queue should be processed until empty
					time.Sleep(time.Duration(t0) * time.Millisecond)
					if len(s.Queue) == 0 {
						if *leader_verbo {
							fmt.Fprintf(file, "%d has quoram and no other notification and elect %d as the leader\n", s.Id, s.Vote.Id)
						}
						s.DeduceLeader(s.Vote.Id)
						return s.Vote
					}
				}

			} else {
				if int(n.Round) == s.Round {
					receivedVotes.Put(int(n.Id), nVote, int(n.Round))
					// if State(n.State) == FOLLOWING {
					// 	log.Printf("server %d sees %d is FOLLOWING %d with vote %v\n", s.Id, n.Id, n.Vote.Id, n.Vote)
					// 	log.Printf("server %d has received these:\n", s.Id)
					// 	for idx, val := range receivedVotes {
					// 		log.Printf("index: %d value: %v\n", idx, val)
					// 	}
					// 	log.Printf("server %d has quorum %d\n", s.Id, receivedVotes.Quorum(nVote))
					// }

					if State(n.State) == LEADING {
						if *leader_verbo {
							fmt.Fprintf(file, "server %d sees that %d whose Zxid=%v is the leader and will follow\n", s.Id, n.Id, nVote.LastZxid)
						}
						s.DeduceLeader(nVote.Id)
						return nVote
					}
					hasQuorum := receivedVotes.HasQuorum(nVote)
					if int(nVote.Id) == s.Id && hasQuorum {
						if *leader_verbo {
							fmt.Fprintf(file, "server %d sees that it is being followed by %d and has achieve quoram so it can prepare to vote for itself\n", s.Id, n.Id)
						}
						s.DeduceLeader(nVote.Id)
						return nVote
					} else if _, ok := outOfElection[int(nVote.Id)]; hasQuorum && ok {
						r := s.ElectNotify(nVote.Id)
						if State(r.GetState()) == LEADING {
							if *leader_verbo {
								fmt.Fprintf(file, "server %d sees the quoram and the new leader %d that %d has voted for so follow the same\n", s.Id, nVote.Id, n.Id)
							}
							s.DeduceLeader(nVote.Id)
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
				// log.Printf("server %d has put %d out of election\n", s.Id, n.Id)
				outOfElection.Put(int(n.Id), nVote, int(n.Round))
				hasQuorum := outOfElection.HasQuorum(nVote)
				if int(nVote.Id) == s.Id && hasQuorum {
					s.Round = int(n.Round)
					if *leader_verbo {
						fmt.Fprintf(file, "server %d sees that it is being followed by %d and has achieve quoram so it can prepare to vote for itself\n", s.Id, n.Id)
					}
					s.DeduceLeader(nVote.Id)
					return nVote
				} else if _, ok := outOfElection[int(nVote.Id)]; hasQuorum && ok {
					r := s.ElectNotify(nVote.Id)
					if State(r.GetState()) == LEADING {
						s.Round = int(n.Round)
						if *leader_verbo {
							fmt.Fprintf(file, "server %d sees that %d is LEADING and will follow\n", s.Id, nVote.Id)
						}
						s.DeduceLeader(nVote.Id)
						return nVote
					}
				}
			}

		}
	}

	return Vote{Id: -1}
}
