package list

import (
	"log"
	"sync"
	"testing"
)

var (
	concurrentListNodePool = NewListTNodePool[int]()
)

func testConcurrentListT(t *testing.T, nonBlock bool, length int32) {
	var (
		cl = func() *ConcurrentListT[int] {
			if length > 0 {
				return NewConcurrentListTWithLength(concurrentListNodePool, length)
			} else {
				return NewConcurrentListT(concurrentListNodePool)
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

func TestConcurrentListTBlock(t *testing.T) {
	testConcurrentListT(t, false, 0)
}

func TestConcurrentListTBlockWithLength(t *testing.T) {
	testConcurrentListT(t, false, 1000000)
}

func TestConcurrentListTNonblock(t *testing.T) {
	testConcurrentListT(t, true, 0)
}

func TestConcurrentListTNonblockWithLength(t *testing.T) {
	testConcurrentListT(t, true, 1000000)
}
