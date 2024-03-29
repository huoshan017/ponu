package cache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/huoshan017/ponu/list"
)

const (
	minCap         = 8
	minExpiredTime = time.Second
)

type node[K comparable, V any] struct {
	Pair[K, V]
	f int32
}

type lfuBase[K comparable, V any] struct {
	cap         int32                   // 容量
	onEvictFun  func(K, V)              // value的回收内存函数
	expiredTime time.Duration           // 超时淘汰时间
	ts          *int32                  // 总大小
	l           list.List               // 数据列表
	k2i         map[K]list.Iterator     // key对应的列表节点
	f2i         map[int32]list.Iterator // 保存訪問次數對應的迭代器，這個迭代器是該訪問次數下最新的，如果有更新的迭代器，則插在它之後
	createTime  time.Time               // 创建时间
	k2t         map[K]time.Duration     // key对应的时间点
}

func newLFUBase[K comparable, V any](cap int32) *lfuBase[K, V] {
	if cap < minCap {
		cap = minCap
	}
	return &lfuBase[K, V]{
		cap:        cap,
		l:          list.List{},
		k2i:        make(map[K]list.Iterator),
		f2i:        make(map[int32]list.Iterator),
		createTime: time.Now(),
		k2t:        make(map[K]time.Duration),
	}
}

func newLFUBaseWithTotalSize[K comparable, V any](cap int32, totalSize *int32) *lfuBase[K, V] {
	if cap < minCap {
		cap = minCap
	}
	return &lfuBase[K, V]{
		cap:        cap,
		ts:         totalSize,
		l:          list.List{},
		k2i:        make(map[K]list.Iterator),
		f2i:        make(map[int32]list.Iterator),
		createTime: time.Now(),
		k2t:        make(map[K]time.Duration),
	}
}

func (lfu *lfuBase[K, V]) WithOnEvict(fun func(K, V)) {
	lfu.onEvictFun = fun
}

func (lfu *lfuBase[K, V]) WithExpiredtime(t time.Duration) {
	if t < minExpiredTime {
		t = minExpiredTime
	}
	lfu.expiredTime = t
}

func (lfu *lfuBase[K, V]) Set(key K, value V) {
	if lfu.k2i == nil {
		lfu.k2i = make(map[K]list.Iterator)
	}
	if lfu.f2i == nil {
		lfu.f2i = make(map[int32]list.Iterator)
	}
	if lfu.k2t == nil {
		lfu.k2t = make(map[K]time.Duration)
	}
	iter, o := lfu.k2i[key]
	if !o { // 插入
		if lfu.ts == nil {
			if lfu.cap <= int32(len(lfu.k2i)) { // 超出容量大小
				lfu.deleteFirst()
			}
			lfu.add(key, value)
		} else {
			if lfu.cap <= atomic.LoadInt32(lfu.ts) { // 原子操作判斷大小已達最大容量
				// 刪除掉一個最不常訪問的，再添加一個，保持總數量不變
				// 如果刪除失敗，表示鏈表已空，總數量加一。
				// 對於一個Shard來説，同一時間保證只有一個goroutine對其Set操作
				if !lfu.deleteFirst() {
					atomic.AddInt32(lfu.ts, 1)
				}
				lfu.add(key, value)
			} else {
				// 前面通過原子操作知道totalSize沒到最大容量cap
				// 再進行原子操作加一后得到當前大小，這時有兩種結果：
				// 1. 大於容量cap，表示這中間有其他goroutine添加了元素，則把當前大小減一(盡可能快，
				//	  因爲會影響到其他goroutine)，刪掉一個再添加一個，保持總數量不變
				// 2. 小於等於cap，則直接添加元素
				if lfu.cap < atomic.AddInt32(lfu.ts, 1) {
					if lfu.deleteFirst() {
						atomic.AddInt32(lfu.ts, -1)
					}
					lfu.add(key, value)
				} else {
					lfu.add(key, value)
				}
			}
		}
	} else { // 更新
		lfu.updateValue(iter, true, value)
	}
	if lfu.expiredTime > 0 {
		lfu.k2t[key] = lfu.getExpiredTimePoint(time.Now(), lfu.expiredTime)
	}
}

