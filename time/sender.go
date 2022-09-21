package time

import (
	"sync/atomic"
	"time"
)

type Sender struct {
	wheel *Wheel
	idx   int32
	ch    chan TimerList
}

func newSender(wheel *Wheel, idx int32) *Sender {
	sender := &Sender{
		wheel: wheel,
		idx:   idx,
	}
	sender.ch = make(chan TimerList, wheel.options.GetSenderListLength())
	wheel.senderMap[idx] = sender
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

func (s *Sender) C() <-chan TimerList {
	return s.ch
}

func (w *Wheel) NewSender() *Sender {
	count := atomic.AddInt32(&w.senderChanListCounter, 1)
	return newSender(w, count-1)
}
