package sort

import (
	"testing"
)

type Int int

func (a Int) Less(e Elem) bool {
	v := e.(Int)
	if a < v {
		return true
	}
	return false
}

func (a Int) Equal(e Elem) bool {
	v := e.(Int)
	if a == v {
		return true
	}
	return false
}

func Test_quicksort(t *testing.T) {
	arr := []Elem{Int(100), Int(90), Int(91), Int(1000), Int(-54), Int(2211), Int(1), Int(22), Int(411), Int(5), Int(333), Int(23421), Int(711), Int(6543), Int(869), Int(3), Int(108), Int(40), Int(64), Int(89), Int(77), Int(918), Int(101), Int(115), Int(609)}
	//HeapSortMin(arr)
	arr2 := MergeSort(arr)
	t.Logf("sorted: %v", arr2)
}
