package skiplist

import (
	"math/rand"
	"time"
)

// MaxSkiplistLayer ...
const MaxSkiplistLayer = 32

type skiplistLayer struct {
	next *skiplistItem
	prev *skiplistItem
	span int32
}

// SkiplistNode ...
type SkiplistNode interface {
	FrontTo(node SkiplistNode) bool
	KeyEqualTo(node SkiplistNode) bool
}

type skiplistItem struct {
	value  SkiplistNode
	layers []*skiplistLayer
}

// Skiplist ...
type Skiplist struct {
	currLayer  int32
	currLength int32
	lengthsNum []int32         // 各层的节点数
	head       *skiplistItem   // 头节点
	tail       *skiplistItem   // 尾节点
	beforeNode []*skiplistItem // 缓存插入之前或删除之前的节点
	rank       []int32         // 缓存排名
	rand       *rand.Rand      // 随机数
}

func (s *Skiplist) randomSkiplistLayer() int32 {
	n := int32(1)
	r := s.rand.Int31()
	//for (rand.Int31()&0xFFFF)%4 == 0 {
	for r%2 == 0 {
		n++
		r /= 2
	}
	if n > MaxSkiplistLayer {
		n = MaxSkiplistLayer
	}
	return n
}

func newSkiplistNode(layer int32, v SkiplistNode) *skiplistItem {
	spLayers := make([]*skiplistLayer, layer)
	for i := int32(0); i < layer; i++ {
		spLayers[i] = &skiplistLayer{}
	}
	return &skiplistItem{
		value:  v,
		layers: spLayers,
	}
}

// NewSkiplist ...
func NewSkiplist() *Skiplist {
	return &Skiplist{
		currLayer:  int32(1),
		lengthsNum: make([]int32, MaxSkiplistLayer),
		head:       newSkiplistNode(MaxSkiplistLayer, nil),
		beforeNode: make([]*skiplistItem, MaxSkiplistLayer),
		rank:       make([]int32, MaxSkiplistLayer),
		rand:       rand.New(rand.NewSource(time.Now().Unix())),
	}
}

// Insert ...
func (s *Skiplist) Insert(v SkiplistNode) int32 {
	node := s.head
	for i := s.currLayer - 1; i >= 0; i-- {
		if i == s.currLayer-1 {
			s.rank[i] = 0
		} else {
			s.rank[i] = s.rank[i+1]
		}
		for node.layers[i].next != nil && node.layers[i].next.value.FrontTo(v) {
			s.rank[i] += node.layers[i].span
			node = node.layers[i].next
		}
		s.beforeNode[i] = node
	}

	newLayer := s.randomSkiplistLayer()
	if newLayer > s.currLayer {
		for i := s.currLayer; i < newLayer; i++ {
			s.rank[i] = 0
			s.beforeNode[i] = s.head
			s.beforeNode[i].layers[i].span = s.currLength
		}
		s.currLayer = newLayer
	}

	newNode := newSkiplistNode(newLayer, v)
	for i := int32(0); i < newLayer; i++ {
		node = s.beforeNode[i]
		newNode.layers[i].next = node.layers[i].next
		newNode.layers[i].prev = node
		if node.layers[i].next != nil {
			node.layers[i].next.layers[i].prev = newNode
		}
		node.layers[i].next = newNode

		newNode.layers[i].span = s.beforeNode[i].layers[i].span - (s.rank[0] - s.rank[i])
		s.beforeNode[i].layers[i].span = (s.rank[0] - s.rank[i]) + 1
	}

	for i := newLayer; i < s.currLayer; i++ {
		s.beforeNode[i].layers[i].span++
	}

	if newNode.layers[0].next == nil {
		s.tail = newNode
	}

	s.lengthsNum[newLayer-1]++
	s.currLength++

	return newLayer
}

// GetNode ...
func (s *Skiplist) getNode(v SkiplistNode) (node *skiplistItem) {
	n := s.head
	for i := s.currLayer - 1; i >= 0; i-- {
		for n.layers[i].next != nil && n.layers[i].next.value.FrontTo(v) {
			n = n.layers[i].next
		}
		s.beforeNode[i] = n
	}

	if n.layers[0].next != nil && n.layers[0].next.value.KeyEqualTo(v) {
		node = n.layers[0].next
	}
	return node
}

// GetNodeByRank ...
func (s *Skiplist) getNodeByRank(rank int32) *skiplistItem {
	var node *skiplistItem
	n := s.head
	currRank := int32(0)
	for i := s.currLayer - 1; i >= 0; i-- {
		for n.layers[i].next != nil && (currRank+n.layers[i].span) <= rank {
			currRank += n.layers[i].span
			n = n.layers[i].next
		}
		if currRank == rank {
			node = n
			break
		}
	}
	return node
}

