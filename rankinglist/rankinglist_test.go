package rankinglist

import (
	"reflect"
	"testing"

	"github.com/huoshan017/ponu/skiplist"
)

type TestType struct {
	key      int32
	level    int32
	serialID int64
}

func (t *TestType) FrontTo(node skiplist.SkiplistNode) bool {
	tt := node.(*TestType)
	if t.level < tt.level {
		return false
	}
	if t.level == tt.level {
		if t.serialID >= tt.serialID {
			return false
		}
	}
	return true
}

func (t *TestType) EqualTo(node skiplist.SkiplistNode) bool {
	tt := node.(*TestType)
	if t.key != tt.key {
		return false
	}
	return true
}

func (t *TestType) GetValue() interface{} {
	return t.level
}

func (t *TestType) SetValue(value interface{}) {
	t.level = value.(int32)
	t.serialID++
}

func Test_one(t *testing.T) {
	rankingList := NewRankingList(10000, reflect.TypeOf(TestType{}))
	for i := 0; i < 100; i++ {
		rankingList.Insert((i+1)*100, i+1)
	}
	for key := 100; key <= 1000; key += 100 {
		value, o := rankingList.GetValue(key)
		if o {
			t.Logf("key is %v, value is %v", key, value)
		}
	}
}
