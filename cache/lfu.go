package cache

import (
	"fmt"
	"sync/atomic"

	"github.com/huoshan017/ponu/list"
)

type node[K keyType, V any] struct {
	k K
	v V
	f int32
}

type lfuShard[K keyType, V any] struct {
	cap       int32
	totalSize *int32
	l         list.List
	k2i       map[K]list.Iterator
	f2i       map[int32]list.Iterator // 保存訪問次數對應的迭代器，這個迭代器是該訪問次數下最新的，如果有更新的迭代器，則插在它之後
}

func newLFUShard[K keyType, V any](cap int32) *lfuShard[K, V] {
	return &lfuShard[K, V]{
		cap: cap,
		l:   list.NewObj(),
		k2i: make(map[K]list.Iterator),
		f2i: make(map[int32]list.Iterator),
	}
}

func newLFUShardWithTotalSize[K keyType, V any](cap int32, totalSize *int32) *lfuShard[K, V] {
	return &lfuShard[K, V]{
		cap:       cap,
		totalSize: totalSize,
		l:         list.NewObj(),
		k2i:       make(map[K]list.Iterator),
		f2i:       make(map[int32]list.Iterator),
	}
}

func (lfu *lfuShard[K, V]) Set(key K, value V) {
	if lfu.k2i == nil {
		lfu.k2i = make(map[K]list.Iterator)
	}
	if lfu.f2i == nil {
		lfu.f2i = make(map[int32]list.Iterator)
	}
	iter, o := lfu.k2i[key]
	if !o { // 插入
		if lfu.totalSize == nil {
			if lfu.cap > 0 && lfu.cap <= int32(len(lfu.k2i)) { // 超出容量大小
				lfu.deleteFirst()
			}
			lfu.add(key, value)
		} else {
			if lfu.cap <= atomic.LoadInt32(lfu.totalSize) {
				lfu.deleteFirst()
				lfu.add(key, value)
			} else {
				if lfu.cap < atomic.AddInt32(lfu.totalSize, 1) {
					lfu.deleteFirst()
					lfu.add(key, value)
					atomic.AddInt32(lfu.totalSize, -1)
				} else {
					lfu.add(key, value)
				}
			}
		}
	} else { // 更新
		lfu.updateValue(iter, true, value)
	}
}

func (lfu *lfuShard[K, V]) Get(key K) (V, bool) {
	var (
		iter list.Iterator
		v    V
		o    bool
	)
	if lfu.k2i == nil {
		return v, false
	}
	iter, o = lfu.k2i[key]
	if !o {
		return v, false
	}
	// iter在update之後有可能會失效，所以要在update之前取到其中的值
	v = iter.Value().(node[K, V]).v
	lfu.update(iter)
	return v, true
}

func (lfu *lfuShard[K, V]) Delete(key K) bool {
	var (
		iter list.Iterator
		o    bool
	)
	iter, o = lfu.k2i[key]
	if !o {
		return false
	}
	n := iter.Value().(node[K, V])
	if iter == lfu.f2i[n.f] {
		lfu.updateFrequency(iter, n.f)
	}
	delete(lfu.k2i, key)
	lfu.l.Delete(iter)
	atomic.AddInt32(lfu.totalSize, -1)
	return true
}

func (lfu *lfuShard[K, V]) Has(key K) bool {
	_, o := lfu.k2i[key]
	return o
}

func (lfu *lfuShard[K, V]) ToList() list.List {
	l := lfu.l.Duplicate()
	return l
}

func (lfu *lfuShard[K, V]) Clear() {
	lfu.l.Clear()
	lfu.k2i = nil
	lfu.f2i = nil
}

func (lfu *lfuShard[K, V]) update(iter list.Iterator) {
	var v V
	lfu.updateValue(iter, false, v)
}

func (lfu *lfuShard[K, V]) updateValue(iter list.Iterator, update bool, val V) {
	var (
		n             node[K, V]
		f             int32
		niter, niter2 list.Iterator
		has, has2     bool
	)

	n = iter.Value().(node[K, V])
	f = n.f
	// 獲取f次最新訪問元素的迭代器
	niter, has = lfu.f2i[n.f]
	if !has {
		panic(fmt.Sprintf("ponu cache: lfu must get iterator for specified frequency %v, lfu.f2i=%v, lfu.k2i=%+v", n.f, lfu.f2i, lfu.k2i))
	}

	// 更新頻次node.f
	n.f += 1
	if update {
		n.v = val
	}

	// 獲取f+1次最新訪問的元素迭代器
	niter2, has2 = lfu.f2i[f+1]

	if iter != niter { // 待更新元素不是最新訪問的元素
		if !lfu.l.Delete(iter) {
			panic("ponu cache: lfu delete iter failed")
		}
		delete(lfu.k2i, n.k)
		if has2 {
			iter = lfu.l.InsertContinue(n, niter2)
		} else {
			iter = lfu.l.InsertContinue(n, niter)
		}
		lfu.k2i[n.k] = iter
	} else { // 待更新的元素就是最新訪問的元素
		lfu.updateFrequency(iter, f)
		if has2 { // 如果有f+1次訪問的元素，説明更新前後的位置是不一樣的，則先刪除再插入
			if !lfu.l.Delete(iter) {
				panic("ponu cache: lfu delete iter failed")
			}
			delete(lfu.k2i, n.k)
			iter = lfu.l.InsertContinue(n, niter2)
			lfu.k2i[n.k] = iter
		} else { // 沒有f+1次訪問的元素，則表示更新前後位置不變，直接修改元素值
			lfu.l.Update(n, iter)
		}
	}

	lfu.f2i[n.f] = iter
}

func (lfu *lfuShard[K, V]) updateFrequency(iter list.Iterator, f int32) {
	prev := iter.Prev()
	if !prev.IsValid() || f != prev.Value().(node[K, V]).f {
		// 前一個元素不存在 或者
		// 待更新元素的訪問次數與它之前一個元素的不一樣，表示此訪問次數下只有一個元素，則直接從lfu.f2i中刪除該訪問次數
		delete(lfu.f2i, f)
	} else {
		// 否則更新此次數對應的最新訪問元素為前一個元素
		lfu.f2i[f] = prev
	}
}

func (lfu *lfuShard[K, V]) deleteFirst() {
	iter := lfu.l.Begin()
	val := lfu.l.Front()
	n := val.(node[K, V])
	// 判斷訪問頻次最低的元素迭代器是否就是頻次映射f2i中保存的迭代器
	if iter == lfu.f2i[n.f] {
		lfu.updateFrequency(iter, n.f)
	}
	delete(lfu.k2i, n.k)
	lfu.l.PopFront()
}

func (lfu *lfuShard[K, V]) add(key K, value V) {
	var iter list.Iterator
	niter, o := lfu.f2i[1]
	if !o { // 沒有訪問次數為1對應的迭代器，則新插入的元素肯定為訪問頻次最低的元素，放在鏈表頭
		lfu.l.PushFront(node[K, V]{k: key, v: value, f: 1})
		iter = lfu.l.Begin()
	} else { // 把該元素插入到同訪問頻次下最近被訪問的元素之後，然後該新插入元素的迭代器為同頻次最近訪問的
		iter = lfu.l.InsertContinue(node[K, V]{k: key, v: value, f: 1}, niter)
	}
	lfu.f2i[1] = iter
	lfu.k2i[key] = iter
}

type LFU[K keyType, V any] struct {
	lfuShard[K, V]
}

func NewLFU[K keyType, V any](cap int32) *LFU[K, V] {
	return &LFU[K, V]{
		*newLFUShard[K, V](cap),
	}
}
