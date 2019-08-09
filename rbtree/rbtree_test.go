package rbtree

import (
	"testing"
)

type KeyValue struct {
	Key   int
	Value int
}

func (a KeyValue) Less(node NodeValue) bool {
	v := node.(*KeyValue)
	if v == nil {
		return false
	}
	if a.Key < v.Key {
		return true
	}
	return false
}

func output_node(node *rbnode, t *testing.T) {
	n := node.value.(*KeyValue)
	if n == nil {
		return
	}
	t.Logf("%v %v", n.Key, node.color)
}

/*func Test_insert_nodes(t *testing.T) {
	var rb RBTree
	for i := 100; i > 0; i-- {
		rb.insert(&KeyValue{
			Key: i,
		})
	}
	t.Logf("Preorder:\n")
	rb.PreorderTraverse(func(node NodeValue) {
		n := node.(*KeyValue)
		if n == nil {
			return
		}
		t.Logf("%v\n", n.Key)
	})
	t.Logf("\nInorder:\n")
	rb.InorderTraverse(func(node NodeValue) {
		n := node.(*KeyValue)
		if n == nil {
			return
		}
		t.Logf("%v\n", n.Key)
	})
	t.Logf("\nPostorder:\n")
	rb.PostorderTraverse(func(node NodeValue) {
		n := node.(*KeyValue)
		if n == nil {
			return
		}
		t.Logf("%v\n", n.Key)
	})
	t.Logf("\nDFS order:\n")
	rb.DFSTraverse(func(node NodeValue) {
		n := node.(*KeyValue)
		if n == nil {
			return
		}
		t.Logf("%v\n", n.Key)
	})
	t.Logf("\nBFS order:\n")
	rb.BFSTraverse(func(node NodeValue) {
		n := node.(*KeyValue)
		if n == nil {
			return
		}
		t.Logf("%v\n", n.Key)
	})
}*/

func output_left_node(node *rbnode, t *testing.T) {
	n := node.value.(*KeyValue)
	if n == nil {
		return
	}
	var pn *KeyValue
	if node.parent != nil {
		pn = (node.parent.value).(*KeyValue)
	}
	if pn != nil {
		t.Logf("left: %v %v, parent: %v %v", n.Key, node.color, pn.Key, node.parent.color)
	} else {
		t.Logf("left: %v %v", n.Key, node.color)
	}
	if node.left != nil && !node.left.is_nil() {
		output_left_node(node.left, t)
	}
	if node.right != nil && !node.right.is_nil() {
		output_right_node(node.right, t)
	}
}

func output_right_node(node *rbnode, t *testing.T) {
	n := node.value.(*KeyValue)
	if n == nil {
		return
	}
	var pn *KeyValue
	if node.parent != nil {
		pn = (node.parent.value).(*KeyValue)
	}
	if pn != nil {
		t.Logf("right: %v %v, parent: %v %v", n.Key, node.color, pn.Key, node.parent.color)
	} else {
		t.Logf("right: %v %v", n.Key, node.color)
	}
	if node.left != nil && !node.left.is_nil() {
		output_left_node(node.left, t)
	}
	if node.right != nil && !node.right.is_nil() {
		output_right_node(node.right, t)
	}
}

func output_nodes(node *rbnode, t *testing.T) {
	if node != nil {
		output_node(node, t)
	}
	if node.left != nil && !node.left.is_nil() {
		output_left_node(node.left, t)
	}
	if node.right != nil && !node.right.is_nil() {
		output_right_node(node.right, t)
	}
}

func Test_ouput_tree(t *testing.T) {
	var rb RBTree
	for i := 29; i >= 1; i-- {
		rb.Insert(&KeyValue{
			Key: i,
		})
	}
	output_nodes(rb.root, t)
}

func Test_output_tree2(t *testing.T) {
	var rb RBTree
	value_list := []int{200, 100, -1, 2, 822, 33, 221, 21, 34, 441, 14, 558, 3333, 44, 457, 1, 32, 4, 9, 101, 8, 71, 564, 22323, 4711, 191, 876, 1222}
	for i := 0; i < len(value_list); i++ {
		rb.Insert(&KeyValue{
			Key: value_list[i],
		})
	}
	output_nodes(rb.root, t)

	delete_value_list := []int{22323, -1, 1, 2}
	for i := 0; i < len(delete_value_list); i++ {
		rb.Delete(&KeyValue{
			Key: delete_value_list[i],
		})
	}
	t.Logf("after delete nodes:")
	output_nodes(rb.root, t)
}
