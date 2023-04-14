package sort

import (
	"testing"
)

type Int int

func (a Int) Less(e Elem) bool {
	v := e.(Int)
	return a < v
}

func (a Int) Equal(e Elem) bool {
	v := e.(Int)
	return a == v
}

var test_arr = []Elem{Int(100), Int(90), Int(91), Int(1000), Int(-54), Int(2211), Int(1), Int(22), Int(411), Int(5), Int(333), Int(23421), Int(711), Int(6543), Int(869), Int(3), Int(108), Int(40), Int(64), Int(89), Int(77), Int(918), Int(101), Int(115), Int(609)}

func Test_quicksort(t *testing.T) {
	tmp := make([]Elem, len(test_arr))
	for i := 0; i < 1000000; i++ {
		copy(tmp, test_arr)
		QSort(tmp, false)
	}
	t.Logf("quick sorted: %v", tmp)
}

func Test_heapsort(t *testing.T) {
	tmp := make([]Elem, len(test_arr))
	for i := 0; i < 1000000; i++ {
		copy(tmp, test_arr)
		HeapSort(tmp, false)
	}
	t.Logf("heap sorted: %v", tmp)
}

func Test_mergesort(t *testing.T) {
	tmp := make([]Elem, len(test_arr))
	for i := 0; i < 1000000; i++ {
		copy(tmp, test_arr)
		tmp = MergeSort(tmp, false)
	}
	t.Logf("merge sorted: %v", tmp)
}
