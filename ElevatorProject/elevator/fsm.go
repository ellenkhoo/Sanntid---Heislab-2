package elevator

import (
	"fmt"
	"sync"
	"time"

	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
)

type FSM struct {
	El      *Elevator
	Od      *ElevOutputDevice
	Fsm_mtx sync.Mutex
}

func (fsm *FSM) SetAllLights() {
	print("Setting all lights\n")
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := ButtonType(0); btn < N_BUTTONS-1; btn++ {
			fsm.Od.RequestButtonLight(floor, btn, fsm.El.GlobalHallRequests[floor][btn])
		}
		fsm.Od.RequestButtonLight(floor, B_Cab, fsm.El.ElevStates.CabRequests[floor])
	}
}

func (fsm *FSM) InitBetweenFloors() {
	fsm.Od.MotorDirection(MD_Up)
	fsm.El.Dirn = MD_Up
	fsm.El.Behaviour = EB_Moving
}

func (fsm *FSM) HandleRequestsToDo(networkChannels *sharedConsts.NetworkChannels, start_timer chan time.Duration) {
	PrintElevator(*fsm.El)
	fsm.SetAllLights()

	switch fsm.El.Behaviour {
	case EB_DoorOpen:
		if ShouldClearImmediately(fsm.El) {
			start_timer <- fsm.El.DoorOpenDuration
		}

	case EB_Idle:
		fsm.Fsm_mtx.Lock()
		pair := ChooseDirection(*fsm.El)
		fsm.El.Dirn = pair.Dirn
		fsm.El.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			SetDoorOpenLamp(true)
			start_timer <- fsm.El.DoorOpenDuration

		case EB_Moving:
			fsm.Od.MotorDirection(fsm.El.Dirn)

		case EB_Idle:
			// Do nothing
		}
		fsm.Fsm_mtx.Unlock()
	}

	fmt.Println("\nNew state:")
	PrintElevator(*fsm.El)
}

func (fsm *FSM) OnFloorArrival(networkChannels *sharedConsts.NetworkChannels, newFloor int, start_timer chan time.Duration) {
	fmt.Printf("\n\n(%d)\n", newFloor)
	PrintElevator(*fsm.El)

	fsm.Fsm_mtx.Lock()
	fsm.El.ElevStates.Floor = newFloor
	fsm.Fsm_mtx.Unlock()

	SetFloorIndicator(newFloor)

	switch fsm.El.Behaviour {
	case EB_Moving:
		if ShouldStop(*fsm.El) {
			fmt.Printf("Elevator stopping at floor %d \n", fsm.El.ElevStates.Floor)
			fsm.Od.MotorDirection(MD_Stop)

			fsm.Fsm_mtx.Lock()
			fsm.El = ClearAtCurrentFloor(fsm.El)

			SetDoorOpenLamp(true)
			start_timer <- fsm.El.DoorOpenDuration
			fmt.Print("Started doorOpen timer")
			fsm.El.Behaviour = EB_DoorOpen

			fsm.Fsm_mtx.Unlock()
		}
	}
	fmt.Println("\nNew state:")
	PrintElevator(*fsm.El)
}

func (fsm *FSM) OnDoorTimeout(timerChan chan time.Duration) {
	PrintElevator(*fsm.El)

	switch fsm.El.Behaviour {
	case EB_DoorOpen:
		pair := ChooseDirection(*fsm.El)
		fsm.El.Dirn = pair.Dirn
		fsm.El.Behaviour = pair.Behaviour

		switch fsm.El.Behaviour {
		case EB_DoorOpen:
			timerChan <- fsm.El.DoorOpenDuration
			fsm.El = ClearAtCurrentFloor(fsm.El)

		case EB_Moving, EB_Idle:
			SetDoorOpenLamp(false)
			fsm.Od.MotorDirection(fsm.El.Dirn)
		}
	}

	fmt.Println("\nNew state:")
	PrintElevator(*fsm.El)
}