func (lfu *lfuBase[K, V]) SetExpired(key K, value V, expiredTime time.Duration) {
	lfu.Set(key, value)
	lfu.k2t[key] = lfu.getExpiredTimePoint(time.Now(), expiredTime)
}

func (lfu *lfuBase[K, V]) Get(key K) (V, bool) {
	var v V
	if lfu.k2i == nil {
		return v, false
	}
	iter, o := lfu.k2i[key]
	if !o {
		return v, false
	}
	if lfu.isKeyExpired(key) {
		lfu.delete(key)
		return v, false
	}
	// iter在update之後有可能會失效，所以要在update之前取到其中的值
	v = iter.Value().(node[K, V]).v
	lfu.update(iter)
	return v, true
}

func (lfu *lfuBase[K, V]) Delete(key K) bool {
	return lfu.delete(key)
}

func (lfu *lfuBase[K, V]) Has(key K) bool {
	_, o := lfu.k2i[key]
	if o {
		if lfu.isKeyExpired(key) {
			lfu.delete(key)
			return false
		}
	}
	return o
}

func (lfu *lfuBase[K, V]) ToList(lis *list.ListT[Pair[K, V]]) {
	for iter := lfu.l.Begin(); iter != lfu.l.End(); {
		n := iter.Value().(node[K, V])
		if lfu.isKeyExpired(n.k) {
			iter = iter.Next()
			lfu.delete(n.k)
			continue
		}
		lis.PushBack(Pair[K, V]{k: n.k, v: n.v})
		iter = iter.Next()
	}
}

func (lfu *lfuBase[K, V]) Clear() {
	lfu.l.Clear()
	lfu.k2i = nil
	lfu.f2i = nil
	lfu.k2t = nil
	if lfu.ts != nil {
		atomic.AddInt32(lfu.ts, -lfu.l.GetLength())
	}
}

func (lfu *lfuBase[K, V]) update(iter list.Iterator) {
	var v V
	lfu.updateValue(iter, false, v)
}

func (lfu *lfuBase[K, V]) updateValue(iter list.Iterator, update bool, val V) {
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
		//delete(lfu.k2i, n.k)
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
			//delete(lfu.k2i, n.k)
			iter = lfu.l.InsertContinue(n, niter2)
			lfu.k2i[n.k] = iter
		} else { // 沒有f+1次訪問的元素，則表示更新前後位置不變，直接修改元素值
			lfu.l.Update(n, iter)
		}
	}
	lfu.f2i[n.f] = iter
}

