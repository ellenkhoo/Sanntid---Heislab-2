package elevator

type DirnBehaviourPair struct {
	Dirn      MotorDirection
	Behaviour ElevatorBehaviour
}

func ClearAllRequests(elevator Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			elevator.RequestsToDo[floor][button] = false
		}
	}
}

func RequestsAbove(elevator Elevator) bool {
	for floor := elevator.ElevStates.CurrentFloor + 1; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			if elevator.RequestsToDo[floor][button] {
				return true
			}
		}
	}
	return false
}

func RequestsBelow(elevator Elevator) bool {
	for floor := 0; floor < elevator.ElevStates.CurrentFloor; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			if elevator.RequestsToDo[floor][button] {
				return true
			}
		}
	}
	return false
}

func RequestsHere(elevator Elevator) bool {
	for button := 0; button < N_BUTTONS; button++ {
		if elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][button] {
			return true
		}
	}
	return false
}

func ChooseDirection(elevator Elevator) DirnBehaviourPair {
	switch elevator.Dirn {
	case MD_Up:
		if RequestsAbove(elevator) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else if RequestsHere(elevator) {
			return DirnBehaviourPair{MD_Down, EB_DoorOpen}
		} else if RequestsBelow(elevator) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}

	case MD_Down:
		if RequestsBelow(elevator) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else if RequestsHere(elevator) {
			return DirnBehaviourPair{MD_Up, EB_DoorOpen}
		} else if RequestsAbove(elevator) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}

	case MD_Stop:
		if RequestsHere(elevator) {
			return DirnBehaviourPair{MD_Stop, EB_DoorOpen}
		} else if RequestsBelow(elevator) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else if RequestsAbove(elevator) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}
	default:
		return DirnBehaviourPair{MD_Stop, EB_Idle}
	}
}

func ShouldStop(elevator Elevator) bool {
	switch elevator.Dirn {
	case MD_Down:
		return elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallDown] ||
			elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_Cab] ||
			!RequestsBelow(elevator)
	case MD_Up:
		return elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallUp] ||
			elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_Cab] ||
			!RequestsAbove(elevator)
	default:
		return true
	}
}

func ShouldClearImmediately(elevator *Elevator) bool {

	if elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_Cab] {
		elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_Cab] = false
		return true
	}
	return (elevator.Dirn == MD_Up && elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallUp]) ||
		(elevator.Dirn == MD_Down && elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallDown]) ||
		elevator.Dirn == MD_Stop
}

func ClearAtCurrentFloor(elevator *Elevator) *Elevator {

	elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_Cab] = false
	elevator.ElevStates.CabRequests[elevator.ElevStates.CurrentFloor] = false

	switch elevator.Dirn {
	case MD_Up:
		if !RequestsAbove(*elevator) && !elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallUp] {
			elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallDown] = false
		}
		elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallUp] = false

	case MD_Down:
		if !RequestsBelow(*elevator) && !elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallDown] {
			elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallUp] = false
		}
		elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallDown] = false

	default:
		elevator.RequestsToDo[elevator.ElevStates.CurrentFloor][B_HallUp] = false
		elevator.Dirn = MD_Up
	}

	return elevator
}
