package zouk

import (
	"log"
	"log/slog"
)

// Interface for messages
type Message interface {
	Error(from int, to int, err error)
	Done(from int, to int)
}

func (m *NewEpoch) Error(from int, to int, err error) {
	slog.Error("NewEpoch", "from", from, "to", to, "err", err)
	// log.Printf("%d error sending epoch to %d: %v", from, to, err)
}

func (m *NewEpoch) Done(from int, to int) {
	slog.Debug("NewEpoch", "from", from, "to", to)
}

func (m *AckEpoch) Error(from int, to int, err error) {
	log.Printf("%d error sending epoch ack to %d: %v", from, to, err)
}

func (m *AckEpoch) Done(from int, to int) {
	slog.Debug("AckEpoch", "from", from, "to", to)
}

func (m *NewLeader) Error(from int, to int, err error) {
	log.Printf("%d error sending leader to %d: %v", from, to, err)
}

func (m *NewLeader) Done(from int, to int) {
	slog.Debug("NewLeader", "from", from, "to", to)
}

func (m *AckLeader) Error(from int, to int, err error) {
	log.Printf("%d error sending leader ack to %d: %v", from, to, err)
}

func (m *AckLeader) Done(from int, to int) {
	// log.Printf("%d sent leader ack to %d", from, to)
	slog.Debug("AckLeader", "from", from, "to", to)
}

func (m *FollowerInfo) Error(from int, to int, err error) {
	slog.Error("FollowerInfo", "from", from, "to", to, "err", err)
}

func (m *FollowerInfo) Done(from int, to int) {
	slog.Debug("FollowerInfo", "from", from, "to", to)
}

func (m *Ping) Error(from int, to int, err error) {
	log.Printf("%d error sending ping to %d: %v", from, to, err)
}

func (m *Ping) Done(from int, to int) {
	slog.Debug("Ping", "from", from, "to", to)
}

func (m *ElectNotification) Error(from int, to int, err error) {
	log.Printf("%d error sending vote notif to %d: %v", from, to, err)
}

func (m *ElectNotification) Done(from int, to int) {
	slog.Debug("ElectNotification", "from", from, "to", to)
}

func (m *ElectResponse) Error(from int, to int, err error) {
	log.Printf("%d error sending vote response to %d: %v", from, to, err)
}

func (m *ElectResponse) Done(from int, to int) {
	slog.Debug("ElectResponse", "from", from, "to", to)
}

func (m *ZabRequest) Error(from int, to int, err error) {
	log.Printf("%d error sending zab request to %d: %v", from, to, err)
}

func (m *ZabRequest) Done(from int, to int) {
	log.Printf("server %d zab request to %d has been completed: %v %s", from, to, m.RequestType, m.Transaction.ExtractLogString())
}

func (m *ZabAck) Error(from int, to int, err error) {
	log.Printf("%d error sending zab ack to %d: %v", from, to, err)
}

func (m *ZabAck) Done(from int, to int) {
	slog.Debug("ZabAck", "from", from, "to", to)
}

func (m *GetChildrenRequest) Error(from int, to int, err error) {
	log.Printf("%d error sending getChildren request to %d: %v", from, to, err)
}

func (m *GetChildrenRequest) Done(from int, to int) {
	slog.Debug("GetChildrenRequest", "from", from, "to", to)
}

func (m *GetChildrenResponse) Error(from int, to int, err error) {
	log.Printf("%d error sending getChildren response to %d: %v", from, to, err)
}

func (m *GetChildrenResponse) Done(from int, to int) {
	slog.Debug("GetChildrenResponse", "from", from, "to", to)
}

func (m *GetDataRequest) Error(from int, to int, err error) {
	log.Printf("%d error sending getData request to %d: %v", from, to, err)
}

func (m *GetDataRequest) Done(from int, to int) {
	slog.Debug("GetDataRequest", "from", from, "to", to)
}

func (m *GetDataResponse) Error(from int, to int, err error) {
	log.Printf("%d error sending getData response to %d: %v", from, to, err)
}

func (m *GetDataResponse) Done(from int, to int) {
	slog.Debug("GetDataResponse", "from", from, "to", to)
}

func (m *GetExistsRequest) Error(from int, to int, err error) {
	log.Printf("%d error sending getExists request to %d: %v", from, to, err)
}

func (m *GetExistsRequest) Done(from int, to int) {
	slog.Debug("GetExistsRequest", "from", from, "to", to)
}

func (m *GetExistsResponse) Error(from int, to int, err error) {
	log.Printf("%d error sending getExists response to %d: %v", from, to, err)
}

func (m *GetExistsResponse) Done(from int, to int) {
	slog.Debug("GetExistsResponse", "from", from, "to", to)
}

func (m *CUDRequest) Error(from int, to int, err error) {
	log.Printf("%d error sending create node request to %d: %v", from, to, err)
}

func (m *CUDRequest) Done(from int, to int) {
	slog.Debug("CUDRequest", "from", from, "to", to)
}

func (m *CUDResponse) Error(from int, to int, err error) {
	log.Printf("%d error sending create node response to %d: %v", from, to, err)
}

func (m *CUDResponse) Done(from int, to int) {
	slog.Debug("CUDResponse", "from", from, "to", to)
}
