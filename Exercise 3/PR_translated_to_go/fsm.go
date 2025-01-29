package main

import (
	
	//"con_load"
	"Driver/elevio"
	"fmt"
)

// // Elevator FSM struct
// type FSM struct {
// 	elevator     Elevator
// 	outputDevice ElevOutputDevice
// }


// // Initializes and returns an FSM instance
// func fsm_init() {

// 	//elevator := elevator_uninitialized()

// 	fsm := &FSM{
// 		elevator:     elevator_uninitialized(),
// 		outputDevice: getOutputDevice(),
// 	}

// 	con_load.LoadConfig("con", map[string]interface{}{
// 		"doorOpenDuration_s":  DoorOpenDuration,
// 		"clearRequestVariant": ClearRequestVariant,
// 	})

// 	//fsm.outputDevice = getOutputDevice()
// }

var elevator Elevator

// Set all elevator lights
func (fsm *FSM) setAllLights() {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			fsm.outputDevice.requestButtonLight(floor, btn, fsm.Requests[floor][btn])
		}
	}
}

// Handle initialization between floors
func (fsm *FSM) fsm_onInitBetweenFloors() {
	fsm.outputDevice.motorDirection(D_Down)
	fsm.Dirn = D_Down
	fsm.Behaviour = EB_Moving
}

// Handle button press event
func (fsm *FSM) fsm_onRequestButtonPress(btn_floor int, btn_type Button) {
	fmt.Printf("\n\n(%d, %s)\n", btn_floor, elevio_button_toString(btn_type))
	elevator_print(fsm.elevator)

	switch fsm.Behaviour {
	case EB_DoorOpen:
		if requests.requests_shouldClearImmediately(fsm.elevator, btn_floor, btn_type) {
			timer.timer_start(doorOpenDuration_s)
		} else {
			fsm.Requests[btn_floor][btn_type] = true
		}

	case EB_Moving:
		fsm.Requests[btn_floor][btn_type] = true

	case EB_Idle:
		fsm.Requests[btn_floor][btn_type] = true
		pair := requests.requests_chooseDirection(fsm.elevator)
		fsm.Dirn = pair.Dirn
		fsm.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			fsm.outputDevice.doorLight(true)
			timer.timer_start(config.doorOpenDuration_s)
			fsm.elevator = requests.requests_clearAtCurrentFloor(fsm.elevator)

		case EB_Moving:
			fsm.outputDevice.motorDirection(fsm.Dirn)

		case EB_Idle:
			// Do nothing
		}
	}

	fsm.setAllLights()
	fmt.Println("\nNew state:")
	fsm.elevator_print()
}

// Handle floor arrival event
func (fsm *FSM) fsm_onFloorArrival(newFloor int) {
	fmt.Printf("\n\n(%d)\n", newFloor)
	fsm.elevator_print()

	fsm.Floor = newFloor
	fsm.outputDevice.floorIndicator(fsm.Floor)

	switch fsm.Behaviour {
	case EB_Moving:
		if requests.requests_shouldStop(fsm.elevator) {
			fsm.outputDevice.motorDirection(D_Stop)
			fsm.outputDevice.doorLight(true)
			fsm.elevator = requests.requests_clearAtCurrentFloor(fsm.elevator)
			timer.timer_start(doorOpenDuration_s)
			fsm.setAllLights()
			fsm.Behaviour = EB_DoorOpen
		}
	}

	fmt.Println("\nNew state:")
	fsm.elevator_print()
}

// Handle door timeout event
func (fsm *FSM) fsm_onDoorTimeout() {
	fsm.elevator_print()

	switch fsm.Behaviour {
	case EB_DoorOpen:
		pair := requests.requests_chooseDirection(fsm.elevator)
		fsm.Dirn = pair.Dirn
		fsm.Behaviour = pair.Behaviour

		switch fsm.Behaviour {
		case EB_DoorOpen:
			timer.timer_start(doorOpenDuration_s)
			fsm.elevator = requests.requests_clearAtCurrentFloor(fsm.elevator)
			fsm.setAllLights()
		case EB_Moving, EB_Idle:
			fsm.outputDevice.doorLight(false)
			fsm.outputDevice.motorDirection(fsm.Dirn)
		}

	}

	fmt.Println("\nNew state:")
	fsm.elevator_print()
}
