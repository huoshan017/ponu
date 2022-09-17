package lockfree

import (
	"sync"
	"testing"
)

func TestQueue(t *testing.T) {
	const (
		count = 100000000
		gn    = 5
	)
	q := NewQueue()
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
