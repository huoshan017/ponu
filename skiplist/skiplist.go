package utils

import (
	"math/rand"
	//"time"
)

const MAX_SKIPLIST_LAYER = 32

type skiplist_layer struct {
	next *skiplist_node
	prev *skiplist_node
	span int32
}

type SkiplistNode interface {
	Less(node interface{}) bool
	Greater(node interface{}) bool
	KeyEqual(key interface{}) bool
	GetKey() interface{}
	GetValue() interface{}
	SetValue(interface{})
	New() SkiplistNode
	Assign(node SkiplistNode)
	CopyDataTo(node interface{})
}

type skiplist_node struct {
	value  SkiplistNode
	layers []*skiplist_layer
}

type Skiplist struct {
	curr_layer  int32
	curr_length int32
	lengths_num []int32          // 各层的节点数
	head        *skiplist_node   // 头节点
	tail        *skiplist_node   // 尾节点
	before_node []*skiplist_node // 缓存插入之前或删除之前的节点
	rank        []int32          // 缓存排名
}

func random_skiplist_layer() int32 {
	n := int32(1)
	for (rand.Int31()&0xFFFF)%4 == 0 {
		n += 1
	}
	if n > MAX_SKIPLIST_LAYER {
		n = MAX_SKIPLIST_LAYER
	}
	return n
}

func new_skiplist_node(layer int32, v SkiplistNode) *skiplist_node {
	sp_layers := make([]*skiplist_layer, layer)
	for i := int32(0); i < layer; i++ {
		sp_layers[i] = &skiplist_layer{}
	}
	return &skiplist_node{
		value:  v,
		layers: sp_layers,
	}
}

func NewSkiplist() *Skiplist {
	return &Skiplist{
		curr_layer:  int32(1),
		lengths_num: make([]int32, MAX_SKIPLIST_LAYER),
		head:        new_skiplist_node(MAX_SKIPLIST_LAYER, nil),
		before_node: make([]*skiplist_node, MAX_SKIPLIST_LAYER),
		rank:        make([]int32, MAX_SKIPLIST_LAYER),
	}
}

func (this *Skiplist) Insert(v SkiplistNode) int32 {
	if this.curr_length == 0 {
		//log.Debug("###[Skiplist]### first node[%v]", v)
	}

	node := this.head
	for i := this.curr_layer - 1; i >= 0; i-- {
		if i == this.curr_layer-1 {
			this.rank[i] = 0
		} else {
			this.rank[i] = this.rank[i+1]
		}
		for node.layers[i].next != nil && node.layers[i].next.value.Greater(v) {
			this.rank[i] += node.layers[i].span
			node = node.layers[i].next
		}
		this.before_node[i] = node
	}

	new_layer := random_skiplist_layer()
	if new_layer > this.curr_layer {
		for i := this.curr_layer; i < new_layer; i++ {
			this.rank[i] = 0
			this.before_node[i] = this.head
			this.before_node[i].layers[i].span = this.curr_length
		}
		this.curr_layer = new_layer
	}

	new_node := new_skiplist_node(new_layer, v)
	for i := int32(0); i < new_layer; i++ {
		node = this.before_node[i]
		new_node.layers[i].next = node.layers[i].next
		new_node.layers[i].prev = node
		if node.layers[i].next != nil {
			node.layers[i].next.layers[i].prev = new_node
		}
		node.layers[i].next = new_node

		new_node.layers[i].span = this.before_node[i].layers[i].span - (this.rank[0] - this.rank[i])
		this.before_node[i].layers[i].span = (this.rank[0] - this.rank[i]) + 1
	}

	for i := new_layer; i < this.curr_layer; i++ {
		this.before_node[i].layers[i].span += 1
	}

	if new_node.layers[0].next == nil {
		this.tail = new_node
	}

	this.lengths_num[new_layer-1] += 1
	this.curr_length += 1

	return new_layer
}

func (this *Skiplist) GetNode(v SkiplistNode) (node *skiplist_node) {
	n := this.head
	for i := this.curr_layer - 1; i >= 0; i-- {
		for n.layers[i].next != nil && n.layers[i].next.value.Greater(v) {
			n = n.layers[i].next
		}
		this.before_node[i] = n
	}
	if n.layers[0].next != nil && n.layers[0].next.value.KeyEqual(v) {
		node = n.layers[0].next
	}
	return
}

func (this *Skiplist) GetNodeByRank(rank int32) (node *skiplist_node) {
	n := this.head
	curr_rank := int32(0)
	for i := this.curr_layer - 1; i >= 0; i-- {
		for n.layers[i].next != nil && (curr_rank+n.layers[i].span) <= rank {
			curr_rank += n.layers[i].span
			n = n.layers[i].next
		}
		if curr_rank == rank {
			node = n
			break
		}
	}
	return
}

func (this *Skiplist) GetByRank(rank int32) (v SkiplistNode) {
	node := this.GetNodeByRank(rank)
	if node == nil {
		return nil
	}
	return node.value
}

func (this *Skiplist) GetRank(v SkiplistNode) (rank int32) {
	node := this.head
	for i := this.curr_layer - 1; i >= 0; i-- {
		for node.layers[i].next != nil && node.layers[i].next.value.Greater(v) {
			rank += node.layers[i].span
			node = node.layers[i].next
		}
		if node.layers[i].next != nil && node.layers[i].next.value.KeyEqual(v) {
			rank += node.layers[i].span
			return
		}
	}
	return 0
}

