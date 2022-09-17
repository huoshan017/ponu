package time

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/huoshan017/ponu/list"
	"github.com/pkg/errors"
)

const (
	minInterval = 5 * time.Millisecond
)

type wheelLayer struct {
	slots  []*list.List
	length int32
}

func createWheelLayer(slot_num int32) *wheelLayer {
	twl := &wheelLayer{}
	twl.slots = make([]*list.List, slot_num)
	return twl
}

func (w *wheelLayer) getSize() int32 {
	return int32(len(w.slots))
}

func (w *wheelLayer) getLength() int32 {
	return w.length
}

func (w *wheelLayer) reset() {
	w.length = 0
}

func (w *wheelLayer) toNextSlot() (toNextLayer bool) {
	// 开始进入此layer和结束进入下一个layer都表示toNextLayer
	if w.length == 0 || int(w.length) >= len(w.slots) {
		w.length = 1
		toNextLayer = true
	} else {
		w.length += 1
	}
	return
}

func (w *wheelLayer) getCurrListAndRemove() *list.List {
	if w.length <= 0 {
		return nil
	}
	l := w.slots[w.length-1]
	if l != nil {
		w.slots[w.length-1] = nil
	}
	return l
}

func (w *wheelLayer) insertTimer(nextNSlot int32, timer *Timer) (list.Iterator, int32) {
	if nextNSlot > int32(len(w.slots)) {
		return list.Iterator{}, -1
	}
	insertSlot := (w.length - 1 + nextNSlot) % int32(len(w.slots))
	l := w.slots[insertSlot]
	if l == nil {
		l = getList()
		w.slots[insertSlot] = l
	}
	l.PushBack(timer)
	return l.RBegin(), insertSlot
}

func (w *wheelLayer) insertTimerWithSlot(slot int32, timer *Timer) list.Iterator {
	l := w.slots[slot]
	if l == nil {
		l = getList()
		w.slots[slot] = l
	}
	l.PushBack(timer)
	return l.RBegin()
}

func (w *wheelLayer) removeTimer(slot int32, iter list.Iterator) bool {
	if w.slots[slot] == nil {
		return false
	}
	return w.slots[slot].Delete(iter)
}

type TimerList struct {
	l *list.List
	m *sync.Map
}

func (t *TimerList) ExecuteFunc() {
	node, o := t.l.PopFront()
	for o {
		timer := node.(*Timer)
		var del bool
		if timer.id > 0 {
			_, del = t.m.LoadAndDelete(timer.id)
		}
		if !del {
			timer.fun(timer.arg)
		}
		putTimer(timer)
		node, o = t.l.PopFront()
	}
	putList(t.l)
	t.l = nil
}

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
	var wheel Wheel
	for _, option := range options {
		option(&wheel.options)
	}
	if timerMaxDuration < wheel.options.GetInterval() {
		return nil
	}

	w := &Wheel{}
	*w = wheel
	w.wheelBase = *newWheelBase(timerMaxDuration, &w.resultSender, &wheel.options)
	w.addCh = make(chan *Timer, wheel.options.GetTimerRecvListLength())
	w.removeCh = make(chan uint32, wheel.options.GetRemoveListLength())
	w.senderChanList = make([]chan TimerList, wheel.options.GetMaxSenderNum())
	w.senderChanList[0] = make(chan TimerList, wheel.options.GetSenderListLength())
	w.C = w.senderChanList[0]
	w.senderChanListCounter = 1
	w.closeCh = make(chan struct{})
	w.id2Pos = make(map[uint32]struct {
		list.Iterator
		uint8
		int8
		int16
	})
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
	w.lastTickTime = time.Now()

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

func (w *Wheel) PostWithDeadline(deadline time.Time, fun TimerFunc, args []any) {
	duration := time.Until(deadline)
	w.Post(duration, fun, args)
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
