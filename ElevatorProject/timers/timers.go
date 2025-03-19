package timers

import (
	"fmt"
	"time"
)

const MaxDuration time.Duration = 1<<63 - 1
const DoorOpenDuration time.Duration = 3.0 * time.Second
const heartbeatTimeout time.Duration = 10 * time.Millisecond // just a suggestion


func Timer_start(timer *time.Timer, start_timer chan time.Duration) {
	for {
		select {
		case duration := <-start_timer:
			fmt.Println("Starting timer with duration: ", duration)
			timer.Reset(duration)
		}
	}
}

