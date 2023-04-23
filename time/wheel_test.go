package time

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "net/http/pprof"
)

func TestWheelOneStep(t *testing.T) {
	const (
		timeUnit         = time.Millisecond
		interval         = 10
		timerMaxDuration = 2000 * interval
	)

	w := NewWheel(time.Duration(timerMaxDuration)*timeUnit, WithInterval(time.Duration(interval)*timeUnit))
	defer w.Stop()

	t.Logf("max steps %v", w.maxStep)

	go w.Run()

	var fun = func(id uint32, args []any) {
		r := args[0].(int)
		startTime := args[1].(time.Time)
		triggerTime := args[2].(time.Time)
		yt := (time.Duration(r) * timeUnit).Milliseconds()
		st := time.Since(startTime).Milliseconds()
		tt := triggerTime.Sub(startTime).Milliseconds()
		if yt > st {
			t.Errorf("yt(%v) > st(%v)", yt, st)
		}
		t.Logf("executed timer func with timeout %+v, cost %v ms (diff %v), trigger %v ms (diff %v)", yt, st, st-yt, tt, tt-yt)
	}
	sender := w.NewSender()
	tid := sender.Add(interval*timeUnit, fun, []any{interval, time.Now()})
	t.Logf("added timer %v", tid)
	var (
		loop bool = true
	)
	for loop {
		tl, o := sender.GetTimerList()
		if o {
			tl.ExecuteFunc()
			loop = false
		}
	}
	time.Sleep(time.Second)
}

