package main

import (
	//"PR_translated_to_go/elevator_io_device"
	"fmt"
	"time"
)

// Må bruke channels, ikke gjort per nå!

// Skal det deklareres her, eller er det definert i en av de inkluderte filene?
// const N_FLOORS = 4
// const N_BUTTONS = 3

// func main() int ?
func main() {
	fmt.Println("Started!\n")

	inputPollRate_ms := 25
	con_load("elevator.con", con_val("inputPollRate_ms", &inputPollRate_ms, "%d"))

	input := elevator_io_device.ElevInputDevice()

	if input.floorSensor() == -1 {
		fsm_onInitBetweenFloors()
	}

	for {
		{ // Request button
			var prev [N_FLOORS][N_BUTTONS]int
			for f := 0; f < N_FLOORS; f++ {
				for b := 0; b < N_BUTTONS; b++ {
					v := input.requestButton(f, b)
					if v && v != prev[f][b] {
						fsm_onRequestButtonPress(f, b)
					}
					prev[f][b] = v
				}
			}
		}

		{ // Floor sensor
			prev := -1
			f := input.floorSensor()
			if f != -1 && f != prev {
				fsm_onFloorArrival(f)
			}
			prev = f
		}

		{ // Timer
			if timer_timedOut() {
				timer_stop()
				fsm_onDoorTimeout()
			}
		}

		time.Sleep(time.Duration(inputPollRate_ms) * time.Millisecond)
	}
}
