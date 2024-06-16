package main

import (
	"context"
	"fmt"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

type Client struct {
	pb.UnimplementedZkCallbackServer
	Conn *pb.ZkCallbackClient
}

func (c *Client) NotifyWatchTrigger(ctx context.Context, in *pb.WatchNotification) (*pb.WatchNotificationResponse, error) {
	fmt.Printf("\nWATCH: %s @ %s\n", in.OperationType, in.Path)
	accepted := true
	return &pb.WatchNotificationResponse{Accept: &accepted}, nil
}
