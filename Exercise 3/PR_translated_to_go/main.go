package main

import (
	//"PR_translated_to_go/elevator_io_device"
	"Driver-go/elevio"
	elevatorpkg "elevator"
	elevator_io_devicepkg "elevator_io_device"
	"fmt"
	fsmpkg "fsm"
	requestpkg "request"
	"time"
	timerpkg "timer"
)

// func main() int ?
func main() {

	fmt.Println("Started!")

	// burde vel ikke måtte definere denne på nytt, er jo definert i elevio
	numFloors := 4
	const maxDuration time.Duration = 1<<63 - 1

	elevio.Init("localhost:15657", numFloors)

	fsm := fsmpkg.FSM{El: elevatorpkg.Elevator_uninitialized(), Od: elevator_io_devicepkg.Elevio_getOutputDevice()}

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

	requestpkg.Clear_all_requests(fsm.El)
	fsm.SetAllLights()

	if elevio.GetFloor() == -1 {
		fsm.Fsm_onInitBetweenFloors()
		fmt.Printf("Init between floor")
	}

	//fsm.SetAllLights()
	fmt.Printf("Current floor: %d \n", fsm.El.Floor)
	fmt.Printf("Current Dirn: %d \n", fsm.El.Dirn)

	var prevFloor = -1
	for {
		select {
		case order := <-buttons_chan:
			fmt.Print("Button pushed")
			if !(fsm.El.Requests[order.Floor][order.Button]) {
				fsm.Fsm_onRequestButtonPress(order.Floor, order.Button, start_timer)
			}

			// Alt dette gjøres av "fsm_onRequestButtonPress"
			// fmt.Printf("Button pushed")
			// elevio.SetButtonLamp(order.Button, order.Floor, true)
			// fmt.Printf("Button light at floor %d set to true", order.Floor)
			// fsm.El.Requests[order.Floor][order.Button] = true
			// fmt.Printf("Request at floor %d added to queue", order.Floor)
			// fsm.Fsm_onRequestButtonPress(order.Floor, order.Button, start_timer)
			// pair := requestpkg.Requests_chooseDirection(fsm.El)
			// elevio.SetMotorDirection(elevio.MotorDirection(pair.Dirn))

		case floor_input := <-floors_chan:
			fmt.Printf("Floor sensor: %d", floor_input)

			if floor_input != -1 && floor_input != prevFloor {
				fsm.Fsm_onFloorArrival(floor_input, start_timer)
			}
			//Floor indicator skal settes i "onFloorArrival"
			// elevio.SetFloorIndicator(floor_input)
			// if floor_input != -1 && floor_input != prevFloor {
			// 	fsm.Fsm_onFloorArrival(floor_input, start_timer)
			// }
			// prevFloor = floor_input

		case obstruction := <-obstruction_chan:
			if obstruction {
				if fsm.El.Behaviour == elevatorpkg.EB_DoorOpen {
					start_timer <- maxDuration
				}
			} else {
				start_timer <- fsm.El.Config.DoorOpenDuration
			}

			// if obstruction {
			// 	elevio.SetMotorDirection(elevio.MD_Stop)
			// 	start_timer <- maxDuration

			// } else {
			// 	main_timer.Stop()
			// 	start_timer <- fsm.El.Config.DoorOpenDuration
			// }

		case <-stop_chan:
			requestpkg.Clear_all_requests(fsm.El)
			fsm.SetAllLights()

		case <-main_timer.C:
			fsm.Fsm_onDoorTimeout(start_timer)
			//elevio.SetDoorOpenLamp(false)
		}
	}
}
