package elevator

import (
	"fmt"
)

type DirnBehaviourPair struct {
	Dirn      MotorDirection
	Behaviour ElevatorBehaviour
}

func ClearAllRequests(e Elevator) {
	fmt.Printf("Clearing all requests!\n")
	for f := 0; f < N_FLOORS; f++ {
		for b := 0; b < N_BUTTONS; b++ {
			e.RequestsToDo[f][b] = false
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
	case MD_Up:
		if RequestsAbove(e) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else if RequestsHere(e) {
			return DirnBehaviourPair{MD_Down, EB_DoorOpen}
		} else if RequestsBelow(e) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}

	case MD_Down:
		if RequestsBelow(e) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else if RequestsHere(e) {
			return DirnBehaviourPair{MD_Up, EB_DoorOpen}
		} else if RequestsAbove(e) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}

	case MD_Stop:
		if RequestsHere(e) {
			return DirnBehaviourPair{MD_Stop, EB_DoorOpen}
		} else if RequestsBelow(e) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else if RequestsAbove(e) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}
	default:
		return DirnBehaviourPair{MD_Stop, EB_Idle}
	}
}

func ShouldStop(e Elevator) bool {
	switch e.Dirn {
	case MD_Down:
		return e.RequestsToDo[e.ElevStates.Floor][B_HallDown] ||
			e.RequestsToDo[e.ElevStates.Floor][B_Cab] ||
			!RequestsBelow(e)
	case MD_Up:
		return e.RequestsToDo[e.ElevStates.Floor][B_HallUp] ||
			e.RequestsToDo[e.ElevStates.Floor][B_Cab] ||
			!RequestsAbove(e)
	default:
		return true
	}
}

func ShouldClearImmediately(e *Elevator) bool {

	if e.RequestsToDo[e.ElevStates.Floor][B_Cab] {
		e.RequestsToDo[e.ElevStates.Floor][B_Cab] = false
		return true
	}
	return (e.Dirn == MD_Up && e.RequestsToDo[e.ElevStates.Floor][B_HallUp]) ||
		(e.Dirn == MD_Down && e.RequestsToDo[e.ElevStates.Floor][B_HallDown]) ||
		e.Dirn == MD_Stop
}

func ClearAtCurrentFloor(e *Elevator) *Elevator {

	e.RequestsToDo[e.ElevStates.Floor][B_Cab] = false
	e.ElevStates.CabRequests[e.ElevStates.Floor] = false

	switch e.Dirn {
	case MD_Up:
		if !RequestsAbove(*e) && !e.RequestsToDo[e.ElevStates.Floor][B_HallUp] {
			e.RequestsToDo[e.ElevStates.Floor][B_HallDown] = false
		}
		e.RequestsToDo[e.ElevStates.Floor][B_HallUp] = false

	case MD_Down:
		if !RequestsBelow(*e) && !e.RequestsToDo[e.ElevStates.Floor][B_HallDown] {
			e.RequestsToDo[e.ElevStates.Floor][B_HallUp] = false
		}
		e.RequestsToDo[e.ElevStates.Floor][B_HallDown] = false

	default:
		e.RequestsToDo[e.ElevStates.Floor][B_HallUp] = false
		e.Dirn = MD_Up
	}

	return e
}
