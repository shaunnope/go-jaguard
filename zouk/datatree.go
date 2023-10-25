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
		Parent:   PATH_SEP,
		Data:     []byte{},
		Eph:      false,
		Id:       0,
	}

	dataTree := DataTree{
		NodeMap: map[string]*Znode{
			PATH_SEP: rootNode,
		},
	}

	return &dataTree
}

func (dataTree *DataTree) CreateNode(path string, data []byte, isEph bool, ephermeralOwner int64, zxid ZxidFragment, isSequence bool) (string, error) {
	log.Printf("Inside CreateNode, data: %d\n", data)

	lastSlashIndex := strings.LastIndex(path, PATH_SEP)
	if lastSlashIndex == -1 {
		return path, errors.New("invalid path")
	}
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
	childNode := NewNode(stat, parentName, data, isEph, 0, isSequence)
	parentNode.AddChild(childName)
	dataTree.NodeMap[path] = &childNode

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

func (dataTree *DataTree) GetData(path string) []byte {
	node := dataTree.NodeMap[path]
	return node.GetData()
}

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
