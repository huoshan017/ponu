package list

import (
	"math/rand"
	"testing"
	"time"
)

var testNodeTPool = NewListTNodePool[int32]()

func TestListT(t *testing.T) {
	var (
		r       = rand.New(rand.NewSource(time.Now().UnixNano()))
		l       = NewListTWithPool(testNodeTPool)
		length  = 10000
		valList []int32
		v       int32
		val     any
		o       bool
	)

	for i := 0; i < length; i++ {
		v = r.Int31n(1000)
		l.PushBack(v)
		valList = append(valList, v)
	}

	for i := 0; i < length; i++ {
		val, o = l.PopFront()
		if !o {
			t.Errorf("pop front failed")
		}
		v = val.(int32)
		if v != valList[i] {
			t.Errorf("compare %v != %v", v, valList[i])
			return
		}
	}
}

func TestIteratorT(t *testing.T) {
	var (
		length  int32 = 1000
		l             = NewListTWithPool(testNodeTPool)
		r             = rand.New(rand.NewSource(time.Now().UnixNano()))
		valList []int32
		o       bool
	)

	for i := int32(0); i < length; i++ {
		v := r.Int31n(length) + 1
		l.PushBack(v)
		valList = append(valList, v)
	}

	c := len(valList) - 1
	iter := l.RBegin()
	for iter != l.REnd() {
		if iter.Value() != valList[c] {
			t.Errorf("iter value %v != list value %v, index %v", iter.Value(), valList[c], c)
			return
		}
		if iter, o = l.DeleteContinuePrev(iter); !o {
			t.Errorf("delete %v failed", iter.Value())
			return
		}
		c -= 1
	}
}

func TestInsertT(t *testing.T) {
	var (
		length  int32 = 1000000
		l             = NewListTWithPool(testNodeTPool)
		r             = rand.New(rand.NewSource(time.Now().UnixNano()))
		valList []int32
	)

	iter := l.Begin()
	for i := int32(0); i < length; i++ {
		v := r.Int31n(10000) + 1
		iter = l.InsertContinue(v, iter)
		valList = append(valList, v)
	}

	iter = l.Begin()
	for n := 0; iter != l.End(); {
		if iter.Value() != valList[n] {
			t.Errorf("compare list value %v to array value %v failed", iter.Value(), valList[n])
			return
		}
		n += 1
		iter = iter.Next()
	}
}

func TestDeleteT(t *testing.T) {
	var (
		length  int32 = 20000
		l             = NewListTWithPool(testNodeTPool)
		r             = rand.New(rand.NewSource(time.Now().UnixNano()))
		valList []int32
		iter    IteratorT[int32]
		o       bool
	)
	for i := int32(0); i < length; i++ {
		v := r.Int31n(100000) + 1
		l.PushBack(v)
		valList = append(valList, v)
	}

	c := len(valList) - 1
	iter = l.RBegin()
	for iter != l.REnd() {
		delVal := iter.Value()
		if delVal != valList[c] {
			t.Errorf("to delete value %v is not right value %v", delVal, valList[c])
		}
		iter, o = l.DeleteContinuePrev(iter)
		if !o {
			t.Errorf("delete iter point value %v", delVal)
			return
		}
		c -= 1
	}
}

func testInsertDeleteT(t *testing.T, rt bool) {
	var (
		length, count int32 = 10000, 10000
		l                   = NewListTWithPool(testNodeTPool)
		r                   = rand.New(rand.NewSource(time.Now().UnixNano()))
		iter          IteratorT[int32]
		o             bool
	)
	for i := int32(0); i < length; i++ {
		l.PushBack(i + 1)
	}

	for c := int32(0); c < count; c++ {
		insertVal := r.Int31n(length) + 1
		t.Logf("to insert value %v", insertVal)
		iter = l.Begin()
		for iter != l.End() {
			if iter.Value() == insertVal {
				t.Logf("insert value %v", insertVal)
				iter = l.InsertContinue(insertVal, iter)
				iter = iter.Next()
			} else {
				iter = iter.Next()
			}
		}

		isDelete := false
		if !rt {
			for iter = l.Begin(); iter != l.End(); {
				if iter.Value() == insertVal {
					iter, o = l.DeleteContinueNext(iter)
					if !o {
						t.Errorf("delete value %v failed", insertVal)
						return
					}
					isDelete = true
					t.Logf("deleted value %v", insertVal)
				}
				iter = iter.Next()
			}
		} else {
			for iter = l.RBegin(); iter != l.REnd(); {
				if iter.Value() == insertVal {
					iter, o = l.DeleteContinuePrev(iter)
					if !o {
						t.Errorf("delete value %v failed", insertVal)
						return
					}
					isDelete = true
					t.Logf("deleted value %v", insertVal)
				}
				iter = iter.Prev()
			}
		}

		if !isDelete {
			t.Errorf("not to delete value %v", insertVal)
			return
		}

		t.Logf("processed %v", c+1)
	}
}

func TestInsertDeleteT(t *testing.T) {
	testInsertDeleteT(t, false)
}

func TestInsertDeleteRevertT(t *testing.T) {
	testInsertDeleteT(t, true)
}
