// Leader approved actions. Called when message is recieved from leader

package server_api

import (
	"fmt"
)

type server interface{
	create()
	delete()
	exists()
	getData()
	setData()
	getChildren()
	sync()
}

func create(path string, data[] string, flags[]string) string{
	// check path
	if exists(path, watch=false){
		getData()
	}
	else{
		// create znodes locally
		setData()
	}
}

func delete(path string, version int64){

}

func exists(path string, watch bool) bool{

}

func getData(path string, watch bool){

}

func setData(path string, data[]string, version int64){
	
}

func getChildren(path string, watch bool){

}

func sync(path string){

}