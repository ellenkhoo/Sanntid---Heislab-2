package elevator

import (
	"github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"fmt"
)

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour ElevatorBehaviour
}


func Clear_all_requests(e Elevator) {
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


func Requests_above(e Elevator) bool {
	for f := e.ElevStates.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.RequestsToDo[f][btn] {
				return true
			}
		}
	}
	return false
}


func Requests_below(e Elevator) bool {
	for f := 0; f < e.ElevStates.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.RequestsToDo[f][btn] {
				return true
			}
		}
	}
	return false
}


func Requests_here(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.RequestsToDo[e.ElevStates.Floor][btn] {
			return true
		}
	}
	return false
}

// Ikke så stor fan av hva som skjer inne i hver case her
// Kanskje det løses bedre med enda en switch case?
func Requests_chooseDirection(e Elevator) DirnBehaviourPair {
	switch e.Dirn {
		// Leter først etter ordre i samme retning
	case elevio.MD_Up:
		if Requests_above(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else if Requests_here(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_DoorOpen}
		} else if Requests_below(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
		}

	case elevio.MD_Down:
		if Requests_below(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		} else if Requests_here(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_DoorOpen}
		} else if Requests_above(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
		}

	case elevio.MD_Stop:
		if Requests_here(e) {
			return DirnBehaviourPair{elevio.MD_Stop, EB_DoorOpen}
		} else if Requests_below(e) {
			return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
		} else if Requests_above(e) {
			return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
		}
	default:
		return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
	}
}

func Requests_shouldStop(e Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return e.RequestsToDo[e.ElevStates.Floor][B_HallDown] ||
			e.RequestsToDo[e.ElevStates.Floor][B_Cab] ||
			!Requests_below(e)
	case elevio.MD_Up:
		return e.RequestsToDo[e.ElevStates.Floor][B_HallUp] ||
			e.RequestsToDo[e.ElevStates.Floor][B_Cab] ||
			!Requests_above(e)
	default:
		return true
	}
}

func Requests_shouldClearImmediately(e Elevator) bool {
	switch e.Config.ClearRequestVariant {
	case "CV_All":
		return e.ElevStates.Floor == btn_floor

	case "CV_InDirn":
		if e.RequestsToDo[e.ElevStates.Floor][B_Cab] {
		   e.RequestsToDo[e.ElevStates.Floor][B_Cab] = false
		   return true
		}
		return (e.Dirn == elevio.MD_Up && e.RequestsToDo[e.ElevStates.Floor][B_HallUp]) ||
			   (e.Dirn == elevio.MD_Down && e.RequestsToDo[e.ElevStates.Floor][B_HallDown])

	default:
		return false
	}
}

func Requests_clearAtCurrentFloor(e Elevator) Elevator {
	switch e.Config.ClearRequestVariant {
	case "CV_All":
		for btn := 0; btn < N_BUTTONS; btn++ {
			e.RequestsToDo[e.ElevStates.Floor][btn] = false
		}

	case "CV_InDirn":
		e.RequestsToDo[e.ElevStates.Floor][B_Cab] = false

		switch e.Dirn {
		case elevio.MD_Up:
			if !Requests_above(e) && !e.RequestsToDo[e.ElevStates.Floor][B_HallUp] {
				e.RequestsToDo[e.ElevStates.Floor][B_HallDown] = false
			}
			e.RequestsToDo[e.ElevStates.Floor][B_HallUp] = false

		case elevio.MD_Down:
			if !Requests_below(e) && !e.RequestsToDo[e.ElevStates.Floor][B_HallDown] {
				e.RequestsToDo[e.ElevStates.Floor][B_HallUp] = false
			}
			e.RequestsToDo[e.ElevStates.Floor][B_HallDown] = false

		default:
			e.RequestsToDo[e.ElevStates.Floor][B_HallUp] = false
			e.RequestsToDo[e.ElevStates.Floor][B_HallDown] = false
		}

	default:
		break
	}

	return e
}
