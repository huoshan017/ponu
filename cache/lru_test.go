package cache

import "testing"

func TestLRU(t *testing.T) {
	l := NewLRU[int, int](10)
	kv := []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13, 14, 14, 15, 15, 16, 16, 17, 17, 18, 18, 19, 19, 20, 20, 21, 21, 22, 22, 23, 23, 24, 24, 25, 25}
	for i := 0; i < len(kv); i += 2 {
		l.Set(kv[i], kv[i+1])
	}
	l.Set(11, 11)
	l.Set(5, 500)
	_, o := l.Get(8)
	if !o {
		t.Logf("cant get value with key 8")
	}
	l.Set(4, 400)
	dl := l.ToList()

	t.Logf("list value is:")
	iter := dl.Begin()
	for iter != dl.End() {
		t.Logf(" %v", iter.Value())
		iter = iter.Next()
	}
}
