package main

import (
	"log"
)

func (s *Server) BasicPing() {
	s.Lock()
	if s.Id == len(config.Servers)-1 {
		s.State = LEADING
		s.Vote = Vote{Id: s.Id}
		log.Printf("server %d is leader", s.Id)
	} else {
		s.State = FOLLOWING
		s.Vote = Vote{Id: len(config.Servers) - 1}
		log.Printf("server %d is following %v", s.Id, s.Vote)
	}
	s.Unlock()

	s.Heartbeat()
}
