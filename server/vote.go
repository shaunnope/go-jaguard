// Vote related definitions
package main

import pb "github.com/shaunnope/go-jaguard/zouk"

type Vote struct {
	LastZxid Zxid
	Id       int
}

type VoteLog struct {
	vote    Vote
	round   int
	version int
}

type VoteMsg struct {
	vote  *Vote
	id    int
	state State
	round int
}

func (v *Vote) LessThan(other Vote) bool {
	return v.LastZxid.LessThan(other.LastZxid) || (v.LastZxid.Equal(other.LastZxid) && v.Id < other.Id)
}

func (v *Vote) Equal(other Vote) bool {
	return v.LastZxid.Equal(other.LastZxid) && v.Id == other.Id
}

func (v *Vote) GreaterThan(other Vote) bool {
	return !v.LessThan(other) && !v.Equal(other)
}

func VoteFrom(raw *pb.Vote) *Vote {
	return &Vote{LastZxid: *ZxidFrom(raw.LastZxid), Id: int(raw.Id)}
}

func (v *Vote) Raw() *pb.Vote {
	return &pb.Vote{LastZxid: v.LastZxid.Raw(), Id: int64(v.Id)}
}

func VoteMsgFrom(in *pb.ElectNotification) *VoteMsg {
	return &VoteMsg{
		vote:  VoteFrom(in.Vote),
		id:    int(in.Id),
		state: State(in.State),
		round: int(in.Round),
	}
}
