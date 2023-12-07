package main

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"log"
	"time"

	pb "github.com/shaunnope/go-jaguard/zouk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Setup of server connections
func (s *Server) Setup() {
	s.Lock()
	defer s.Unlock()

	for idx := range config.Servers {
		if idx == s.Id {
			continue
		}
		// TODO: make this async
		// issue: concurrent map writes
		s.EstablishConnection(idx, *maxTimeout)
	}
}

// Simulate state evolution - a client (which needs to be moved to a separate thread) sending a request to zookeeper *Server
func Simulate(s *Server, path string) {
	addr := fmt.Sprintf("%s:%d", config.Servers[s.Id].Host, config.Servers[s.Id].Port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("c%d failed to connect: %v", s.Id, err)
	}
	c := pb.NewNodeClient(conn)

	time.Sleep(time.Duration(5000) * time.Millisecond)

	// Preparing client request with context and request (calling ZabRequest as a gRPC from the fake "client")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*maxTimeout)*time.Millisecond)
		defer cancel()

		data := make([]byte, 10)
		_, err := crand.Read(data)
		if err != nil {
			log.Printf("c%d error generating random data: %v", s.Id, err)
		}

		req := &pb.CUDSRequest{
			Path:          path,
			Data:          data,
			Flags:         &pb.Flag{IsSequential: false, IsEphemeral: false},
			OperationType: pb.OperationType_WRITE,
		}

		cudReply, err := c.HandleClientCUDS(ctx, req)
		if err != nil {
			log.Printf("%d error sending zab request: %v", s.Id, err)
		} else {
			log.Printf("Writing new node in path %s is :%t", path, *cudReply.Accept)
		}

		// Preparing client to try and getChildren
		//TODO: Have to use a call as a client using IP and port of zk server instead of calling by connections[to]
		getChildrenReply, err := SendGrpc(pb.NodeClient.GetChildren, s, s.Vote.Id, &pb.GetChildrenRequest{Path: "/", SetWatch: false}, *maxTimeout)
		fmt.Printf("READ: %s are the children of '/'\n", getChildrenReply.Children)
		if err != nil {
			log.Printf("%d error sending read request: %v\n", s.Id, err)
			return
		}

		path = "/foo"
		getExistReply, err := SendGrpc(pb.NodeClient.GetExists, s, s.Vote.Id, &pb.GetExistsRequest{Path: path, SetWatch: false}, *maxTimeout)
		fmt.Printf("READ: %s exists: %t\n", path, getExistReply.Exists)
		if err != nil {
			log.Printf("%d error sending read request: %v\n", s.Id, err)
			return
		}

		getDataReply, err := SendGrpc(pb.NodeClient.GetData, s, s.Vote.Id, &pb.GetDataRequest{Path: path, SetWatch: false}, *maxTimeout)
		fmt.Printf("READ: Data of %s is: %v\n", path, getDataReply.Data)
		if err != nil {
			log.Printf("%d error sending read request: %v\n", s.Id, err)
			return
		}
	}()

}
