package time

import (
	"fmt"
	"log"
	"time"

	"github.com/huoshan017/ponu/list"
)

type TimerFunc func(arg []any)

type Timer struct {
	id          uint32
	fun         TimerFunc
	arg         []any
	timeout     time.Duration
	expireTime  time.Time
	senderIndex int32
	leftStep    int32
}

func (t *Timer) Clean() {
	t.id = 0
	t.arg = nil
	t.fun = nil
	t.leftStep = 0
}

type wheelBase struct {
	maxDuration    time.Duration
	resultSender   IResultSender
	options        *Options
	layers         [2][]*wheelLayer
	periodIndex    int8
	prevLayersSize []int32
	lastTickTime   time.Time
	maxStep        int32
	step           int32
	state          int32
	currId         uint32
	id2Pos         map[uint32]struct {
		list.Iterator
		uint8
		int8
		int16
	}
	index2List map[int32]*list.List
}

func newWheelBase(timerMaxDuration time.Duration, resultSender IResultSender, options *Options) *wheelBase {
	wheel := &wheelBase{resultSender: resultSender, options: options}
	if wheel.options.GetInterval() < minInterval {
		wheel.options.SetInterval(minInterval)
	}

	if wheel.options.GetTimerRecvListLength() <= 0 {
		wheel.options.SetTimerRecvListLength(defaultTimerRecvListLength)
	}
	if wheel.options.GetRemoveListLength() <= 0 {
		wheel.options.SetRemoveListLength(defaultRemoveListLength)
	}
	if wheel.options.GetMaxSenderNum() <= 0 {
		wheel.options.SetMaxSenderNum(defaultSendNum)
	}
	if wheel.options.GetSenderListLength() <= 0 {
		wheel.options.SetSenderListLength(defaultSenderListLength)
	}
	var (
		layers         [2][]*wheelLayer
		prevLayersSize []int32
		maxStep, ll    int32
	)
	n := int32((timerMaxDuration + wheel.options.GetInterval() - 1) / wheel.options.GetInterval())
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
	wheel.layers = layers
	wheel.prevLayersSize = prevLayersSize
	wheel.maxDuration = timerMaxDuration
	wheel.maxStep = maxStep
	wheel.index2List = make(map[int32]*list.List)
	return wheel
}

func (w *wheelBase) addTimeout(t *Timer) {
	// todo 计算timer的step
	cost := time.Until(t.expireTime)
	step := int32(cost / w.options.GetInterval())
	t.leftStep = step
	w.addTimer(t, false)
}

func (w *wheelBase) addTimer(t *Timer, update bool) {
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
		panic(fmt.Sprintf("time wheel: Not found suitable position to insert time (id: %v,  step: %v,  w.step: %v,  w.periodIndex: %v,  periodIndex: %v,  forwardSlots: %v,  leftSteps: %v)",
			t.id, steps, w.step, w.periodIndex, periodIndex, forwardSlots, t.leftStep))
	}

	var iter list.Iterator
	var pos int32
	if periodIndex == w.periodIndex {
		iter, pos = layer.insertTimer(forwardSlots, t)
		if pos < 0 {
			putTimer(t)
			panic(fmt.Sprintf("time wheel: insert time with forwardSlots %v failed", forwardSlots))
		}
	} else {
		iter = layer.insertTimerWithSlot(forwardSlots-1, t)
	}

	if t.id > 0 {
		w.id2Pos[t.id] = struct {
			list.Iterator
			uint8
			int8
			int16
		}{iter, uint8(periodIndex), int8(layerN), int16(pos)}
	}
}

func (w *wheelBase) adjustTimer(t *Timer) bool {
	d := time.Until(t.expireTime)
	if d <= 0 { // 已超时不需要调整
		return false
	}
	// 尽量调少，忽略掉余数
	step := int32(d / w.options.GetInterval())
	if step == 0 {
		step = 1
	}
	t.leftStep = step
	w.addTimer(t, true)
	return true
}

func (w *wheelBase) remove(id uint32) bool {
	value, o := w.id2Pos[id]
	if !o {
		log.Printf("time: wheel remove timer %v failed", id)
		return false
	}
	delete(w.id2Pos, id)
	return w.layers[value.uint8][value.int8].removeTimer(int32(value.int16), value.Iterator)
}

func (w *wheelBase) stepOne() {
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
	}

	// 移动一个step，并处理对应slot的timer
	for i := 0; i < len(w.layers[w.periodIndex]); i++ {
		// 移动到下一个slot，返回是否到下一层
		toNextLayer := w.layers[w.periodIndex][i].toNextSlot()
		// 不包括第一层，把该层对应slot的Timer链表取出
		if i > 0 {
			l := w.layers[w.periodIndex][i].getCurrListAndRemove()
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
					//delete(w.id2Pos, t.id)
				}
				putList(l)
			}
		}
		// 不会进入下一层则跳出
		if !toNextLayer {
			break
		}
	}
}

func (w *wheelBase) handleStep() {
	w.stepOne()

	// 取出当前层slot中的链表，处理Timer
	var tlist = w.layers[w.periodIndex][0].getCurrListAndRemove()
	if tlist == nil || tlist.GetLength() == 0 {
		return
	}

	var haveTimer bool
	now := time.Now()
	for iter := tlist.Begin(); iter != tlist.End(); {
		t := iter.Value().(*Timer)
		//var toDel bool
		//if t.id > 0 {
		//	_, toDel = w.toDelIdMap.LoadAndDelete(t.id)
		//}
		//if toDel { // 是否需要删除
		//	iter, _ = tlist.DeleteContinueNext(iter)
		//	putTimer(t)
		//} else {
		// 未到超时时间
		if now.Sub(t.expireTime) < 0 {
			if w.adjustTimer(t) {
				iter, _ = tlist.DeleteContinueNext(iter)
				continue
			}
		}
		// 处理不同sender的timer
		if t.senderIndex > 0 {
			l := w.index2List[t.senderIndex]
			if l == nil {
				l = getList()
				w.index2List[t.senderIndex] = l
			}
			l.PushBack(t)
			iter, _ = tlist.DeleteContinueNext(iter)
			if haveTimer {
				haveTimer = true
			}
			continue
		}
		iter = iter.Next()
		//}
		// 删除掉map中缓存的timer id
		if t.id > 0 {
			delete(w.id2Pos, t.id)
		}
	}
	if tlist.GetLength() == 0 {
		putList(tlist)
	} else {
		w.index2List[0] = tlist
		haveTimer = true
	}

	if haveTimer {
		for idx, l := range w.index2List {
			if l != nil {
				w.resultSender.Send(idx, l)
				w.index2List[idx] = nil
			}
		}
	}
}

func (w *wheelBase) handleTick() {
	lastTickTime := time.Now()
	d := lastTickTime.Sub(w.lastTickTime)
	for d >= w.options.GetInterval() {
		w.handleStep()
		d -= w.options.GetInterval()
	}
	if d > 0 {
		w.lastTickTime = lastTickTime.Add(-d)
	}
}
