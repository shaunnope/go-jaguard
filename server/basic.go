package main

import (
	"log"
)

// Manual setup of server states
func (s *Server) Setup() {
	s.Lock()
	defer s.Unlock()
	if s.Id == len(config.Servers)-1 {
		s.State = LEADING
		s.Vote = Vote{Id: s.Id}
		log.Printf("server %d is leader", s.Id)
	} else {
		s.State = FOLLOWING
		s.Vote = Vote{Id: len(config.Servers) - 1}
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
