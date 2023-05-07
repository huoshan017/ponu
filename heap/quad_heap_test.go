package heap

import (
	"math/rand"
	"testing"
	"time"
)

func TestMaxQuadHeap(t *testing.T) {
	const (
		rmax int32 = 10000
		n    int   = 50000
	)
	var (
		h   = NewMaxQuadHeap[int32]()
		r   = rand.New(rand.NewSource(time.Now().Unix()))
		num int32
	)

	for i := 0; i < n; i++ {
		num = r.Int31n(rmax)
		h.Set(num)
	}

	t.Logf("h.array length is %v", len(h.array))
}

func TestMinQuadHeap(t *testing.T) {
	const (
		rmax int32 = 100000000
		n    int   = 5000000
	)
	var (
		h   = NewMinQuadHeap[int32]()
		r   = rand.New(rand.NewSource(time.Now().Unix()))
		num int32
	)

	for i := 0; i < n; i++ {
		num = r.Int31n(rmax)
		h.Set(num)
	}

	t.Logf("h.array length is %v", len(h.array))
}
