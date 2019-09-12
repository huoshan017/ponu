package main

import (
	"fmt"

	"github.com/huoshan017/ponu/sort"
)

/*type Int int

func (this *Int) Less(e sort.Elem) bool {
	a := e.(*Int)
	if *a < *this {
		return true
	}
	return false
}

func (this *Int) Greater(e sort.Elem) bool {
	a := e.(*Int)
	if *a > *this {
		return true
	}
	return false
}

func (this *Int) Equal(e sort.Elem) bool {
	a := e.(*Int)
	if *a == *this {
		return true
	}
	return false
}*/

func main() {
	arr := []int{100, 90, 91, 1000, 1, 22, 411, 5, 333, 2211}
	sort.QSort(arr, 0, len(arr)-1)
	fmt.Printf("sorted: %v", arr)
}
