package tmap

import (
	"testing"
)

func Test_one(t *testing.T) {
	var key_value_list = []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13, 14, 14, 15, 15, 16, 16, 17, 17, 18, 18, 19, 19, 20, 20, 21, 21, 22, 22, 23, 23, 24, 24}
	var m TMap
	for i := 0; i < len(key_value_list)/2; i++ {
		m.Insert(key_value_list[2*i], uint64(key_value_list[2*i+1]))
	}
	for i := 0; i < len(key_value_list)/2; i++ {
		v := m.Get(key_value_list[2*i])
		t.Logf("key is %v, value is %v", key_value_list[2*i], v)
	}
}
