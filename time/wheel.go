package time

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/huoshan017/ponu/list"
	"github.com/pkg/errors"
)

type wheelLayer struct {
	slots     []*list.List
	curr_slot int32
}

func createWheelLayer(slot_num int64) *wheelLayer {
	twl := &wheelLayer{}
	twl.slots = make([]*list.List, slot_num)
	return twl
}

func (w *wheelLayer) toNextSlot() (nextPeriod bool) {
	w.curr_slot += 1
	if int(w.curr_slot) >= len(w.slots) {
		w.curr_slot = 0
		nextPeriod = true
	}
	return
}

func (w *wheelLayer) getCurrListAndRemove() *list.List {
	l := w.slots[w.curr_slot]
	if l != nil {
		w.slots[w.curr_slot] = nil
	}
	return l
}

func (w *wheelLayer) insertTimer(nextNSlot int32, timer *Timer) bool {
	if int(nextNSlot) >= len(w.slots) {
		return false
	}
	insertSlot := int(w.curr_slot+nextNSlot) % len(w.slots)
	l := w.slots[insertSlot]
	if l == nil {
		l = getList()
		w.slots[insertSlot] = l
	}
	l.PushBack(timer)
	return true
}

func (w *wheelLayer) insertTimerWithSlot(slot int32, timer *Timer) list.Iterator {
	l := w.slots[slot]
	if l == nil {
		l = getList()
		w.slots[slot] = l
	}
	l.PushBack(timer)
	return l.End()
}

