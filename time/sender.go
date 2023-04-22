package time

import (
	"sync/atomic"
	"time"

	"github.com/huoshan017/ponu/lockfree"
)

type Sender struct {
	wheel *Wheel
	idx   int32
	tlist *lockfree.QueueT[TimerList]
}

func newSender(wheel *Wheel, idx int32) *Sender {
	sender := &Sender{
		wheel: wheel,
		idx:   idx,
		tlist: lockfree.NewQueueT[TimerList](),
	}
	wheel.resultSenderCh <- sender
	return sender
}

func (s *Sender) Add(timeout time.Duration, fun TimerFunc, args []any) uint32 {
	if timeout < s.wheel.options.GetInterval() || timeout > s.wheel.maxDuration {
		return 0
	}
	newId := atomic.AddUint32(&s.wheel.currId, 1)
	s.wheel.add(s.idx, newId, timeout, fun, args)
	return newId
}

func (s *Sender) Post(timeout time.Duration, fun TimerFunc, args []any) bool {
	if timeout < s.wheel.options.GetInterval() || timeout > s.wheel.maxDuration {
		return false
	}
	s.wheel.add(s.idx, 0, timeout, fun, args)
	return true
}

func (s *Sender) Cancel(timerId uint32) {
	s.wheel.Cancel(timerId)
}

func (s *Sender) GetTimerList() (TimerList, bool) {
	return s.tlist.Dequeue()
}

func (w *Wheel) NewSender() *Sender {
	count := atomic.AddInt32(&w.senderChanListCounter, 1)
	return newSender(w, count-1)
}
