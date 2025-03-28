package elevator

import (
	"sync"
	"time"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

// Finite state machine
type FSM struct {
	Elevator     *Elevator
	OutputDevice *ElevOutputDevice
	FSM_mutex    sync.Mutex
}

func (fsm *FSM) SetAllLights() {
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := ButtonType(0); button < N_BUTTONS-1; button++ {
			fsm.OutputDevice.RequestButtonLight(floor, button, fsm.Elevator.GlobalHallRequests[floor][button])
		}
		fsm.OutputDevice.RequestButtonLight(floor, B_Cab, fsm.Elevator.ElevStates.CabRequests[floor])
	}
}

func (fsm *FSM) InitBetweenFloors() {
	fsm.OutputDevice.MotorDirection(MD_Up)
	fsm.Elevator.Dirn = MD_Up
	fsm.Elevator.Behaviour = EB_Moving
}

func (fsm *FSM) HandleRequestsToDo(networkChannels *sharedConsts.NetworkChannels, timerChan chan time.Duration) {
	fsm.SetAllLights()

	switch fsm.Elevator.Behaviour {
	case EB_DoorOpen:
		if ShouldClearImmediately(fsm.Elevator) {
			timerChan <- fsm.Elevator.DoorOpenDuration
		}

	case EB_Idle:
		fsm.FSM_mutex.Lock()
		pair := ChooseDirection(*fsm.Elevator)
		fsm.Elevator.Dirn = pair.Dirn
		fsm.Elevator.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			SetDoorOpenLamp(true)
			timerChan <- fsm.Elevator.DoorOpenDuration

		case EB_Moving:
			fsm.OutputDevice.MotorDirection(fsm.Elevator.Dirn)

		case EB_Idle:
			// Do nothing
		}
		fsm.FSM_mutex.Unlock()
	}
}

func (fsm *FSM) OnFloorArrival(networkChannels *sharedConsts.NetworkChannels, newFloor int, timerChan chan time.Duration) {
	fsm.FSM_mutex.Lock()
	fsm.Elevator.ElevStates.CurrentFloor = newFloor
	fsm.FSM_mutex.Unlock()

	SetFloorIndicator(newFloor)

	switch fsm.Elevator.Behaviour {
	case EB_Moving:
		if ShouldStop(*fsm.Elevator) {
			fsm.OutputDevice.MotorDirection(MD_Stop)

			fsm.FSM_mutex.Lock()
			fsm.Elevator = ClearAtCurrentFloor(fsm.Elevator)

			SetDoorOpenLamp(true)
			timerChan <- fsm.Elevator.DoorOpenDuration
			fsm.Elevator.Behaviour = EB_DoorOpen

			fsm.FSM_mutex.Unlock()
		}
	}
}

func (fsm *FSM) OnDoorTimeout(timerChan chan time.Duration) {
	switch fsm.Elevator.Behaviour {
	case EB_DoorOpen:
		pair := ChooseDirection(*fsm.Elevator)
		fsm.Elevator.Dirn = pair.Dirn
		fsm.Elevator.Behaviour = pair.Behaviour

		switch fsm.Elevator.Behaviour {
		case EB_DoorOpen:
			timerChan <- fsm.Elevator.DoorOpenDuration
			fsm.Elevator = ClearAtCurrentFloor(fsm.Elevator)

		case EB_Moving, EB_Idle:
			SetDoorOpenLamp(false)
			fsm.OutputDevice.MotorDirection(fsm.Elevator.Dirn)
		}
	}
}
