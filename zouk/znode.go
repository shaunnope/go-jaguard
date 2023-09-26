package zouk

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

	// remember to declare and assign array as slice
	// example:
	// type house struct {
	// 	s []string
	// }

	// func main() {
	// 	h := house{}
	// 	a := make([]string, 3)
	// 	h.s = a
	// }
	Stat     Stat
	Children []int64
	Parent   int64
	Data     []byte
	Eph      bool
	Id       int64
}