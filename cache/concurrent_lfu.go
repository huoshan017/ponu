package cache

import (
	"reflect"
	"sync"

	"github.com/huoshan017/ponu/list"
)

const (
	shardSize = 64
)

type keyType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~string
}

type lfuShardWithLock[K keyType, V any] struct {
	*lfuShard[K, V]
	rwlock sync.RWMutex
}

func newLFUShardWithLock[K keyType, V any](cap int32, totalSize *int32) *lfuShardWithLock[K, V] {
	return &lfuShardWithLock[K, V]{
		lfuShard: newLFUShardWithTotalSize[K, V](cap, totalSize),
	}
}

func (shard *lfuShardWithLock[K, V]) Set(key K, value V) {
	shard.rwlock.Lock()
	defer shard.rwlock.Unlock()
	shard.lfuShard.Set(key, value)
}

func (shard *lfuShardWithLock[K, V]) Get(key K) (V, bool) {
	shard.rwlock.Lock()
	defer shard.rwlock.Unlock()
	return shard.lfuShard.Get(key)
}

func (shard *lfuShardWithLock[K, V]) Has(key K) bool {
	shard.rwlock.RLock()
	defer shard.rwlock.RUnlock()
	return shard.lfuShard.Has(key)
}

func (shard *lfuShardWithLock[K, V]) Delete(key K) bool {
	shard.rwlock.Lock()
	defer shard.rwlock.Unlock()
	return shard.lfuShard.Delete(key)
}

func (shard *lfuShardWithLock[K, V]) ToList() list.List {
	shard.rwlock.RLock()
	defer shard.rwlock.RUnlock()
	return shard.lfuShard.ToList()
}

func (shard *lfuShardWithLock[K, V]) Clear() {
	shard.rwlock.Lock()
	defer shard.rwlock.Unlock()
	shard.lfuShard.Clear()
}

type ConcurrentLFU[K keyType, V any] struct {
	typ      reflect.Type
	currSize int32
	shards   []*lfuShardWithLock[K, V]
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
		shards: make([]*lfuShardWithLock[K, V], shardSize),
	}
	for i := 0; i < len(lfu.shards); i++ {
		lfu.shards[i] = newLFUShardWithLock[K, V](cap, &lfu.currSize)
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
			index = defaultLfuIntHashFunc(v.Elem().Int(), shardSize)
		} else {
			index = lfuIntHashFunc(v.Elem().Int(), shardSize)
		}
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Uintptr:
		if lfuUintHashFunc == nil {
			index = defaultLfuUintHashFunc(v.Elem().Uint(), shardSize)
		} else {
			index = lfuUintHashFunc(v.Elem().Uint(), shardSize)
		}
	case reflect.String:
		if lfuStringHashFunc == nil {
			index = defaultLfuStringHashFunc(v.Elem().String(), shardSize)
		} else {
			index = lfuStringHashFunc(v.Elem().String(), shardSize)
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
