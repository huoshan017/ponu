package list

import (
	"sync"
)

type node struct {
	value      any
	prev, next *node
}

var (
	nullNode = node{}
	nodePool *sync.Pool
)

func init() {
	nodePool = &sync.Pool{
		New: func() any {
			return &node{}
		},
	}
}

func getNode() *node {
	return nodePool.Get().(*node)
}

func putNode(n *node) {
	n.prev = nil
	n.next = nil
	n.value = nil
	nodePool.Put(n)
}

type Iterator struct {
	n *node
}

func (iter *Iterator) Value() any {
	return iter.n.value
}

func (iter *Iterator) Next() Iterator {
	if iter.n.next == nil {
		return Iterator{n: &nullNode}
	}
	return Iterator{n: iter.n.next}
}

func (iter *Iterator) Prev() Iterator {
	if iter.n.prev == nil {
		return Iterator{n: &nullNode}
	}
	return Iterator{n: iter.n.prev}
}

func (iter *Iterator) IsValid() bool {
	return iter.n != nil && iter.n != &nullNode
}

func (iter *Iterator) IsHead(l *List) bool {
	return iter.n == l.head
}

func (iter *Iterator) IsTail(l *List) bool {
	return iter.n == l.tail
}

func (iter *Iterator) IsEqualTo(i Iterator) bool {
	return *(iter.n) == *(i.n)
}

type List struct {
	head, tail *node
	length     int32
}

func New() *List {
	return &List{}
}

func (l List) GetLength() int32 {
	return l.length
}

func (l List) IsEmpty() bool {
	return l.length == 0
}

func (l *List) PushFront(val any) {
	n := getNode()
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

func (l *List) PushBack(val any) {
	n := getNode()
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

func (l *List) PopFront() (any, bool) {
	if l.length == 0 {
		return nil, false
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
	putNode(n)
	return value, true
}

func (l *List) PopBack() (any, bool) {
	if l.length == 0 {
		return nil, false
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
	putNode(n)
	return value, true
}

func (l *List) InsertContinue(val any, after Iterator) Iterator {
	var iter Iterator
	n := l.insert(val, after)
	iter.n = n
	return iter
}

func (l *List) Insert(val any, after Iterator) {
	l.insert(val, after)
}

func (l *List) InsertBeforeContinue(val any, before Iterator) Iterator {
	var iter Iterator
	n := l.insertBefore(val, before)
	iter.n = n
	return iter
}

func (l *List) InsertBefore(val any, before Iterator) {
	l.insertBefore(val, before)
}

func (l *List) insertBefore(val any, before Iterator) *node {
	after := before.Prev()
	return l.insert(val, after)
}

func (l *List) insert(val any, after Iterator) *node {
	n := getNode()
	n.value = val
	if after.n == nil || after.n == &nullNode {
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

func (l *List) Update(val any, iter Iterator) bool {
	if iter.n == nil || iter.n == &nullNode {
		return false
	}
	iter.n.value = val
	return true
}

func (l *List) DeleteContinueNext(iter Iterator) (Iterator, bool) {
	if iter.n == nil || iter.n == &nullNode {
		return Iterator{}, false
	}
	nn := iter.n.next
	l.delete(iter)
	if nn == nil {
		nn = &nullNode
	}
	return Iterator{n: nn}, true
}

func (l *List) DeleteContinuePrev(iter Iterator) (Iterator, bool) {
	if iter.n == nil || iter.n == &nullNode {
		return Iterator{}, false
	}
	np := iter.n.prev
	l.delete(iter)
	if np == nil {
		np = &nullNode
	}
	return Iterator{n: np}, true
}

func (l *List) Delete(iter Iterator) bool {
	if iter.n == nil || iter.n == &nullNode {
		return false
	}
	l.delete(iter)
	return true
}

func (l *List) delete(iter Iterator) {
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
	putNode(iter.n)
}

func (l *List) Clear() {
	n := l.head
	for n != nil {
		nn := n.next
		putNode(n)
		n = nn
	}
	l.head = nil
	l.tail = nil
	l.length = 0
}

func (l *List) Front() any {
	if l.head == nil {
		return nil
	}
	return l.head.value
}

func (l *List) Back() any {
	if l.tail == nil {
		return nil
	}
	return l.tail.value
}

func (l *List) Begin() Iterator {
	if l.head == nil {
		return l.End()
	}
	return Iterator{n: l.head}
}

func (l *List) End() Iterator {
	return Iterator{n: &nullNode}
}

func (l *List) RBegin() Iterator {
	if l.tail == nil {
		return l.REnd()
	}
	return Iterator{n: l.tail}
}

func (l *List) REnd() Iterator {
	return Iterator{n: &nullNode}
}

func (l *List) Duplicate() List {
	var nl List
	n := l.Begin()
	for n != l.End() {
		nl.PushBack(n.Value())
		n = n.Next()
	}
	return nl
}

func (l *List) CopyTo(li *List) {
	n := l.Begin()
	for n != li.End() {
		li.PushBack(n.Value())
		n = n.Next()
	}
}

func (l *List) Merge(li List) {
	n := li.Begin()
	for n != li.End() {
		l.PushBack(n.Value())
		n = n.Next()
	}
}
