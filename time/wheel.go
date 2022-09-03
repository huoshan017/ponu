package time

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/huoshan017/ponu/list"
	"github.com/pkg/errors"
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

func (w *wheelLayer) insertTimer(nextNSlot int32, timer *Timer) (list.Iterator, bool) {
	if nextNSlot > int32(len(w.slots)) {
		return list.Iterator{}, false
	}
	insertSlot := int(w.length-1+nextNSlot) % len(w.slots)
	l := w.slots[insertSlot]
	if l == nil {
		l = getList()
		w.slots[insertSlot] = l
	}
	l.PushBack(timer)
	return l.RBegin(), true
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
	prevLayersSize []int32
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
		args []any
	}
	removeCh     chan uint32
	closeCh      chan struct{}
	closeOnce    sync.Once
	lastTickTime time.Time
	lastStepTime time.Time
	maxStep      int32
	step         int32
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
		prevLayersSize []int32
		maxStep, ll    int32
	)
	n := int32((timerMaxDuration + interval - 1) / interval)
	for i := 0; i < len(layers); i++ {
		if n <= 256 {
			layers[i] = []*wheelLayer{createWheelLayer(n)}
			prevLayersSize = []int32{1}
			maxStep = n
		} else if n <= 256*64 {
			ll = (n + 255) / 256
			layers[i] = []*wheelLayer{createWheelLayer(256), createWheelLayer(ll)}
			prevLayersSize = []int32{1, 256}
			maxStep = ll * 256
		} else if n <= 256*64*64 {
			ll = (n + 256*64 - 1) / (256 * 64)
			layers[i] = []*wheelLayer{createWheelLayer(256), createWheelLayer(64), createWheelLayer(ll)}
			prevLayersSize = []int32{1, 256, 256 * 64}
			maxStep = ll * 256 * 64
		} else if n <= 256*64*64*64 {
			ll = (n + 256*64*64 - 1) / (256 * 64 * 64)
			layers[i] = []*wheelLayer{createWheelLayer(256), createWheelLayer(64), createWheelLayer(64), createWheelLayer(ll)}
			prevLayersSize = []int32{1, 256, 256 * 64, 256 * 64 * 64}
			maxStep = ll * 256 * 64 * 64
		} else if n <= 256*64*64*64*16 {
			ll = (n + 256*64*64*64 - 1) / (256 * 64 * 64 * 64)
			layers[i] = []*wheelLayer{createWheelLayer(256), createWheelLayer(64), createWheelLayer(64), createWheelLayer(64), createWheelLayer(ll)}
			prevLayersSize = []int32{1, 256, 256 * 64, 256 * 64 * 64, 256 * 64 * 64 * 64}
			maxStep = ll * 256 * 64 * 64 * 64
		} else {
			panic("ponu: greater than time wheel range")
		}
	}
	w := &Wheel{
		layers:         layers,
		prevLayersSize: prevLayersSize,
		interval:       interval,
		maxDuration:    timerMaxDuration,
		maxStep:        maxStep,
		getCh:          make(chan TimerList, 256),
		addCh: make(chan struct {
			uint32
			time.Time
			time.Duration
			TimerFunc
			args []any
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
	defer func() {
		if err := recover(); err != nil {
			er := errors.Errorf("%v", err)
			log.Fatalf("\n%+v", er)
		}
	}()

	now := time.Now()
	w.stepTicker = time.NewTicker(w.interval)
	w.lastTickTime = now
	w.lastStepTime = now
	atomic.StoreInt32(&w.state, 1)
	for atomic.LoadInt32(&w.state) > 0 {
		select {
		case <-w.closeCh:
			close(w.getCh)
			w.getCh = nil
			atomic.StoreInt32(&w.state, 0)
		case v, o := <-w.addCh:
			if o {
				w.addTimeout(v.uint32, v.Time, v.Duration, v.TimerFunc, v.args)
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
	if timeout < w.interval || timeout > w.maxDuration {
		return 0
	}
	newId := atomic.AddUint32(&w.currId, 1)
	d := struct {
		uint32
		time.Time
		time.Duration
		TimerFunc
		args []any
	}{newId, time.Now(), timeout, fun, args}
	w.addCh <- d
	return newId
}

func (w *Wheel) AddNoId(timeout time.Duration, fun TimerFunc, args []any) bool {
	if timeout < w.interval || timeout > w.maxDuration {
		return false
	}
	w.addCh <- struct {
		uint32
		time.Time
		time.Duration
		TimerFunc
		args []any
	}{0, time.Now(), timeout, fun, args}
	return true
}

func (w *Wheel) AddWithDeadline(deadline time.Time, fun TimerFunc, args []any) uint32 {
	duration := time.Until(deadline)
	return w.Add(duration, fun, args)
}

func (w *Wheel) Remove(id uint32) {
	w.toDelIdMap.LoadOrStore(id, true)
	w.removeCh <- id
}

func (w *Wheel) ReadTimerList() TimerList {
	tl := <-w.getCh
	return tl
}

func (w *Wheel) stepOne() {
	var l *list.List
	for i := 0; i < len(w.layers[w.periodIndex]); i++ {
		// 不会进入下一层
		if w.layers[w.periodIndex][i].toNextSlot() {
			continue
		}
		// 不包括第一层，把该层对应slot的Timer链表取出
		if i > 0 {
			l = w.layers[w.periodIndex][i].getCurrListAndRemove()
			// 处理取出的链表，根据剩余的step数计算出合适的位置放入，没有剩余则插入第一层的当前slot链表中
			if l != nil {
				node, o := l.PopFront()
				for o {
					t := node.(*Timer)
					if t.leftStep <= 0 { // 插入到第一层
						w.layers[w.periodIndex][0].insertTimer(0, t)
					} else { // 继续插入到合适的位置
						w.addTimer(t, true)
					}
					node, o = l.PopFront()
					delete(w.id2Pos, t.id)
				}
				putList(l)
			}
		}
		break
	}
}

func (w *Wheel) handleStep() {
	w.stepOne()

	// 计步
	w.step += 1
	if w.step > w.maxStep {
		// 重置layer的数据
		for i := 0; i < len(w.layers[w.periodIndex]); i++ {
			w.layers[w.periodIndex][i].reset()
		}
		w.step = 1
		// 切换period
		w.periodIndex = (w.periodIndex + 1) % 2
		for i := 0; i < len(w.layers[w.periodIndex]); i++ {
			if !w.layers[w.periodIndex][i].toNextSlot() {
				break
			}
		}
	}
	now := time.Now()
	w.lastStepTime = now

	// 取出当前层slot中的链表，处理Timer
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
				// no timeout
				if now.Sub(t.expireTime) < 0 {
					if w.adjustTimer(t) {
						iter, _ = tlist.DeleteContinueNext(iter)
						continue
					}
				}
				iter = iter.Next()
			}
			if t.id > 0 {
				delete(w.id2Pos, t.id)
			}
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

func (w *Wheel) addTimeout(id uint32, start time.Time, timeout time.Duration, fun TimerFunc, arg []any) {
	// todo 计算timer的step
	expireTime := start.Add(timeout)
	cost := expireTime.Sub(w.lastStepTime)
	step := int32(cost / w.interval)
	t := getTimer()
	t.id = id
	t.fun = fun
	t.arg = arg
	t.timeout = timeout
	t.leftStep = step
	t.expireTime = expireTime
	w.addTimer(t, false)
}

func (w *Wheel) addTimer(t *Timer, update bool) {
	var (
		layer               *wheelLayer
		layerN              int32
		forwardSlots, steps int32 = t.leftStep, t.leftStep
		periodIndex         int8
		currLayers          []*wheelLayer
		cum                 int32
	)

	if w.step+forwardSlots <= w.maxStep {
		periodIndex = w.periodIndex
		currLayers = w.layers[periodIndex]
		for layerN = int32(0); layerN < int32(len(currLayers)); layerN++ {
			layer = currLayers[layerN]
			cum += layer.getLength() * w.prevLayersSize[layerN]
			if forwardSlots <= layer.getSize() {
				t.leftStep -= (layer.getLength()+forwardSlots)*w.prevLayersSize[layerN] - cum
				break
			}
			a := layer.getSize() - layer.getLength()
			forwardSlots = (forwardSlots - a + layer.getSize() - 1) / layer.getSize()
		}
	} else {
		forwardSlots = (forwardSlots + w.step) - w.maxStep
		t.leftStep = forwardSlots
		periodIndex = (w.periodIndex + 1) % 2
		currLayers = w.layers[periodIndex]
		for layerN = int32(0); layerN < int32(len(currLayers)); layerN++ {
			layer = currLayers[layerN]
			forwardSlots = (forwardSlots + w.prevLayersSize[layerN] - 1) / w.prevLayersSize[layerN]
			if forwardSlots <= layer.getSize() {
				t.leftStep -= (forwardSlots - 1) * w.prevLayersSize[layerN]
				break
			}
		}
	}

	if layerN >= int32(len(currLayers)) {
		for i := 0; i < len(w.layers[w.periodIndex]); i++ {
			log.Printf("      w.layers[%v][%v] length %v", w.periodIndex, i, w.layers[w.periodIndex][i].length)
		}
		panic(fmt.Sprintf("time wheel: Not found suitable position to insert time (id: %v,  step: %v,  w.step: %v,  w.periodIndex: %v,  periodIndex: %v,  forwardSlots: %v,  leftSteps: %v)",
			t.id, steps, w.step, w.periodIndex, periodIndex, forwardSlots, t.leftStep))
	}

	var iter list.Iterator
	if periodIndex == w.periodIndex {
		var o bool
		iter, o = layer.insertTimer(forwardSlots, t)
		if !o {
			putTimer(t)
			panic(fmt.Sprintf("time wheel: insert time with forwardSlots %v failed", forwardSlots))
		}
	} else {
		iter = layer.insertTimerWithSlot(forwardSlots-1, t)
	}

	if t.id > 0 {
		w.id2Pos[t.id] = struct {
			list.Iterator
			int32
		}{iter, (int32(periodIndex)<<24&0x7f000000 | layerN<<16&0x7fff0000 | (layer.getLength()-1)&0x0000ffff)}
	}

	/*
		if !update {
			log.Printf("add timer  id %v  w.periodIndex %v  periodIndex %v  timeout %v  layerN %v  forwardSlots %v  w.step %v  steps %v  leftSteps %v", t.id, w.periodIndex, periodIndex, t.timeout, layerN, forwardSlots, w.step, steps, t.leftStep)
		} else {
			log.Printf("update timer  id %v  w.periodIndex %v  periodIndex %v  timeout %v  layerN %v  forwardSlots %v  w.step %v  steps %v  leftSteps %v", t.id, w.periodIndex, periodIndex, t.timeout, layerN, forwardSlots, w.step, steps, t.leftStep)
		}
	*/
}

func (w *Wheel) adjustTimer(t *Timer) bool {
	d := time.Until(t.expireTime)
	if d == 0 {
		return false
	}
	step := int32(d / w.interval)
	if step == 0 {
		step = 1
	}
	t.leftStep = step
	w.addTimer(t, true)
	return true
}

func (w *Wheel) remove(id uint32) bool {
	value, o := w.id2Pos[id]
	if !o {
		return false
	}
	delete(w.id2Pos, id)
	w.toDelIdMap.Delete(id)
	periodIndex, layer, cn := value.int32>>24, value.int32>>16&0xffff, value.int32&0xffff
	return w.layers[periodIndex][layer].removeTimer(cn, value.Iterator)
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
