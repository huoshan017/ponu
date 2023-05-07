package heap

import "golang.org/x/exp/constraints"

type QuadHeap[T constraints.Ordered] struct {
	array []T
	t     HeapType
}

func NewQuadHeap[T constraints.Ordered](t HeapType) *QuadHeap[T] {
	return &QuadHeap[T]{
		t: t,
	}
}

func NewMaxQuadHeap[T constraints.Ordered]() *QuadHeap[T] {
	return &QuadHeap[T]{
		t: HeapType_Max,
	}
}

func NewMinQuadHeap[T constraints.Ordered]() *QuadHeap[T] {
	return &QuadHeap[T]{
		t: HeapType_Min,
	}
}

func (h *QuadHeap[T]) Set(v T) {
	l := len(h.array)
	h.array = append(h.array, v)
	h.adjustUp(l)
}

func (h *QuadHeap[T]) Get() (T, bool) {
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

func (h *QuadHeap[T]) Peek() (T, bool) {
	if len(h.array) <= 0 {
		var v T
		return v, false
	}
	return h.array[0], true
}

func (h *QuadHeap[T]) Length() int32 {
	return int32(len(h.array))
}

func (h *QuadHeap[T]) adjustUp(n int) {
	p := (n - 1) / 4
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
		p = (p - 1) / 4
	}
}

func (h *QuadHeap[T]) adjustDown(n int) {
	var (
		s, m int
		c    = [4]int{s*4 + 1, s*4 + 2, s*4 + 3, s*4 + 4}
	)

	for c[0] <= n {
		s = c[0]
		if h.t == HeapType_Max {
			if c[1] <= n && h.array[m] < h.array[c[1]] {
				m = c[1]
			}
			if c[2] <= n && h.array[m] < h.array[c[2]] {
				m = c[2]
			}
			if c[3] <= n && h.array[m] < h.array[c[3]] {
				m = c[3]
			}
			if h.array[s] >= h.array[m] {
				break
			}
			h.array[s], h.array[m] = h.array[m], h.array[s]
			s = m
		} else {
			if c[1] <= n && h.array[m] > h.array[c[1]] {
				m = c[1]
			}
			if c[2] <= n && h.array[m] > h.array[c[2]] {
				m = c[2]
			}
			if c[3] <= n && h.array[m] > h.array[c[3]] {
				m = c[3]
			}
			if h.array[s] >= h.array[m] {
				break
			}
			h.array[s], h.array[m] = h.array[m], h.array[s]
			s = m
		}
		c[0] = s*4 + 1
		c[1] = s*4 + 2
		c[2] = s*4 + 3
		c[3] = s*4 + 4
	}
}
