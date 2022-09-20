package time

import (
	"math/rand"
	"testing"
	"time"
)

func TestWheelTimerMaxDuration(t *testing.T) {
	const (
		timeUnit               = time.Millisecond
		interval         int32 = 5
		timerMaxDuration int32 = 20000 * interval
	)
	var (
		w          = NewWheel(time.Duration(timerMaxDuration)*timeUnit, WithInterval(time.Duration(interval)*timeUnit))
		en, tn, rn uint32
	)

	var fun = TimerFunc(func(args []any) {
		en += 1
		r := args[0].(int32)
		startTime := args[1].(time.Time)
		triggerTime := args[2].(time.Time)
		yt := (time.Duration(r) * timeUnit).Milliseconds()
		st := time.Since(startTime).Milliseconds()
		tt := triggerTime.Sub(startTime).Milliseconds()
		if yt > st {
			t.Fatalf("yt(%v) > st(%v)", yt, st)
		}
		t.Logf("executed (total count: %v, count: %v, to remove count: %v) timer func with timeout %+v, cost %v ms (diff %v), trigger %v ms (diff %v)", tn, en, rn, yt, st, st-yt, tt, tt-yt)
	})

	for i := 0; i < len(w.layers); i++ {
		for j := 0; j < len(w.layers[i]); j++ {
			t.Logf("layer[%v][%v] %+v", i, j, w.layers[i][j])
		}
	}

	t.Logf("max steps %v", w.maxStep)

	defer w.Stop()
	go w.Run()

	var count int32 = 1
	var c int32
	var ticker = time.NewTicker(time.Second)
	var flag bool
	for c < count*1 {
		select {
		case d := <-w.C:
			d.ExecuteFunc()
			c += 1
		case <-ticker.C:
			if flag {
				break
			}
			now := time.Now()
			for i := int32(0); i < count; i++ {
				duration := timerMaxDuration - 10*i*interval
				id := w.Add(time.Duration(duration)*timeUnit, fun, []any{duration, now})
				if id == 0 {
					t.Fatalf("@@@ Add failed with timeout %v", duration)
					return
				}
				/*duration = timerMaxDuration / 2
				id = w.Add(time.Duration(duration)*timeUnit, fun, []any{duration, now})
				if id == 0 {
					t.Fatalf("@@@ Add failed with timeout %v", duration)
					return
				}*/
			}
			flag = true
		}
	}

	t.Logf("complete!!!")
}

