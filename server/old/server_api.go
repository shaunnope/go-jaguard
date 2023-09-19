// Leader approved actions. Called when message is recieved from leader

package server

type server interface {
	create()
	delete()
	exists()
	getData()
	setData()
	getChildren()
	sync()
}

func create(path string, data []string, flags []string) string {
	// check path
	if exists(path, false) {
		getData(path, false)
	} else {
		// create znodes locally
		setData(path, data, 0)
	}

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
