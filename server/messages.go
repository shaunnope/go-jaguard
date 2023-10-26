package main

import (
	"context"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
)

// Perform a gRPC call to another server
//
// Attempt to send a message up to 3 times before returning a response/error
func SendGrpc[T pb.Message, R pb.Message](
	F func(pb.NodeClient, context.Context, T, ...grpc.CallOption) (R, error),
	s *Server,
	to int,
	msg T,
	timeout int,
) (R, error) {
	var err error = nil
	var r R
	for count := 0; err == nil && count < 3; count++ {
		ctx, cancel := s.EstablishConnection(to, timeout)
		conn := s.Connections[to]
		defer cancel()
		r, err = F(*conn, ctx, msg)
	}
	if err != nil {
		msg.Error(s.Id, to, err)
		return r, err
	}
	msg.Done(s.Id, to)
	return r, nil

}
