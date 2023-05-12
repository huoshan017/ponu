package list

import (
	"sync"
)

type ConcurrentList struct {
	List
	maxLength                 int32
	mutex                     sync.Mutex
	notFullCond, notEmptyCond *sync.Cond
}

func NewConcurrentList() *ConcurrentList {
	cl := &ConcurrentList{
		List: List{},
	}
	cl.notFullCond = sync.NewCond(&cl.mutex)
	cl.notEmptyCond = sync.NewCond(&cl.mutex)
	return cl
}

func NewConcurrentListWithLength(maxLength int32) *ConcurrentList {
	if maxLength <= 0 {
		panic("ponu.list ConcurrentList need maxLength greater to zero")
	}
	cl := NewConcurrentList()
	cl.maxLength = maxLength
	return cl
}

func (l *ConcurrentList) PushBack(value any) bool {
	return l.pushBack(value, false)
}

func (l *ConcurrentList) PushBackNonBlock(value any) bool {
	return l.pushBack(value, true)
}

func (l *ConcurrentList) PopFront() (any, bool) {
	return l.popFront(false)
}

func (l *ConcurrentList) PopFrontNonBlock() (any, bool) {
	return l.popFront(true)
}

func (l *ConcurrentList) Length() int32 {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.List.GetLength()
}

func (l *ConcurrentList) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.List.Clear()
}

func (l *ConcurrentList) pushBack(value any, nonBlock bool) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.maxLength > 0 {
		for l.List.GetLength() >= l.maxLength {
			if nonBlock { // 长度受限且非阻塞则返回失败
				return false
			}
			l.notFullCond.Wait()
		}
	}
	l.List.PushBack(value)
	l.notEmptyCond.Signal()
	return true
}

func (l *ConcurrentList) popFront(nonBlock bool) (any, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for l.List.GetLength() == 0 {
		if nonBlock {
			return nil, false
		}
		l.notEmptyCond.Wait()
	}
	val, _ := l.List.PopFront()
	if l.maxLength > 0 {
		l.notFullCond.Signal()
	}
	return val, true
}
