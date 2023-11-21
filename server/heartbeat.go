package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func (s *Server) ReelectListener() {
	for {
		select {
		case _, ok := <-s.Stop:
			if ok {
				panic(fmt.Sprintf("%d: unexpected data on Stop", s.Id))
			}
			return
		case <-s.Reelect:
			if vote := s.FastElection(*maxTimeout); vote.Id == -1 {
				slog.Error("Election failed", "s", s.Id)
				s.Stop <- true
				return
			}
		}
	}
}

func (s *Server) Heartbeat() {
	if s.State == DOWN {
		return
	}

	// trigger shutdown
	// defer func() {
	// 	close(s.Stop)
	// }()
	go s.ReelectListener()

	for {
		// TODO: consider if concurrent state reads are safe
		// i.e. not possible for 2 servers to be leading at any point in time
		switch s.State {
		case DOWN:
			return
		case LEADING:
			failed := make(map[int]bool)
			wg := sync.WaitGroup{}
			wg.Add(len(config.Servers) - 1)
			for idx := range config.Servers {
				if idx == s.Id {
					continue
				}

				go func(i int) {
					_, err := SendGrpc[*pb.Ping, *pb.Ping](pb.NodeClient.SendPing, s, i, &pb.Ping{Data: int64(s.Id)}, *maxTimeout)
					if err != nil {
						failed[i] = true
					}
					wg.Done()
				}(idx)
			}
			// check for failed nodes
			wg.Wait()
			if len(failed) > len(config.Servers)/2 {
				slog.Info("Lost quorum", "s", s.Id, "failed", failed)
				s.Reelect <- true
				return
			}

		case FOLLOWING:
			// Simulate failure
			// fail := rand.Intn(100) < 10
			// if fail {
			// 	log.Printf("%d failed", s.Id)
			// 	return
			// }

			// Send heartbeat to leader
			_, err := SendGrpc[*pb.Ping, *pb.Ping](pb.NodeClient.SendPing, s, s.Vote.Id, &pb.Ping{Data: int64(s.Id)}, *maxTimeout)
			if err != nil {
				slog.Info("Lost leader", "s", s.Id, "leader", s.Vote)
				s.Reelect <- true
				return
			}

		}
		slog.Debug("Heartbeat", "s", s.Id, "state", s.State, "vote", s.Vote)
		time.Sleep(time.Duration(*maxTimeout) * time.Millisecond)
	}
}
