package main

import (
	"context"
	"fmt"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

type Client struct {
	pb.UnimplementedZkCallbackServer
}

func (c *Client) NotifyWatchTrigger(ctx context.Context, in *pb.WatchNotification) (*pb.WatchNotificationResponse, error) {
	fmt.Printf("\n Watch Triggered: Watch triggered by %s operation in path %s\n", in.OperationType, in.Path)
	accepted := true
	return &pb.WatchNotificationResponse{Accept: &accepted}, nil
}
