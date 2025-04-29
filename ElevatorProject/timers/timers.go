package timers

import (
	"time"
)

const MaxDuration time.Duration = 1<<63 - 1
const DoorOpenDuration time.Duration = 3.0 * time.Second

func StartTimer(timer *time.Timer, timerChan chan time.Duration) {
	for {
		select {
		case duration := <-timerChan:
			timer.Reset(duration)
		}
	}
}