func (this *Skiplist) GetByRankRange(rank_start, rank_num int32, values []SkiplistNode) bool {
	node := this.GetNodeByRank(rank_start)
	if node == nil || rank_num <= 0 || values == nil {
		return false
	}

	if len(values) < int(rank_num) {
		return false
	}

	values[0] = node.value
	for i := int32(1); i < rank_num; i++ {
		if node.layers[0].next == nil {
			break
		}
		values[i] = node.layers[0].next.value
		node = node.layers[0].next
	}
	return true
}

func (this *Skiplist) DeleteNode(node *skiplist_node) {
	for n := int32(0); n < this.curr_layer; n++ {
		if len(node.layers) > int(n) {
			if node.layers[n].prev != nil {
				node.layers[n].prev.layers[n].next = node.layers[n].next
				node.layers[n].prev.layers[n].span += (node.layers[n].span - 1)
			}
			if node.layers[n].next != nil {
				node.layers[n].next.layers[n].prev = node.layers[n].prev
			}
		} else {
			this.before_node[n].layers[n].span -= 1
		}
	}

	if this.tail == node && node != nil {
		this.tail = node.layers[0].prev
	}

	// 更新当前最大层数
	if this.curr_layer > 1 && this.head.layers[this.curr_layer-1].next == nil {
		this.curr_layer -= 1
	}

	if this.lengths_num[len(node.layers)-1] > 0 {
		this.lengths_num[len(node.layers)-1] -= 1
	}
	if this.curr_length > 0 {
		this.curr_length -= 1
	}
}

func (this *Skiplist) Delete(v SkiplistNode) bool {
	if this.curr_length == 0 {
		return false
	}

	node := this.GetNode(v)
	if node == nil {
		//log.Error("###[Skiplist]### get node %v failed", v)
		return false
	}

	this.DeleteNode(node)

	return true
}

func (this *Skiplist) DeleteByRank(rank int32) bool {
	if this.curr_length == 0 {
		return false
	}
	node := this.GetNodeByRank(rank)
	if node == nil {
		//log.Error("###[Skiplist]### get node by rank[%v] failed", rank)
		return false
	}

	this.DeleteNode(node)
	return true
}

func (this *Skiplist) PullList() (nodes []SkiplistNode) {
	node := this.head
	for node.layers[0].next != nil {
		nodes = append(nodes, node.layers[0].next.value)
		node = node.layers[0].next
	}
	return
}

func (this *Skiplist) GetLength() int32 {
	return this.curr_length
}

func (this *Skiplist) GetLayer() int32 {
	return this.curr_layer
}

func (this *Skiplist) GetLayerLength(layer int32) int32 {
	if layer < 1 || layer > this.curr_layer {
		return -1
	}
	return this.lengths_num[layer-1]
}

type Int32Value int32

func (this Int32Value) Less(id interface{}) bool {
	if this < id.(Int32Value) {
		return true
	}
	return false
}

func (this Int32Value) Greater(id interface{}) bool {
	if this > id.(Int32Value) {
		return true
	}
	return false
}

func (this Int32Value) KeyEqual(id interface{}) bool {
	if this == id {
		return true
	}
	return false
}

func (this Int32Value) GetKey() interface{} {
	return this
}

func (this Int32Value) GetValue() interface{} {
	return this
}

func (this Int32Value) SetValue(value interface{}) {

}

func (this Int32Value) New() SkiplistNode {
	return this
}

func (this Int32Value) Assign(node SkiplistNode) {
}

func (this Int32Value) CopyDataTo(node interface{}) {

}

type PlayerInfo struct {
	PlayerId    int32
	PlayerLevel int32
	PlayerScore int32
}

func (this *PlayerInfo) Less(info interface{}) bool {
	item := info.(*PlayerInfo)
	if item == nil {
		return false
	}
	if this.PlayerScore < item.PlayerScore {
		return true
	}
	if this.PlayerScore == item.PlayerScore {
		if this.PlayerLevel < item.PlayerLevel {
			return true
		}
		if this.PlayerLevel == item.PlayerLevel {
			if this.PlayerId < item.PlayerId {
				return true
			}
		}
	}
	return false
}

func (this *PlayerInfo) Greater(info interface{}) bool {
	item := info.(*PlayerInfo)
	if item == nil {
		return false
	}
	if this.PlayerScore > item.PlayerScore {
		return true
	}
	if this.PlayerScore == item.PlayerScore {
		if this.PlayerLevel > item.PlayerLevel {
			return true
		}
		if this.PlayerLevel == item.PlayerLevel {
			if this.PlayerId > item.PlayerId {
				return true
			}
		}
	}
	return false
}

func (this *PlayerInfo) KeyEqual(info interface{}) bool {
	item := info.(*PlayerInfo)
	if item == nil {
		return false
	}
	if this.PlayerId == item.PlayerId {
		return true
	}
	return false
}

func (this *PlayerInfo) GetKey() interface{} {
	return this.PlayerId
}

func (this *PlayerInfo) GetValue() interface{} {
	return this.PlayerId
}

func (this *PlayerInfo) SetValue(value interface{}) {

}

func (this *PlayerInfo) New() SkiplistNode {
	return &PlayerInfo{}
}

func (this *PlayerInfo) Assign(node SkiplistNode) {
	n := node.(*PlayerInfo)
	if n == nil {
		return
	}
	this.PlayerId = n.PlayerId
	this.PlayerLevel = n.PlayerLevel
	this.PlayerScore = n.PlayerScore
}

func (this *PlayerInfo) CopyDataTo(node interface{}) {

}
