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
	arr := []Elem{Int(100), Int(90), Int(91), Int(1000), Int(1), Int(22), Int(411), Int(5), Int(333), Int(2211)}
	QSort(arr, 0, len(arr)-1)
	t.Logf("sorted: %v", arr)
}
