package list

import (
	"sync"
	"unsafe"
)

var (
	nullTNodePtr = unsafe.Pointer(&node_t[struct{}]{})
)

type ListTNodePool[T any] struct {
	pool *sync.Pool
}

func NewListTNodePool[T any]() *ListTNodePool[T] {
	return &ListTNodePool[T]{
		pool: &sync.Pool{
			New: func() any {
				return &node_t[T]{}
			},
		},
	}
}

func (pool *ListTNodePool[T]) get() *node_t[T] {
	return pool.pool.Get().(*node_t[T])
}

func (pool *ListTNodePool[T]) put(n *node_t[T]) {
	n.prev = nil
	n.next = nil
	pool.pool.Put(n)
}

type node_t[T any] struct {
	value      T
	prev, next *node_t[T]
}

type IteratorT[T any] struct {
	n *node_t[T]
}

func (iter *IteratorT[T]) Value() T {
	return iter.n.value
}

func (iter *IteratorT[T]) Next() IteratorT[T] {
	if iter.n.next == nil {
		return IteratorT[T]{n: (*node_t[T])(nullTNodePtr)}
	}
	return IteratorT[T]{n: iter.n.next}
}

func (iter *IteratorT[T]) Prev() IteratorT[T] {
	if iter.n.prev == nil {
		return IteratorT[T]{n: (*node_t[T])(nullTNodePtr)}
	}
	return IteratorT[T]{n: iter.n.prev}
}

func (iter *IteratorT[T]) IsValid() bool {
	return iter.n != nil && iter.n != (*node_t[T])(nullTNodePtr)
}

func (iter *IteratorT[T]) IsHead(l *ListT[T]) bool {
	return iter.n == l.head
}

func (iter *IteratorT[T]) IsTail(l *ListT[T]) bool {
	return iter.n == l.tail
}

func (iter *IteratorT[T]) IsEqualTo(i IteratorT[T]) bool {
	return iter.n == i.n
}

type ListT[T any] struct {
	head, tail *node_t[T]
	length     int32
	nodePool   *ListTNodePool[T]
}

func NewListT[T any](pool *ListTNodePool[T]) *ListT[T] {
	return &ListT[T]{
		nodePool: pool,
	}
}

func NewListTObj[T any](pool *ListTNodePool[T]) ListT[T] {
	return ListT[T]{
		nodePool: pool,
	}
}

func (l ListT[T]) GetLength() int32 {
	return l.length
}

func (l ListT[T]) IsEmpty() bool {
	return l.length == 0
}

func (l *ListT[T]) PushFront(val T) {
	n := l.nodePool.get()
	n.value = val
	n.next = l.head
	if l.head == nil {
		l.tail = n
	} else {
		l.head.prev = n
	}
	l.head = n
	l.length += 1
}

func (l *ListT[T]) PushBack(val T) {
	n := l.nodePool.get()
	n.value = val
	n.prev = l.tail
	if l.head == nil {
		l.head = n
	} else {
		l.tail.next = n
	}
	l.tail = n
	l.length += 1
}

func (l *ListT[T]) PopFront() (T, bool) {
	if l.length == 0 {
		var t T
		return t, false
	}
	n := l.head
	l.head = l.head.next
	if n.next != nil {
		n.next.prev = nil
	}
	if l.length == 1 {
		l.tail = nil
	}
	l.length -= 1
	value := n.value
	l.nodePool.put(n)
	return value, true
}

func (l *ListT[T]) PopBack() (T, bool) {
	var t T
	if l.length == 0 {
		return t, false
	}
	n := l.tail
	l.tail = l.tail.prev
	if l.tail != nil {
		l.tail.next = nil
	} else {
		l.head = nil
	}
	l.length -= 1
	value := n.value
	l.nodePool.put(n)
	return value, true
}

func (l *ListT[T]) InsertContinue(val T, after IteratorT[T]) IteratorT[T] {
	var iter IteratorT[T]
	n := l.insert(val, after)
	iter.n = n
	return iter
}

func (l *ListT[T]) Insert(val T, after IteratorT[T]) {
	l.insert(val, after)
}

func (l *ListT[T]) InsertBeforeContinue(val T, before IteratorT[T]) IteratorT[T] {
	var iter IteratorT[T]
	n := l.insertBefore(val, before)
	iter.n = n
	return iter
}

func (l *ListT[T]) InsertBefore(val T, before IteratorT[T]) {
	l.insertBefore(val, before)
}

