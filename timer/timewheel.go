package timer

import (
	"container/list"
	"fmt"
	"os"
	"time"
)

type timeWheelLayer struct {
	slots     []*list.List
	curr_slot int
}

func createTimeWheelLayer(slot_num uint32) *timeWheelLayer {
	twl := &timeWheelLayer{}
	twl.slots = make([]*list.List, slot_num)
	for i := uint32(0); i < slot_num; i++ {
		twl.slots[i] = list.New()
	}
	return twl
}

func (this *timeWheelLayer) addWithTimeout(time_span, timeout time.Duration, fun TimerFunc, args ...interface{}) (*Timer, *list.Element) {
	slot_n := int((timeout + time_span - 1) / time_span)
	if slot_n >= len(this.slots) {
		return nil, nil
	}
	slot_n = (slot_n + this.curr_slot) % len(this.slots)
	timer := &Timer{
		fun:  fun,
		args: args,
		slot: slot_n,
	}
	e := this.slots[slot_n].PushBack(timer)
	return timer, e
}

func (this *timeWheelLayer) remove(timer *Timer, e *list.Element) bool {
	if timer.slot < 0 || timer.slot >= len(this.slots) {
		return false
	}
	this.slots[timer.slot].Remove(e)
	return true
}

type TimeSingleWheel struct {
	layer      *timeWheelLayer
	time_span  time.Duration
	timers_map map[*Timer]*list.Element
	last_time  time.Time
}

func CreateTimeSingleWheel(slot_num uint32, time_span time.Duration) *TimeSingleWheel {
	twl := createTimeWheelLayer(slot_num)
	tw := &TimeSingleWheel{}
	tw.layer = twl
	tw.time_span = time_span
	tw.timers_map = make(map[*Timer]*list.Element)
	tw.last_time = time.Now()
	return tw
}

func (this *TimeSingleWheel) AddWithDeadline(deadline time.Time, fun TimerFunc, args ...interface{}) *Timer {
	duration := time.Until(deadline)
	return this.AddWithTimeout(duration, fun, args)
}

func (this *TimeSingleWheel) AddWithTimeout(timeout time.Duration, fun TimerFunc, args ...interface{}) *Timer {
	timer, e := this.layer.addWithTimeout(this.time_span, timeout, fun, args...)
	if timer == nil {
		return nil
	}
	this.timers_map[timer] = e
	return timer
}

func (this *TimeSingleWheel) Remove(timer *Timer) bool {
	e := this.timers_map[timer]
	if e == nil {
		return false
	}
	return this.layer.remove(timer, e)
}

func (this *TimeSingleWheel) Update() {
	now := time.Now()
	duration := now.Sub(this.last_time)
	slot_n := int(duration / this.time_span)
	if slot_n < 1 {
		return
	}
	if slot_n > len(this.layer.slots)-1 {
		slot_n = len(this.layer.slots) - 1
	}
	for i := this.layer.curr_slot; i < this.layer.curr_slot+slot_n; i++ {
		l := this.layer.slots[i%len(this.layer.slots)]
		e := l.Front()
		for e != nil {
			t, o := (e.Value).(*Timer)
			if !o {
				break
			}
			t.fun(t.args)
			delete(this.timers_map, t)
			n := e.Next()
			l.Remove(e)
			e = n
		}
	}
	this.last_time = this.last_time.Add(time.Duration(slot_n) * this.time_span)
}

type TimeWheel struct {
	layers     []*timeWheelLayer
	time_span  time.Duration
	span_num   uint32
	timers_map map[*Timer]*list.Element
	last_time  time.Time
}

func CreateTimeWheel(slots_num []uint32, time_span time.Duration) *TimeWheel {
	tw := &TimeWheel{}
	tw.layers = make([]*timeWheelLayer, len(slots_num))
	tw.span_num = 1
	for i := 0; i < len(slots_num); i++ {
		tw.layers[i] = createTimeWheelLayer(slots_num[i])
		if i == 0 {
			tw.layers[i].curr_slot = -1
		}
		tw.span_num *= slots_num[i]
	}
	tw.time_span = time_span
	tw.timers_map = make(map[*Timer]*list.Element)
	tw.last_time = time.Now()
	return tw
}

func (this *TimeWheel) AddWithDeadline(deadline time.Time, fun TimerFunc, args ...interface{}) *Timer {
	duration := time.Until(deadline)
	return this.AddWithTimeout(duration, fun, args)
}

func (this *TimeWheel) AddWithTimeout(timeout time.Duration, fun TimerFunc, args ...interface{}) *Timer {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}()

	n := uint32((timeout + this.time_span - 1) / this.time_span)
	n = (n + this.span_num) % this.span_num
	var left_time time.Duration = timeout - time.Duration(n)*this.time_span
	if left_time < 0 {
		left_time = 0
	}
	var layer *timeWheelLayer
	var layer_n int
	for layer_n = 0; layer_n < len(this.layers); layer_n++ {
		if layer_n > 0 {
			n /= uint32(len(this.layers[layer_n-1].slots))
		}
		if this.layers[layer_n].curr_slot+int(n) < len(this.layers[layer_n].slots) {
			layer = this.layers[layer_n]
			n = uint32(this.layers[layer_n].curr_slot) + n
			break
		}
	}
	if layer == nil {
		panic("cant get layer by timeout")
	}
	timer := &Timer{
		fun:       fun,
		args:      args,
		slot:      int(n),
		layer:     layer_n,
		left_time: left_time,
	}
	e := layer.slots[n].PushBack(timer)
	this.timers_map[timer] = e
	return timer
}

func (this *TimeWheel) Remove(timer *Timer) bool {
	e := this.timers_map[timer]
	if e == nil {
		return false
	}
	return this.layers[timer.layer].slots[timer.slot].Remove(e) != nil
}

func (this *TimeWheel) Update() {
	duration := time.Now().Sub(this.last_time)
	n := uint32(duration / this.time_span)
	if n == 0 {
		return
	}
	if n > this.span_num {
		panic("update too late over than time wheel life")
	}

	ln := 0
	for i := uint32(0); i < n; i++ {
		layer := this.layers[ln]
		for {
			if layer.curr_slot+1 < len(layer.slots) {
				layer.curr_slot += 1
				break
			}
			layer.curr_slot = 0
			ln = (ln + 1) % len(this.layers)
			layer = this.layers[ln]
		}

		l := layer.slots[layer.curr_slot]
		e := l.Front()
		for e != nil {
			t, o := (e.Value).(*Timer)
			if !o {
				break
			}
			delete(this.timers_map, t)
			if t.left_time > 0 {
				this.AddWithTimeout(t.left_time, t.fun, t.args...)
			} else {
				t.fun(t.args)
			}
			ne := e.Next()
			l.Remove(e)
			e = ne
		}
	}

	this.last_time.Add(this.time_span * time.Duration(n))
}

var (
	DefaultSlotsNum = []uint32{256, 64, 64, 64, 64}
	DefaultTimeSpan = time.Millisecond * 10
)

func CreateDefaultTimeWheel() *TimeWheel {
	return CreateTimeWheel(DefaultSlotsNum, DefaultTimeSpan)
}
