package server

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
	children []int64
	parent   int64
	data     []string
	eph      bool
	id       int64
}
