package cache

import (
	"sync"

	"github.com/huoshan017/ponu/list"
)

type Pair[K comparable, V any] struct {
	k K
	v V
}

func (pair Pair[K, V]) GetKey() K {
	return pair.k
}

func (pair Pair[K, V]) GetValue() V {
	return pair.v
}

type LRU[K comparable, V any] struct {
	cap      int32
	l        list.ListT[Pair[K, V]]
	m        map[K]list.IteratorT[Pair[K, V]]
	pairPool *list.ListTNodePool[Pair[K, V]]
}

func NewLRU[K comparable, V any](cap int32) *LRU[K, V] {
	lru := &LRU[K, V]{
		cap:      cap,
		m:        make(map[K]list.IteratorT[Pair[K, V]]),
		pairPool: list.NewListTNodePool[Pair[K, V]](),
	}
	lru.l = list.NewListTObjWithPool(lru.pairPool)
	return lru
}

func (lru *LRU[K, V]) Set(key K, value V) bool {
	if lru.m == nil {
		lru.m = make(map[K]list.IteratorT[Pair[K, V]])
	}
	iter, o := lru.m[key]
	if !o {
		if lru.cap > 0 && lru.cap <= lru.l.GetLength() {
			kv, o := lru.l.PopFront()
			if !o { // pop front failed
				return false
			}
			delete(lru.m, kv.k)
		}
		lru.l.PushBack(Pair[K, V]{k: key, v: value})
		lru.m[key] = lru.l.RBegin()
	} else {
		lru.update(iter, true, value)
	}
	return true
}

func (lru *LRU[K, V]) Get(key K) (V, bool) {
	var (
		iter list.IteratorT[Pair[K, V]]
		o    bool
		v    V
	)
	if lru.m == nil {
		return v, false
	}
	iter, o = lru.m[key]
	if !o {
		return v, false
	}
	lru.update(iter, false, v)
	return iter.Value().v, true
}

func (lru *LRU[K, V]) Has(key K) bool {
	if lru.m == nil {
		return false
	}
	_, o := lru.m[key]
	return o
}

func (lru *LRU[K, V]) Delete(key K) bool {
	if lru.m == nil {
		return false
	}
	iter, o := lru.m[key]
	if !o {
		return false
	}
	delete(lru.m, key)
	lru.l.Delete(iter)
	return true
}

func (lru *LRU[K, V]) ToList() list.ListT[V] {
	var l list.ListT[V]
	dl := lru.l.Duplicate()
	for iter := dl.Begin(); iter != dl.End(); iter = iter.Next() {
		l.PushBack(iter.Value().v)
	}
	dl.Clear()
	return l
}

func (lru *LRU[K, V]) Clear() {
	lru.l.Clear()
	lru.m = nil
}

func (lru *LRU[K, V]) update(iter list.IteratorT[Pair[K, V]], update bool, value V) {
	if iter == lru.l.RBegin() { // keep the origin position
		return
	}
	n := iter.Value()
	if !lru.l.Delete(iter) { // delete failed by iter
		return
	}
	if update {
		n.v = value
	}
	lru.l.PushBack(n)
}

type LRUWithLock[K comparable, V any] struct {
	*LRU[K, V]
	locker sync.RWMutex
}

func NewLRUWithLock[K comparable, V any](cap int32) *LRUWithLock[K, V] {
	return &LRUWithLock[K, V]{
		LRU: NewLRU[K, V](cap),
	}
}

func (lru *LRUWithLock[K, V]) Set(key K, value V) {
	lru.locker.Lock()
	defer lru.locker.Unlock()
	lru.LRU.Set(key, value)
}

func (lru *LRUWithLock[K, V]) Get(key K) (V, bool) {
	lru.locker.Lock()
	defer lru.locker.Unlock()
	return lru.LRU.Get(key)
}

func (lru *LRUWithLock[K, V]) Delete(key K) bool {
	lru.locker.Lock()
	defer lru.locker.Unlock()
	return lru.LRU.Delete(key)
}

func (lru *LRUWithLock[K, V]) Has(key K) bool {
	lru.locker.RLock()
	defer lru.locker.RUnlock()
	return lru.LRU.Has(key)
}

func (lru *LRUWithLock[K, V]) ToList() list.ListT[V] {
	lru.locker.RLock()
	defer lru.locker.RUnlock()
	return lru.LRU.ToList()
}
