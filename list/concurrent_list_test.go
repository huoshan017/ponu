package list

import (
	"sync"
	"testing"
)

func testConcurrentList(t *testing.T, length int32) {
	const (
		count = 100000000
	)
	var (
		cl = NewConcurrentList()
		wg = sync.WaitGroup{}
	)
	go func() {
		for i := 0; i < count; i++ {
			cl.PushBack(i)
		}
	}()

	wg.Add(count)
	go func() {
		for i := 0; i < count; i++ {
			cl.PopFront()
			wg.Done()
		}
	}()

	wg.Wait()

	cl.Clear()
}

func TestConcurrentList(t *testing.T) {
	testConcurrentList(t, 0)
}

func TestConcurrentListLength(t *testing.T) {
	testConcurrentList(t, 10000)
}
