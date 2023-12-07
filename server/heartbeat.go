package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func (s *Server) ReelectListener() {
	done := false
	for {
		select {
		case _, ok := <-s.Stop:
			if ok {
				panic(fmt.Sprintf("%d: unexpected data on Stop", s.Id))
			}
			return
		case <-s.Reelect:
			if done {
				return
			}
			done = true
			slog.Info("Reelecting", "s", s.Id)
			if vote := s.FastElection(*maxTimeout); vote.Id == -1 {
				slog.Error("Election failed", "s", s.Id)
				close(s.Stop)
				return
			} else {
				slog.Info("Elected", "s", s.Id, "L", vote.Id)
				defer s.Startup()
				return
			}
		}
	}
}

func (s *Server) Heartbeat() {
	if s.State == DOWN {
		return
	}

	go s.ReelectListener()

	for {
		// TODO: consider if concurrent state reads are safe
		// i.e. not possible for 2 servers to be leading at any point in time
		switch s.State {
		case DOWN:
			return
		case LEADING:
			failed := make([]int, 0)
			lock := sync.Mutex{}
			wg := sync.WaitGroup{}
			SendPing := func(i int) {
				_, err := SendGrpc(pb.NodeClient.SendPing, s, i, &pb.Ping{Data: int64(s.Id)}, *maxTimeout)
				if err != nil {
					lock.Lock()
					failed = append(failed, i)
					lock.Unlock()
				}
				wg.Done()
			}
			wg.Add(len(config.Servers) - 1)
			for idx := range config.Servers {
				if idx == s.Id {
					continue
				}
				go SendPing(idx)
			}

			wg.Wait()
			if len(failed) > len(config.Servers)/2 {
				slog.Info("Lost quorum", "s", s.Id, "failed", failed)
				s.Reelect <- true
				return
			}

		case FOLLOWING:
			// Send heartbeat to leader
			_, err := SendGrpc(pb.NodeClient.SendPing, s, s.Vote.Id, &pb.Ping{Data: int64(s.Id)}, *maxTimeout)
			if err != nil {
				slog.Info("Lost leader", "s", s.Id, "leader", s.Vote)
				s.Reelect <- true
				return
			}

		}
		slog.Debug("Heartbeat", "s", s.Id, "state", s.State, "vote", s.Vote)
		select {
		case _, ok := <-s.Stop:
			if ok {
				panic(fmt.Sprintf("%d: unexpected data on Stop", s.Id))
			}
			return
		case <-time.After(time.Duration(*maxTimeout) * time.Millisecond):
		}
	}
}

// Listener to wait for liveness of ensemble
func (s *Server) WaitForLive() {
	done := make(map[int]bool)
	// TODO: consider how many to wait for
	// TODO: try concurrent broadcast
	for len(done) < len(config.Servers)/2+1 {
		for idx := range config.Servers {
			if idx == s.Id {
				continue
			}
			if _, ok := done[idx]; ok {
				continue
			}
			if r, err := SendGrpc(pb.NodeClient.SendPing, s, idx, &pb.Ping{Data: int64(s.Id)}, *maxTimeout); err == nil && State(r.Data) > ELECTION {
				done[idx] = true
			}
		}
		time.Sleep(time.Duration(*maxTimeout/2) * time.Millisecond)
	}
}
