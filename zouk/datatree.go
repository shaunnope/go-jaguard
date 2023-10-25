package zouk

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	PATH_SEP = "/"
)

type DataTree struct {
	NodeMap map[string]*Znode
}

// NewDataTree initializes a new DataTree with a root node.
func NewDataTree() *DataTree {
	rootNode := &Znode{
		//TODO: Change zxid and ephemeral owner
		Stat:     CreateStat(ZxidFragment{}, time.Now().Unix(), 0),
		Children: map[string]bool{},
		Parent:   "",
		Data:     []byte{},
		Eph:      false,
		Id:       PATH_SEP,
	}

	dataTree := DataTree{
		NodeMap: map[string]*Znode{
			PATH_SEP: rootNode,
		},
	}

	return &dataTree
}

// .
// .
// .
// .
// .
// .
//// Basic Operations

// CreateNode creates a node by its path.
func (dataTree *DataTree) CreateNode(path string, data []byte, isEph bool, ephermeralOwner int64, zxid ZxidFragment, isSequence bool) (string, error) {
	lastSlashIndex := strings.LastIndex(path, PATH_SEP)
	parentName := getParentName(path, lastSlashIndex)

	// Check if parent node is ephemeral, return error if ephemeral
	parentNode, ok := dataTree.NodeMap[parentName]
	if !ok {
		return parentName, errors.New("invalid parent name")
	}
	if parentNode.IsEphemeral() {
		log.Printf("%s cannot have a child node as it is ephemeral", parentName)
		return parentName, errors.New("invalid parent name")
	}

	// Sequence node suffix
	if isSequence {
		path = addSequenceNumber(parentNode, path, isSequence)
	}

	childName := path[lastSlashIndex:]
	if parentNode.ChildExists(childName) {
		return childName, errors.New("invalid children as it already exists")
	}

	// fmt.Printf("lastslashindex{%d}, parentName{%s}, childName:{%s}\n", lastSlashIndex, parentName, childName)
	stat := CreateStat(zxid, time.Now().Unix(), ephermeralOwner)
	childNode := NewNode(stat, parentName, data, isEph, path, isSequence)
	parentNode.AddChild(childName)
	dataTree.NodeMap[path] = &childNode
	// fmt.Printf("Inside CreateNode, parentName:%s, childName:%s, data: %d\n", parentName, childName, data)

	return path, nil
}

// DeleteNode deletes a node by its path.
func (dataTree *DataTree) DeleteNode(path string, zxid int64) (string, error) {
	nodeToDelete, ok := dataTree.NodeMap[path]
	if !ok {
		return path, errors.New("node does not exist")
	}
	if len(nodeToDelete.Children) > 0 {
		return path, errors.New("node not empty")
	}

	lastSlashIndex := strings.LastIndex(path, PATH_SEP)
	parentName := path[:lastSlashIndex]
	childName := path[lastSlashIndex:]
	parentNode, ok := dataTree.NodeMap[parentName]
	if !ok {
		return parentName, errors.New("invalid parent name")
	}

	if !parentNode.ChildExists(childName) {
		return childName, errors.New("invalid children as it does not exists")
	}

	parentNode.RemoveChild(childName)
	delete(dataTree.NodeMap, path)

	return "Removed", nil
}

// SetData sets the data of a node by its path.
func (dataTree *DataTree) SetData(path string, data []byte, version int64, zxid ZxidFragment) Stat {
	node := dataTree.NodeMap[path]
	node.SetData(data)

	stat := node.GetStat()
	stat.Mtime = time.Now().Unix()
	stat.Mzxid = zxid
	stat.Version = version

	outStat := CopyStat(stat)
	return outStat
}

// GetNodeChildren gets all children of a node
func (dataTree *DataTree) GetNodeChildren(path string) (map[string]bool, error) {
	parentNode, ok := dataTree.NodeMap[path]
	if !ok {
		return nil, errors.New("node does not exist")
	}

	return parentNode.GetChildren(), nil
}

// GetNode gets the node
func (dataTree *DataTree) GetNode(path string) (*Znode, error) {
	node, ok := dataTree.NodeMap[path]
	if !ok {
		return nil, errors.New("node does not exist")
	}

	return node, nil
}

// GetData gets all data of a node
func (dataTree *DataTree) GetData(path string) ([]byte, error) {
	node := dataTree.NodeMap[path]
	return node.GetData(), nil
}

