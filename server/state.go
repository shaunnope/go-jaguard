package main

import (
	mu "sync"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

const (
	maxElectionNotifQueueSize = 20
)

type State int

const (
	DOWN State = iota
	ELECTION
	LEADING
	FOLLOWING
)

type ZabLeader struct {
	FollowerEpochs map[int]int
	HasQuorum      chan bool
}

type StateVector struct {
	mu.Mutex
	Id       int
	State    State
	Round    int
	LastZxid pb.ZxidFragment // last proposal
	Vote     Vote
	Queue    chan VoteMsg

	Connections map[int]*pb.NodeClient

	History       []Transaction
	AcceptedEpoch int // last NewEpoch
	CurrentEpoch  int // last NewLeader

	Leader ZabLeader
}

func newStateVector(idx int) StateVector {
	return StateVector{
		Id:          idx,
		Queue:       make(chan VoteMsg, maxElectionNotifQueueSize),
		Connections: make(map[int]*pb.NodeClient),
		Leader:      ZabLeader{FollowerEpochs: make(map[int]int)},
	}
}

func (sv *StateVector) SetState(state State) {
	sv.Lock()
	defer sv.Unlock()
	sv.State = state
}

func (sv *StateVector) GetState() State {
	sv.Lock()
	defer sv.Unlock()
	return sv.State
}

// Election data

func (sv *StateVector) SetVote(atomic bool) {
	if atomic {
		sv.Lock()
		defer sv.Unlock()
	}
	sv.Vote = Vote{LastZxid: sv.LastZxid, Id: sv.Id}
}

func (sv *StateVector) GetVote() Vote {
	sv.Lock()
	defer sv.Unlock()
	return sv.Vote
}

func (sv *StateVector) IncRound(atomic bool) int {
	if atomic {
		sv.Lock()
		defer sv.Unlock()
	}
	sv.Round++
	return sv.Round
}
