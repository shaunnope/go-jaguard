package zouk

type Stat struct {
	czxid          int64
	mzxid          int64
	pzxid          int64
	ctime          int64
	mtime          int64
	version        int64
	cversion       int64
	aversion       int64
	ephemeralOwner int64
	dataLength     int32
	numChildren    int32
}

type znode struct {

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
	stat     Stat
	children []int64
	parent   int64
	data     []byte
	eph      bool
	id       int64
}
