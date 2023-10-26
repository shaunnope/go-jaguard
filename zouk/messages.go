package zouk

import "log"

// Interface for messages
type Message interface {
	Error(from int, to int, err error)
	Done(from int, to int)
}

func (m *NewEpoch) Error(from int, to int, err error) {
	log.Printf("%d error sending epoch to %d: %v", from, to, err)
}

func (m *NewEpoch) Done(from int, to int) {
	log.Printf("%d received epoch from %d", from, to)
}

func (m *FollowerInfo) Error(from int, to int, err error) {
	log.Printf("%d error sending follower info to %d: %v", from, to, err)
}

func (m *FollowerInfo) Done(from int, to int) {
	log.Printf("%d received follower info from %d", from, to)
}

func (m *Ping) Error(from int, to int, err error) {
	log.Printf("%d error sending ping to %d: %v", from, to, err)
}

func (m *Ping) Done(from int, to int) {
	log.Printf("%d received ping from %d", from, to)
}
