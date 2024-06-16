package zouk

type EventType int

const (
	Create EventType = iota
	Delete
	Change
	Child
)

type Event struct {
	UserId int64
	Type   EventType
	Path   string
	Zxid   int64
}
