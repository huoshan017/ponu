package time

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/huoshan017/ponu/list"
	"github.com/huoshan017/ponu/lockfree"
	"github.com/pkg/errors"
)

const (
	reqAdd    int32 = iota
	reqCancel int32 = 1
)

type Requester struct {
	wheel *WheelX
	index int32
	queue *lockfree.QueueT[TimerList]
}

func newRequester(w *WheelX, idx int32) *Requester {
	return &Requester{
		wheel: w,
		index: idx,
		queue: lockfree.NewQueueT[TimerList](),
	}
}

func (r *Requester) Add(timeout time.Duration, fun TimerFunc, args []any) uint32 {
	return r.wheel.add(r.index, timeout, fun, args)
}

func (r *Requester) AddWithDeadline(deadline time.Time, fun TimerFunc, args []any) uint32 {
	return r.wheel.addWithDeadline(r.index, deadline, fun, args)
}

func (r *Requester) Post(timeout time.Duration, fun TimerFunc, args []any) bool {
	return r.wheel.post(r.index, timeout, fun, args)
}

func (r *Requester) PostWithDeadline(deadline time.Time, fun TimerFunc, args []any) bool {
	return r.wheel.postWithDeadline(r.index, deadline, fun, args)
}

func (r *Requester) Cancel(timerId uint32) {
	r.wheel.cancel(timerId)
}

func (r *Requester) GetResult() (TimerList, bool) {
	t, o := r.queue.Dequeue()
	for !o {
		return TimerList{}, false
	}
	return t, true
}

func (r *Requester) Update() {
	t, o := r.queue.Dequeue()
	for o {
		t.ExecuteFunc()
		t, o = r.queue.Dequeue()
	}
}

type resultQueueSender struct {
	w *WheelX
}

func (s *resultQueueSender) Send(index int32, tlist *list.ListT[*Timer]) {
	requester := s.w.requesterMap[index]
	if requester != nil {
		requester.queue.Enqueue(TimerList{l: tlist, m: &s.w.toDelIdMap})
	}
}

type WheelX struct {
	*wheelBase
	options          Options
	reqList          *list.ConcurrentList
	requesterCounter int32
	resultSender     IResultSender
	state            int32
	toDelIdMap       sync.Map
	requesterMap     map[int32]*Requester
}

func NewWheelX(timerMaxDuration time.Duration, options ...Option) *WheelX {
	var ops Options
	for _, option := range options {
		option(&ops)
	}
	if timerMaxDuration < ops.GetInterval() {
		return nil
	}

	w := &WheelX{}
	w.options = ops
	w.resultSender = &resultQueueSender{w: w}
	w.wheelBase = newWheelBase(timerMaxDuration, w.resultSender, &w.options)
	w.reqList = list.NewConcurrentList()
	w.requesterMap = make(map[int32]*Requester)
	return w
}

func (w *WheelX) NewRequester() *Requester {
	count := atomic.AddInt32(&w.requesterCounter, 1)
	requester := newRequester(w, count-1)
	w.requesterMap[count-1] = requester
	return requester
}

func (w *WheelX) Run() {
	defer func() {
		if err := recover(); err != nil {
			er := errors.Errorf("%v", err)
			log.Fatalf("\n%+v", er)
		}
	}()

	w.start()

	var req any
	atomic.StoreInt32(&w.state, 1)
	for atomic.LoadInt32(&w.state) > 0 {
		req, _ = w.reqList.PopFront()
		if req != nil {
			d := req.(struct {
				typ  int32
				data any
			})
			if d.typ == reqAdd {
				w.addTimeout(d.data.(*Timer))
			} else if d.typ == reqCancel {
				w.remove(d.data.(uint32))
			} else {
				log.Printf("ponu.time.WheelX unknown request type %v", d.typ)
			}
			w.handleTick()
		}
	}
}

func (w *WheelX) Stop() {
	atomic.StoreInt32(&w.state, 0)
}

func (w *WheelX) add(index int32, timeout time.Duration, fun TimerFunc, args []any) uint32 {
	if timeout < w.options.GetInterval() || timeout > w.maxDuration {
		return 0
	}
	newId := atomic.AddUint32(&w.currId, 1)
	w.request(index, newId, timeout, fun, args)
	return newId
}

func (w *WheelX) post(index int32, timeout time.Duration, fun TimerFunc, args []any) bool {
	if timeout < w.options.GetInterval() || timeout > w.maxDuration {
		return false
	}
	w.request(index, 0, timeout, fun, args)
	return true
}

func (w *WheelX) addWithDeadline(index int32, deadline time.Time, fun TimerFunc, args []any) uint32 {
	duration := time.Until(deadline)
	return w.add(index, duration, fun, args)
}

func (w *WheelX) postWithDeadline(index int32, deadline time.Time, fun TimerFunc, args []any) bool {
	duration := time.Until(deadline)
	return w.post(index, duration, fun, args)
}

func (w *WheelX) cancel(id uint32) {
	w.toDelIdMap.LoadOrStore(id, true)
	w.reqList.PushBack(struct {
		typ  int32
		data any
	}{reqCancel, id})
}

func (w *WheelX) request(idx int32, id uint32, timeout time.Duration, fun TimerFunc, args []any) {
	t := getTimer()
	t.senderIndex = idx
	t.id = id
	t.timeout = timeout
	t.fun = fun
	t.args = args
	t.expireTime = time.Now().Add(timeout)
	w.reqList.PushBack(struct {
		typ  int32
		data any
	}{typ: reqAdd, data: t})
}

func (w *WheelX) remove(id uint32) {
	w.wheelBase.remove(id)
	w.toDelIdMap.Delete(id)
}