func (lfu *lfuBase[K, V]) updateFrequency(iter list.Iterator, f int32) {
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

func (lfu *lfuBase[K, V]) deleteFirst() bool {
	if lfu.l.GetLength() <= 0 {
		return false
	}
	iter := lfu.l.Begin()
	val := lfu.l.Front()
	n := val.(node[K, V])
	// 判斷訪問頻次最低的元素迭代器是否就是頻次映射f2i中保存的迭代器
	if iter == lfu.f2i[n.f] {
		lfu.updateFrequency(iter, n.f)
	}
	delete(lfu.k2i, n.k)
	lfu.l.PopFront()
	delete(lfu.k2t, n.k)
	if lfu.onEvictFun != nil {
		lfu.onEvictFun(n.k, n.v)
	}
	return true
}

func (lfu *lfuBase[K, V]) delete(key K) bool {
	iter, o := lfu.k2i[key]
	if !o {
		return false
	}
	n := iter.Value().(node[K, V])
	if iter == lfu.f2i[n.f] {
		lfu.updateFrequency(iter, n.f)
	}
	delete(lfu.k2i, key)
	lfu.l.Delete(iter)
	delete(lfu.k2t, n.k)
	if lfu.onEvictFun != nil {
		lfu.onEvictFun(n.k, n.v)
	}
	if lfu.ts != nil {
		atomic.AddInt32(lfu.ts, -1)
	}
	return true
}

func (lfu *lfuBase[K, V]) add(key K, value V) {
	var iter list.Iterator
	niter, o := lfu.f2i[1]
	if !o { // 沒有訪問次數為1對應的迭代器，則新插入的元素肯定為訪問頻次最低的元素，放在鏈表頭
		lfu.l.PushFront(node[K, V]{Pair: Pair[K, V]{k: key, v: value}, f: 1})
		iter = lfu.l.Begin()
	} else { // 把該元素插入到同訪問頻次下最近被訪問的元素之後，然後該新插入元素的迭代器為同頻次最近訪問的
		iter = lfu.l.InsertContinue(node[K, V]{Pair: Pair[K, V]{k: key, v: value}, f: 1}, niter)
	}
	lfu.f2i[1] = iter
	lfu.k2i[key] = iter
	if lfu.expiredTime > 0 {
		lfu.k2t[key] = lfu.getExpiredTimePoint(time.Now(), lfu.expiredTime)
	}
}

func (lfu *lfuBase[K, V]) isKeyExpired(key K) bool {
	if d, o := lfu.k2t[key]; o {
		if lfu.isExpired(time.Now(), d) {
			return true
		}
	}
	return false
}

func (lfu *lfuBase[K, V]) getExpiredTimePoint(now time.Time, duration time.Duration) time.Duration {
	return now.Sub(lfu.createTime) + duration
}

func (lfu *lfuBase[K, V]) isExpired(now time.Time, timePoint time.Duration) bool {
	return now.Sub(lfu.createTime) >= timePoint
}

type LFU[K comparable, V any] struct {
	*lfuBase[K, V]
}

func NewLFU[K comparable, V any](cap int32) *LFU[K, V] {
	return &LFU[K, V]{
		newLFUBase[K, V](cap),
	}
}

type LFUWithLock[K comparable, V any] struct {
	*lfuBase[K, V]
	rwlock sync.RWMutex
}

func NewLFUWithLock[K comparable, V any](cap int32) *LFUWithLock[K, V] {
	return &LFUWithLock[K, V]{
		lfuBase: newLFUBase[K, V](cap),
	}
}

func newLFUWithLockAndTotalSize[K comparable, V any](cap int32, totalSize *int32) *LFUWithLock[K, V] {
	return &LFUWithLock[K, V]{
		lfuBase: newLFUBaseWithTotalSize[K, V](cap, totalSize),
	}
}

func (lfu *LFUWithLock[K, V]) WithOnEvict(fun func(K, V)) {
	lfu.rwlock.Lock()
	defer lfu.rwlock.Unlock()
	lfu.lfuBase.WithOnEvict(fun)
}

func (lfu *LFUWithLock[K, V]) Set(key K, value V) {
	lfu.rwlock.Lock()
	defer lfu.rwlock.Unlock()
	lfu.lfuBase.Set(key, value)
}

func (lfu *LFUWithLock[K, V]) Get(key K) (V, bool) {
	lfu.rwlock.Lock()
	defer lfu.rwlock.Unlock()
	return lfu.lfuBase.Get(key)
}

func (lfu *LFUWithLock[K, V]) Has(key K) bool {
	lfu.rwlock.RLock()
	defer lfu.rwlock.RUnlock()
	return lfu.lfuBase.Has(key)
}

func (lfu *LFUWithLock[K, V]) Delete(key K) bool {
	lfu.rwlock.Lock()
	defer lfu.rwlock.Unlock()
	return lfu.lfuBase.Delete(key)
}

func (lfu *LFUWithLock[K, V]) ToList(lis *list.ListT[Pair[K, V]]) {
	lfu.rwlock.RLock()
	defer lfu.rwlock.RUnlock()
	lfu.lfuBase.ToList(lis)
}

func (lfu *LFUWithLock[K, V]) Clear() {
	lfu.rwlock.Lock()
	defer lfu.rwlock.Unlock()
	lfu.lfuBase.Clear()
}
