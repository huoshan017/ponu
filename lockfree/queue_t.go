package lockfree

import (
	"sync/atomic"
	"unsafe"
)

type node_t[T any] struct {
	value T
	next  unsafe.Pointer
}

type QueueT[T any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

func NewQueueT[T any]() *QueueT[T] {
	n := unsafe.Pointer(&node_t[T]{})
	return &QueueT[T]{head: n, tail: n}
}

func (q *QueueT[T]) Enqueue(v T) {
	n := &node_t[T]{value: v}
	for {
		tail := load_t[T](&q.tail)
		next := load_t[T](&tail.next)
		if tail == load_t[T](&q.tail) {
			if next == nil {
				if cas_t(&tail.next, next, n) {
					cas_t(&q.tail, tail, n)
					return
				}
			} else {
				cas_t(&q.tail, tail, next)
			}
		}
	}
}

func (q *QueueT[T]) Dequeue() (T, bool) {
	for {
		head := load_t[T](&q.head)
		tail := load_t[T](&q.tail)
		next := load_t[T](&head.next)
		if head == load_t[T](&q.head) {
			if head == tail {
				if next == nil {
					var t T
					return t, false
				}
				cas_t(&q.tail, tail, next)
			} else {
				v := next.value
				if cas_t(&q.head, head, next) {
					return v, true
				}
			}
		}
	}
}

func load_t[T any](p *unsafe.Pointer) *node_t[T] {
	return (*node_t[T])(atomic.LoadPointer(p))
}

func cas_t[T any](p *unsafe.Pointer, old, new *node_t[T]) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