// AddWatcher adds watcher based on event type to a node
func (dataTree *DataTree) AddWatchToNode(path string, watch *Watch) (string, error) {
	nodeAddWatch, ok := dataTree.NodeMap[path]
	if !ok {
		return path, errors.New("node does not exist")
	}

	nodeAddWatch.AddWatch(watch)
	return "ok", nil
}

func (dataTree *DataTree) CheckWatchTrigger(event *Event) {
	// based on the path of the event, the client, check the parent, check what kind of event it is - like create or delete etc
	// remove the watch
	lastSlashIndex := strings.LastIndex(event.Path, "/")
	parentName := getParentName(event.Path, lastSlashIndex)
	nodeName := event.Path[lastSlashIndex:]
	fmt.Printf("Checking triggers with parentName:%s, nodeName:%s for zxid:%d\n", parentName, nodeName, event.Zxid)

	parentNode, _ := dataTree.GetNode(parentName)
	node, _ := dataTree.GetNode(event.Path)

	// Function to remove triggered watches
	removeTriggeredWatches := func(watches []*Watch, watchType WatchType) []*Watch {
		var remainingWatches []*Watch
		for _, watch := range watches {
			if watch.Type == watchType {
				fmt.Printf("Triggered: %s\n", watch.PrintWatch())
			} else {
				remainingWatches = append(remainingWatches, watch)
			}
		}
		return remainingWatches
	}

	switch event.Type {
	case Create:
		// Check for current node
		node.SetWatches(removeTriggeredWatches(node.GetWatches(), Exists))
		// Check for parent node
		parentNode.SetWatches(removeTriggeredWatches(parentNode.GetWatches(), GetChildren))

		// for _, watch := range node.GetWatches() {
		// 	if watch.Type == Exists {
		// 		fmt.Printf("Triggered: %s\n", watch.PrintWatch())
		// 	}
		// }
		// for _, watch := range parentNode.GetWatches() {
		// 	if watch.Type == GetChildren {
		// 		fmt.Printf("Triggered: %s\n", watch.PrintWatch())
		// 	}
		// }
	case Delete:
		// Check for current node
		node.SetWatches(removeTriggeredWatches(node.GetWatches(), Exists))
		node.SetWatches(removeTriggeredWatches(node.GetWatches(), GetData))
		// Check for parent node
		parentNode.SetWatches(removeTriggeredWatches(parentNode.GetWatches(), GetChildren))

		// for _, watch := range node.GetWatches() {
		// 	if watch.Type == Exists || watch.Type == GetData {
		// 		fmt.Printf("Triggered: %s\n", watch.PrintWatch())
		// 	}
		// }
		// for _, watch := range parentNode.GetWatches() {
		// 	if watch.Type == GetChildren {
		// 		fmt.Printf("Triggered: %s\n", watch.PrintWatch())
		// 	}
		// }
	case Change:
		// Check for current node
		node.SetWatches(removeTriggeredWatches(node.GetWatches(), Exists))
		node.SetWatches(removeTriggeredWatches(node.GetWatches(), GetData))
		// for _, watch := range node.GetWatches() {
		// 	if watch.Type == Exists || watch.Type == GetData {
		// 		fmt.Printf("Triggered: %s\n", watch.PrintWatch())
		// 	}
		// }
	case Child:
		parentNode.SetWatches(removeTriggeredWatches(parentNode.GetWatches(), GetChildren))
		// for _, watch := range parentNode.GetWatches() {
		// 	if watch.Type == GetChildren {
		// 		fmt.Printf("Triggered: %s\n", watch.PrintWatch())
		// 	}
		// }
	}
	fmt.Println("Done checking watches")
}

// .
// .
// .
// .
// .
// .
// Utility

// Helper function to extract the parent name from a path.
func getParentName(path string, lastSlashIndex int) string {
	if lastSlashIndex == 0 {
		return PATH_SEP
	}
	return path[:lastSlashIndex]
}

// Helper function to add a sequence number to the path.
func addSequenceNumber(parentNode *Znode, path string, isSequence bool) string {
	i := parentNode.SequenceNum
	parentNode.SequenceNum++
	padded := fmt.Sprintf("%010d", i)
	return path + padded
}

func (dataTree *DataTree) GetPaths() string {
	keys := make([]string, 0, len(dataTree.NodeMap))
	for key := range dataTree.NodeMap {
		keys = append(keys, key)
	}
	return strings.Join(keys, ", ")
}

// addWatch
// getEphermerals
// removeWatch
