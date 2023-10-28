package main

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"log"
	"math/rand"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// Simulate state evolution
func Simulate(s *Server) {
	addr := fmt.Sprintf("%s:%d", config.Servers[s.Id].Host, config.Servers[s.Id].Port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("c%d failed to connect: %v", s.Id, err)
	}
	c := pb.NewNodeClient(conn)

	for {
		if rand.Intn(100) < 10 {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*maxTimeout)*time.Millisecond)
				defer cancel()

				data := make([]byte, 10)
				_, err := crand.Read(data)
				if err != nil {
					log.Printf("c%d error generating random data: %v", s.Id, err)
				}

				req := &pb.ZabRequest{
					Transaction: &pb.Transaction{
						Path: "/foo",
						Data: data,
					},
					RequestType: pb.RequestType_CLIENT,
				}

				_, err = c.SendZabRequest(ctx, req)
				if err != nil {
					log.Printf("c%d error sending zab request: %v", s.Id, err)
				}
			}()
		}
		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
	}
}
