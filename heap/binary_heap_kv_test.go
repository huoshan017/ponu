package heap

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestBinaryHeapSetGet(t *testing.T) {
	const (
		count int32 = 100
	)
	var (
		h      = NewMaxBinaryHeapKV[int32, int32]()
		source = rand.NewSource(time.Now().Unix())
	)
	r := rand.New(source)
	for i := int32(0); i < count; i++ {
		v := r.Int31n(1000)
		h.Set(i, v)
	}

	t.Logf("h length is %v", h.Length())

	var (
		o = true
		v int32
		s string
	)
	for o {
		_, v, o = h.Get()
		if o {
			s = fmt.Sprintf("%v %v", s, v)
		}
	}

	t.Logf("after get, h length is %v, get value list: %v", h.Length(), s)
}

func TestBinaryMinHeapKV(t *testing.T) {
	const (
		rmax int32 = 100000000
		n    int32 = 5000000
	)
	var (
		h        = NewMinBinaryHeapKV[int32, int32]()
		r        = rand.New(rand.NewSource(time.Now().Unix()))
		key, num int32
	)

	for i := int32(0); i < n; i++ {
		key = i
		num = r.Int31n(rmax)
		h.Set(key, num)
	}

	for i := 0; i < 100000; i++ {
		key = r.Int31n(n)
		h.Delete(key)
	}

	for i := 0; i < 100000; i++ {
		h.Get()
	}

	t.Logf("h.array length is %v", len(h.array))
}

func TestBinaryMaxHeapKV(t *testing.T) {
	const (
		rmax int32 = 100000000
		n    int32 = 5000000
	)
	var (
		h        = NewMaxBinaryHeapKV[int32, int32]()
		r        = rand.New(rand.NewSource(time.Now().Unix()))
		key, num int32
	)

	for i := int32(0); i < n; i++ {
		key = i
		num = r.Int31n(rmax)
		h.Set(key, num)
	}

	for i := 0; i < 100000; i++ {
		key = r.Int31n(n)
		h.Delete(key)
	}

	t.Logf("h.array length is %v", len(h.array))
}
