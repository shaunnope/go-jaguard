package zouk

type Stat struct {
	Czxid          ZxidFragment // created zxid
	Mzxid          ZxidFragment // last modified zxid
	Pzxid          ZxidFragment // last modified children
	Ctime          int64        // created time
	Mtime          int64        // last modified time
	Version        int64        // version
	Cversion       int64        // child version
	Aversion       int64        // acl version
	EphemeralOwner int64        // owner id if ephermeral, 0 otw
	DataLength     int32        // length of the data in the node
	NumChildren    int32        // number of children of this node
}

type Znode struct {
	Stat        Stat
	Children    map[string]bool
	Parent      string
	Data        []byte
	Eph         bool
	Id          int64
	SequenceNum int64
}

func NewNode(stat Stat, parent string, data []byte, isEphemeral bool, id int64, isSequence bool) Znode {
	node := Znode{
		Stat:     stat,
		Children: map[string]bool{},
		Parent:   parent,
		Data:     data,
		Eph:      isEphemeral,
		//TODO: What is Id for and how is the id of a znode determined
		Id:          id,
		SequenceNum: 0,
	}
	return node
}

func (znode *Znode) AddChild(child string) map[string]bool {
	znode.Children[child] = true
	return znode.Children
}

func (znode *Znode) RemoveChild(child string) map[string]bool {
	delete(znode.Children, child)
	return znode.Children
}

func (znode *Znode) ChildExists(childName string) bool {
	_, exists := znode.Children[childName]
	return exists
}

func (znode *Znode) GetChildren() map[string]bool {
	copyChildren := map[string]bool{}
	for key, value := range znode.Children {
		copyChildren[key] = value
	}
	return copyChildren
}

func CreateStat(zxid ZxidFragment, time int64, ephemeralOwner int64) Stat {
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

// Return true if the Znode is ephemeral.
func (znode *Znode) IsEphemeral() bool {
	return znode.Eph
}

// Return the ID of the Znode.
func (znode *Znode) GetID() int64 {
	return znode.Id
}

// Return the sequence number.
func (znode *Znode) GetSequenceNum() int64 {
	return znode.SequenceNum
}

// Return the parent path.
func (znode *Znode) GetParent() string {
	return znode.Parent
}

// Return a copy of the stat.
func (znode *Znode) GetStat() Stat {
	return CopyStat(znode.Stat)
}

// Return a copy of the data.
func (znode *Znode) GetData() []byte {
	copiedData := make([]byte, len(znode.Data))
	copy(copiedData, znode.Data)
	return copiedData
}

// Set the data to a copy of the data input.
func (znode *Znode) SetData(data []byte) {
	updatedData := make([]byte, len(data))
	copy(updatedData, data)
	znode.Data = updatedData
}
