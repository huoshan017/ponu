package cache

import (
	"reflect"

	"github.com/huoshan017/ponu/list"
)

const (
	shardSize = 64
)

type keyType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~string
}

type ConcurrentLFU[K keyType, V any] struct {
	typ      reflect.Type
	currSize int32
	shards   []*LFUWithLock[K, V]
}

func NewConcurrentLFU[K keyType, V any](cap int32) *ConcurrentLFU[K, V] {
	var k K
	t := reflect.TypeOf(k)
	switch t.Kind() {
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Uintptr:
	case reflect.String:
	default:
		panic("cache ponu: not supporting concurrent lfu key type")
	}
	lfu := &ConcurrentLFU[K, V]{
		typ:    t,
		shards: make([]*LFUWithLock[K, V], shardSize),
	}
	for i := 0; i < len(lfu.shards); i++ {
		lfu.shards[i] = newLFUWithLockAndTotalSize[K, V](cap, &lfu.currSize)
	}
	return lfu
}

func (lfu *ConcurrentLFU[K, V]) Set(key K, value V) {
	index := lfu.getHashIndex(key)
	shard := lfu.shards[index]
	shard.Set(key, value)
}

func (lfu *ConcurrentLFU[K, V]) Get(key K) (V, bool) {
	index := lfu.getHashIndex(key)
	shard := lfu.shards[index]
	return shard.Get(key)
}

func (lfu *ConcurrentLFU[K, V]) Has(key K) bool {
	index := lfu.getHashIndex(key)
	shard := lfu.shards[index]
	return shard.Has(key)
}

func (lfu *ConcurrentLFU[K, V]) Delete(key K) bool {
	index := lfu.getHashIndex(key)
	shard := lfu.shards[index]
	return shard.Delete(key)
}

func (lfu *ConcurrentLFU[K, V]) Clear() {
	for i := 0; i < len(lfu.shards); i++ {
		lfu.shards[i].Clear()
	}
}

func (lfu *ConcurrentLFU[K, V]) ToList() list.List {
	var l list.List
	for i := 0; i < len(lfu.shards); i++ {
		lfu.shards[i].l.CopyTo(&l)
	}
	return l
}

func (lfu *ConcurrentLFU[K, V]) getHashIndex(key K) int32 {
	var index int32
	v := reflect.ValueOf(key)
	switch lfu.typ.Kind() {
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		if lfuIntHashFunc == nil {
			index = defaultLfuIntHashFunc(v.Int(), shardSize)
		} else {
			index = lfuIntHashFunc(v.Int(), shardSize)
		}
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Uintptr:
		if lfuUintHashFunc == nil {
			index = defaultLfuUintHashFunc(v.Uint(), shardSize)
		} else {
			index = lfuUintHashFunc(v.Uint(), shardSize)
		}
	case reflect.String:
		if lfuStringHashFunc == nil {
			index = defaultLfuStringHashFunc(v.String(), shardSize)
		} else {
			index = lfuStringHashFunc(v.String(), shardSize)
		}
	default:
		panic("cache ponu: unsupported key type")
	}
	return index
}

var (
	lfuIntHashFunc    func(int64, int32) int32
	lfuUintHashFunc   func(uint64, int32) int32
	lfuStringHashFunc func(string, int32) int32
)

func SetLfuIntHashFunc(f func(key int64, shardSize int32) int32) {
	lfuIntHashFunc = f
}

func SetUintHashLfuFunc(f func(key uint64, shardSize int32) int32) {
	lfuUintHashFunc = f
}

func SetStringFunc(f func(key string, shardSize int32) int32) {
	lfuStringHashFunc = f
}

func defaultLfuIntHashFunc(key int64, shardSize int32) int32 {
	return int32(key % int64(shardSize))
}

func defaultLfuUintHashFunc(key uint64, shardSize int32) int32 {
	return int32(key % uint64(shardSize))
}

func defaultLfuStringHashFunc(key string, shardSize int32) int32 {
	h := 0
	for i := 0; i < len(key); i++ {
		h = 31*h + int(key[i])
	}
	return int32(int32(h) % int32(shardSize))
}
