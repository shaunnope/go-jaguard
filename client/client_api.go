package client_api

import (
	"fmt"
)

type client interface{
	create()
	delete()
	exists()
	getData()
	setData()
	getChildren()
	sync()
}

func create(path string, data[] string, flags[]string) string{
	
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