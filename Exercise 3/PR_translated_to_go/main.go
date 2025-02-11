package main

import (
	//"PR_translated_to_go/elevator_io_device"
	"Driver-go/elevio"
	"elevator"
	"elevator_io_device"
	"fmt"
	"fsm"
	requestpkg "request"
	"time"
	"timer"
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
	const maxDuration time.Duration = 1<<63-1

	elevio.Init("localhost:15657", numFloors)



	
	// Initialize fsm
	// Tror ikke dette egt. er riktig, og at dette med fsm ikke var nødvendig for én heis
	fsm := fsmpkg.FSM{El: elevatorpkg.Elevator_uninitialized(), Od: elevator_io_devicepkg.Elevio_getOutputDevice()}

	//Kanskje bruk denne under om heisen skal init-es med floor = 1
	// elevatorpkg.Elevator{Floor: 0, Dirn: elevator_io_devicepkg.D_Stop, Behaviour: elevatorpkg.EB_Idle}

	// fsm.El.Behaviour = elevatorpkg.EB_Idle
	// fsm.El.Dirn = elevator_io_devicepkg.D_Stop
	var d elevio.MotorDirection
	fsm.Fsm_onInitBetweenFloors()
	fmt.Printf("Init between floor")

	fsm.SetAllLights()
	fmt.Printf("Current floor: %d \n", fsm.El.Floor)
	fmt.Printf("Current Dirn: %d \n", fsm.El.Dirn)

	// for f := fsm.El.Floor + 1; f < elevatorpkg.N_FLOORS; f++ {
	// 	for btn := 0; btn < elevatorpkg.N_BUTTONS; btn++ {
	// 		fsm.El.Requests[f][btn] = false
	// 		elevio.SetButtonLamp(elevator_io_devicepkg.Button(btn), f, false)
	// 	}
	// }

	// Skulle man hatt en "clear all requests"-funksjon? Når jeg prøver å kjøre programmet, kjører heisen bare opp.

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
	main_timer := time.NewTimer(time.Duration(fsm.El.Config.DoorOpenDuration))
	main_timer.Stop()

	// Start goroutines
	go elevio.PollButtons(buttons_chan)
	go elevio.PollFloorSensor(floors_chan)
	go elevio.PollObstructionSwitch(obstruction_chan)
	go elevio.PollStopButton(stop_chan)
	go timerpkg.Timer_start(main_timer, start_timer)

	//floor_input := <-floors_chan
	//input := elevio_getInputDevice()

	// if elevio_getInputDevice().FloorSensor() == -1 {
	// 	fsm.fsm_onInitBetweenFloors()
	// }

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
			fmt.Printf("Button pushed")
			elevio.SetButtonLamp(button_pushed.Button, button_pushed.Floor, true)
			fmt.Printf("Button light at floor %d set to true", button_pushed.Floor)
			fsm.El.Requests[button_pushed.Floor][button_pushed.Button] = true
			fmt.Printf("Request at floor %d added to queue", button_pushed.Floor)
			fsm.Fsm_onRequestButtonPress(button_pushed.Floor, button_pushed.Button, start_timer)
			pair := requestpkg.Requests_chooseDirection(fsm.El)
			elevio.SetMotorDirection(elevio.MotorDirection(pair.Dirn))


		case floor_input := <-floors_chan:
			elevio.SetFloorIndicator(floor_input)

			prev := -1
			if floor_input != -1 && floor_input != prev {
				fsm.Fsm_onFloorArrival(floor_input, start_timer)
			}

		case obstruction := <-obstruction_chan:
			if obstruction {
				elevio.SetMotorDirection(elevio.MD_Stop)
				start_timer <- maxDuration
			} else {
				start_timer <- fsm.El.Config.DoorOpenDuration
				elevio.SetMotorDirection(d)
			}

		// case stop := <-stop_chan:
		// 	for f := 0; f < numFloors; f++ {
		// 		for b := elevio.ButtonType(0); b < 3; b++ {	
		// 			elevio.SetButtonLamp(b, f, false) 
		// 		}
		// 	}
		case <-main_timer.C:
			fsm.Fsm_onDoorTimeout(start_timer)

		}
	}
}