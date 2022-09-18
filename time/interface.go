package time

import (
	"time"

	"github.com/huoshan017/ponu/list"
)

type IResultSender interface {
	Send(int32, *list.List)
}

type IWheel interface {
	Add(timeout time.Duration, fun TimerFunc, args []any) uint32
	Post(timeout time.Duration, fun TimerFunc, args []any) bool
	AddWithDeadline(deadline time.Time, fun TimerFunc, args []any) uint32
	PostWithDeadline(deadline time.Time, fun TimerFunc, args []any) bool
	Cancel(id uint32)
	Run()
	Stop()
}
