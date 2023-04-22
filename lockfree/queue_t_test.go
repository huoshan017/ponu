package lockfree

import (
	"sync"
	"testing"
)

func TestQueueT(t *testing.T) {
	const (
		count = 100000000
		gn    = 5
	)
	q := NewQueueT[int]()
	for n := 0; n < gn; n++ {
		go func() {
			for i := 0; i < count; i++ {
				q.Enqueue(i)
			}
		}()
	}

	var wg = sync.WaitGroup{}
	wg.Add(count * gn)
	go func() {
		for i := 0; i < count*gn; i++ {
			q.Dequeue()
			wg.Done()
		}
	}()

	wg.Wait()
}
