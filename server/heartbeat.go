package main

import (
	"log"
	"sync"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func (s *Server) Heartbeat() {
	for {
		state := s.GetState()
		switch state {
		case LEADING:
			failed := make(map[int]bool)
			wg := sync.WaitGroup{}
			wg.Add(len(config.Servers) - 1)
			for idx := range config.Servers {
				if idx == s.Id {
					continue
				}
				go func(i int) {
					ctx, cancel := s.EstablishConnection(i, *maxTimeout)
					defer cancel()
					msg := &pb.Ping{Data: int64(s.Id)}
					_, err := (*s.Connections[i]).SendPing(ctx, msg)
					if err != nil {
						failed[i] = true
					}
					wg.Done()
				}(idx)
			}
			// check for failed nodes
			wg.Wait()
			if len(failed) > len(config.Servers)/2 {
				// s.SetState(ELECTION)
				log.Printf("%d lost quorum", s.Id)
				// s.FastElection(1000)
				return
			} else {
				log.Printf("%d has quorum", s.Id)
			}

		case FOLLOWING:
			// Simulate failure
			// fail := rand.Intn(100) < 10
			// if fail {
			// 	log.Printf("%d failed", s.Id)
			// 	return
			// }

			ctx, cancel := s.EstablishConnection(s.Vote.Id, *maxTimeout)
			defer cancel()
			msg := &pb.Ping{Data: int64(s.Id)}
			_, err := (*s.Connections[s.Vote.Id]).SendPing(ctx, msg)
			if err != nil {
				// s.SetState(ELECTION)
				log.Printf("%d lost leader", s.Id)
				// s.FastElection(1000)
				return
			} else {
				log.Printf("%d has leader", s.Id)
			}

		}

		time.Sleep(time.Duration(*maxTimeout) * time.Millisecond)
	}
}
