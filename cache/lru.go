package cache

import "github.com/huoshan017/ponu/list"

type pair[K comparable, V any] struct {
	k K
	v V
}

type LRU[K comparable, V any] struct {
	cap int32
	l   list.List
	m   map[K]list.Iterator
}

func NewLRU[K comparable, V any](cap int32) *LRU[K, V] {
	return &LRU[K, V]{
		cap: cap,
		l:   list.NewObj(),
		m:   make(map[K]list.Iterator),
	}
}

func (lru *LRU[K, V]) Set(key K, value V) bool {
	if lru.m == nil {
		lru.m = make(map[K]list.Iterator)
	}
	iter, o := lru.m[key]
	if !o {
		if lru.cap > 0 && lru.cap <= lru.l.GetLength() {
			kv, o := lru.l.PopFront()
			if !o { // pop front failed
				return false
			}
			delete(lru.m, kv.(pair[K, V]).k)
		}
		lru.l.PushBack(pair[K, V]{k: key, v: value})
		lru.m[key] = lru.l.RBegin()
	} else {
		lru.update(iter, true, value)
	}
	return true
}

func (lru *LRU[K, V]) Get(key K) (V, bool) {
	var (
		iter list.Iterator
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
	return iter.Value().(pair[K, V]).v, true
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

func (lru *LRU[K, V]) ToList() list.List {
	return lru.l.Duplicate()
}

func (lru *LRU[K, V]) Clear() {
	lru.l.Clear()
	lru.m = nil
}

func (lru *LRU[K, V]) update(iter list.Iterator, update bool, value V) {
	if iter == lru.l.RBegin() { // keep the origin position
		return
	}
	n := iter.Value().(pair[K, V])
	if !lru.l.Delete(iter) { // delete failed by iter
		return
	}
	if update {
		n.v = value
	}
	lru.l.PushBack(n)
}
