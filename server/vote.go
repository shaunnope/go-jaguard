// Vote related definitions
package main

import pb "github.com/shaunnope/go-jaguard/zouk"

type VoteLog struct {
	Vote    *pb.Vote
	Round   int
	Version int
}

type VoteMsg = *pb.ElectNotification
