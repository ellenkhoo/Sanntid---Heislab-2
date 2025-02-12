package requestpkg

import (
	"Driver-go/elevio"
	elevatorpkg "elevator"
	// "elevator_io_device"
)

// Endrer funksjonene til å returnere bools

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour elevatorpkg.ElevatorBehaviour
}

func Clear_all_requests() {
	for f := 0; f < elevatorpkg.N_FLOORS; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}
}

// Leter etter bestillinger i høyere etasjer
func Requests_above(e elevatorpkg.Elevator) bool {
	for f := e.Floor + 1; f < elevatorpkg.N_FLOORS; f++ {
		for btn := 0; btn < elevatorpkg.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

// Lete etter bestillinger i lavere etasjer
func Requests_below(e elevatorpkg.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevatorpkg.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

// Leter etter bestillinger i etasjen hvor heisen befinner seg
func Requests_here(e elevatorpkg.Elevator) bool {
	for btn := 0; btn < elevatorpkg.N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

// Ikke så stor fan av hva som skjer inne i hver case her
// Kanskje det løses bedre med enda en switch case?
func Requests_chooseDirection(e elevatorpkg.Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case elevio.MD_Up:
		//Leter først etter ordre over for å prioritere å reise i samme retning
		if Requests_above(e){
			return DirnBehaviourPair{elevio.MD_Up, elevatorpkg.EB_Moving}
		} else if Requests_here(e){
			return DirnBehaviourPair{elevio.MD_Down, elevatorpkg.EB_DoorOpen}
		} else if Requests_below(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevatorpkg.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevatorpkg.EB_Idle}
		}

	case elevio.MD_Down:
		//Leter ned først av samme grunn
		if Requests_below(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevatorpkg.EB_Moving}
		} else if Requests_here(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevatorpkg.EB_DoorOpen}
		} else if Requests_above(e){
			return DirnBehaviourPair{elevio.MD_Up, elevatorpkg.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevatorpkg.EB_Idle}
		}

	case elevio.MD_Stop:
		if Requests_here(e){
			return DirnBehaviourPair{elevio.MD_Stop, elevatorpkg.EB_DoorOpen}
		} else if Requests_below(e){
			return DirnBehaviourPair{elevio.MD_Down, elevatorpkg.EB_Moving}
		} else if Requests_above(e){
			return DirnBehaviourPair{elevio.MD_Up, elevatorpkg.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevatorpkg.EB_Idle}
		}
	default:
		return DirnBehaviourPair{elevio.MD_Stop, elevatorpkg.EB_Idle}
	}
}

// Sjekker om heisen bør stoppe eller ikke
func Requests_shouldStop(e elevatorpkg.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return e.Requests[e.Floor][elevatorpkg.B_HallDown] ||
			e.Requests[e.Floor][elevatorpkg.B_Cab] ||
			!Requests_below(e)
	case elevio.MD_Up:
		return e.Requests[e.Floor][elevatorpkg.B_HallUp] ||
			e.Requests[e.Floor][elevatorpkg.B_Cab] ||
			!Requests_above(e)
	default:
		return true
	}
}

func Requests_shouldClearImmediately(e elevatorpkg.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case "CV_All":
		return e.Floor == btn_floor

	case "CV_InDirn":
		return e.Floor == btn_floor && ((e.Dirn == elevio.MD_Up && btn_type == elevatorpkg.B_HallUp) ||
			(e.Dirn == elevio.MD_Down && btn_type == elevatorpkg.B_HallDown) ||
			e.Dirn == elevio.MD_Stop ||
			btn_type == elevatorpkg.B_Cab)

	default:
		return false
	}
}

// Fjerner alle bestillinger i etasjen hvor heisen befinner seg
func Requests_clearAtCurrentFloor(e elevatorpkg.Elevator) elevatorpkg.Elevator {
	switch e.Config.ClearRequestVariant {
	case "CV_All":
		for btn := 0; btn < elevatorpkg.N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}

	case "CV_InDirn":
		e.Requests[e.Floor][elevatorpkg.B_Cab] = false

		switch e.Dirn {
		case elevio.MD_Up:
			if !Requests_above(e) && !e.Requests[e.Floor][elevatorpkg.B_HallUp] {
				e.Requests[e.Floor][elevatorpkg.B_HallDown] = false
			}
			e.Requests[e.Floor][elevatorpkg.B_HallUp] = false

		case elevio.MD_Down:
			if !Requests_below(e) && !e.Requests[e.Floor][elevatorpkg.B_HallDown] {
				e.Requests[e.Floor][elevatorpkg.B_HallUp] = false
			}
			e.Requests[e.Floor][elevatorpkg.B_HallDown] = false

		default:
			e.Requests[e.Floor][elevatorpkg.B_HallUp] = false
			e.Requests[e.Floor][elevatorpkg.B_HallDown] = false
		}

	default:
		break
	}

	return e
}
