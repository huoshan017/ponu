package time

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/huoshan017/ponu/list"
	"github.com/pkg/errors"
)

type resultChanSender struct {
	w *Wheel
}

func (s *resultChanSender) Send(index int32, tlist *list.ListT[*Timer]) {
	sender := s.w.senderMap[index]
	if sender == nil {
		// 等待通道傳過來
		for sender == nil || sender.idx != index {
			sender = <-s.w.resultSenderCh
			s.w.senderMap[sender.idx] = sender
		}
	}
	sender.tlist.Enqueue(TimerList{l: tlist, m: &s.w.toDelIdMap})
}

type Wheel struct {
	wheelBase
	options               Options
	stepTicker            *time.Ticker
	addCh                 chan *Timer
	removeCh              chan uint32
	resultSenderCh        chan *Sender
	senderChanListCounter int32
	closeCh               chan struct{}
	closeOnce             sync.Once
	toDelIdMap            sync.Map
	resultSender          resultChanSender
	senderMap             map[int32]*Sender
}

func NewWheel(timerMaxDuration time.Duration, options ...Option) *Wheel {
	var ops Options
	for _, option := range options {
		option(&ops)
	}
	if timerMaxDuration < ops.GetInterval() {
		return nil
	}

	w := &Wheel{}
	w.options = ops
	w.wheelBase = *newWheelBase(timerMaxDuration, &w.resultSender, &w.options)
	w.addCh = make(chan *Timer, w.options.GetTimerRecvListLength())
	w.removeCh = make(chan uint32, w.options.GetRemoveListLength())
	w.resultSenderCh = make(chan *Sender)
	w.closeCh = make(chan struct{})
	w.resultSender = resultChanSender{w: w}
	w.senderMap = make(map[int32]*Sender)
	return w
}

func (w *Wheel) Run() {
	defer func() {
		if err := recover(); err != nil {
			er := errors.Errorf("%v", err)
			log.Fatalf("\n%+v", er)
		}
	}()

	w.stepTicker = time.NewTicker(w.options.GetInterval())
	<-w.stepTicker.C
	w.start()

	var loop bool = true
	for loop {
		select {
		case <-w.closeCh:
			atomic.StoreInt32(&w.senderChanListCounter, 1)
			loop = false
		case v, o := <-w.addCh:
			if o {
				w.addTimeout(v)
			}
		case id, o := <-w.removeCh:
			if o {
				w.remove(id)
			}
		case sender, o := <-w.resultSenderCh:
			if o {
				if w.senderMap[sender.idx] == nil {
					w.senderMap[sender.idx] = sender
				}
			}
		case <-w.stepTicker.C:
			w.handleTick()
		}
	}
	w.stepTicker.Stop()
	w.stepTicker = nil
}

func (w *Wheel) Stop() {
	w.closeOnce.Do(func() {
		close(w.closeCh)
	})
}

func (w *Wheel) Add(timeout time.Duration, fun TimerFunc, args []any) uint32 {
	if timeout < w.options.GetInterval() || timeout > w.maxDuration {
		return 0
	}
	newId := atomic.AddUint32(&w.currId, 1)
	w.add(0, newId, timeout, fun, args)
	return newId
}

func (w *Wheel) Post(timeout time.Duration, fun TimerFunc, args []any) bool {
	if timeout < w.options.GetInterval() || timeout > w.maxDuration {
		return false
	}
	w.add(0, 0, timeout, fun, args)
	return true
}

func (w *Wheel) AddWithDeadline(deadline time.Time, fun TimerFunc, args []any) uint32 {
	duration := time.Until(deadline)
	return w.Add(duration, fun, args)
}

func (w *Wheel) PostWithDeadline(deadline time.Time, fun TimerFunc, args []any) bool {
	duration := time.Until(deadline)
	return w.Post(duration, fun, args)
}

func (w *Wheel) Cancel(id uint32) {
	w.toDelIdMap.LoadOrStore(id, true)
	w.removeCh <- id
}

func (w *Wheel) add(idx int32, id uint32, timeout time.Duration, fun TimerFunc, args []any) {
	t := getTimer()
	t.senderIndex = idx
	t.id = id
	t.timeout = timeout
	t.fun = fun
	t.arg = args
	t.expireTime = time.Now().Add(timeout)
	w.addCh <- t
}

func (w *Wheel) remove(id uint32) {
	w.wheelBase.remove(id)
	w.toDelIdMap.Delete(id)
}

var (
	timerPool, listPool sync.Pool
	nodePool            = list.NewListTNodePool[*Timer]()
)

func init() {
	timerPool = sync.Pool{
		New: func() any {
			return &Timer{}
		},
	}
	listPool = sync.Pool{
		New: func() any {
			return list.NewListT(nodePool)
		},
	}
}

func getTimer() *Timer {
	return timerPool.Get().(*Timer)
}

func putTimer(t *Timer) {
	t.Clean()
	timerPool.Put(t)
}

func getList() *list.ListT[*Timer] {
	return listPool.Get().(*list.ListT[*Timer])
}

func putList(l *list.ListT[*Timer]) {
	l.Clear()
	listPool.Put(l)
}
