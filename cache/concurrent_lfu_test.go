package cache

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/huoshan017/ponu/log"
)

func TestConcurrentLFU(t *testing.T) {
	var (
		cap     int32 = 50
		loopNum int32 = 500000
		maxKey  int32 = 5000
		l             = NewConcurrentLFU[int32, int32](cap)
	)

	var wg sync.WaitGroup

	wg.Add(20)
	for g := 0; g < 20; g++ {
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
			for n := 0; n < 1000; n++ {
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
		for n := 0; n < 10; n++ {
			k := r.Int31n(maxKey) + 1
			l.Delete(k)
		}
		wg.Done()
	}()

	wg.Wait()

	if l.currSize > cap {
		t.Fatalf("curr size %v cant greater to cap %v", l.currSize, cap)
	}

	for i := 0; i < len(l.shards); i++ {
		k2i_size := int32(len(l.shards[i].k2i))
		list_size := l.shards[i].l.GetLength()
		if k2i_size != list_size {
			t.Fatalf("shard %v k2i_size(%v) not equal to list_size(%v)", i, k2i_size, list_size)
		}
	}

	l.Clear()
}

func TestConcurrentLFUWithString(t *testing.T) {
	var l = NewConcurrentLFU[string, int32](100)
	l.Set("aaa", 1)
	l.Set("bbb", 2)
}
