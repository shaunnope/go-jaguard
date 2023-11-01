package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/shaunnope/go-jaguard/zouk"
)

const (
	PATH_SEP    = "/"
	LINE        = "│   "
	MIDDLE_NODE = "├── "
	LAST_NODE   = "└── "
)

func printTree(dataTree *zouk.DataTree, w io.Writer, path string, prefix string) {
	node, ok := dataTree.NodeMap[path]
	if !ok {
		return
	}

	childrenKeys := make([]string, 0, len(node.Children))
	for key := range node.Children {
		childrenKeys = append(childrenKeys, key)
	}

	// Sort children for consistent output
	sort.Strings(childrenKeys)

	for i, childName := range childrenKeys {
		childPath := path + PATH_SEP + childName
		newPrefix := prefix

		// Print depending on whether the child is the last in the list
		if i == len(childrenKeys)-1 {
			fmt.Fprint(w, prefix+LAST_NODE+childName+"\n") // added the childName
			newPrefix += "    "                            // For last node, we just add spaces
		} else {
			fmt.Fprint(w, prefix+MIDDLE_NODE+childName+"\n") // added the childName
			newPrefix += LINE
		}

		printTree(dataTree, w, childPath, newPrefix)
	}
}
