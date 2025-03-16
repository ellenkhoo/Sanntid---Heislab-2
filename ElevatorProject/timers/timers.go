package timers

import (
	"fmt"
	"time"
)

const MaxDuration time.Duration = 1<<63 - 1


func Timer_start(timer *time.Timer, start_timer chan time.Duration) {
	for {
		select {
		case duration := <-start_timer:
			fmt.Println("Starting timer with duration: ", duration)
			timer.Reset(duration)
		}
	}

	// Bør håndtere dør-operasjoner her, siden det bare er avhengig av timeren?
}

