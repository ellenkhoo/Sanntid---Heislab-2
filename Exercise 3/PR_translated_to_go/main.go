package main

import (
	//"PR_translated_to_go/elevator_io_device"
	"Driver/elevio"
	"fmt"
	"time"
)

// Må bruke channels, ikke gjort per nå!

// Skal det deklareres her, eller er det definert i en av de inkluderte filene?
// const N_FLOORS = 4
// const N_BUTTONS = 3

// func main() int ?
func main() {
	fmt.Println("Started!")

	// burde vel ikke måtte definere denne på nytt, er jo definert i elevio
	numFloors := 4
	elevio.Init("localhost:15657", numFloors)

	// Initialize fsm
	// Tror ikke dette egt. er riktig, og at dette med fsm ikke var nødvendig for én heis
	fsm := FSM{el: Elevator{Floor: 0, Dirn: D_Stop, Behaviour: EB_Idle}, od: elevio_getOutputDevice()}

	//inputPollRate_ms := 25
	// Hva gjør dette egt?
	//con_load("elevator.con", con_val("inputPollRate_ms", &inputPollRate_ms, "%d"))

	// Initialize channels
	buttons_chan := make(chan elevio.ButtonEvent)
	floors_chan := make(chan int)
	obstruction_chan := make(chan bool)
	stop_chan := make(chan bool)
	start_timer := make(chan time.Duration)

	// Initialize timer, stop it until needed
	main_timer := time.NewTimer(time.Duration(fsm.el.Config.DoorOpenDuration))
	main_timer.Stop()

	// Start goroutines
	go elevio.PollButtons(buttons_chan)
	go elevio.PollFloorSensor(floors_chan)
	go elevio.PollObstructionSwitch(obstruction_chan)
	go elevio.PollStopButton(stop_chan)
	go timer_start(main_timer, start_timer)

	//floor_input := <-floors_chan
	//input := elevio_getInputDevice()

	if elevio_getInputDevice().FloorSensor() == -1 {
		fsm.fsm_onInitBetweenFloors()
	}

	// Erstatter med channels
	// for {
	// 	{ // Request button
	// 		var prev [N_FLOORS][N_BUTTONS]int
	// 		for floor := 0; floor < N_FLOORS; floor++ {
	// 			for b := 0; b < N_BUTTONS; b++ {
	// 				button := Button(b)
	// 				v := input.RequestButton(floor, button)
	// 				if v != 0 && v != prev[floor][button] {
	// 					fsm.fsm_onRequestButtonPress(floor, button)
	// 				}
	// 				prev[floor][button] = v
	// 			}
	// 		}
	// 	}

	// 	{ // Floor sensor
	// 		prev := -1
	// 		f := input.FloorSensor()
	// 		if f != -1 && f != prev {
	// 			fsm.fsm_onFloorArrival(f)
	// 		}
	// 		prev = f
	// 	}

	// 	{ // Timer
	// 		if timer_timedOut() {
	// 			timer_stop()
	// 			fsm.fsm_onDoorTimeout()
	// 		}
	// 	}

	// 	time.Sleep(time.Duration(inputPollRate_ms) * time.Millisecond)

	// }

	for {
		select {
		case button_pushed := <-buttons_chan:
			elevio.SetButtonLamp(button_pushed.Button, button_pushed.Floor, true)
			fsm.el.Requests[button_pushed.Floor][button_pushed.Button] = true
			fsm.fsm_onRequestButtonPress(button_pushed.Floor, button_pushed.Button, start_timer)

		case floor_input := <-floors_chan:
			elevio.SetFloorIndicator(floor_input)

			prev := -1
			if floor_input != -1 && floor_input != prev {
				fsm.fsm_onFloorArrival(floor_input, start_timer)
			}

		case <-main_timer.C:
			fsm.fsm_onDoorTimeout(start_timer)

		}
	}
}
