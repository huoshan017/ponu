package cache

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestLFU(t *testing.T) {
	var (
		cap     uint32 = 500
		loopNum int32  = 5000
		maxKey  int32  = 5000
		l              = NewLFU[int32, int32](cap)
		s              = rand.NewSource(time.Now().Unix())
		r              = rand.New(s)
		n, k, v int32
		str     string
	)

	for n = 0; n < loopNum; n++ {
		k = r.Int31n(maxKey) + 1
		v = r.Int31n(maxKey) + 1
		l.Set(k, v)
	}

	for n = 0; n < 10; n++ {
		k = r.Int31n(maxKey) + 1
		_, o := l.Get(k)
		if !o {
			t.Logf("cant get value with key %v", k)
		}
	}

	for n = 0; n < 10; n++ {
		k = r.Int31n(maxKey) + 1
		if !l.Delete(k) {
			t.Logf("cant delete with key %v", k)
		} else {
			t.Logf("deleted key %v", k)
		}
	}

	dl := l.ToList()

	for k, v := range l.k2i {
		str += fmt.Sprintf("  key: %+v, iterator: %+v", k, v.Value().(node[int32, int32]))
	}
	t.Logf("k2i(len:%v) values: %v", len(l.k2i), str)

	iter := l.l.Begin()
	for iter != l.l.End() {
		str += fmt.Sprintf(" %v", iter.Value())
		iter = iter.Next()
	}
	t.Logf("original list(len: %v) value is: %v", l.l.GetLength(), str)

	str = ""
	iter = dl.Begin()
	for iter != dl.End() {
		str += fmt.Sprintf(" %v", iter.Value())
		delete(l.k2i, iter.Value().(node[int32, int32]).k)
		iter = iter.Next()
	}
	t.Logf("duplicate list(len: %v) value is: %v", dl.GetLength(), str)

	t.Logf("k2i(diff len to list: %v):", len(l.k2i))
	for k, v := range l.k2i {
		t.Logf("	ke: %+v, iterator: %+v", k, v.Value().(node[int32, int32]))
	}

	t.Logf("f2i(len:%v):", len(l.f2i))
	for k, v := range l.f2i {
		t.Logf("    frequecy: %+v, iterator: %+v", k, v.Value().(node[int32, int32]))
	}
}
