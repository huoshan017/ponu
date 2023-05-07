package heap

type HeapType int

const (
	HeapType_Max HeapType = iota
	HeapType_Min
)

type BinaryHeapOrdered[T Ordered] struct {
	array []T
	t     HeapType
}

func NewBinaryHeapOrdered[T Ordered](t HeapType) *BinaryHeapOrdered[T] {
	if t != HeapType_Max && t != HeapType_Min {
		panic("ponu: heap type invalid")
	}
	return &BinaryHeapOrdered[T]{
		t: t,
	}
}

func NewMaxBinaryHeapOrdered[T Ordered]() *BinaryHeapOrdered[T] {
	return &BinaryHeapOrdered[T]{
		t: HeapType_Max,
	}
}

func NewMinBinaryHeapOrdered[T Ordered]() *BinaryHeapOrdered[T] {
	return &BinaryHeapOrdered[T]{
		t: HeapType_Min,
	}
}

func (h *BinaryHeapOrdered[T]) Set(v T) {
	l := len(h.array)
	h.array = append(h.array, v)
	h.adjustUp(l)
}

func (h *BinaryHeapOrdered[T]) Get() (T, bool) {
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

func (h *BinaryHeapOrdered[T]) Peek() (T, bool) {
	var v T
	if len(h.array) <= 0 {
		return v, false
	}
	return h.array[0], true
}

func (h *BinaryHeapOrdered[T]) Length() int32 {
	return int32(len(h.array))
}

func (h *BinaryHeapOrdered[T]) adjustUp(n int) {
	p := (n - 1) / 2
	for p >= 0 {
		if h.t == HeapType_Max {
			if h.array[n].LessEqual(h.array[p]) {
				break
			}
		} else {
			if h.array[n].GreaterEqual(h.array[p]) {
				break
			}
		}
		h.array[n], h.array[p] = h.array[p], h.array[n]
		n = p
		p = (p - 1) / 2
	}
}

func (h BinaryHeapOrdered[T]) adjustDown(n int) {
	var (
		c, m int
		l    = c*2 + 1
		r    = c*2 + 2
	)

	for l <= n {
		m = l
		if h.t == HeapType_Max { // 大顶堆
			if r <= n && h.array[m].Less(h.array[r]) {
				m = r
			}
			if h.array[c].GreaterEqual(h.array[m]) {
				break
			}
			h.array[c], h.array[m] = h.array[m], h.array[c]
			c = m
		} else { // 小顶堆
			if r <= n && h.array[m].Greater(h.array[r]) {
				m = r
			}
			if h.array[c].LessEqual(h.array[m]) {
				break
			}
			h.array[c], h.array[m] = h.array[m], h.array[c]
			c = m
		}
		l = c*2 + 1 // left child
		r = c*2 + 2 // right child
	}
}
