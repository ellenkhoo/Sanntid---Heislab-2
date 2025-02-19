package fsmpkg

import (
	"Driver-go/elevio"
	elevatorpkg "elevator"
	elevator_io_devicepkg "elevator_io_device"
	"fmt"
	requestpkg "request"
	"time"
)

// Elevator FSM struct
type FSM struct {
	El elevatorpkg.Elevator
	Od elevator_io_devicepkg.ElevOutputDevice
}

// Set all elevator lights
func (fsm *FSM) SetAllLights() {
	print("Setting all lights\n")
	for floor := 0; floor < elevatorpkg.N_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < elevatorpkg.N_BUTTONS; btn++ {
			fsm.Od.RequestButtonLight(floor, btn, fsm.El.Requests[floor][btn])
		}
	}
}

// Handle initialization between floors
func (fsm *FSM) Fsm_onInitBetweenFloors() {
	fsm.Od.MotorDirection(elevio.MD_Up)
	fsm.El.Dirn = elevio.MD_Up
	fsm.El.Behaviour = elevatorpkg.EB_Moving
}

// Handle button press event
func (fsm *FSM) Fsm_onRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, start_timer chan time.Duration) {
	fmt.Printf("\n\n(%d, %s)\n", btn_floor, btn_type)
	elevatorpkg.Elevator_print(fsm.El)

	switch fsm.El.Behaviour {
	case elevatorpkg.EB_DoorOpen:
		if requestpkg.Requests_shouldClearImmediately(fsm.El, btn_floor, btn_type) {
			start_timer <- fsm.El.Config.DoorOpenDuration
		} else {
			fsm.El.Requests[btn_floor][btn_type] = true
		}

	case elevatorpkg.EB_Moving:
		fsm.El.Requests[btn_floor][btn_type] = true

	case elevatorpkg.EB_Idle:
		fsm.El.Requests[btn_floor][btn_type] = true
		pair := requestpkg.Requests_chooseDirection(fsm.El)
		fsm.El.Dirn = pair.Dirn
		fsm.El.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case elevatorpkg.EB_DoorOpen:
			elevio.SetDoorOpenLamp(false)
			start_timer <- fsm.El.Config.DoorOpenDuration
			fsm.El = requestpkg.Requests_clearAtCurrentFloor(fsm.El)

		case elevatorpkg.EB_Moving:
			fsm.Od.MotorDirection(fsm.El.Dirn)

		case elevatorpkg.EB_Idle:
			// Do nothing
		}
	}

	fsm.SetAllLights()
	fmt.Println("\nNew state:")
	elevatorpkg.Elevator_print(fsm.El)
}

// Handle floor arrival event
func (fsm *FSM) Fsm_onFloorArrival(newFloor int, start_timer chan time.Duration) {
	fmt.Printf("\n\n(%d)\n", newFloor)
	elevatorpkg.Elevator_print(fsm.El)

	// fsm.El.PrevFloor = fsm.El.Floor
	fsm.El.Floor = newFloor

	elevio.SetFloorIndicator(newFloor)

	switch fsm.El.Behaviour {
	case elevatorpkg.EB_Moving:
		if requestpkg.Requests_shouldStop(fsm.El) {
			fmt.Printf("Elevator stopping at floor %d \n", fsm.El.Floor)
			fsm.Od.MotorDirection(elevio.MD_Stop)
			fsm.El = requestpkg.Requests_clearAtCurrentFloor(fsm.El)
			elevio.SetDoorOpenLamp(true)
			fsm.SetAllLights()
			start_timer <- fsm.El.Config.DoorOpenDuration
			fmt.Print("Started doorOpen timer")
			fsm.El.Behaviour = elevatorpkg.EB_DoorOpen
		}
	}

	fmt.Println("\nNew state:")
	elevatorpkg.Elevator_print(fsm.El)
}

// Handle door timeout event
func (fsm *FSM) Fsm_onDoorTimeout(start_timer chan time.Duration) {
	elevatorpkg.Elevator_print(fsm.El)
	
	// fsm.SetAllLights()

	switch fsm.El.Behaviour {
	case elevatorpkg.EB_DoorOpen:
		pair := requestpkg.Requests_chooseDirection(fsm.El)
		fsm.El.Dirn = pair.Dirn
		fsm.El.Behaviour = pair.Behaviour

		switch fsm.El.Behaviour {
		case elevatorpkg.EB_DoorOpen:
			start_timer <- fsm.El.Config.DoorOpenDuration
			fsm.El = requestpkg.Requests_clearAtCurrentFloor(fsm.El)
			fsm.SetAllLights()
		case elevatorpkg.EB_Moving, elevatorpkg.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			fsm.Od.MotorDirection(fsm.El.Dirn)
		}

	}

	fmt.Println("\nNew state:")
	elevatorpkg.Elevator_print(fsm.El)
}
