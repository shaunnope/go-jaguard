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
	mu.Mutex
	FollowerEpochs map[int]int
	QuorumReady    chan bool
	HasQuorum      bool
}

func (l *ZabLeader) Reset() {
	l.Lock()
	defer l.Unlock()
	l.FollowerEpochs = make(map[int]int)
	l.QuorumReady = make(chan bool)
	l.HasQuorum = false
}

type Transactions = pb.TransactionFragments

type StateVector struct {
	mu.Mutex
	Id       int
	State    State
	Round    int
	LastZxid pb.ZxidFragment // last proposal
	Vote     Vote
	Queue    chan VoteMsg

	Connections map[int]*pb.NodeClient

	History       Transactions
	AcceptedEpoch int // last NewEpoch
	CurrentEpoch  int // last NewLeader

	// channel to stop server
	Stop chan bool

	Leader ZabLeader

	// TODO: save data tree to disk
	Data *pb.DataTree
}

func NewStateVector(idx int) StateVector {
	return StateVector{
		Id:          idx,
		Queue:       make(chan VoteMsg, maxElectionNotifQueueSize),
		Connections: make(map[int]*pb.NodeClient),
		Leader:      ZabLeader{FollowerEpochs: make(map[int]int), QuorumReady: make(chan bool)},
		Data:        pb.NewDataTree(),
		Stop:        make(chan bool),
		Vote: pb.VoteFragment{
			LastZxid: pb.ZxidFragment{
				Epoch:   0,
				Counter: 0,
			},
			Id: idx,
		},
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
