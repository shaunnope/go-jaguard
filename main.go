package main

import (
	"fmt"

	"github.com/shaunnope/go-jaguard/zouk"
)

func hello() {
	fmt.Println("Hello, World!")
	var b zouk.Stat
	b.Czxid =1
	a := zouk.Znode{}
	a.Parent = 2

}

func start() {

}
func AddChild(parent zouk.Znode, child zouk.Znode) zouk.Znode{
	parent.Children = append(parent.Children, 1)
	fmt.Println(parent.Children)
	return parent
}

func main() {
	hello()
	a := zouk.Znode{}
	b := zouk.Znode{}
	AddChild(a,b)
}
