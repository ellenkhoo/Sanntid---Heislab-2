package timerpkg

import (
	"fmt"
	"time"
)

//Her har jeg noe usikkerhet om varaibeltypene blir riktige

func get_wall_time() time.Time {
	// now := time.Now()
	// seconds := float64(now.Unix())
	// microseconds := float64(now.Nanosecond()) / 1e9
	// return seconds + microseconds
	return time.Now()
}

var timerEndTime time.Time
var timerActive bool

// Erstatter dette med en funksjon som bruker channels
// func timer_start(duration float64) {
// 	timerEndTime = get_wall_time() + duration
// 	timerActive = 1
// }

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
	return (timerActive == true && get_wall_time().After(timerEndTime))
}
