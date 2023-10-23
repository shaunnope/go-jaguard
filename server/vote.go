// Vote related definitions
package main

import pb "github.com/shaunnope/go-jaguard/zouk"

type Vote = pb.VoteFragment

type VoteLog struct {
	Vote    Vote
	Round   int
	Version int
}

type VoteMsg = *pb.ElectNotification
