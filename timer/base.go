package timer

import (
	"time"
)

type TimerFunc func(args ...interface{})

type Timer struct {
	fun       TimerFunc
	args      []interface{}
	slot      int
	layer     int
	left_time time.Duration
}
