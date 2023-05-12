package heap

import "golang.org/x/exp/constraints"

type pair[K comparable, V constraints.Ordered] struct {
	k K
	v V
}

type BinaryHeapKV[K comparable, V constraints.Ordered] struct {
	array []pair[K, V]
	k2n   map[K]int32
	t     HeapType
}

func NewBinaryHeapKV[K comparable, V constraints.Ordered](t HeapType) *BinaryHeapKV[K, V] {
	if t != HeapType_Max && t != HeapType_Min {
		panic("ponu: heap type invalid")
	}
	return &BinaryHeapKV[K, V]{
		t:   t,
		k2n: make(map[K]int32),
	}
}

func NewMaxBinaryHeapKV[K comparable, V constraints.Ordered]() *BinaryHeapKV[K, V] {
	return NewBinaryHeapKV[K, V](HeapType_Max)
}

func NewMinBinaryHeapKV[K comparable, V constraints.Ordered]() *BinaryHeapKV[K, V] {
	return NewBinaryHeapKV[K, V](HeapType_Min)
}

func (h *BinaryHeapKV[K, V]) Set(k K, v V) {
	l := len(h.array)
	h.array = append(h.array, pair[K, V]{k, v})
	l += 1
	h.k2n[k] = int32(l - 1)
	h.adjustUp(l - 1)
}

func (h *BinaryHeapKV[K, V]) Get() (K, V, bool) {
	l := len(h.array)
	if l <= 0 {
		var (
			k K
			v V
		)
		return k, v, false
	}
	kv := h.array[0]
	h.array[0] = h.array[l-1]
	h.array = h.array[:l-1]
	l -= 1
	delete(h.k2n, kv.k)
	h.k2n[h.array[0].k] = 0
	h.adjustDown(0, l-1)
	return kv.k, kv.v, true
}

func (h *BinaryHeapKV[K, V]) Peek() (K, V, bool) {
	if len(h.array) <= 0 {
		var (
			k K
			v V
		)
		return k, v, false
	}
	return h.array[0].k, h.array[0].v, true
}

func (h *BinaryHeapKV[K, V]) Length() int32 {
	return int32(len(h.array))
}

func (h *BinaryHeapKV[K, V]) Delete(k K) (V, bool) {
	n, o := h.k2n[k]
	if !o {
		var v V
		return v, false
	}
	v := h.array[n].v
	l := len(h.array)
	if n != int32(l-1) {
		h.array[n] = h.array[l-1]
		h.k2n[h.array[n].k] = n
	}
	delete(h.k2n, k)
	h.array = h.array[:l-1]
	l -= 1
	h.adjustDown(int(n), l-1)
	return v, true
}

func (h *BinaryHeapKV[K, V]) DeleteCallback(k K, onDelete func(K, V)) bool {
	var (
		v V
		o bool
	)
	if v, o = h.Delete(k); !o {
		return false
	}
	onDelete(k, v)
	return true
}

func (h *BinaryHeapKV[K, V]) adjustUp(n int) {
	p := (n - 1) / 2
	for p >= 0 {
		if h.t == HeapType_Max {
			if h.array[n].v <= h.array[p].v {
				break
			}
		} else {
			if h.array[n].v >= h.array[p].v {
				break
			}
		}
		h.array[n], h.array[p] = h.array[p], h.array[n]
		kn, kp := h.array[n].k, h.array[p].k
		h.k2n[kn] = int32(n)
		h.k2n[kp] = int32(p)
		n = p
		p = (p - 1) / 2
	}
}

func (h BinaryHeapKV[K, V]) adjustDown(c, n int) {
	var (
		m int
		l = c*2 + 1 // left child
		r = c*2 + 2 // right child
	)

	for l <= n {
		m = l
		if h.t == HeapType_Max { // 大顶堆
			if r <= n && h.array[m].v < h.array[r].v {
				m = r
			}
			if h.array[c].v >= h.array[m].v {
				break
			}
		} else { // 小顶堆
			if r <= n && h.array[m].v > h.array[r].v {
				m = r
			}
			if h.array[c].v <= h.array[m].v {
				break
			}
		}
		h.array[c], h.array[m] = h.array[m], h.array[c]
		kc, km := h.array[c].k, h.array[m].k
		h.k2n[kc] = int32(c)
		h.k2n[km] = int32(m)
		c = m
		l = c*2 + 1
		r = c*2 + 2
	}
}
