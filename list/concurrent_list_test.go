package list

import (
	"log"
	"sync"
	"testing"
)

const (
	pushCount = 100000000
)

func testConcurrentList(t *testing.T, nonBlock bool, length int32) {
	var (
		cl = func() *ConcurrentList {
			if length > 0 {
				return NewConcurrentListWithLength(length)
			} else {
				return NewConcurrentList()
			}
		}()
		wg = sync.WaitGroup{}
	)
	go func() {
		for i := 0; i < pushCount; {
			if !nonBlock {
				if cl.PushBack(i) {
					i += 1
				}
			} else {
				if cl.PushBackNonBlock(i) {
					i += 1
				}
			}
		}
	}()

	wg.Add(pushCount)
	go func() {
		for i := 0; i < pushCount; {
			var (
				done  bool
				value any
				o     bool
			)
			if !nonBlock {
				if value, o = cl.PopFront(); o {
					if value.(int) != i {
						log.Fatalf("block op front value %v not equal to %v", value, i)
					}
					done = true
				}
			} else {
				if value, o = cl.PopFrontNonBlock(); o {
					if value.(int) != i {
						log.Fatalf("nonblok pop front value %v not equal to %v", value, i)
					}
					done = true
				}
			}
			if done {
				wg.Done()
				i += 1
			}
		}
	}()

	wg.Wait()

	cl.Clear()
}

func TestConcurrentListBlock(t *testing.T) {
	testConcurrentList(t, false, 0)
}

func TestConcurrentListBlockWithLength(t *testing.T) {
	testConcurrentList(t, false, 1000000)
}

func TestConcurrentListNonblock(t *testing.T) {
	testConcurrentList(t, true, 0)
}

func TestConcurrentListNonblockWithLength(t *testing.T) {
	testConcurrentList(t, true, 1000000)
}
