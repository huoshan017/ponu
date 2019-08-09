package tmap

import (
	"fmt"
	"os"

	"github.com/huoshan017/ponu/rbtree"
)

type KeyValue struct {
	key   interface{}
	value interface{}
}

func _get_value_from(i interface{}) (t int8, v1 int64, v2 uint64, v3 float32, v4 float64) {
	switch tt := i.(type) {
	case int8:
		v1 = int64(i.(int8))
	case uint8:
		v1 = int64(i.(uint8))
	case int16:
		v1 = int64(i.(int16))
	case uint16:
		v1 = int64(i.(uint16))
	case int32:
		v1 = int64(i.(int32))
	case uint32:
		v1 = int64(i.(uint32))
	case int:
		v1 = int64(i.(int))
	case uint:
		v1 = int64(i.(uint))
	case int64:
		v1 = i.(int64)
	case uint64:
		v2 = i.(uint64)
		t = 1
	case float32:
		v3 = i.(float32)
		t = 2
	case float64:
		v4 = i.(float64)
		t = 3
	default:
		fmt.Fprintf(os.Stderr, "unsupported type %v get value from", tt)
		t = -1
	}
	return t, v1, v2, v3, v4
}

func (this *KeyValue) Less(node rbtree.NodeValue) bool {
	n := node.(*KeyValue)
	if n == nil {
		return false
	}

	kt, kv1, kv2, kv3, kv4 := _get_value_from(this.key)
	nt, nv1, nv2, nv3, nv4 := _get_value_from(n.key)

	if kt < 0 || nt < 0 {
		if func() int8 {
			if kt < 0 {
				return kt
			} else {
				return nt
			}
		}() < 0 {
			fmt.Fprintf(os.Stderr, "tmap: unsupported type %v to compare")
			return false
		}
	}

	if kt != nt {
		fmt.Fprintf(os.Stderr, "tmap: compare with different")
		return false
	}

	if kt == 0 {
		if kv1 < nv1 {
			return true
		}
	} else if kt == 1 {
		if kv2 < nv2 {
			return true
		}
	} else if kt == 2 {
		if kv3 < nv3 {
			return true
		}
	} else if kt == 3 {
		if kv4 < nv4 {
			return true
		}
	}

	return false
}

type TMap struct {
	t rbtree.RBTree
}

func (this *TMap) Insert(key, value interface{}) {
	key_value := &KeyValue{
		key:   key,
		value: value,
	}
	this.t.Insert(key_value)
}

func (this *TMap) Delete(key interface{}) bool {
	return this.t.Delete(&KeyValue{
		key: key,
	})
}

func (this *TMap) Get(key interface{}) interface{} {
	var key_value = KeyValue{
		key: key,
	}
	kv := this.t.Get(&key_value)
	if kv == nil {
		return nil
	}
	return (kv.(*KeyValue)).value
}

func (this *TMap) Has(key interface{}) bool {
	var kv = KeyValue{
		key: key,
	}
	return this.t.Has(&kv)
}
