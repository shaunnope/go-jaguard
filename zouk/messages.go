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
	// log.Printf("%d sent epoch to %d", from, to)
}

func (m *AckEpoch) Error(from int, to int, err error) {
	log.Printf("%d error sending epoch ack to %d: %v", from, to, err)
}

func (m *AckEpoch) Done(from int, to int) {
	// log.Printf("%d sent epoch ack to %d", from, to)
}

func (m *NewLeader) Error(from int, to int, err error) {
	log.Printf("%d error sending leader to %d: %v", from, to, err)
}

func (m *NewLeader) Done(from int, to int) {
	// log.Printf("%d sent leader to %d", from, to)
}

func (m *AckLeader) Error(from int, to int, err error) {
	log.Printf("%d error sending leader ack to %d: %v", from, to, err)
}

func (m *AckLeader) Done(from int, to int) {
	// log.Printf("%d sent leader ack to %d", from, to)
}

func (m *FollowerInfo) Error(from int, to int, err error) {
	log.Printf("%d error sending follower info to %d: %v", from, to, err)
}

func (m *FollowerInfo) Done(from int, to int) {
	// log.Printf("%d sent follower info to %d", from, to)
}

func (m *Ping) Error(from int, to int, err error) {
	log.Printf("%d error sending ping to %d: %v", from, to, err)
}

func (m *Ping) Done(from int, to int) {
	// log.Printf("%d received ping from %d", from, to)
}

func (m *ElectNotification) Error(from int, to int, err error) {
	log.Printf("%d error sending vote notif to %d: %v", from, to, err)
}

func (m *ElectNotification) Done(from int, to int) {
	// log.Printf("%d sent vote notif to %d", from, to)
}

func (m *ElectResponse) Error(from int, to int, err error) {
	log.Printf("%d error sending vote response to %d: %v", from, to, err)
}

func (m *ElectResponse) Done(from int, to int) {
	// log.Printf("%d sent vote response to %d", from, to)
}

func (m *ZabRequest) Error(from int, to int, err error) {
	log.Printf("%d error sending zab request to %d: %v", from, to, err)
}

func (m *ZabRequest) Done(from int, to int) {
	log.Printf("%d sent zab request to %d: %v %v", from, to, m.RequestType, m.Transaction)
}

func (m *ZabAck) Error(from int, to int, err error) {
	log.Printf("%d error sending zab ack to %d: %v", from, to, err)
}

func (m *ZabAck) Done(from int, to int) {
	// log.Printf("%d sent zab ack to %d", from, to)
}
