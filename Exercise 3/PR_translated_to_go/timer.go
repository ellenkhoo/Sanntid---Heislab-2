package main

import (
	"time"
)

//Her har jeg noe usikkerhet om varaibeltypene blir riktige

func get_wall_time() float64 {
	now := time.Now()
	seconds := float64(now.Unix())
	microseconds := float64(now.Nanosecond()) / 1e9
	return seconds + microseconds
}

var timerEndTime float64
var timerActive int

func timer_start(duration float64) {
	timerEndTime = get_wall_time() + duration
	timerActive = 1
}

func timer_stop() {
	timerActive = 0
}

func timer_timedOut() bool {
	return (timerActive == 1 && get_wall_time() > timerEndTime)
}
