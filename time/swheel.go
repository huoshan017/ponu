package time

import (
	"sync/atomic"
	"time"

	"github.com/huoshan017/ponu/list"
)

type resultExecutor struct {
}

func (e *resultExecutor) Send(index int32, tlist *list.ListT[*Timer]) {
	timer, o := tlist.PopFront()
	for o {
		timer.args = append(timer.args, timer.triggerTime)
		timer.fun(timer.id, timer.args)
		timer, o = tlist.PopFront()
	}
	putList(tlist)
}

type SWheel struct {
	*wheelBase
	options        Options
	resultExecutor iresultSender
}

func NewSWheel(timerMaxDuration time.Duration, options ...Option) *SWheel {
	var ops Options
	for _, option := range options {
		option(&ops)
	}
	if timerMaxDuration < ops.GetInterval() {
		return nil
	}
	w := &SWheel{}
	w.options = ops
	w.resultExecutor = &resultExecutor{}
	w.wheelBase = newWheelBase(timerMaxDuration, w.resultExecutor, &w.options)
	return w
}

func (w *SWheel) Start() {
	w.start()
}

func (w *SWheel) Update() bool {
	return w.handleTick()
}

func (w *SWheel) Add(timeout time.Duration, fun TimerFunc, args []any) uint32 {
	if timeout < w.options.GetInterval() || timeout > w.maxDuration {
		return 0
	}
	newId := atomic.AddUint32(&w.currId, 1)
	w.add(0, newId, timeout, fun, args)
	return newId
}

func (w *SWheel) Post(timeout time.Duration, fun TimerFunc, args []any) bool {
	if timeout < w.options.GetInterval() || timeout > w.maxDuration {
		return false
	}
	w.add(0, 0, timeout, fun, args)
	return true
}

func (w *SWheel) AddWithDeadline(deadline time.Time, fun TimerFunc, args []any) uint32 {
	duration := time.Until(deadline)
	return w.Add(duration, fun, args)
}

func (w *SWheel) PostWithDeadline(deadline time.Time, fun TimerFunc, args []any) bool {
	duration := time.Until(deadline)
	return w.Post(duration, fun, args)
}

func (w *SWheel) Cancel(id uint32) bool {
	return w.wheelBase.remove(id)
}

func (w *SWheel) add(index int32, id uint32, timeout time.Duration, fun TimerFunc, args []any) bool {
	t := getTimer()
	t.senderIndex = 0
	t.id = id
	t.timeout = timeout
	t.fun = fun
	t.args = args
	t.expireTime = time.Now().Add(timeout)
	if !w.addTimeout(t) {
		putTimer(t)
		return false
	}
	return true
}
