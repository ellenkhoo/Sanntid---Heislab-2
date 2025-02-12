package timerpkg

import (
	"fmt"
	"time"
)

var timerEndTime time.Time
var timerActive bool


func Timer_start(timer *time.Timer, start_timer chan time.Duration) {
	for {
		select {
		case duration := <-start_timer:
			fmt.Println("Duration: ", duration)
			timer.Reset(duration)
		}
	}

	// Bør håndtere dør-operasjoner her, siden det bare er avhengig av timeren?
}

func Timer_stop() {
	timerActive = false
}

func Timer_timedOut() bool {
	return (timerActive && time.Now().After(timerEndTime))
}
