package main

import (
	"context"
	"fmt"

	"github.com/shaunnope/go-jaguard/zouk"
	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		if err == nil {
			break
		}
	}
	if err != nil {
		msg.Error(s.Id, to, err)
		return r, err
	}
	msg.Done(s.Id, to)
	return r, nil
}

func TriggerWatch(watch *zouk.Watch, operationType pb.OperationType) {
	fmt.Printf("Sending watch gRPC call\n")
	callbackAddr := fmt.Sprintf("%s:%s", watch.ClientAddr.Host, watch.ClientAddr.Port)
	conn, err := grpc.Dial(callbackAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		fmt.Println("Couldnt connect to zkclient")
	}
	defer conn.Close()
	client := pb.NewZkCallbackClient(conn)
	client.NotifyWatchTrigger(context.Background(), &pb.WatchNotification{Path: watch.Path, OperationType: operationType})
}
