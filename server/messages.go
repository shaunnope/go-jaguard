package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ContextKey string

// Establish connection to another server if it does not already exist. Returns a context and cancel function
func (s *Server) EstablishConnection(to int, timeout int) (context.Context, context.CancelFunc, *pb.NodeClient) {
	if to == s.Id {
		return nil, nil, nil
	}
	addr := fmt.Sprintf("%s:%d", config.Servers[to].Host, config.Servers[to].Port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("%d failed to connect to %d: %v", s.Id, to, err)
	}
	c := pb.NewNodeClient(conn)
	ctx := context.WithValue(context.Background(), ContextKey("from"), s.Id)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	return ctx, cancel, &c
}

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
	send := func() (R, error) {
		ctx, cancel, conn := s.EstablishConnection(to, timeout)
		// conn := s.Connections[to]
		defer cancel()
		return F(*conn, ctx, msg)
	}
	for count := 0; err == nil && count < maxRetries; count++ {
		r, err = send()
		time.Sleep(100 * time.Microsecond)
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

// func TriggerWatch(watch *pb.Watch, operationType pb.OperationType) {
// 	fmt.Printf("Sending watch gRPC call\n")
// 	callbackAddr := fmt.Sprintf("%s:%s", watch.ClientAddr.Host, watch.ClientAddr.Port)
// 	conn, err := grpc.Dial(callbackAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		slog.Error("TriggerWatch", "err", err)
// 		return
// 	}
// 	defer conn.Close()

// 	client := pb.NewZkCallbackClient(conn)
// 	_, err = client.NotifyWatchTrigger(context.Background(), &pb.WatchNotification{Path: watch.Path, OperationType: operationType})
// 	if err != nil {
// 		slog.Error("TriggerWatch", "err", err)
// 	}
// }

func TriggerWatch(watch *pb.Watch, operationType pb.OperationType) {
	fmt.Printf("Sending watch gRPC call\n")
	callbackAddr := fmt.Sprintf("%s:%s", watch.ClientAddr.Host, watch.ClientAddr.Port)

	// Maximum number of attempts
	maxAttempts := 3
	var err error

	conn, dialErr := grpc.Dial(callbackAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if dialErr != nil {
		slog.Error("TriggerWatch", "err", dialErr)
		return
	}
	defer conn.Close()

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		client := pb.NewZkCallbackClient(conn)
		_, err = client.NotifyWatchTrigger(context.Background(), &pb.WatchNotification{Path: watch.Path, OperationType: operationType})
		if err != nil {
			slog.Error("TriggerWatch", "err", err)
			// Sleep for some time before retrying
			time.Sleep(time.Second * 2)
		} else {
			// The gRPC call was successful, break out of the loop
			break
		}
	}

	// Print error after maximum attempts have been reached
	if err != nil {
		slog.Error("TriggerWatch", "err", "Max attempts reached, unable to perform gRPC call")
	}
}
