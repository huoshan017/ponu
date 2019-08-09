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

func main() {
	var rb rbtree.RBTree
	/*for i := 40; i >= 1; i-- {
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

	value_list := []int{200, 100, -1, 2, 822, 33, 221, 21, 34, 441, 14, 558, 3333, 44, 457, 1, 32, 4, 9, 101, 8, 71, 564, 22323, 4711, 191, 876, 1222}
	for i := 0; i < len(value_list); i++ {
		rb.Insert(&KeyValue{
			Key: value_list[i],
		})
	}

	delete_value_list := []int{22323, -1, 1, 2, 100, 221}
	for i := 0; i < len(delete_value_list); i++ {
		rb.Delete(&KeyValue{
			Key: delete_value_list[i],
		})
	}
}