// GetByRank ...
func (s *Skiplist) GetByRank(rank int32) (v SkiplistNode) {
	node := s.getNodeByRank(rank)
	if node == nil {
		return nil
	}
	return node.value
}

// GetRank ...
func (s *Skiplist) GetRank(v SkiplistNode) (rank int32) {
	node := s.head
	for i := s.currLayer - 1; i >= 0; i-- {
		for node.layers[i].next != nil && node.layers[i].next.value.FrontTo(v) {
			rank += node.layers[i].span
			node = node.layers[i].next
		}
		if node.layers[i].next != nil && node.layers[i].next.value.KeyEqualTo(v) {
			rank += node.layers[i].span
			return
		}
	}
	return 0
}

// GetByRankRange ...
func (s *Skiplist) GetByRankRange(rankStart, rankNum int32, values []SkiplistNode) bool {
	node := s.getNodeByRank(rankStart)
	if node == nil || rankNum <= 0 || values == nil {
		return false
	}

	if len(values) < int(rankNum) {
		return false
	}

	values[0] = node.value
	for i := int32(1); i < rankNum; i++ {
		if node.layers[0].next == nil {
			break
		}
		values[i] = node.layers[0].next.value
		node = node.layers[0].next
	}
	return true
}

// DeleteNode ...
func (s *Skiplist) DeleteNode(node *skiplistItem) {
	for n := int32(0); n < s.currLayer; n++ {
		if len(node.layers) > int(n) {
			if node.layers[n].prev != nil {
				node.layers[n].prev.layers[n].next = node.layers[n].next
				node.layers[n].prev.layers[n].span += (node.layers[n].span - 1)
			}
			if node.layers[n].next != nil {
				node.layers[n].next.layers[n].prev = node.layers[n].prev
			}
		} else {
			s.beforeNode[n].layers[n].span--
		}
	}

	if s.tail == node && node != nil {
		s.tail = node.layers[0].prev
	}

	// 更新当前最大层数
	if s.currLayer > 1 && s.head.layers[s.currLayer-1].next == nil {
		s.currLayer--
	}

	if s.lengthsNum[len(node.layers)-1] > 0 {
		s.lengthsNum[len(node.layers)-1]--
	}
	if s.currLength > 0 {
		s.currLength--
	}
}

// Delete ...
func (s *Skiplist) Delete(v SkiplistNode) bool {
	if s.currLength == 0 {
		return false
	}

	node := s.getNode(v)
	if node == nil {
		return false
	}

	s.DeleteNode(node)

	return true
}

// DeleteByRank ...
func (s *Skiplist) DeleteByRank(rank int32) bool {
	if s.currLength == 0 {
		return false
	}
	node := s.getNodeByRank(rank)
	if node == nil {
		return false
	}

	s.DeleteNode(node)
	return true
}

// DeleteTail ...
func (s *Skiplist) DeleteTail() (SkiplistNode, bool) {
	var node SkiplistNode
	if s.tail == nil {
		node = nil
		return node, false
	}

	node = s.tail.value

	for i := int(0); i < len(s.tail.layers); i++ {
		n := s.tail.layers[i].prev
		if n != nil {
			n.layers[i].next = nil
		}
	}

	if s.currLayer > 1 && s.head.layers[s.currLayer-1].next == nil {
		s.currLayer--
	}

	if s.lengthsNum[len(s.tail.layers)-1] > 0 {
		s.lengthsNum[len(s.tail.layers)-1]--
	}

	if s.currLength > 0 {
		s.currLength--
	}

	s.tail = s.tail.layers[0].prev
	return node, true
}

// GetFirst ...
func (s *Skiplist) GetFirst() (SkiplistNode, bool) {
	if s.currLength == 0 {
		return nil, false
	}

	node := s.head.layers[0].next.value
	if node == nil {
		return nil, false
	}

	return node, true
}

// GetTail ...
func (s *Skiplist) GetTail() (SkiplistNode, bool) {
	if s.tail == nil {
		return nil, false
	}

	node := s.tail.value
	if node == nil {
		return nil, false
	}

	return node, true
}

// PullList ...
func (s *Skiplist) PullList() (nodes []SkiplistNode) {
	node := s.head
	for node.layers[0].next != nil {
		nodes = append(nodes, node.layers[0].next.value)
		node = node.layers[0].next
	}
	return
}

// GetLength ...
func (s *Skiplist) GetLength() int32 {
	return s.currLength
}

// GetLayer ...
func (s *Skiplist) GetLayer() int32 {
	return s.currLayer
}

// GetLayerLength ...
func (s *Skiplist) GetLayerLength(layer int32) int32 {
	if layer < 1 || layer > s.currLayer {
		return -1
	}
	return s.lengthsNum[layer-1]
}
