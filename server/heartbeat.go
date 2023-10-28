package main

import (
	"log"
	"sync"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func (s *Server) Heartbeat() {

	// trigger shutdown
	// defer func() {
	// 	s.Stop <- true
	// }()

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
				log.Printf("%d lost quorum", s.Id)
				s.FastElection(*maxTimeout)
				return
				// } else {
				// 	log.Printf("%d has quorum", s.Id)
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
				log.Printf("%d lost leader", s.Id)
				s.FastElection(*maxTimeout)
				return
				// } else {
				// 	log.Printf("%d has leader", s.Id)
			}

		}

		time.Sleep(time.Duration(*maxTimeout) * time.Millisecond)
	}
}
