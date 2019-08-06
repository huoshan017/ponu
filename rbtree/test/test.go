package main

import (
	"fmt"
	"os"

	"github.com/huoshan017/ponu/rbtree"
)

type KeyValue struct {
	Key   int
	Value int
}

func (a *KeyValue) Less(node rbtree.NodeValue) bool {
	v := node.(*KeyValue)
	if v == nil {
		return false
	}
	if a.Key < v.Key {
		return true
	}
	return false
}

func (a *KeyValue) Greater(node rbtree.NodeValue) bool {
	v := node.(*KeyValue)
	if v == nil {
		return false
	}
	if a.Key > v.Key {
		return true
	}
	return false
}

func (a *KeyValue) Equal(node rbtree.NodeValue) bool {
	v := node.(*KeyValue)
	if v == nil {
		return false
	}
	if a.Key == v.Key {
		return true
	}
	return false
}

func main() {
	var rb rbtree.RBTree
	for i := 40; i >= 1; i-- {
		rb.Insert(&KeyValue{
			Key: i,
		})
	}
	/*for i := 100; i > 0; i -= 2 {
		rb.Insert(&KeyValue{
			Key: i,
		})
	}
	for i := 99; i >= 1; i -= 2 {
		rb.Insert(&KeyValue{
			Key: i,
		})
	}*/
	fmt.Fprintf(os.Stdout, "node num: %v\n", rb.NodeNum())
	fmt.Fprintf(os.Stdout, "preorder traverse:\n")
	rb.PreorderTraverse(func(node rbtree.NodeValue) {
		n := node.(*KeyValue)
		if n == nil {
			return
		}
		fmt.Fprintf(os.Stdout, "%v\n", n.Key)
	})
	fmt.Fprintf(os.Stdout, "inorder traverse:\n")
	rb.InorderTraverse(func(node rbtree.NodeValue) {
		n := node.(*KeyValue)
		if n == nil {
			return
		}
		fmt.Fprintf(os.Stdout, "%v\n", n.Key)
	})
	fmt.Fprintf(os.Stdout, "postorder traverse:\n")
	rb.PostorderTraverse(func(node rbtree.NodeValue) {
		n := node.(*KeyValue)
		if n == nil {
			return
		}
		fmt.Fprintf(os.Stdout, "%v\n", n.Key)
	})
}
