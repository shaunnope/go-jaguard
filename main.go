package main

import (
	"fmt"

	pb "github.com/shaunnope/go-jaguard/zouk"
)

func main() {

	// Initialise datatree with root node
	dataTree := pb.NewDataTree()
	fmt.Println("--------Data tree after initialisation:--------")
	for k := range dataTree.NodeMap {
		fmt.Printf("Node path has %s\n", k)
	}

	// Create node on top of existing root node
	dataTree.CreateNode("/node1", []byte{1, 2, 3, 4}, false, 1, pb.ZxidFragment{}, false)
	fmt.Println("--------Data tree before deletion:--------")
	for k := range dataTree.NodeMap {
		fmt.Printf("Node path has %s\n", k)
	}

	// Create node on top of existing root node
	dataTree.CreateNode("/node1", []byte{1, 2, 3, 4}, false, 1, pb.ZxidFragment{}, true)
	fmt.Println("--------Data tree before deletion:--------")
	for k := range dataTree.NodeMap {
		fmt.Printf("Node path has %s\n", k)
	}

	// Create node on top of existing root node
	fmt.Println("--------Data tree before deletion:--------")
	for k := range dataTree.NodeMap {
		fmt.Printf("Node path has %s\n", k)
	}

	// Set Data
	fmt.Println("--------Updating data:--------")
	fmt.Printf("Root node added has data: %d\n", dataTree.GetData("/node1"))
	dataTree.SetData("/node1", []byte{2, 3, 4, 5}, 2, pb.ZxidFragment{})
	fmt.Printf("Root node added has new data: %d\n", dataTree.GetData("/node1"))

	// Delete Node
	dataTree.DeleteNode("/node1", 0)
	fmt.Println("--------Data tree after deletion:--------")
	for k := range dataTree.NodeMap {
		fmt.Printf("Node path has %s\n", k)
	}

}
