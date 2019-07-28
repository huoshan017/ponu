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

func (a KeyValue) Greater(node NodeValue) bool {
	v := node.(*KeyValue)
	if v == nil {
		return false
	}
	if a.Key > v.Key {
		return true
	}
	return false
}

func (a KeyValue) Equal(node NodeValue) bool {
	v := node.(*KeyValue)
	if v == nil {
		return false
	}
	if a.Key == v.Key {
		return true
	}
	return false
}

func output_node(node *rbnode, t *testing.T) {
	n := node.value.(*KeyValue)
	if n == nil {
		return
	}
	t.Logf("%v\n", n.Key)
}

func Test_insert_nodes(t *testing.T) {
	var rb RBTree
	for i := 0; i < 100; i++ {
		rb.insert(&KeyValue{
			Key: i + 1,
		})
	}
	rb.Preorder_traverse(&rb, t)
}

func Benchmark_insert_nodes(b *testing.B) {
	var rb RBTree
	b.StartTimer()
	for i := 0; i < 100; i++ {
		rb.insert(&KeyValue{
			Key: i + 1,
		})
	}

	b.StopTimer()
}
