package zouk

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
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
		Watches:  []*Watch{},
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
		return childName, errors.New("invalid child. already exists")
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
	if parentName == "" {
		parentName = "/"
	}
	childName := path[lastSlashIndex:]
	parentNode, ok := dataTree.NodeMap[parentName]
	if !ok {
		fmt.Printf("parentName:%s, childName:%s", parentName, childName)
		return parentName, errors.New("invalid parent name")
	}

	if !parentNode.ChildExists(childName) {
		return childName, errors.New("invalid child. does not exist")
	}

	parentNode.RemoveChild(childName)
	delete(dataTree.NodeMap, path)

	return "Removed", nil
}

// SetData sets the data of a node by its path.
func (dataTree *DataTree) SetData(path string, data []byte, version int64, zxid ZxidFragment) (Stat, error) {
	node, exists := dataTree.NodeMap[path]
	if !exists {
		return Stat{}, errors.New("node does not exist")
	}
	node.SetData(data)

	stat := node.GetStat()
	stat.Mtime = time.Now().Unix()
	stat.Mzxid = zxid
	stat.Version = version

	outStat := CopyStat(stat)
	return outStat, nil
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
	node, exists := dataTree.NodeMap[path]
	if !exists {
		return nil, errors.New("node does not exist")
	}
	return node.GetData(), nil
}

// AddWatcher adds watcher based on event type to a node
func (dataTree *DataTree) AddWatchToNode(path string, watch *Watch) (string, error) {
	nodeAddWatch, ok := dataTree.NodeMap[path]
	if !ok {
		return path, errors.New("node does not exist")
	}

	nodeAddWatch.AddWatch(watch)
	fmt.Printf("Watch %s added\n", watch.PrintWatch())
	return "ok", nil
}

func (dataTree *DataTree) CheckWatchTrigger(transactionFragment *TransactionFragment) []*Watch {
	// Extract parentName and nodeName from the path
	lastSlashIndex := strings.LastIndex(transactionFragment.Path, PATH_SEP)
	parentName := getParentName(transactionFragment.Path, lastSlashIndex)
	nodeName := transactionFragment.Path[lastSlashIndex:]

	// Print debug information
	fmt.Printf("Checking watches on parent:%s, node:%s for transaction:%s\n", parentName, nodeName, transactionFragment)

	// Get parent and current nodes from the data tree
	parentNode, err := dataTree.GetNode(parentName)
	// check if parentNode exists, if no, return empty slice
	if err != nil {
		slog.Info("parentNode does not exist", "path", parentName)
		return []*Watch{}
	}
	node, err := dataTree.GetNode(transactionFragment.Path)
	// check if node exists, if no, return empty slice
	if err != nil {
		slog.Info("Node does not exist", "path", transactionFragment.Path)
		return []*Watch{}
	}

	// Function to remove triggered watches
	removeTriggeredWatches := func(watches []*Watch, watchType WatchType) ([]*Watch, []*Watch) {
		var remainingWatches []*Watch
		var triggeredWatches []*Watch

		// Iterate over watches and filter triggered and remaining watches
		for _, watch := range watches {
			if watch.Type == watchType {
				fmt.Printf("Triggered: %s\n", watch.PrintWatch())
				triggeredWatches = append(triggeredWatches, watch)
			} else {
				remainingWatches = append(remainingWatches, watch)
			}
		}
		return remainingWatches, triggeredWatches
	}

	var remainingWatches []*Watch
	var triggeredWatches []*Watch
	var triggered []*Watch

	switch transactionFragment.Type {
	case OperationType_WRITE, OperationType_DELETE, OperationType_UPDATE:
		// For DELETE and UPDATE, also check GetData watch
		if transactionFragment.Type != OperationType_WRITE {
			remainingWatches, triggered = removeTriggeredWatches(node.GetWatches(), GetData)
			node.SetWatches(remainingWatches)
			triggeredWatches = append(triggeredWatches, triggered...)
		}

		// Check for current node
		remainingWatches, triggered = removeTriggeredWatches(node.GetWatches(), Exists)
		node.SetWatches(remainingWatches)
		triggeredWatches = append(triggeredWatches, triggered...)

		// Check for parent node
		remainingWatches, triggered = removeTriggeredWatches(parentNode.GetWatches(), GetChildren)
		parentNode.SetWatches(remainingWatches)
		triggeredWatches = append(triggeredWatches, triggered...)
	}

	// Print debug information
	return triggeredWatches
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