func TestWheel(t *testing.T) {
	const (
		timeUnit                = time.Millisecond
		interval          int32 = 5
		timerMaxDuration  int32 = 20000 * interval
		addTickerDuration int32 = 200 * interval
		rmTickerDuration  int32 = 200 * interval
		testDuration      int32 = 30000 * interval
		resetDuration     int32 = timerMaxDuration * 2
	)
	var (
		w     = NewWheel(time.Duration(timerMaxDuration)*timeUnit, WithInterval(time.Duration(interval)*timeUnit))
		timer = time.NewTimer(time.Duration(testDuration) * timeUnit)

		ran                        = rand.New(rand.NewSource(time.Now().Unix()))
		n                   uint32 = uint32(interval) * 20
		ac                         = 0
		loop                       = true
		pauseTicker                = false
		timerReset                 = false
		minIdCount, idCount uint32
		en, tn, rn          uint32
	)

	for i := 0; i < len(w.layers); i++ {
		for j := 0; j < len(w.layers[i]); j++ {
			t.Logf("layer[%v][%v] %+v", i, j, w.layers[i][j])
		}
	}

	t.Logf("max steps %v", w.maxStep)

	go w.Run()
	ticker := time.NewTicker(time.Duration(addTickerDuration) * timeUnit)
	rmTicker := time.NewTicker(time.Duration(rmTickerDuration) * timeUnit)
	rmTicker.C = nil

	for loop {
		select {
		case d := <-w.C:
			d.ExecuteFunc()
		case <-ticker.C:
			if pauseTicker {
				break
			}
			var fun = TimerFunc(func(args []any) {
				en += 1
				r := args[0].(int32)
				startTime := args[1].(time.Time)
				triggerTime := args[2].(time.Time)
				yt := (time.Duration(r) * timeUnit).Milliseconds()
				st := time.Since(startTime).Milliseconds()
				tt := triggerTime.Sub(startTime).Milliseconds()
				if yt > st {
					t.Fatalf("yt(%v) > st(%v)", yt, st)
				}
				t.Logf("executed (total count: %v, count: %v, to remove count: %v) timer func with timeout %+v, cost %v ms (diff %v), trigger %v ms (diff %v)", tn, en, rn, yt, st, st-yt, tt, tt-yt)
			})
			now := time.Now()
			for i := 0; i < int(n); i++ {
				r := interval + ran.Int31n(timerMaxDuration-interval)
				cc := ran.Int31n(2)
				if cc == 0 {
					if !w.Post(time.Duration(r)*timeUnit, fun, []any{r, now}) {
						t.Fatalf("@@@ Post failed with timeout %v", r)
						continue
					}
					ac += 1
					//t.Logf("@@@ Post timer func with timeout %+v and steps %v, added %v timer", time.Duration(r)*timeUnit, r/interval, ac)
				} else {
					id := w.Add(time.Duration(r)*timeUnit, fun, []any{r, now})
					if id == 0 {
						t.Fatalf("@@@ Add failed with timeout %v", r)
						continue
					}
					ac += 1
					//t.Logf("@@@ Add timer func with id %v timeout %+v and steps %v, added %v timer", id, time.Duration(r)*timeUnit, r/interval, ac)
					if minIdCount == 0 || minIdCount > id {
						minIdCount = id
					}
					idCount = id
				}
			}
			tn += n
		case <-rmTicker.C:
			if minIdCount > 0 && idCount-minIdCount >= 10 {
				id := minIdCount + uint32(ran.Int63n(int64(idCount-minIdCount)))
				w.Cancel(id)
				minIdCount = 0
				rn += 1
				t.Logf("@@@ to remove timer %v", id)
			}
		case <-timer.C:
			if !timerReset {
				timer.Reset(time.Duration(resetDuration) * timeUnit)
				timerReset = true
				pauseTicker = true
				t.Logf("timer reset, and ticker pause")
			} else {
				ticker.Stop()
				w.Stop()
				loop = false
			}
		}
	}

	timer.Stop()

	for i := 0; i < len(w.layers); i++ {
		for j := 0; j < len(w.layers[i]); j++ {
			t.Logf("Wheel layers:  i %v,  j %v,  length %v,  slots %+v", i, j, w.layers[i][j].length, w.layers[i][j].slots)
		}
	}
	t.Logf("Wheel length id2Pos %v", len(w.id2Pos))
}

func TestWheelCancelTimer(t *testing.T) {
	const (
		interval = time.Second
		period   = time.Minute
	)
	var w = NewWheel(period, WithInterval(interval))
	defer w.Stop()
	go w.Run()

	timeout := 5 * time.Second
	var tid = w.Add(timeout, func(args []any) {
		t.Logf("timer timeout after %v", timeout)
	}, nil)

	time.Sleep(time.Second)

	w.Cancel(tid)

	time.Sleep(time.Second)
}

