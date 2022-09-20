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

func (s *resultChanSender) Send(index int32, tlist *list.List) {
	if s.w.senderChanList[index] == nil {
		s.w.senderChanList[index] = make(chan TimerList, s.w.options.GetSenderListLength())
	}
	s.w.senderChanList[index] <- TimerList{l: tlist, m: &s.w.toDelIdMap}
}

type Wheel struct {
	wheelBase
	options               Options
	stepTicker            *time.Ticker
	addCh                 chan *Timer
	removeCh              chan uint32
	senderChanList        []chan TimerList
	senderChanListCounter int32
	C                     <-chan TimerList
	closeCh               chan struct{}
	closeOnce             sync.Once
	toDelIdMap            sync.Map
	resultSender          resultChanSender
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
	w.senderChanList = make([]chan TimerList, w.options.GetMaxSenderNum())
	w.senderChanList[0] = make(chan TimerList, w.options.GetSenderListLength())
	w.C = w.senderChanList[0]
	w.senderChanListCounter = 1
	w.closeCh = make(chan struct{})
	w.resultSender = resultChanSender{w: w}
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

	atomic.StoreInt32(&w.state, 1)
	for atomic.LoadInt32(&w.state) > 0 {
		select {
		case <-w.closeCh:
			for i := 0; i < len(w.senderChanList); i++ {
				close(w.senderChanList[i])
			}
			w.senderChanList = []chan TimerList{nil}
			atomic.StoreInt32(&w.senderChanListCounter, 1)
			atomic.StoreInt32(&w.state, 0)
		case v, o := <-w.addCh:
			if o {
				w.addTimeout(v)
			}
		case id, o := <-w.removeCh:
			if o {
				w.remove(id)
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

func (w *Wheel) ReadTimerList() TimerList {
	tl := <-w.senderChanList[0]
	return tl
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
)

func init() {
	timerPool = sync.Pool{
		New: func() any {
			return &Timer{}
		},
	}
	listPool = sync.Pool{
		New: func() any {
			return list.New()
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

func getList() *list.List {
	return listPool.Get().(*list.List)
}

func putList(l *list.List) {
	l.Clear()
	listPool.Put(l)
}
