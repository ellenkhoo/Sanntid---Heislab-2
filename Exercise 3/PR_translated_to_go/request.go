package main

// Endrer funksjonene til å returnere bools

type DirnBehaviourPair struct {
	dirn      Dirn
	behaviour ElevatorBehaviour
}

//Leter etter bestillinger i høyere etasjer
func requests_above(e Elevator) bool {
	for f := e.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

//Lete etter bestillinger i lavere etasjer
func requests_below(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

//Leter etter bestillinger i etasjen hvor heisen befinner seg
func requests_here(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

//Ikke så stor fan av hva som skjer inne i hver case her
//Kanskje det løses bedre med enda en switch case?
func requests_chooseDirection(e Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case D_Up:
		//Leter først etter ordre over for å prioritere å reise i samme retning
		if requests_above(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else if requests_here(e) {
			return DirnBehaviourPair{D_Down, EB_DoorOpen}
		} else if requests_below(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{D_Stop, EB_Idle}
		}

	case D_Down:
		//Leter ned først av samme grunn
		if requests_below(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		} else if requests_here(e) {
			return DirnBehaviourPair{D_Up, EB_DoorOpen}
		} else if requests_above(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{D_Stop, EB_Idle}
		}

	case D_Stop:
		if requests_here(e) {
			return DirnBehaviourPair{D_Stop, EB_DoorOpen}
		} else if requests_below(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else if requests_above(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{D_Stop, EB_Idle}
		}
	default:
		return DirnBehaviourPair{D_Stop, EB_Idle}
	}
}

//Sjekker om heisen bør stoppe eller ikke
func requests_shouldStop(e Elevator) bool {
	switch e.Dirn {
	case D_Down:
		return e.Requests[e.Floor][B_HallDown] ||
			e.Requests[e.Floor][B_Cab] ||
			!requests_below(e)
	case D_Up:
		return e.Requests[e.Floor][B_HallUp] ||
			e.Requests[e.Floor][B_Cab] ||
			!requests_above(e)
	default:
		return true
	}
}

func requests_shouldClearImmediately(e Elevator, btn_floor int, btn_type Button) bool {
	switch e.Config.ClearRequestVariant {
	case "CV_All":
		return e.Floor == btn_floor

	case "CV_InDirn":
		return e.Floor == btn_floor && ((e.Dirn == D_Up && btn_type == B_HallUp) ||
			(e.Dirn == D_Down && btn_type == B_HallDown) ||
			e.Dirn == D_Stop ||
			btn_type == B_Cab)

	default:
		return false
	}
}

//Fjerner alle bestillinger i etasjen hvor heisen befinner seg
func requests_clearAtCurrentFloor(e Elevator) Elevator {
	switch e.Config.ClearRequestVariant {
	case "CV_All":
		for btn := 0; btn < N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}

	case "CV_InDirn":
		e.Requests[e.Floor][B_Cab] = false

		switch e.Dirn {
		case D_Up:
			if !requests_above(e) && !e.Requests[e.Floor][B_HallUp] {
				e.Requests[e.Floor][B_HallDown] = false
			}
			e.Requests[e.Floor][B_HallDown] = false

		case D_Down:
			if !requests_below(e) && !e.Requests[e.Floor][B_HallDown] {
				e.Requests[e.Floor][B_HallUp] = false
			}
			e.Requests[e.Floor][B_HallDown] = false

		default:
			e.Requests[e.Floor][B_HallUp] = false
			e.Requests[e.Floor][B_HallDown] = false
		}

	default:
		break
	}

	return e
}
