package main

import (
	"log"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

// Manual setup of server states
func (s *Server) Setup(vote pb.VoteFragment) {
	s.Lock()
	defer s.Unlock()
	s.Vote = vote
	if s.Id == vote.Id {
		s.State = LEADING
		log.Printf("server %d is leader", s.Id)
	} else {
		s.State = FOLLOWING
		log.Printf("server %d is following %v", s.Id, s.Vote)
	}

	for idx := range config.Servers {
		if idx == s.Id {
			continue
		}
		// TODO: make this async
		// issue: concurrent map writes
		s.EstablishConnection(idx, *maxTimeout)
	}
}
