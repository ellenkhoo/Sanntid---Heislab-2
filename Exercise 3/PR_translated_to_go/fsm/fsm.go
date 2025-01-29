package fsm

import (
	"PR_translated_to_go/elevator"
	"con_load"
	"elevator_io_device"
	"fmt"
)

// Elevator FSM struct
type FSM struct {
	elevator     elevator.Elevator
	outputDevice elevator_io_device.ElevOutputDevice
}

// Initializes and returns an FSM instance
func fsm_init() {

	//elevator := elevator.elevator_uninitialized()

	fsm := &FSM{
		elevator:     elevator.elevator_uninitialized(),
		outputDevice: elevator_io_device.getOutputDevice(),
	}

	con_load.LoadConfig("elevator.con", map[string]interface{}{
		"doorOpenDuration_s":  &fsm.elevator.Config.DoorOpenDuration,
		"clearRequestVariant": &fsm.elevator.Config.ClearRequestVariant,
	})

	//fsm.outputDevice = elevator_io_device.getOutputDevice()
}

// Set all elevator lights
func (fsm *FSM) setAllLights() {
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			fsm.outputDevice.requestButtonLight(floor, btn, fsm.elevator.Requests[floor][btn])
		}
	}
}

// Handle initialization between floors
func (fsm *FSM) fsm_onInitBetweenFloors() {
	fsm.outputDevice.motorDirection(D_Down)
	fsm.elevator.Dirn = D_Down
	fsm.elevator.Behaviour = EB_Moving
}

// Handle button press event
func (fsm *FSM) fsm_onRequestButtonPress(btn_floor int, btn_type elevator_io_device.Button) {
	fmt.Printf("\n\n(%d, %s)\n", btn_floor, elevator_io_device.elevio_button_toString(btn_type))
	elevator_io_device.elevator_print(fsm.elevator)

	switch fsm.elevator.Behaviour {
	case elevator.EB_DoorOpen:
		if requests.requests_shouldClearImmediately(fsm.elevator, btn_floor, btn_type) {
			timer.timer_start(fsm.elevator.Config.doorOpenDuration_s)
		} else {
			fsm.elevator.Requests[btn_floor][btn_type] = true
		}

	case elevator.EB_Moving:
		fsm.elevator.Requests[btn_floor][btn_type] = true

	case elevator.EB_Idle:
		fsm.elevator.Requests[btn_floor][btn_type] = true
		pair := requests.requests_chooseDirection(fsm.elevator)
		fsm.elevator.Dirn = pair.Dirn
		fsm.elevator.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			fsm.outputDevice.doorLight(true)
			timer.timer_start(elevator.config.doorOpenDuration_s)
			fsm.elevator = requests.requests_clearAtCurrentFloor(fsm.elevator)

		case elevator.EB_Moving:
			fsm.outputDevice.motorDirection(fsm.elevator.Dirn)

		case elevator.EB_Idle:
			// Do nothing
		}
	}

	fsm.setAllLights()
	fmt.Println("\nNew state:")
	fsm.elevator.elevator_print()
}

// Handle floor arrival event
func (fsm *FSM) fsm_onFloorArrival(newFloor int) {
	fmt.Printf("\n\n(%d)\n", newFloor)
	fsm.elevator.elevator_print()

	fsm.elevator.Floor = newFloor
	fsm.outputDevice.floorIndicator(fsm.elevator.Floor)

	switch fsm.elevator.Behaviour {
	case elevator.EB_Moving:
		if requests.requests_shouldStop(fsm.elevator) {
			fsm.outputDevice.motorDirection(elevator.D_Stop)
			fsm.outputDevice.doorLight(true)
			fsm.elevator = requests.requests_clearAtCurrentFloor(fsm.elevator)
			timer.timer_start(fsm.elevator.Config.doorOpenDuration_s)
			fsm.setAllLights()
			fsm.elevator.Behaviour = elevator.EB_DoorOpen
		}
	}

	fmt.Println("\nNew state:")
	fsm.elevator.elevator_print()
}

// Handle door timeout event
func (fsm *FSM) fsm_onDoorTimeout() {
	fsm.elevator.elevator_print()

	switch fsm.elevator.Behaviour {
	case elevator.EB_DoorOpen:
		pair := requests.requests_chooseDirection(fsm.elevator)
		fsm.elevator.Dirn = pair.Dirn
		fsm.elevator.Behaviour = pair.Behaviour

		switch fsm.elevator.Behaviour {
		case elevator.EB_DoorOpen:
			timer.timer_start(fsm.elevator.Config.doorOpenDuration_s)
			fsm.elevator = requests.requests_clearAtCurrentFloor(fsm.elevator)
			fsm.setAllLights()
		case elevator.EB_Moving, elevator.EB_Idle:
			fsm.outputDevice.doorLight(false)
			fsm.outputDevice.motorDirection(fsm.elevator.Dirn)
		}

	}

	fmt.Println("\nNew state:")
	fsm.elevator.elevator_print()
}
