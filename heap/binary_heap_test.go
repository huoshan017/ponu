package heap

import (
	"math/rand"
	"testing"
	"time"
)

func TestMaxBinaryHeap(t *testing.T) {
	const (
		rmax int32 = 10000
		n    int   = 50000
	)
	var (
		h   = NewMaxBinaryHeap[int32]()
		r   = rand.New(rand.NewSource(time.Now().Unix()))
		num int32
	)

	for i := 0; i < n; i++ {
		num = r.Int31n(rmax)
		h.Set(num)
	}

	t.Logf("h.array length is %v", len(h.array))
}

func TestMinBinaryHeap(t *testing.T) {
	const (
		rmax int32 = 100000000
		n    int   = 5000000
	)
	var (
		h   = NewMinBinaryHeap[int32]()
		r   = rand.New(rand.NewSource(time.Now().Unix()))
		num int32
	)

	for i := 0; i < n; i++ {
		num = r.Int31n(rmax)
		h.Set(num)
	}

	t.Logf("h.array length is %v", len(h.array))
}

func TestMaxBinaryHeapOrdered(t *testing.T) {
	const (
		rmax int32 = 10000
		n    int   = 50000
	)
	var (
		h   = NewMaxBinaryHeapOrdered[Int32]()
		r   = rand.New(rand.NewSource(time.Now().Unix()))
		num Int32
	)

	for i := 0; i < n; i++ {
		num = Int32(r.Int31n(rmax))
		h.Set(num)
	}

	t.Logf("h.array length is %v", len(h.array))
}

func TestMinBinaryHeapOrdered(t *testing.T) {
	const (
		rmax int32 = 100000000
		n    int   = 5000000
	)
	var (
		h   = NewMinBinaryHeapOrdered[Int32]()
		r   = rand.New(rand.NewSource(time.Now().Unix()))
		num Int32
	)

	for i := 0; i < n; i++ {
		num = Int32(r.Int31n(rmax))
		h.Set(num)
	}

	t.Logf("h.array length is %v", len(h.array))
}
