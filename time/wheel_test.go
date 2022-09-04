package time

import (
	"math/rand"
	"testing"
	"time"
)

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
		w        = NewWheel(time.Duration(interval)*timeUnit, time.Duration(timerMaxDuration)*timeUnit)
		ticker   = time.NewTicker(time.Duration(addTickerDuration) * timeUnit)
		rmTicker = time.NewTicker(time.Duration(rmTickerDuration) * timeUnit)
		timer    = time.NewTimer(time.Duration(testDuration) * timeUnit)

		ran                        = rand.New(rand.NewSource(time.Now().Unix()))
		n                   uint32 = uint32(interval)
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
				yt := (time.Duration(r) * timeUnit).Milliseconds()
				st := time.Since(startTime).Milliseconds()
				if yt > st {
					t.Fatalf("yt(%v) > st(%v)", yt, st)
				}
				t.Logf("executed (total count: %v, count: %v, remove count: %v) timer func with timeout %+v, cost %v ms", tn, en, rn, yt, st)
			})
			for i := 0; i < int(n); i++ {
				r := interval + ran.Int31n(timerMaxDuration-interval)
				cc := ran.Int31n(2)
				now := time.Now()
				if cc == 0 {
					if !w.AddNoId(time.Duration(r)*timeUnit, fun, []any{r, now}) {
						t.Fatalf("@@@ AddNoId failed with timeout %v", r)
						continue
					}
					ac += 1
					//t.Logf("@@@ AddNoId timer func with timeout %+v and steps %v, added %v timer", time.Duration(r)*timeUnit, r/interval, ac)
				} else {
					id := w.Add(time.Duration(r)*timeUnit, fun, []any{r, time.Now()})
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
				w.Remove(minIdCount + uint32(ran.Int63n(int64(idCount-minIdCount))))
				minIdCount = 0
				rn += 1
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
			/*
				for k := 0; k < len(w.layers[i][j].slots); k++ {
					if w.layers[i][j].slots[k] != nil {
						t.Logf("     w.layers[%v][%v].slots[%v] = %+v", i, j, k, w.layers[i][j].slots[k])
						var iter = w.layers[i][j].slots[k].Begin()
						for iter != w.layers[i][j].slots[k].End() {
							n := iter.Value().(*Timer)
							t.Logf("			node %+v", *n)
							iter = iter.Next()
						}
					}
				}
			*/
		}
	}
	t.Logf("Wheel length id2Pos %v", len(w.id2Pos))
}
