package rankinglist

import (
	"reflect"

	"github.com/huoshan017/ponu/skiplist"
)

// RankItem ...
type RankItem interface {
	skiplist.SkiplistNode
	GetValue() interface{}
	SetValue(interface{})
}

// RankingList ...
type RankingList struct {
	_list      *skiplist.Skiplist
	_key2Value map[interface{}]RankItem
	_maxLength int
	_type      reflect.Type
}

// NewRankingList ...
func NewRankingList(maxLength int, typ reflect.Type) *RankingList {
	return &RankingList{
		_list:      skiplist.NewSkiplist(),
		_key2Value: make(map[interface{}]RankItem, maxLength),
		_maxLength: maxLength,
		_type:      typ,
	}
}

func (r *RankingList) insert(key interface{}, value interface{}) bool {
	// RankItem.SetValue function
	v := reflect.New(r._type)
	m := v.Elem().MethodByName("SetValue")
	m.Call([]reflect.Value{reflect.ValueOf(value)})
	// convert to RankItem
	var rankItem RankItem = (v.Elem().Interface()).(RankItem)
	if rankItem == nil {
		panic("RankingList.insert: reflect value %v cant be convert to RankItem interface type")
	}
	r._list.Insert(rankItem)
	r._key2Value[key] = rankItem
	return true
}

// Insert ...
func (r *RankingList) Insert(key interface{}, value interface{}) bool {
	_, o := r._key2Value[key]
	if !o {
		return false
	}
	return r.insert(key, value)
}

func (r *RankingList) delete(key interface{}, item RankItem) bool {
	delete(r._key2Value, key)
	return r._list.Delete(item)
}

// Delete ...
func (r *RankingList) Delete(key interface{}) bool {
	item, o := r._key2Value[key]
	if !o {
		return false
	}
	return r.delete(key, item)
}

func (r *RankingList) update(key interface{}, value interface{}, item RankItem) bool {
	if !r.delete(key, item) {
		return false
	}
	return r.insert(key, value)
}

// Update ...
func (r *RankingList) Update(key interface{}, value interface{}) bool {
	item, o := r._key2Value[key]
	if !o {
		return false
	}
	return r.update(key, value, item)
}

// InsertOrUpdate ...
func (r *RankingList) InsertOrUpdate(key interface{}, value interface{}) bool {
	var res bool
	item, o := r._key2Value[key]
	if !o {
		res = r.insert(key, value)
	} else {
		res = r.update(key, value, item)
	}
	return res
}

// GetValue ...
func (r *RankingList) GetValue(key interface{}) (interface{}, bool) {
	n, o := r._key2Value[key]
	if !o {
		return nil, false
	}
	v := n.GetValue()
	return v, true
}

// GetValueAndRank ...
func (r *RankingList) GetValueAndRank(key interface{}) (interface{}, int32, bool) {
	n, o := r._key2Value[key]
	if !o {
		return nil, 0, false
	}
	value := n.GetValue()
	rank := r._list.GetRank(n)
	return value, rank, true
}

// GetValueByRank ...
func (r *RankingList) GetValueByRank(rank int32) (interface{}, bool) {
	node := r._list.GetByRank(rank)
	if node == nil {
		return nil, false
	}
	item := node.(RankItem)
	if item == nil {
		panic("RankingList.GetValueByRank: the node cant convert to RankItem interface type")
	}
	return item.GetValue(), true
}

// HasKey ...
func (r *RankingList) HasKey(key interface{}) bool {
	_, o := r._key2Value[key]
	return o
}
