package main

import (
	"fmt"
	"log/slog"
	mu "sync"

	"github.com/shaunnope/go-jaguard/utils"
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
	FollowerInfoQ  chan *pb.FollowerInfo
	QuorumReady    chan bool
	BroadcastReady chan bool
	HasQuorum      bool
	Abort          chan bool
}

func NewZabSession() ZabSession {
	return ZabSession{
		FollowerEpochs: make(map[int]int),
		FollowerInfoQ:  make(chan *pb.FollowerInfo, 10),
		QuorumReady:    make(chan bool),
		BroadcastReady: make(chan bool),
		HasQuorum:      false,
		Abort:          make(chan bool),
	}
}

func (l *ZabSession) Reset() {
	l.Lock()
	defer l.Unlock()
	slog.Info("Resetting zab session")
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

	// channels to trigger events
	// Stop the server
	Stop chan bool
	// Elect a new leader
	Reelect chan bool

	// states related to Zab Session
	Zab  ZabSession
	Data *pb.DataTree

	// path to memory
	Path string
}

func NewStateVector(idx int) StateVector {
	return StateVector{
		Id:          idx,
		Queue:       make(chan VoteMsg, maxElectionNotifQueueSize),
		Connections: make(map[int]*pb.NodeClient),
		Zab:         NewZabSession(),
		Data:        pb.NewDataTree(),
		Path:        fmt.Sprintf("%s/s%d/", *logDir, idx),
		Stop:        make(chan bool),
		Reelect:     make(chan bool),
		Vote: pb.VoteFragment{
			LastZxid: pb.ZxidFragment{
				Epoch:   0,
				Counter: 0,
			},
			Id: idx,
		},
	}
}

// State and Vote can only be modified in FLE with lock
func (sv *StateVector) SetStateAndVote(state State, vote *Vote) {
	sv.Lock()
	defer sv.Unlock()
	if state == ELECTION {
		if vote != nil {
			panic("vote should be nil when setting ELECTION state")
		}
		sv.Round++
	}
	if vote == nil {
		vote = &Vote{LastZxid: sv.LastZxid, Id: sv.Id}
	}
	if err := sv.SaveState(data_STATE, vote.Marshal()); err != nil {
		slog.Error("SetStateAndVote", "err", err)
	} else {
		if sv.State == LEADING && state != LEADING {
			close(sv.Zab.Abort)
		}
		sv.State = state
		sv.Vote = *vote
	}

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
		vote = &Vote{LastZxid: sv.LastZxid, Id: sv.Id}
	}
	sv.SaveState(data_STATE, vote.Marshal())
	sv.Vote = *vote
}

func (sv *StateVector) GetVote() Vote {
	sv.Lock()
	defer sv.Unlock()
	return sv.Vote
}

func (sv *StateVector) IncRound(atomic bool) int {
	sv.Lock()
	defer sv.Unlock()
	sv.Round++
	return sv.Round
}

func (sv *StateVector) SetAcceptedEpoch(epoch int) {
	data := make([]byte, 16)
	copy(data[0:8], utils.MarshalInt(epoch))
	copy(data[8:16], utils.MarshalInt(sv.CurrentEpoch))
	if err := sv.SaveState(data_EPOCH, data); err != nil {
		slog.Error("SetAcceptedEpoch", "err", err)
	} else {
		slog.Debug("SetAcceptedEpoch", "epoch", epoch)
		sv.AcceptedEpoch = epoch
	}
}

func (sv *StateVector) SetEpochs(accepted *int, current *int) {
	data := make([]byte, 16)
	if accepted == nil {
		accepted = &sv.AcceptedEpoch
	}
	if current == nil {
		current = &sv.CurrentEpoch
	}

	copy(data[0:8], utils.MarshalInt(*accepted))
	copy(data[8:16], utils.MarshalInt(*current))
	if err := sv.SaveState(data_EPOCH, data); err != nil {
		slog.Error("SetEpochs", "s", sv.Id, "err", err)
	} else {
		slog.Info("SetEpochs", "s", sv.Id, "a", *accepted, "c", *current)
		sv.AcceptedEpoch = *accepted
		sv.CurrentEpoch = *current
	}
}

func (sv *StateVector) SetLastZxid(zxid pb.ZxidFragment) {
	if err := sv.SaveState(data_ZXID, zxid.Marshal()); err != nil {
		slog.Error("SetLastZxid", "err", err)
	} else {
		slog.Debug("SetLastZxid", "zxid", zxid)
		sv.LastZxid = zxid
	}
}

func (sv *StateVector) ReplaceHistory(history []*pb.Transaction) {
	sv.History.Set(pb.Transactions(history).ExtractAll())
}
