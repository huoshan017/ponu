package time

import (
	"math/rand"
	"testing"
	"time"
)

func TestWheel(t *testing.T) {
	var (
		interval = 20 * time.Millisecond
		w        = NewWheel(interval, time.Minute)
		ticker   = time.NewTicker(100 * interval)
		timer    = time.NewTimer(10000 * interval)

		ran         = rand.New(rand.NewSource(time.Now().Unix()))
		n           = 100000
		c           = 0
		loop        = true
		pauseTicker = false
		timerReset  = false
	)

	for loop {
		select {
		case d := <-w.C:
			d.ExecuteFunc()
		case <-ticker.C:
			if pauseTicker {
				break
			}
			for i := 0; i < n; i++ {
				r := time.Duration(1+ran.Int31n(2999)) * interval
				w.AddNoId(r, TimerFunc(func(arg any) {
					c += 1
				}), nil)
			}
		case <-timer.C:
			if !timerReset {
				timer.Reset(time.Minute)
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

	t.Logf("executed %v", c)

	for i := 0; i < len(w.layers); i++ {
		for j := 0; j < len(w.layers[i]); j++ {
			t.Logf("Wheel layers:  i %v,  j %v,  curr_slot %v,  slots %+v", i, j, w.layers[i][j].curr_slot, w.layers[i][j].slots)
			for k := 0; k < len(w.layers[i][j].slots); k++ {
				if w.layers[i][j].slots[k] != nil {
					t.Logf("     w.layers[%v][%v].slots[%v] = %+v", i, j, k, w.layers[i][j].slots[k])
				}
			}
		}
	}
	t.Logf("Wheel id2Pos %+v", w.id2Pos)
}
