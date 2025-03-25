package elevator

import (
	"fmt"
	"sync"
	"time"

	elevio "github.com/ellenkhoo/ElevatorProject/elevator/Driver"
	"github.com/ellenkhoo/ElevatorProject/sharedConsts"
	"github.com/ellenkhoo/ElevatorProject/timers"
)

// Elevator FSM struct
type FSM struct {
	El      *Elevator
	Od      *ElevOutputDevice
	Fsm_mtx sync.Mutex
}

// Set all elevator lights
func (fsm *FSM) SetAllLights() {
	print("Setting all lights\n")
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < N_BUTTONS-1; btn++ {
			fsm.Od.RequestButtonLight(floor, btn, fsm.El.GlobalHallRequests[floor][btn])
		}
		fsm.Od.RequestButtonLight(floor, B_Cab, fsm.El.ElevStates.CabRequests[floor])
	}
}

// Handle initialization between floors
func (fsm *FSM) InitBetweenFloors() {
	fsm.Od.MotorDirection(elevio.MD_Up)
	fsm.El.Dirn = elevio.MD_Up
	fsm.El.Behaviour = EB_Moving
}

func (fsm *FSM) HandleRequestsToDo(networkChannels *sharedConsts.NetworkChannels, start_timer chan time.Duration) {
	PrintElevator(*fsm.El)

	switch fsm.El.Behaviour {
	case EB_DoorOpen:
		if ShouldClearImmediately(fsm.El) {
			fmt.Println("Should clear order immediately")
			start_timer <- timers.DoorOpenDuration
			SendCurrentState(networkChannels, *fsm.El)
		} else {
			fmt.Println("Shouldn't clear order immediately")
		}

	case EB_Idle:
		fsm.Fsm_mtx.Lock()
		pair := ChooseDirection(*fsm.El)
		fmt.Println("Chose direction:", pair.Dirn)
		fsm.El.Dirn = pair.Dirn
		fsm.El.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			start_timer <- timers.DoorOpenDuration
			//fsm.El = Requests_clearAtCurrentFloor(fsm.El)

		case EB_Moving:
			fsm.Od.MotorDirection(fsm.El.Dirn)

		case EB_Idle:
			// Do nothing
		}
		fsm.Fsm_mtx.Unlock()
	}

	fsm.SetAllLights()
	fmt.Println("\nNew state:")
	PrintElevator(*fsm.El)
}

// Handle floor arrival event
func (fsm *FSM) OnFloorArrival(networkChannels *sharedConsts.NetworkChannels, newFloor int, start_timer chan time.Duration) {
	fmt.Printf("\n\n(%d)\n", newFloor)
	PrintElevator(*fsm.El)

	fsm.Fsm_mtx.Lock()
	fsm.El.ElevStates.Floor = newFloor
	fsm.Fsm_mtx.Unlock()

	elevio.SetFloorIndicator(newFloor)

	switch fsm.El.Behaviour {
	case EB_Moving:
		if ShouldStop(*fsm.El) {
			fmt.Printf("Elevator stopping at floor %d \n", fsm.El.ElevStates.Floor)
			fsm.Od.MotorDirection(elevio.MD_Stop)

			fsm.Fsm_mtx.Lock()
			fsm.El = ClearAtCurrentFloor(fsm.El)

			elevio.SetDoorOpenLamp(true)
			start_timer <- timers.DoorOpenDuration
			fmt.Print("Started doorOpen timer")
			fsm.El.Behaviour = EB_DoorOpen

			fsm.Fsm_mtx.Unlock()
		}
	}
	fmt.Println("\nNew state:")
	PrintElevator(*fsm.El)
}

// Handle door timeout event
func (fsm *FSM) OnDoorTimeout(timerChan chan time.Duration) {
	PrintElevator(*fsm.El)

	switch fsm.El.Behaviour {
	case EB_DoorOpen:
		pair := ChooseDirection(*fsm.El)
		fsm.El.Dirn = pair.Dirn
		fsm.El.Behaviour = pair.Behaviour

		switch fsm.El.Behaviour {
		case EB_DoorOpen:
			timerChan <- timers.DoorOpenDuration
			// fsm.El = Requests_clearAtCurrentFloor(fsm.El)
			//fsm.SetAllLights()
		case EB_Moving, EB_Idle:
			elevio.SetDoorOpenLamp(false)
			fsm.Od.MotorDirection(fsm.El.Dirn)
		}

	}

	fmt.Println("\nNew state:")
	PrintElevator(*fsm.El)
}
