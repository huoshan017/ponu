package main

import (
	"fmt"

	"github.com/huoshan017/ponu/sort"
)

type Int int

func (a Int) Less(e sort.Elem) bool {
	v := e.(Int)
	if a < v {
		return true
	}
	return false
}

func (a Int) Equal(e sort.Elem) bool {
	v := e.(Int)
	if a == v {
		return true
	}
	return false
}

func main() {
	arr := []sort.Elem{Int(100), Int(90), Int(91), Int(1000), Int(1), Int(22), Int(411), Int(5), Int(333), Int(2211)}
	sort.QSort(arr, 0, len(arr)-1)
	fmt.Printf("sorted: %v", arr)
}
