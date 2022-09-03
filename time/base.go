package time

import "time"

type TimerFunc func(arg []any)

type Timer struct {
	id         uint32
	fun        TimerFunc
	arg        []any
	timeout    time.Duration
	expireTime time.Time
	leftStep   int32
}

func (t *Timer) Clean() {
	t.id = 0
	t.arg = nil
	t.fun = nil
	t.leftStep = 0
}