func TestWheelX(t *testing.T) {
	const (
		timeUnit                = time.Millisecond
		interval          int32 = 10
		timerMaxDuration  int32 = 2000 * interval
		addTickerDuration int32 = 200 * interval
		rmTickerDuration  int32 = 200 * interval
		testDuration      int32 = 3000 * interval
		resetDuration     int32 = timerMaxDuration * 2
	)

	var (
		ran                        = rand.New(rand.NewSource(time.Now().Unix()))
		n                   uint32 = uint32(interval) * 10
		ac                         = 0
		loop                       = true
		pauseTicker                = false
		timerReset                 = false
		minIdCount, idCount uint32
		en, tn, rn          uint32
	)

	var fun = TimerFunc(func(args []any) {
		en += 1
		r := args[0].(int32)
		startTime := args[1].(time.Time)
		triggerTime := args[2].(time.Time)
		yt := (time.Duration(r) * timeUnit).Milliseconds()
		st := time.Since(startTime).Milliseconds()
		ts := triggerTime.Sub(startTime).Milliseconds()
		if yt > st {
			t.Fatalf("yt(%v) > st(%v)", yt, st)
		}
		t.Logf("executed (total count: %v, count: %v, to remove count: %v) timer func with timeout %+v, cost %v ms (diff %v), trigger %v ms (diff %v)", tn, en, rn, yt, st, st-yt, ts, ts-yt)
	})

	wheelX := NewWheelX(time.Duration(timerMaxDuration)*timeUnit, WithInterval(time.Duration(interval)*timeUnit))
	defer wheelX.Stop()
	go wheelX.Run()

	for i := 0; i < len(wheelX.layers); i++ {
		for j := 0; j < len(wheelX.layers[i]); j++ {
			t.Logf("layer[%v][%v] %+v", i, j, wheelX.layers[i][j])
		}
	}

	t.Logf("max steps %v", wheelX.maxStep)

	ticker := time.NewTicker(time.Duration(addTickerDuration) * timeUnit)
	rmTicker := time.NewTicker(time.Duration(rmTickerDuration) * timeUnit)
	rmTicker.C = nil
	timer := time.NewTimer(time.Duration(testDuration) * timeUnit)

	requester := wheelX.NewRequester()
	for loop {
		select {
		case <-ticker.C:
			if pauseTicker {
				break
			}
			now := time.Now()
			for i := 0; i < int(n); i++ {
				r := interval + ran.Int31n(timerMaxDuration-interval)
				cc := ran.Int31n(2)
				if cc == 0 {
					if !requester.Post(time.Duration(r)*timeUnit, fun, []any{r, now}) {
						t.Fatalf("@@@ Post failed with timeout %v", r)
						continue
					}
					ac += 1
					//t.Logf("@@@ Post timer func with timeout %+v and steps %v, added %v timer", time.Duration(r)*timeUnit, r/interval, ac)
				} else {
					id := requester.Add(time.Duration(r)*timeUnit, fun, []any{r, now})
					if id == 0 {
						t.Fatalf("@@@ Add failed with timeout %v", r)
						continue
					}
					ac += 1
					//t.Logf("@@@ Add timer func with id %v timeout %+v and steps %v, added %v timer", id, time.Duration(r)*timeUnit, r/interval, ac)
					if minIdCount == 0 || minIdCount > id {
						minIdCount = id
					}
					idCount = id
				}
			}
			tn += n
		case <-rmTicker.C:
			if minIdCount > 0 && idCount-minIdCount >= 10 {
				id := minIdCount + uint32(ran.Int63n(int64(idCount-minIdCount)))
				requester.Cancel(id)
				minIdCount = 0
				rn += 1
				t.Logf("@@@ to remove timer %v", id)
			}
		case <-timer.C:
			if !timerReset {
				timer.Reset(time.Duration(resetDuration) * timeUnit)
				timerReset = true
				pauseTicker = true
				t.Logf("timer reset, and ticker pause")
			} else {
				ticker.Stop()
				wheelX.Stop()
				loop = false
			}
		default:
			r, o := requester.GetResult()
			if o {
				r.ExecuteFunc()
			} else {
				time.Sleep(time.Microsecond)
			}
		}
	}

	timer.Stop()

	for i := 0; i < len(wheelX.layers); i++ {
		for j := 0; j < len(wheelX.layers[i]); j++ {
			t.Logf("Wheel layers:  i %v,  j %v,  length %v,  slots %+v", i, j, wheelX.layers[i][j].length, wheelX.layers[i][j].slots)
		}
	}
	t.Logf("Wheel length id2Pos %v", len(wheelX.id2Pos))
}

