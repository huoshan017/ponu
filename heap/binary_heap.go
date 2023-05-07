package heap

import "golang.org/x/exp/constraints"

type BinaryHeap[T constraints.Ordered] struct {
	array []T
	t     HeapType
}

func NewBinaryHeap[T constraints.Ordered](t HeapType) *BinaryHeap[T] {
	if t != HeapType_Max && t != HeapType_Min {
		panic("ponu: heap type invalid")
	}
	return &BinaryHeap[T]{
		t: t,
	}
}

func NewMaxBinaryHeap[T constraints.Ordered]() *BinaryHeap[T] {
	return &BinaryHeap[T]{
		t: HeapType_Max,
	}
}

func NewMinBinaryHeap[T constraints.Ordered]() *BinaryHeap[T] {
	return &BinaryHeap[T]{
		t: HeapType_Min,
	}
}

func (h *BinaryHeap[T]) Set(v T) {
	l := len(h.array)
	h.array = append(h.array, v)
	h.adjustUp(l)
}

func (h *BinaryHeap[T]) Get() (T, bool) {
	var v T
	l := len(h.array)
	if l <= 0 {
		return v, false
	}
	v = h.array[0]
	h.array[0] = h.array[l-1]
	h.array = h.array[:l-1]
	l -= 1
	h.adjustDown(l - 1)
	return v, true
}

func (h *BinaryHeap[T]) Peek() (T, bool) {
	if len(h.array) <= 0 {
		var v T
		return v, false
	}
	return h.array[0], true
}

func (h *BinaryHeap[T]) Length() int32 {
	return int32(len(h.array))
}

func (h *BinaryHeap[T]) adjustUp(n int) {
	p := (n - 1) / 2
	for p >= 0 {
		if h.t == HeapType_Max {
			if h.array[n] <= h.array[p] {
				break
			}
		} else {
			if h.array[n] >= h.array[p] {
				break
			}
		}
		h.array[n], h.array[p] = h.array[p], h.array[n]
		n = p
		p = (p - 1) / 2
	}
}

func (h BinaryHeap[T]) adjustDown(n int) {
	var (
		c, m int
		l    = c*2 + 1
		r    = c*2 + 2
	)

	for l <= n {
		m = l
		if h.t == HeapType_Max { // 大顶堆
			if r <= n && h.array[m] < h.array[r] {
				m = r
			}
			if h.array[c] >= h.array[m] {
				break
			}
			h.array[c], h.array[m] = h.array[m], h.array[c]
			c = m
		} else { // 小顶堆
			if r <= n && h.array[m] > h.array[r] {
				m = r
			}
			if h.array[c] <= h.array[m] {
				break
			}
			h.array[c], h.array[m] = h.array[m], h.array[c]
			c = m
		}
		l = c*2 + 1 // left child
		r = c*2 + 2 // right child
	}
}