func (l *ListT[T]) insertBefore(val T, before IteratorT[T]) *node_t[T] {
	after := before.Prev()
	return l.insert(val, after)
}

func (l *ListT[T]) insert(val T, after IteratorT[T]) *node_t[T] {
	n := l.nodePool.get()
	n.value = val
	if after.n == nil || after.n == (*node_t[T])(nullTNodePtr) {
		if l.head == nil {
			l.head = n
			l.tail = n
		} else {
			l.head.prev = n
			n.next = l.head
			l.head = n
		}
	} else {
		n.prev = after.n
		n.next = after.n.next
		if n.next != nil {
			n.next.prev = n
		} else {
			l.tail = n
		}
		n.prev.next = n
	}
	l.length += 1
	return n
}

func (l *ListT[T]) Update(val T, iter IteratorT[T]) bool {
	if iter.n == nil || unsafe.Pointer(iter.n) == nullTNodePtr {
		return false
	}
	iter.n.value = val
	return true
}

func (l *ListT[T]) DeleteContinueNext(iter IteratorT[T]) (IteratorT[T], bool) {
	if iter.n == nil || unsafe.Pointer(iter.n) == nullTNodePtr {
		return IteratorT[T]{}, false
	}
	nn := iter.n.next
	l.delete(iter)
	if nn == nil {
		nn = (*node_t[T])(nullTNodePtr)
	}
	return IteratorT[T]{n: nn}, true
}

func (l *ListT[T]) DeleteContinuePrev(iter IteratorT[T]) (IteratorT[T], bool) {
	if iter.n == nil || unsafe.Pointer(iter.n) == nullTNodePtr {
		return IteratorT[T]{}, false
	}
	np := iter.n.prev
	l.delete(iter)
	if np == nil {
		np = (*node_t[T])(unsafe.Pointer(nullTNodePtr))
	}
	return IteratorT[T]{n: np}, true
}

func (l *ListT[T]) Delete(iter IteratorT[T]) bool {
	if iter.n == nil || unsafe.Pointer(iter.n) == nullTNodePtr {
		return false
	}
	l.delete(iter)
	return true
}

func (l *ListT[T]) delete(iter IteratorT[T]) {
	prev := iter.n.prev
	next := iter.n.next
	if prev != nil {
		prev.next = next
	} else { // delete head
		l.head = next
	}
	if next != nil {
		next.prev = prev
	} else { // delete tail
		l.tail = prev
	}
	l.length -= 1
	if l.length == 0 {
		l.tail = nil
	}
	l.nodePool.put(iter.n)
}

func (l *ListT[T]) Clear() {
	n := l.head
	for n != nil {
		nn := n.next
		l.nodePool.put(n)
		n = nn
	}
	l.head = nil
	l.tail = nil
	l.length = 0
}

func (l *ListT[T]) Front() (T, bool) {
	if l.head == nil {
		var t T
		return t, false
	}
	return l.head.value, true
}

func (l *ListT[T]) Back() (T, bool) {
	if l.tail == nil {
		var t T
		return t, false
	}
	return l.tail.value, true
}

func (l *ListT[T]) Begin() IteratorT[T] {
	if l.head == nil {
		return l.End()
	}
	return IteratorT[T]{n: l.head}
}

func (l *ListT[T]) End() IteratorT[T] {
	return IteratorT[T]{n: (*node_t[T])(nullTNodePtr)}
}

func (l *ListT[T]) RBegin() IteratorT[T] {
	if l.tail == nil {
		return l.REnd()
	}
	return IteratorT[T]{n: l.tail}
}

func (l *ListT[T]) REnd() IteratorT[T] {
	return IteratorT[T]{n: (*node_t[T])(nullTNodePtr)}
}

func (l *ListT[T]) Duplicate() ListT[T] {
	nl := NewListTObj(l.nodePool)
	n := l.Begin()
	for n != l.End() {
		nl.PushBack(n.Value())
		n = n.Next()
	}
	return nl
}

func (l *ListT[T]) CopyTo(li *ListT[T]) {
	n := l.Begin()
	for n != li.End() {
		li.PushBack(n.Value())
		n = n.Next()
	}
}

func (l *ListT[T]) Merge(li ListT[T]) {
	n := li.Begin()
	for n != li.End() {
		l.PushBack(n.Value())
		n = n.Next()
	}
}
