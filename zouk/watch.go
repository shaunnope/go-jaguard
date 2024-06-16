package zouk

import "fmt"

// WatchType represents the type of a ZooKeeper watch event.
type WatchType int

var WatchTypeMap = map[WatchType]string{
	GetData:     "GetData",
	GetChildren: "GetChildren",
	Exists:      "Exists",
}

const (
	GetData WatchType = iota
	GetChildren
	Exists
)

type Addr struct {
	Host string
	Port string
}

// Watch represents a ZooKeeper watch event with relevant information.
type Watch struct {
	Type       WatchType // The type of the watch event.
	Path       string    // The path to the znode associated with the event.
	ClientAddr Addr
}

func (watch *Watch) PrintWatch() string {
	return fmt.Sprintf("[[[[Watch Type: %s, Path: %s, Client Addr: %s]]]]", WatchTypeMap[watch.Type], watch.Path, watch.ClientAddr)
}