func TestSWheel(t *testing.T) {
	const (
		timeUnit                = time.Millisecond
		interval          int32 = 10
		timerMaxDuration  int32 = 2000000 * interval
		addTickerDuration int32 = 100 * interval
		rmTickerDuration  int32 = 100 * interval
		testDuration      int32 = 3000 * interval
		resetDuration     int32 = timerMaxDuration / 1000 * 2
	)

	var (
		ran                        = rand.New(rand.NewSource(time.Now().Unix()))
		n                   uint32 = uint32(interval) * 100
		ac                         = 0
		loop                       = true
		pauseTicker                = false
		timerReset                 = false
		minIdCount, idCount uint32
		en, tn, rn          uint32
	)

	var fun = TimerFunc(func(args []any) {
		en += 1
		r := args[0].(int32)
		startTime := args[1].(time.Time)
		triggerTime := args[2].(time.Time)
		yt := (time.Duration(r) * timeUnit).Milliseconds()
		st := time.Since(startTime).Milliseconds()
		ts := triggerTime.Sub(startTime).Milliseconds()
		if yt > st {
			t.Fatalf("yt(%v) > st(%v)", yt, st)
		}
		t.Logf("executed (total count: %v, count: %v, to remove count: %v) timer func with timeout %+v, cost %v ms (diff %v), trigger %v ms (diff %v)", tn, en, rn, yt, st, st-yt, ts, ts-yt)
	})

	wheel := NewSWheel(time.Duration(timerMaxDuration)*timeUnit, WithInterval(time.Duration(interval)*timeUnit))

	for i := 0; i < len(wheel.layers); i++ {
		for j := 0; j < len(wheel.layers[i]); j++ {
			t.Logf("layer[%v][%v] %+v", i, j, wheel.layers[i][j])
		}
	}

	t.Logf("max steps %v", wheel.maxStep)

	ticker := time.NewTicker(time.Duration(addTickerDuration) * timeUnit)
	rmTicker := time.NewTicker(time.Duration(rmTickerDuration) * timeUnit)
	rmTicker.C = nil
	timer := time.NewTimer(time.Duration(testDuration) * timeUnit)

	wheel.Start()

	for loop {
		select {
		case <-ticker.C:
			if pauseTicker {
				break
			}
			now := time.Now()
			for i := 0; i < int(n); i++ {
				r := interval + ran.Int31n(timerMaxDuration/1000-interval)
				cc := ran.Int31n(2)
				if cc == 0 {
					if !wheel.Post(time.Duration(r)*timeUnit, fun, []any{r, now}) {
						t.Fatalf("@@@ Post failed with timeout %v", r)
						continue
					}
					ac += 1
					//t.Logf("@@@ Post timer func with timeout %+v and steps %v, added %v timer", time.Duration(r)*timeUnit, r/interval, ac)
				} else {
					id := wheel.Add(time.Duration(r)*timeUnit, fun, []any{r, now})
					if id == 0 {
						t.Fatalf("@@@ Add failed with timeout %v", r)
						continue
					}
					ac += 1
					//t.Logf("@@@ Add timer func with id %v timeout %+v and steps %v, added %v timer", id, time.Duration(r)*timeUnit, r/interval, ac)
					if minIdCount == 0 || minIdCount > id {
						minIdCount = id
					}
					idCount = id
				}
			}
			tn += n
		case <-rmTicker.C:
			if minIdCount > 0 && idCount-minIdCount >= 10 {
				id := minIdCount + uint32(ran.Int63n(int64(idCount-minIdCount)))
				wheel.Cancel(id)
				minIdCount = 0
				rn += 1
				t.Logf("@@@ to remove timer %v", id)
			}
		case <-timer.C:
			if !timerReset {
				timer.Reset(time.Duration(resetDuration) * timeUnit)
				timerReset = true
				pauseTicker = true
				t.Logf("timer reset, and ticker pause")
			} else {
				ticker.Stop()
				loop = false
			}
		default:
			if !wheel.Update() {
				time.Sleep(time.Microsecond)
			}
		}
	}

	timer.Stop()

	for i := 0; i < len(wheel.layers); i++ {
		for j := 0; j < len(wheel.layers[i]); j++ {
			t.Logf("Wheel layers:  i %v,  j %v,  length %v,  slots %+v", i, j, wheel.layers[i][j].length, wheel.layers[i][j].slots)
		}
	}
	t.Logf("Wheel length id2Pos %v", len(wheel.id2Pos))
}
