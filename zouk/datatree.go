package zouk

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type DataTree struct {
	NodeMap map[string]*Znode
}

func NewDataTree() DataTree {
	rootNode := Znode{
		//TODO: Change zxid and ephemeral owner
		Stat:     CreateStat(0, time.Now().Unix(), 0),
		Children: map[string]bool{},
		Parent:   "/",
		Data:     []byte{},
		Eph:      false,
		Id:       0,
	}

	dataTree := DataTree{
		NodeMap: map[string]*Znode{
			"/root": &rootNode,
		},
	}

	return dataTree
}

func (dataTree *DataTree) CreateNode(path string, data []byte, ephermeralOwner int64, time int64, zxid int64) (string, error) {
	fmt.Printf("Inside CreateNode, data: %d\n", data)
	lastSlashIndex := strings.LastIndex(path, "/")
	parentName := path[:lastSlashIndex]
	childName := path[lastSlashIndex:]
	// fmt.Printf("lastslashindex{%d}, parentName{%s}, childName:{%s}\n", lastSlashIndex, parentName, childName)
	stat := CreateStat(zxid, time, ephermeralOwner)
	parentNode, ok := dataTree.NodeMap[parentName]
	if !ok {
		return parentName, errors.New("invalid parent name")
	}
	children := parentNode.GetChildren()
	_, ok = children[childName] // check for existence
	if ok {
		return childName, errors.New("invalid children as it already exists")
	}
	childNode := Znode{
		Data: data,
		Stat: stat,
	}
	parentNode.AddChild(childName)
	dataTree.NodeMap[path] = &childNode
	return path, nil
}

func (dataTree *DataTree) DeleteNode(path string, zxid int64) (string, error) {
	lastSlashIndex := strings.LastIndex(path, "/")
	parentName := path[:lastSlashIndex]
	childName := path[lastSlashIndex:]
	parentNode, ok := dataTree.NodeMap[parentName]
	if !ok {
		return parentName, errors.New("invalid parent name")
	}
	children := parentNode.GetChildren()
	_, ok = children[childName] // check for existence
	if !ok {
		return childName, errors.New("invalid children as it doesn't exists")
	}
	parentNode.RemoveChild(childName)

	delete(dataTree.NodeMap, path)
	return "Removed", nil
}

func (dataTree *DataTree) SetData(path string, data []byte, version int64, zxid int64, time int64) Stat {
	node := dataTree.NodeMap[path]
	node.SetData(data)

	stat := node.Stat
	stat.Mtime = time
	stat.Mzxid = zxid
	stat.Version = version

	outStat := CopyStat(stat)
	return outStat
}

func (dataTree *DataTree) GetData(path string) []byte {
	node := dataTree.NodeMap[path]
	return node.GetData()
}

// addWatch
// getEphermerals
// removeWatch
