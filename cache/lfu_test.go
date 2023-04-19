package cache

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/huoshan017/ponu/list"
	"github.com/huoshan017/ponu/log"
)

func TestLFU(t *testing.T) {
	var (
		cap     int32 = 50
		loopNum int32 = 5000000
		maxKey  int32 = 50000
		l             = NewLFU[int32, int32](cap)
		s             = rand.NewSource(time.Now().Unix())
		r             = rand.New(s)
		n, k, v int32
		iter    list.Iterator
	)

	for n = 0; n < loopNum; n++ {
		k = r.Int31n(maxKey) + 1
		v = r.Int31n(maxKey) + 1
		l.Set(k, v)
	}

	for n = 0; n < 1000; n++ {
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
		}
	}

	for n = 0; n < 100; n++ {
		k = r.Int31n(maxKey) + 1
		if !l.Has(k) {
			t.Logf("has no key %v", k)
		}
	}

	dl := l.ToList()

	iter = dl.Begin()
	for iter != dl.End() {
		delete(l.k2i, iter.Value().(node[int32, int32]).k)
		iter = iter.Next()
	}

	t.Logf("k2i(diff len to list: %v):", len(l.k2i))

	t.Logf("f2i(len:%v):", len(l.f2i))
	for k, v := range l.f2i {
		t.Logf("    frequecy: %+v, iterator: %+v", k, v.Value().(node[int32, int32]))
	}

	l.Clear()
}

func TestLFUWithLock(t *testing.T) {
	var (
		cap     int32 = 50
		loopNum int32 = 500000
		maxKey  int32 = 5000
		l             = NewLFUWithLock[int32, int32](cap)
	)

	var wg sync.WaitGroup

	wg.Add(50)
	for g := 0; g < 50; g++ {
		go func() {
			var (
				n, k, v int32
				s       = rand.NewSource(time.Now().Unix())
				r       = rand.New(s)
			)
			defer func() {
				if err := recover(); err != nil {
					log.WithStack(fmt.Errorf("n: %v, k: %v, v: %v, err: %v", n, k, v, err))
				}
			}()
			for n = int32(0); n < loopNum; n++ {
				k = r.Int31n(maxKey) + 1
				v = r.Int31n(maxKey) + 1
				l.Set(k, v)
			}
			wg.Done()
		}()
	}

	wg.Add(10)
	for g := 0; g < 10; g++ {
		go func() {
			s := rand.NewSource(time.Now().Unix())
			r := rand.New(s)
			for n := int32(0); n < loopNum; n++ {
				k := r.Int31n(maxKey) + 1
				l.Get(k)
			}
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s)
		for n := int32(0); n < loopNum; n++ {
			k := r.Int31n(maxKey) + 1
			l.Delete(k)
		}
		wg.Done()
	}()

	wg.Wait()

	k2i_size := int32(len(l.k2i))
	list_size := l.l.GetLength()
	if k2i_size != list_size {
		t.Fatalf("k2i_size(%v) not equal to list_size(%v)", k2i_size, list_size)
	}

	l.Clear()
}
