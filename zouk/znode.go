package zouk

import (
	"errors"
	"fmt"
)

type Stat struct {
	Czxid          int64 // created zxid
	Mzxid          int64 // last modified zxid
	Pzxid          int64 // last modified children
	Ctime          int64 // created time
	Mtime          int64 // last modified time
	Version        int64 // version
	Cversion       int64 // child version
	Aversion       int64 // acl version
	EphemeralOwner int64 // owner id if ephermeral, 0 otw
	DataLength     int32 // length of the data in the node
	NumChildren    int32 // number of children of this node
}

type Znode struct {
	stat        Stat
	children    map[string]bool
	parent      string
	data        []byte
	eph         bool
	id          string
	sequenceNum int64
	watches     []*Watch
}

func NewNode(stat Stat, parent string, data []byte, isEphemeral bool, id string, isSequence bool) Znode {
	node := Znode{
		stat:        stat,
		children:    map[string]bool{},
		parent:      parent,
		data:        data,
		eph:         isEphemeral,
		id:          id,
		sequenceNum: 0,
		watches:     []*Watch{},
	}
	return node
}

func (znode *Znode) AddChild(child string) map[string]bool {
	znode.children[child] = true
	return znode.children
}

func (znode *Znode) RemoveChild(child string) map[string]bool {
	delete(znode.children, child)
	return znode.children
}

func (znode *Znode) ChildExists(childName string) bool {
	_, exists := znode.children[childName]
	return exists
}

func (znode *Znode) GetChildren() map[string]bool {
	copyChildren := map[string]bool{}
	for key, value := range znode.children {
		copyChildren[key] = value
	}
	return copyChildren
}

func CreateStat(zxid int64, time int64, ephemeralOwner int64) Stat {
	stat := Stat{
		Czxid:          zxid,
		Mzxid:          zxid,
		Pzxid:          zxid,
		Ctime:          time,
		Mtime:          time,
		Cversion:       0,
		Aversion:       0,
		EphemeralOwner: ephemeralOwner,
	}
	return stat
}

func CopyStat(stat Stat) Stat {
	copyStat := Stat{
		Czxid:          stat.Czxid,
		Mzxid:          stat.Mzxid,
		Pzxid:          stat.Pzxid,
		Ctime:          stat.Ctime,
		Mtime:          stat.Mtime,
		Cversion:       stat.Cversion,
		Aversion:       stat.Aversion,
		EphemeralOwner: stat.EphemeralOwner,
	}
	return copyStat
}

// IsEphemeral returns true if the Znode is ephemeral.
func (znode *Znode) IsEphemeral() bool {
	return znode.eph
}

// GetID returns the ID of the Znode.
func (znode *Znode) GetID() string {
	return znode.id
}

// GetSequenceNum returns the sequence number.
func (znode *Znode) GetSequenceNum() int64 {
	return znode.sequenceNum
}

// GetParent returns the parent path.
func (znode *Znode) GetParent() string {
	return znode.parent
}

// GetStat returns a copy of the stat.
func (znode *Znode) GetStat() Stat {
	return CopyStat(znode.stat)
}

// GetParent returns a copy of the data.
func (znode *Znode) GetData() []byte {
	copiedData := make([]byte, len(znode.data))
	copy(copiedData, znode.data)
	return copiedData
}

// SetData sets the data to a copy of the data input.
func (znode *Znode) SetData(data []byte) []byte {
	updatedData := make([]byte, len(data))
	copy(updatedData, data)
	znode.data = updatedData
	return data
}

// Add Watch to the Znode + check if watch already exists
func (znode *Znode) AddWatch(watch *Watch) (string, error) {
	for _, clientWatch := range znode.watches {
		if clientWatch.ClientId == watch.ClientId && clientWatch.Type == watch.Type {
			return "", errors.New("watch already exists")
		}
	}
	znode.watches = append(znode.watches, watch)
	return "ok", nil
}

func (znode *Znode) GetWatches() []*Watch {
	return znode.watches
}

func (znode *Znode) SetWatches(watches []*Watch) {
	znode.watches = watches
}

// PrintZnode returns a string with information about a Znode, including statistics and watches.
func (znode *Znode) PrintZnode() string {
	znodeInfo := "	Znode Information:\n"
	znodeInfo += fmt.Sprintf("		Path: %s \n", znode.id)
	znodeInfo += fmt.Sprintf("		Statistics: %+v \n", znode.stat)
	znodeInfo += fmt.Sprintf("		Data: %s \n", string(znode.data))
	znodeInfo += fmt.Sprintf("		Ephemeral: %v \n", znode.eph)
	znodeInfo += fmt.Sprintf("		Sequence Number: %d\n", znode.sequenceNum)

	// watchInfo := "Watches:\n"
	// for _, watch := range znode.watches {
	// 	watchInfo += fmt.Sprintf("- %s\n", watch.PrintWatch())
	// }
	// return znodeInfo + watchInfo
	return znodeInfo
}
