package main

import (
	"fmt"
)

// Elevator FSM struct
type FSM struct {
	el Elevator
	od ElevOutputDevice
}

// FLytte til et annet sted?
func convertBoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Set all elevator lights
func (fsm *FSM) setAllLights() {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			fsm.od.RequestButtonLight(floor, Button(btn), convertBoolToInt(fsm.el.Requests[floor][btn]))
		}
	}
}

// Handle initialization between floors
func (fsm *FSM) fsm_onInitBetweenFloors() {
	fsm.od.MotorDirection(D_Down)
	fsm.el.Dirn = elevio_drin_toString(D_Down) //skal det være string eller int?
	fsm.el.Behaviour = elevatorBehaviourToString[EB_Moving]
}

// Handle button press event
func (fsm *FSM) fsm_onRequestButtonPress(btn_floor int, btn_type Button) {
	fmt.Printf("\n\n(%d, %s)\n", btn_floor, elevio_button_toString(btn_type))
	elevator_print(fsm.el)

	switch fsm.el.Behaviour {
	case elevatorBehaviourToString[EB_DoorOpen]:
		if requests_shouldClearImmediately(fsm.el, btn_floor, btn_type) {
			timer_start(fsm.el.Config.DoorOpenDuration)
		} else {
			fsm.el.Requests[btn_floor][btn_type] = true
		}

	case elevatorBehaviourToString[EB_Moving]:
		fsm.el.Requests[btn_floor][btn_type] = true

	case elevatorBehaviourToString[EB_Idle]:
		fsm.el.Requests[btn_floor][btn_type] = true
		pair := requests_chooseDirection(fsm.el)
		fsm.el.Dirn = pair.Dirn
		fsm.el.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case elevatorBehaviourToString[EB_DoorOpen]:
			fsm.od.DoorLight(1)
			timer_start(fsm.el.Config.DoorOpenDuration)
			fsm.el = requests_clearAtCurrentFloor(fsm.el)

		case EB_Moving:
			fsm.od.MotorDirection(stringToDirn[fsm.el.Dirn])

		case EB_Idle:
			// Do nothing
		}
	}

	fsm.setAllLights()
	fmt.Println("\nNew state:")
	elevator_print(fsm.el)
}

// Handle floor arrival event
func (fsm *FSM) fsm_onFloorArrival(newFloor int) {
	fmt.Printf("\n\n(%d)\n", newFloor)
	elevator_print(fsm.el)

	fsm.el.Floor = newFloor
	fsm.od.FloorIndicator(newFloor)

	switch fsm.el.Behaviour {
	case elevatorBehaviourToString[EB_Moving]:
		if requests_shouldStop(fsm.el) {
			fsm.od.MotorDirection(D_Stop)
			fsm.od.DoorLight(1)
			fsm.el = requests_clearAtCurrentFloor(fsm.el)
			timer_start(fsm.el.Config.DoorOpenDuration)
			fsm.setAllLights()
			fsm.el.Behaviour = elevatorBehaviourToString[EB_DoorOpen]
		}
	}

	fmt.Println("\nNew state:")
	elevator_print(fsm.el)
}

// Handle door timeout event
func (fsm *FSM) fsm_onDoorTimeout() {
	elevator_print(fsm.el)

	switch fsm.el.Behaviour {
	case elevatorBehaviourToString[EB_DoorOpen]:
		pair := requests_chooseDirection(fsm.el)
		fsm.el.Dirn = pair.Dirn
		fsm.el.Behaviour = pair.Behaviour

		switch fsm.el.Behaviour {
		case elevatorBehaviourToString[EB_DoorOpen]:
			timer_start(fsm.el.Config.DoorOpenDuration)
			fsm.el = requests_clearAtCurrentFloor(fsm.el)
			fsm.setAllLights()
		case elevatorBehaviourToString[EB_Moving], elevatorBehaviourToString[EB_Idle]:
			fsm.od.DoorLight(0)
			fsm.od.MotorDirection(stringToDirn[fsm.el.Dirn])
		}

	}

	fmt.Println("\nNew state:")
	elevator_print(fsm.el)
}
