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

func (s State) String() string {
	switch s {
	case DOWN:
		return "DOWN"
	case ELECTION:
		return "ELECTION"
	case LEADING:
		return "LEADING"
	case FOLLOWING:
		return "FOLLOWING"
	default:
		return "UNKNOWN"
	}
}

// state information relating to zab session
type ZabSession struct {
	mu.Mutex
	FollowerEpochs map[int]int
	QuorumReady    chan bool
	BroadcastReady chan bool
	HasQuorum      bool
	Abort          chan bool
}

func NewZabSession() ZabSession {
	return ZabSession{
		FollowerEpochs: make(map[int]int),
		QuorumReady:    make(chan bool),
		BroadcastReady: make(chan bool),
		HasQuorum:      false,
		Abort:          make(chan bool),
	}
}

func (l *ZabSession) Reset() {
	l.Lock()
	defer l.Unlock()
	clear(l.FollowerEpochs)
	l.QuorumReady = make(chan bool)
	l.BroadcastReady = make(chan bool)
	l.HasQuorum = false
	l.Abort = make(chan bool)
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

	// states related to Zab Session
	Zab ZabSession

	// TODO: save data tree to non-volatile memory
	Data *pb.DataTree
}

func NewStateVector(idx int) StateVector {
	return StateVector{
		Id:          idx,
		Queue:       make(chan VoteMsg, maxElectionNotifQueueSize),
		Connections: make(map[int]*pb.NodeClient),
		Zab:         NewZabSession(),
		Data:        pb.NewDataTree(),
		Stop:        make(chan bool),
		// LastZxid:    pb.ZxidFragment{Epoch: 1, Counter: 0},
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

func (sv *StateVector) SetVote(vote *Vote) {
	sv.Lock()
	defer sv.Unlock()
	if vote == nil {
		sv.Vote = Vote{LastZxid: sv.LastZxid, Id: sv.Id}
	} else {
		sv.Vote = *vote
	}
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