func TestWheel(t *testing.T) {
	const (
		timeUnit                = time.Millisecond
		interval          int32 = 10
		timerMaxDuration  int32 = 2000 * interval
		addTickerDuration int32 = 40 * interval
		rmTickerDuration  int32 = 400 * interval
		testDuration      int32 = 3000 * interval
		resetDuration     int32 = timerMaxDuration
	)

	/*go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()*/

	w := NewWheel(time.Duration(timerMaxDuration)*timeUnit, WithInterval(time.Duration(interval)*timeUnit), WithTimerRecvListLength(4096*10), WithSenderListLength(1024))
	defer w.Stop()

	t.Logf("max steps %v", w.maxStep)

	go w.Run()

	var (
		tcn, ycn, cn, lcn int32
		c                 int = 20
		wg                sync.WaitGroup
	)
	wg.Add(c)
	for i := 0; i < c; i++ {
		go func(index int) {
			var (
				timer                      = time.NewTimer(time.Duration(testDuration) * timeUnit)
				ran                        = rand.New(rand.NewSource(time.Now().Unix()))
				n                   uint32 = uint32(interval) * 100
				ac                         = 0
				loop                       = true
				pauseTicker                = false
				timerReset                 = false
				minIdCount, idCount uint32
				en, tn, rn          uint32
			)

			ticker := time.NewTicker(time.Duration(addTickerDuration) * timeUnit)
			rmTicker := time.NewTicker(time.Duration(rmTickerDuration) * timeUnit)
			rmTicker.C = nil
			sender := w.NewSender()

			var fun = TimerFunc(func(id uint32, args []any) {
				en += 1
				r := args[0].(int32)
				startTime := args[1].(time.Time)
				triggerTime := args[2].(time.Time)
				yt := (time.Duration(r) * timeUnit).Milliseconds()
				st := time.Since(startTime).Milliseconds()
				tt := triggerTime.Sub(startTime).Milliseconds()
				exec_diff := st - yt
				trigger_diff := tt - yt
				if yt > st {
					t.Errorf("yt(%v) > st(%v)", yt, st)
				}
				if exec_diff > int64(interval) {
					atomic.AddInt32(&ycn, 1)
					if exec_diff > trigger_diff*2 && trigger_diff > 2*int64(interval) {
						//t.Logf("timer id (%v) index(%v) executed (total count: %v, count: %v, to remove count: %v) timer func with timeout %+v, cost %v ms (diff %v), trigger %v ms (diff %v)",
						//	id, index, tn, en, rn, yt, st, st-yt, tt, tt-yt)
						atomic.AddInt32(&cn, 1)
						if exec_diff > trigger_diff*5 && trigger_diff > 2*int64(interval) {
							atomic.AddInt32(&lcn, 1)
						}
					}
				}
				atomic.AddInt32(&tcn, 1)
			})

			for loop {
				select {
				case <-ticker.C:
					if pauseTicker {
						break
					}

					for i := 0; i < int(n); i++ {
						r := interval + ran.Int31n(timerMaxDuration-interval)
						now := time.Now()
						id := sender.Add(time.Duration(r)*timeUnit, fun, []any{r, now})
						if id == 0 {
							t.Errorf("@@@ Add failed with timeout %v", r)
							continue
						}
						ac += 1
						if minIdCount == 0 || minIdCount > id {
							minIdCount = id
						}
						idCount = id
					}
					tn += n
				case <-rmTicker.C:
					if minIdCount > 0 && idCount-minIdCount >= 10 {
						id := minIdCount + uint32(ran.Int63n(int64(idCount-minIdCount)))
						sender.Cancel(id)
						minIdCount = 0
						rn += 1
						t.Logf("@@@ index(%v) to remove timer %v", index, id)
					}
				case <-timer.C:
					if !timerReset {
						timer.Reset(time.Duration(resetDuration) * timeUnit)
						timerReset = true
						pauseTicker = true
						t.Logf("index(%v)  timer reset, and ticker pause", index)
					} else {
						ticker.Stop()
						loop = false
					}
				default:
					var (
						o  bool = true
						tl TimerList
					)
					for o {
						tl, o = sender.GetTimerList()
						if !o {
							break
						}
						tl.ExecuteFunc()
					}
				}
			}

			timer.Stop()
			wg.Done()
		}(i)
	}

	wg.Wait()

	for i := 0; i < len(w.layers); i++ {
		for j := 0; j < len(w.layers[i]); j++ {
			t.Logf("Wheel layers:  i %v,  j %v,  length %v,  slots %+v", i, j, w.layers[i][j].length, w.layers[i][j].slots)
		}
	}
	t.Logf("Wheel length id2Pos %v", len(w.id2Pos))
	t.Logf("@@@ tcn = %v,  ycn = %v,  cn = %v,  lcn = %v", tcn, ycn, cn, lcn)
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
	var tid = w.Add(timeout, func(id uint32, args []any) {
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
		resetDuration     int32 = timerMaxDuration
	)

	wheelX := NewWheelX(time.Duration(timerMaxDuration)*timeUnit, WithInterval(time.Duration(interval)*timeUnit))
	defer wheelX.Stop()
	go wheelX.Run()

	for i := 0; i < len(wheelX.layers); i++ {
		for j := 0; j < len(wheelX.layers[i]); j++ {
			t.Logf("layer[%v][%v] %+v", i, j, wheelX.layers[i][j])
		}
	}

	t.Logf("max steps %v", wheelX.maxStep)

	var c int = 1
	var wg sync.WaitGroup
	wg.Add(c)
	for i := 0; i < c; i++ {
		go func(t *testing.T, index int) {
			var (
				ran                        = rand.New(rand.NewSource(time.Now().Unix()))
				n                   uint32 = uint32(interval)
				ac                         = 0
				loop                       = true
				pauseTicker                = false
				timerReset                 = false
				minIdCount, idCount uint32
				en, tn, rn          uint32
			)
			var fun = TimerFunc(func(id uint32, args []any) {
				en += 1
				r := args[0].(int32)
				startTime := args[1].(time.Time)
				triggerTime := args[2].(time.Time)
				yt := (time.Duration(r) * timeUnit).Milliseconds()
				st := time.Since(startTime).Milliseconds()
				ts := triggerTime.Sub(startTime).Milliseconds()
				if yt > st {
					t.Errorf("index(%v)  yt(%v) > st(%v)", index, yt, st)
				}
				if st-yt > int64(1.5*float32(interval)) {
					t.Logf("index(%v)  executed (total count: %v, count: %v, to remove count: %v) timer func with timeout %+v, cost %v ms (diff %v), trigger %v ms (diff %v)", index, tn, en, rn, yt, st, st-yt, ts, ts-yt)
				}
			})
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
					for i := 0; i < int(n); i++ {
						r := interval + ran.Int31n(timerMaxDuration-interval)
						cc := ran.Int31n(2)
						if cc == 0 {
							now := time.Now()
							if !requester.Post(time.Duration(r)*timeUnit, fun, []any{r, now}) {
								t.Errorf("@@@ Post failed with timeout %v", r)
								continue
							}
							ac += 1
						} else {
							now := time.Now()
							id := requester.Add(time.Duration(r)*timeUnit, fun, []any{r, now})
							if id == 0 {
								t.Errorf("@@@ Add failed with timeout %v", r)
								continue
							}
							ac += 1
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
						t.Logf("index(%v) timer reset, and ticker pause", index)
					} else {
						ticker.Stop()
						loop = false
					}
				default:
					requester.Update()
				}
			}
			timer.Stop()
			wg.Done()
		}(t, i)
	}

	wg.Wait()

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
		n                   uint32 = uint32(interval) * 1000
		ac                         = 0
		loop                       = true
		pauseTicker                = false
		timerReset                 = false
		minIdCount, idCount uint32
		en, tn, rn          uint32
	)

	var fun = TimerFunc(func(id uint32, args []any) {
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
		if st-yt > int64(1.5*float32(interval)) {
			t.Logf("executed (total count: %v, count: %v, to remove count: %v) timer func with timeout %+v, cost %v ms (diff %v), trigger %v ms (diff %v)", tn, en, rn, yt, st, st-yt, ts, ts-yt)
		}
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
				} else {
					id := wheel.Add(time.Duration(r)*timeUnit, fun, []any{r, now})
					if id == 0 {
						t.Fatalf("@@@ Add failed with timeout %v", r)
						continue
					}
					ac += 1
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
				time.Sleep(time.Nanosecond)
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
