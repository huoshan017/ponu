package list

import (
	"sync"
)

type ConcurrentListT[T any] struct {
	ListT[T]
	maxLength                 int32
	mutex                     sync.Mutex
	notFullCond, notEmptyCond *sync.Cond
}

func NewConcurrentListT[T any](pool *ListTNodePool[T]) *ConcurrentListT[T] {
	cl := &ConcurrentListT[T]{
		ListT: NewListTObjWithPool(pool),
	}
	cl.notFullCond = sync.NewCond(&cl.mutex)
	cl.notEmptyCond = sync.NewCond(&cl.mutex)
	return cl
}

func NewConcurrentListTWithLength[T any](pool *ListTNodePool[T], maxLength int32) *ConcurrentListT[T] {
	if maxLength <= 0 {
		panic("ponu.list ConcurrentList need maxLength greater to zero")
	}
	cl := NewConcurrentListT(pool)
	cl.maxLength = maxLength
	return cl
}

func (l *ConcurrentListT[T]) PushBack(value T) bool {
	return l.pushBack(value, false)
}

func (l *ConcurrentListT[T]) PushBackNonBlock(value T) bool {
	return l.pushBack(value, true)
}

func (l *ConcurrentListT[T]) PopFront() (T, bool) {
	return l.popFront(false)
}

func (l *ConcurrentListT[T]) PopFrontNonBlock() (T, bool) {
	return l.popFront(true)
}

func (l *ConcurrentListT[T]) Length() int32 {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.ListT.GetLength()
}

func (l *ConcurrentListT[T]) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.ListT.Clear()
}

func (l *ConcurrentListT[T]) pushBack(value T, nonBlock bool) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.maxLength > 0 {
		for l.ListT.GetLength() >= l.maxLength {
			if nonBlock { // 长度受限且非阻塞则返回失败
				return false
			}
			l.notFullCond.Wait()
		}
	}
	l.ListT.PushBack(value)
	l.notEmptyCond.Signal()
	return true
}

func (l *ConcurrentListT[T]) popFront(nonBlock bool) (T, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for l.ListT.GetLength() == 0 {
		if nonBlock {
			var t T
			return t, false
		}
		l.notEmptyCond.Wait()
	}
	val, _ := l.ListT.PopFront()
	if l.maxLength > 0 {
		l.notFullCond.Signal()
	}
	return val, true
}