func (w *wheelLayer) removeTimer(slot int32, iter list.Iterator) bool {
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
			_, del = t.m.Load(timer.id)
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

type Wheel struct {
	layers         [2][]*wheelLayer
	periodIndex    int8
	prevLayersStep []int64
	interval       time.Duration
	maxDuration    time.Duration
	stepTicker     *time.Ticker
	getCh          chan TimerList
	C              <-chan TimerList
	addCh          chan struct {
		uint32
		time.Time
		time.Duration
		TimerFunc
		any
	}
	removeCh     chan uint32
	closeCh      chan struct{}
	closeOnce    sync.Once
	startTime    time.Time
	lastTickTime time.Time
	maxStep      int64
	step         int64
	state        int32
	currId       uint32
	id2Pos       map[uint32]struct {
		list.Iterator
		int32
	}
	toDelIdMap sync.Map
}

func NewWheel(interval, timerMaxDuration time.Duration) *Wheel {
	w := newWheel(interval, timerMaxDuration)
	go w.run()
	return w
}

func newWheel(interval, timerMaxDuration time.Duration) *Wheel {
	if interval <= 0 || timerMaxDuration < interval {
		return nil
	}
	var (
		layers         [2][]*wheelLayer
		prevLayersStep []int64
		maxStep        int64
	)
	n := int64((timerMaxDuration + interval - 1) / interval)
	for i := 0; i < len(layers); i++ {
		if n < 256 {
			nn := n + 1
			layers[i] = []*wheelLayer{createWheelLayer(nn)}
			prevLayersStep = []int64{0}
			maxStep = nn - 1 + 1
		} else if n < 256*64 {
			nn := (n-255+256-1)/256 + 1
			layers[i] = []*wheelLayer{createWheelLayer(256), createWheelLayer(nn)}
			prevLayersStep = []int64{0, 256}
			maxStep = 255 + 256*(nn-1) + 1
		} else if n < 256*64*64 {
			nn := (n-255-256*63+256*64-1)/(256*64) + 1
			layers[i] = []*wheelLayer{createWheelLayer(256), createWheelLayer(64), createWheelLayer(nn)}
			prevLayersStep = []int64{0, 256, 256 * 64}
			maxStep = 255 + 256*63 + 256*64*(nn-1) + 1
		} else if n < 256*64*64*64 {
			nn := (n-255-256*63-256*64*63+256*64*64-1)/(256*64*64) + 1
			layers[i] = []*wheelLayer{createWheelLayer(256), createWheelLayer(64), createWheelLayer(64), createWheelLayer(nn)}
			prevLayersStep = []int64{0, 256, 256 * 64, 256 * 64 * 64}
			maxStep = 255 + 256*63 + 256*64*63 + 256*64*64*(nn-1) + 1
		} else if n < 256*64*64*64*64 {
			nn := (n - 255 - 256*63 - 256*64*63 - 256*64*64*63 + 256*64*64*64 - 1) / (256 * 64 * 64 * 64)
			layers[i] = []*wheelLayer{createWheelLayer(256), createWheelLayer(64), createWheelLayer(64), createWheelLayer(64), createWheelLayer(nn)}
			prevLayersStep = []int64{0, 256, 256 * 64, 256 * 64 * 64, 256 * 64 * 64 * 64}
			maxStep = 255 + 256*63 + 256*64*63 + 256*64*64*63 + 256*64*64*64*(nn-1) + 1
		} else {
			panic("ponu: greater than time wheel range")
		}
	}
	w := &Wheel{
		layers:         layers,
		prevLayersStep: prevLayersStep,
		interval:       interval,
		maxDuration:    timerMaxDuration,
		maxStep:        maxStep,
		getCh:          make(chan TimerList, 256),
		addCh: make(chan struct {
			uint32
			time.Time
			time.Duration
			TimerFunc
			any
		}, 256),
		removeCh: make(chan uint32, 256),
		closeCh:  make(chan struct{}),
		id2Pos: make(map[uint32]struct {
			list.Iterator
			int32
		}),
	}
	w.C = w.getCh
	return w
}

func (w *Wheel) run() {
	now := time.Now()
	w.startTime = now
	w.stepTicker = time.NewTicker(w.interval)
	w.lastTickTime = now
	atomic.StoreInt32(&w.state, 1)
	defer func() {
		if err := recover(); err != nil {
			er := errors.Errorf("%v", err)
			log.Fatalf("\n%+v", er)
		}
	}()
	for atomic.LoadInt32(&w.state) > 0 {
		select {
		case <-w.closeCh:
			close(w.getCh)
			w.getCh = nil
			atomic.StoreInt32(&w.state, 0)
		case v, o := <-w.addCh:
			if o {
				w.addTimeout(v.uint32, v.Time, v.Duration, v.TimerFunc, v.any)
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

func (w *Wheel) Add(timeout time.Duration, fun TimerFunc, arg any) uint32 {
	if timeout == 0 || timeout > w.maxDuration-w.interval {
		return 0
	}
	newId := atomic.AddUint32(&w.currId, 1)
	w.addCh <- struct {
		uint32
		time.Time
		time.Duration
		TimerFunc
		any
	}{newId, time.Now(), timeout, fun, arg}
	return newId
}

func (w *Wheel) AddNoId(timeout time.Duration, fun TimerFunc, arg any) bool {
	if timeout == 0 || timeout > w.maxDuration-w.interval {
		return false
	}
	w.addCh <- struct {
		uint32
		time.Time
		time.Duration
		TimerFunc
		any
	}{0, time.Now(), timeout, fun, arg}
	return true
}

func (w *Wheel) AddWithDeadline(deadline time.Time, fun TimerFunc, args any) uint32 {
	duration := time.Until(deadline)
	return w.Add(duration, fun, args)
}

func (w *Wheel) Remove(id uint32) {
	w.toDelIdMap.Store(id, true)
	w.removeCh <- id
}

func (w *Wheel) ReadTimerList() TimerList {
	tl := <-w.getCh
	return tl
}

func (w *Wheel) handleStep() {
	var l *list.List
	for i := 0; i < len(w.layers[w.periodIndex]); i++ {
		if !w.layers[w.periodIndex][i].toNextSlot() {
			if i > 0 {
				l = w.layers[w.periodIndex][i].getCurrListAndRemove()
			}
			break
		}
	}
	if l != nil {
		node, o := l.PopFront()
		for o {
			t := node.(*Timer)
			if t.leftStep <= 0 { // 立即执行
				w.layers[w.periodIndex][0].insertTimer(0, t)
			} else { // 继续插入到合适的位置
				w.addTimeoutWithSteps(t.id, int64(t.leftStep), t.fun, t.arg)
				putTimer(t)
			}
			node, o = l.PopFront()
			delete(w.id2Pos, t.id)
		}
		putList(l)
	}
	w.step += 1
	if w.step >= w.maxStep {
		w.step = 0
		w.periodIndex = (w.periodIndex + 1) % 2
	}
	var tlist = w.layers[w.periodIndex][0].getCurrListAndRemove()
	if tlist != nil && tlist.GetLength() > 0 {
		for iter := tlist.Begin(); iter != tlist.End(); {
			t := iter.Value().(*Timer)
			var toDel bool
			if t.id > 0 {
				_, toDel = w.toDelIdMap.Load(t.id)
			}
			if toDel {
				iter, _ = tlist.DeleteContinueNext(iter)
				putTimer(t)
			} else {
				iter = iter.Next()
			}
			delete(w.id2Pos, t.id)
		}
		w.getCh <- TimerList{l: tlist, m: &w.toDelIdMap}
	}
}

func (w *Wheel) handleTick() {
	lastTickTime := time.Now()
	d := lastTickTime.Sub(w.lastTickTime)
	for d >= w.interval {
		w.handleStep()
		d -= w.interval
	}
	if d > 0 {
		w.lastTickTime = lastTickTime.Add(-d)
	}
}

func (w *Wheel) addTimeout(id uint32, start time.Time, timeout time.Duration, fun TimerFunc, arg any) {
	var n = w.step + int64((timeout-time.Since(start)+w.interval-1)/w.interval)
	w.addTimeoutWithSteps(id, n, fun, arg)
}

func (w *Wheel) addTimeoutWithSteps(id uint32, n int64, fun TimerFunc, arg any) {
	var (
		layer  int32
		nn     int64 = n
		cs, ls int32
		timer  *Timer
		iter   list.Iterator
	)

	var periodIndex int8
	if nn < w.maxStep {
		periodIndex = w.periodIndex
	} else {
		periodIndex = (w.periodIndex + 1) % 2
		nn -= w.maxStep
	}
	for layer = int32(len(w.layers[periodIndex])) - 1; layer >= 0; layer-- {
		if w.prevLayersStep[layer] > 0 {
			cs = int32(nn / w.prevLayersStep[layer])
			ls = int32(nn % w.prevLayersStep[layer])
		} else {
			cs = int32(nn)
			ls = 0
		}
		if cs > 0 && cs != int32(w.layers[periodIndex][layer].curr_slot) || (layer == 0 && cs == 0) {
			timer = getTimer()
			timer.id = id
			timer.fun = fun
			timer.arg = arg
			timer.leftStep = ls
			iter = w.layers[periodIndex][layer].insertTimerWithSlot(cs, timer)
			break
		}
		nn = int64(ls)
	}

	if !iter.IsValid() {
		log.Printf("add timer failed,  id %v  n %v  periodIndex %v  cs %v  ls %v", id, n, periodIndex, cs, ls)
	}

	if id > 0 {
		w.id2Pos[id] = struct {
			list.Iterator
			int32
		}{iter, (layer<<16&0x7fff0000 | cs&0x0000ffff)}
	}
}

func (w *Wheel) remove(id uint32) bool {
	value, o := w.id2Pos[id]
	if !o {
		return false
	}
	delete(w.id2Pos, id)
	w.toDelIdMap.Delete(id)
	layer, cn := value.int32>>16&0xffff, value.int32&0xffff
	return w.layers[w.periodIndex][layer].removeTimer(cn, value.Iterator)
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
