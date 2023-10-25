package main

import (
	"fmt"

	"github.com/shaunnope/go-jaguard/zouk"
)

func main() {

	// Initialise datatree with root node
	dataTree := zouk.NewDataTree()
	// Check creation of watch
	// getChildren(dataTree, "/", 0, true)

	// node, _ := dataTree.GetNode("/")
	// Check if node has watch
	// watches := node.GetWatches()
	// for _, watch := range watches {
	// 	fmt.Printf("%s is a watch for node %s\n", watch.PrintWatch(), "/")
	// }

	dataTree.CreateNode("/node1", []byte{1, 2, 3, 4}, false, 1, 1, false)
	dataTree.CreateNode("/node1/node2", []byte{1, 2, 3, 4}, false, 1, 1, false)

	for _, node := range dataTree.NodeMap {
		fmt.Printf("Node path has %s\n", node.PrintZnode())
	}

	getChildren(dataTree, "/node1", 0, true)

	event := zouk.Event{
		UserId: 0,
		Type:   zouk.EventType(zouk.Create),
		Path:   "/node1/node2",
	}

	dataTree.CheckWatchTrigger(&event)

	// // Create node on top of existing root node
	// dataTree.CreateNode("/node1", []byte{1, 2, 3, 4}, false, 1, 1, false)
	// fmt.Printf("Node path has %s\n", dataTree.GetPaths())

	// // Create node on top of existing root node
	// dataTree.CreateNode("/node1", []byte{1, 2, 3, 4}, false, 1, 1, true)
	// dataTree.CreateNode("/node1", []byte{1, 2, 3, 4}, false, 1, 1, true)
	// dataTree.CreateNode("/node1", []byte{1, 2, 3, 4}, false, 1, 1, true)
	// fmt.Println("--------Data tree before deletion:--------")
	// fmt.Printf("Node path has %s\n", dataTree.GetPaths())

	// // Create node on top of existing root node
	// fmt.Println("--------Data tree before deletion:--------")
	// fmt.Printf("Node path has %s\n", dataTree.GetPaths())

	// // Set Data
	// fmt.Println("--------Updating data:--------")
	// data, _ := dataTree.GetData("/node1")
	// fmt.Printf("Root node added has data: %d\n", data)
	// dataTree.SetData("/node1", []byte{2, 3, 4, 5}, 2, 1)
	// data, _ = dataTree.GetData("/node1")
	// fmt.Printf("Root node added has new data: %d\n", data)

	// // Delete Node
	// dataTree.DeleteNode("/node1", 0)
	// fmt.Println("--------Data tree after deletion:--------")
	// fmt.Printf("Node path has %s\n", dataTree.GetPaths())

}

// .
// .
// .
// .
// .
// .
//// clientAPI (including setting of watch)

// ls /node -w -s
func getChildren(dataTree *zouk.DataTree, path string, clientID int64, setWatch bool) (map[string]bool, error) {
	if setWatch {
		watch := &zouk.Watch{
			Type:     zouk.GetChildren,
			Path:     path,
			ClientId: clientID,
		}
		dataTree.AddWatchToNode(path, watch)
	}
	return dataTree.GetNodeChildren(path)
}

// get /node
func getData(dataTree *zouk.DataTree, path string, clientID int64, setWatch bool) ([]byte, error) {
	if setWatch {
		watch := &zouk.Watch{
			Type:     zouk.GetData,
			Path:     path,
			ClientId: clientID,
		}
		dataTree.AddWatchToNode(path, watch)
	}
	return dataTree.GetData(path)
}

// check if exists /node
func exists(dataTree *zouk.DataTree, path string, clientID int64, setWatch bool) (*zouk.Znode, error) {
	if setWatch {
		watch := &zouk.Watch{
			Type:     zouk.Exists,
			Path:     path,
			ClientId: clientID,
		}
		dataTree.AddWatchToNode(path, watch)
	}
	return dataTree.GetNode(path)
}
