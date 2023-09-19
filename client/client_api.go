package client

type client interface {
	create()
	delete()
	exists()
	getData()
	setData()
	getChildren()
	sync()
}

func create(path string, data []string, flags []string) string {
	// marshal command + params
	// send to server
	// recieve from server
	// unmarshal
	// return to application

	return ""
}

func delete(path string, version int64) {

}

func exists(path string, watch bool) bool {

	return false
}

func getData(path string, watch bool) {

}

func setData(path string, data []string, version int64) {

}

func getChildren(path string, watch bool) {

}

func sync(path string) {

}
