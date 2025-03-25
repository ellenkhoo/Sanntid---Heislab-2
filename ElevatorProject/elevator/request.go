package elevator

import (
	"fmt"

	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
)

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour ElevatorBehaviour
}

func ClearAllRequests(e Elevator) {
	fmt.Printf("Clearing all requests!\n")
	for f := 0; f < N_FLOORS; f++ {
		for b := 0; b < N_BUTTONS; b++ {
			e.RequestsToDo[f][b] = false
		}
	}
	for f := 0; f < N_FLOORS; f++ {
		for b := 0; b < N_BUTTONS; b++ {
			if e.RequestsToDo[f][b] {
				fmt.Printf("Order at floor %d not cleared", f)
			}
		}
	}
}

func RequestsAbove(e Elevator) bool {
	for f := e.ElevStates.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.RequestsToDo[f][btn] {
				return true
			}
		}
	}
	return false
}

func RequestsBelow(e Elevator) bool {
	for f := 0; f < e.ElevStates.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.RequestsToDo[f][btn] {
				return true
			}
		}
	}
	return false
}

func RequestsHere(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.RequestsToDo[e.ElevStates.Floor][btn] {
			return true
		}
	}
	return false
}

func ChooseDirection(e Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case elevio.MD_Up:
		if RequestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else if RequestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_DoorOpen}
		} else if RequestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
		}

	case elevio.MD_Down:
		if RequestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		} else if RequestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_DoorOpen}
		} else if RequestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
		}

	case elevio.MD_Stop:
		if RequestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Stop, EB_DoorOpen}
		} else if RequestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		} else if RequestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
		}
	default:
		return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
	}
}

func ShouldStop(e Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return e.RequestsToDo[e.ElevStates.Floor][B_HallDown] ||
			e.RequestsToDo[e.ElevStates.Floor][B_Cab] ||
			!RequestsBelow(e)
	case elevio.MD_Up:
		return e.RequestsToDo[e.ElevStates.Floor][B_HallUp] ||
			e.RequestsToDo[e.ElevStates.Floor][B_Cab] ||
			!RequestsAbove(e)
	default:
		return true
	}
}

func ShouldClearImmediately(e *Elevator) bool {
	switch e.Config.ClearRequestVariant {
	// case "CV_All":
	// 	return e.ElevStates.Floor == btn_floor

	case "CV_InDirn":
		if e.RequestsToDo[e.ElevStates.Floor][B_Cab] {
			fmt.Println("Cab call at", e.ElevStates.Floor)
			e.RequestsToDo[e.ElevStates.Floor][B_Cab] = false
			return true
		}
		return (e.Dirn == elevio.MD_Up && e.RequestsToDo[e.ElevStates.Floor][B_HallUp]) ||
			(e.Dirn == elevio.MD_Down && e.RequestsToDo[e.ElevStates.Floor][B_HallDown]) ||
			e.Dirn == elevio.MD_Stop

	default:
		return false
	}
}

func ClearAtCurrentFloor(e *Elevator) *Elevator {
	switch e.Config.ClearRequestVariant {
	case "CV_All":
		for btn := 0; btn < N_BUTTONS; btn++ {
			e.RequestsToDo[e.ElevStates.Floor][btn] = false
		}

	case "CV_InDirn":
		fmt.Println("Setting cab request to false at floor", e.ElevStates.Floor)
		e.RequestsToDo[e.ElevStates.Floor][B_Cab] = false
		e.ElevStates.CabRequests[e.ElevStates.Floor] = false
		fmt.Print("Cab at floor is", e.ElevStates.CabRequests[e.ElevStates.Floor])

		switch e.Dirn {
		case elevio.MD_Up:
			if !RequestsAbove(*e) && !e.RequestsToDo[e.ElevStates.Floor][B_HallUp] {
				e.RequestsToDo[e.ElevStates.Floor][B_HallDown] = false
			}
			e.RequestsToDo[e.ElevStates.Floor][B_HallUp] = false

		case elevio.MD_Down:
			if !RequestsBelow(*e) && !e.RequestsToDo[e.ElevStates.Floor][B_HallDown] {
				e.RequestsToDo[e.ElevStates.Floor][B_HallUp] = false
			}
			e.RequestsToDo[e.ElevStates.Floor][B_HallDown] = false

		default:
			e.RequestsToDo[e.ElevStates.Floor][B_HallUp] = false
			//e.RequestsToDo[e.ElevStates.Floor][B_HallDown] = false
		}

	default:
		break
	}

	return e
}
